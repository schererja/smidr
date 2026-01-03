package runtime

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/intrik8-labs/smidr/internal/agent/config"
	"github.com/intrik8-labs/smidr/internal/agent/job"
	"github.com/intrik8-labs/smidr/internal/agent/runtime/plugins"
	commandplugin "github.com/intrik8-labs/smidr/internal/agent/runtime/plugins/command"
	yoctoplugin "github.com/intrik8-labs/smidr/internal/agent/runtime/plugins/yocto"
	"github.com/intrik8-labs/smidr/internal/logging"
)

type Runtime struct {
	cfg       *config.Config
	log       *logging.Logger
	heartbeat *Heartbeat
	poller    *Poller
	plugins   map[string]interface{}
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
		}, cfg.AgentConfig.Demo),
		poller: NewPoller(cfg.AgentConfig.ID, cfg.ControlPlane.URI, 5, 1, []string{"yocto", "command"}, cfg.AgentConfig.Demo),
		plugins: map[string]interface{}{
			"command": commandplugin.NewCommandPlugin(),
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
	if err := r.Register(ctx, r.cfg.AgentConfig.Demo); err != nil {
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
func (r *Runtime) executeJob(ctx context.Context, j job.Job) {
	defer r.wg.Done()

	log := logging.FromContext(ctx)

	r.activeMu.Lock()
	r.activeJobs++
	r.activeMu.Unlock()

	log.InfoContext(ctx, "job execution started", logging.String("job_id", j.ID))

	// Execute job using plugin
	plugin, ok := r.plugins[j.PluginType]
	if !ok {
		log.ErrorContext(ctx, "unsupported plugin type", logging.String("plugin_type", j.PluginType))
		r.activeMu.Lock()
		r.activeJobs--
		r.activeMu.Unlock()
		return
	}

	// Execute based on plugin type with proper request unmarshaling
	var response *job.JobResponse
	var execErr error

	switch j.PluginType {
	case "yocto":
		yoctoPlugin := plugin.(*yoctoplugin.YoctoPlugin)
		var req plugins.YoctoBuildRequest
		if unmarshalErr := j.UnmarshalRequest(&req); unmarshalErr != nil {
			log.ErrorContext(ctx, "failed to unmarshal request", logging.String("job_id", j.ID), logging.Err(unmarshalErr))
			execErr = unmarshalErr
		} else {
			response, execErr = yoctoPlugin.Execute(ctx, j, req)
		}
	case "command":
		commandPlugin := plugin.(*commandplugin.CommandPlugin)
		var req plugins.CommandRequest
		if unmarshalErr := j.UnmarshalRequest(&req); unmarshalErr != nil {
			log.ErrorContext(ctx, "failed to unmarshal request", logging.String("job_id", j.ID), logging.Err(unmarshalErr))
			execErr = unmarshalErr
		} else {
			response, execErr = commandPlugin.Execute(ctx, j, req)
		}
	default:
		execErr = fmt.Errorf("unsupported plugin type: %s", j.PluginType)
	}

	if execErr != nil {
		log.ErrorContext(ctx, "job execution failed", logging.String("job_id", j.ID), logging.Err(execErr))
		// Report failure using response if available
		status := "failed"
		message := execErr.Error()
		artifacts := []string{}
		if response != nil {
			status = response.Status
			message = response.Message
			if response.ErrorMessage != "" {
				message = response.ErrorMessage
			}
			artifacts = response.Artifacts
		}
		if reportErr := r.poller.ReportJobCompletion(ctx, j.ID, status, message, artifacts); reportErr != nil {
			log.ErrorContext(ctx, "failed to report job failure", logging.String("job_id", j.ID), logging.Err(reportErr))
		}
	} else {
		// Report completion with response data
		status := "completed"
		message := "job executed successfully"
		artifacts := []string{}
		if response != nil {
			status = response.Status
			message = response.Message
			artifacts = response.Artifacts
		}
		if err := r.poller.ReportJobCompletion(ctx, j.ID, status, message, artifacts); err != nil {
			log.ErrorContext(ctx, "failed to report job completion", logging.String("job_id", j.ID), logging.Err(err))
		}
	}

	r.activeMu.Lock()
	r.activeJobs--
	r.activeMu.Unlock()

	log.InfoContext(ctx, "job execution completed", logging.String("job_id", j.ID))
}
