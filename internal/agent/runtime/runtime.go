package runtime

import (
	"context"
	"sync"
	"time"

	"github.com/intrik8-labs/smidr/internal/agent/config"
	"github.com/intrik8-labs/smidr/internal/agent/job"
	"github.com/intrik8-labs/smidr/internal/agent/runtime/plugins"
	buildplugin "github.com/intrik8-labs/smidr/internal/agent/runtime/plugins/build"
	monitorplugin "github.com/intrik8-labs/smidr/internal/agent/runtime/plugins/monitor"
	yoctoplugin "github.com/intrik8-labs/smidr/internal/agent/runtime/plugins/yocto"
	"github.com/intrik8-labs/smidr/internal/logging"
)

type Runtime struct {
	cfg       *config.Config
	log       *logging.Logger
	heartbeat *Heartbeat
	poller    *Poller
	plugins   map[string]plugins.Plugin
	// Add fields as necessary
	activeJobs int
	activeMu   sync.Mutex

	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func New(cfg *config.Config, log *logging.Logger) *Runtime {
	ctx, cancel := context.WithCancel(context.Background())

	return &Runtime{
		ctx:    ctx,
		cancel: cancel,
		cfg:    cfg,
		log:    log,
		heartbeat: NewHeartbeat(cfg.AgentConfig.HeartBeatInterval, &HeartbeatConfig{
			AgentID:         cfg.AgentConfig.ID,
			ControlPlaneURI: cfg.ControlPlane.URI,
		}),
		poller: NewPoller(cfg.AgentConfig.ID, cfg.ControlPlane.URI, 5, 1, []string{"build", "monitor", "yocto"}), // Poll every 5s, max 1 job
		plugins: map[string]plugins.Plugin{
			"build":   buildplugin.NewBuildPlugin(),
			"monitor": monitorplugin.NewMonitorPlugin(),
			"yocto":   yoctoplugin.NewYoctoPlugin(),
		},
		activeJobs: 0,
	}
}

func (r *Runtime) Start() error {
	ctx := logging.WithExecutor(r.ctx, r.cfg.AgentConfig.ID)
	r.log.InfoContext(ctx, "starting agent runtime",
		logging.String("executor_id", r.cfg.AgentConfig.ID),
		logging.String("control_plane", r.cfg.ControlPlane.URI),
	)

	// Register with control plane
	if err := r.Register(ctx); err != nil {
		r.log.ErrorContext(ctx, "failed to register with control plane", logging.Err(err))
		return err
	}

	// Start heartbeat loop
	r.wg.Add(1)
	go r.runHeartbeat(ctx)

	// Start job poller loop
	r.log.InfoContext(ctx, "starting job poller")
	r.wg.Add(1)
	go r.runPoller(ctx)

	r.log.InfoContext(ctx, "agent runtime started")
	return nil
}

func (r *Runtime) Stop() {
	ctx := logging.WithExecutor(r.ctx, r.cfg.AgentConfig.ID)
	r.log.InfoContext(ctx, "stopping agent runtime")
	r.cancel()
	r.wg.Wait()
	r.log.InfoContext(ctx, "agent runtime stopped")
}

// runHeartbeat sends periodic health signals to control plane
func (r *Runtime) runHeartbeat(ctx context.Context) {
	defer r.wg.Done()

	ticker := time.NewTicker(time.Duration(r.heartbeat.intervalSeconds) * time.Second)
	defer ticker.Stop()

	r.log.InfoContext(ctx, "heartbeat started",
		logging.Int("interval_seconds", r.heartbeat.intervalSeconds),
	)

	// Send initial heartbeat immediately
	r.activeMu.Lock()
	active := r.activeJobs
	r.activeMu.Unlock()
	if err := r.heartbeat.Send(ctx, active); err != nil {
		r.log.WarnContext(ctx, "initial heartbeat failed", logging.Err(err))
	}

	for {
		select {
		case <-ctx.Done():
			r.log.InfoContext(ctx, "heartbeat stopped")
			return
		case <-ticker.C:
			r.activeMu.Lock()
			active := r.activeJobs
			r.activeMu.Unlock()
			if err := r.heartbeat.Send(ctx, active); err != nil {
				r.log.WarnContext(ctx, "heartbeat failed", logging.Err(err))
			}
		}
	}
}

// runPoller polls for new jobs from the control plane
func (r *Runtime) runPoller(ctx context.Context) {
	defer r.wg.Done()

	log := logging.FromContext(ctx)
	log.InfoContext(ctx, "job poller started")

	ticker := time.NewTicker(time.Duration(5) * time.Second) // Poll every 5 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.InfoContext(ctx, "job poller stopped")
			return
		case <-ticker.C:
			r.activeMu.Lock()
			active := r.activeJobs
			r.activeMu.Unlock()
			jobs, err := r.poller.Poll(ctx, active)
			if err != nil {
				log.WarnContext(ctx, "job poll failed", logging.Err(err))
				continue
			}

			log.InfoContext(ctx, "job poll completed", logging.Int("jobs_found", len(jobs)))

			// Process available jobs
			for _, job := range jobs {
				if err := r.poller.ClaimJob(ctx, job.ID); err != nil {
					log.ErrorContext(ctx, "failed to claim job", logging.String("job_id", job.ID), logging.Err(err))
					continue
				}

				// Execute job in separate goroutine
				r.wg.Add(1)
				go r.executeJob(ctx, job)
			}
		}
	}

}

// executeJob simulates job execution
func (r *Runtime) executeJob(ctx context.Context, job job.Job) {
	defer r.wg.Done()

	log := logging.FromContext(ctx)

	r.activeMu.Lock()
	r.activeJobs++
	r.activeMu.Unlock()

	log.InfoContext(ctx, "job execution started", logging.String("job_id", job.ID))

	// Execute job using plugin
	plugin, ok := r.plugins[job.PluginType]
	if !ok {
		log.ErrorContext(ctx, "unsupported plugin type", logging.String("plugin_type", job.PluginType))
		r.activeMu.Lock()
		r.activeJobs--
		r.activeMu.Unlock()
		return
	}

	if err := plugin.Execute(ctx, job); err != nil {
		log.ErrorContext(ctx, "job execution failed", logging.String("job_id", job.ID), logging.Err(err))
		// Still report completion with failure status
		if reportErr := r.poller.ReportJobCompletion(ctx, job.ID, "failed", err.Error(), []string{}); reportErr != nil {
			log.ErrorContext(ctx, "failed to report job failure", logging.String("job_id", job.ID), logging.Err(reportErr))
		}
	} else {
		// Report completion
		if err := r.poller.ReportJobCompletion(ctx, job.ID, "completed", "job executed successfully", []string{}); err != nil {
			log.ErrorContext(ctx, "failed to report job completion", logging.String("job_id", job.ID), logging.Err(err))
		}
	}

	r.activeMu.Lock()
	r.activeJobs--
	r.activeMu.Unlock()

	log.InfoContext(ctx, "job execution completed", logging.String("job_id", job.ID))
}
