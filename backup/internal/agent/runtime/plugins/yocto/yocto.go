package yocto

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"time"

	"github.com/intrik8-labs/smidr/internal/agent/job"
	"github.com/intrik8-labs/smidr/internal/agent/runtime/plugins"
	"github.com/intrik8-labs/smidr/internal/logging"
)

type YoctoPlugin struct{}

func NewYoctoPlugin() *YoctoPlugin {
	return &YoctoPlugin{}
}

func (p *YoctoPlugin) Execute(ctx context.Context, j job.Job, req plugins.YoctoBuildRequest) (*job.JobResponse, error) {
	logger := logging.FromContext(ctx).With("plugin", "yocto", "job_id", j.ID)

	logger.Info("Starting Yocto build job execution")

	// Get the config YAML from one of the sources
	configYAML, err := p.getConfig(ctx, req)
	if err != nil {
		return &job.JobResponse{
			Status:       "failed",
			Message:      "Failed to retrieve build config",
			ErrorMessage: err.Error(),
		}, err
	}

	logger.Info("Retrieved build config", "config_length", len(configYAML))

	// Parse config and extract key parameters (simplified for now)
	// In a real implementation, you'd parse the YAML here
	image := "core-image-minimal"
	machine := "qemux86-64"
	workDir := "/build"

	logger.Info("Yocto build parameters",
		"image", image,
		"machine", machine,
		"workdir", workDir,
		"cache_enabled", req.CacheEnabled,
		"clean_build", req.CleanBuild)

	startTime := time.Now()

	// Simulate Yocto build steps
	steps := []struct {
		name     string
		duration time.Duration
	}{
		{"Initialize build environment", 1 * time.Second},
		{"Configure local.conf", 500 * time.Millisecond},
		{"Configure bblayers.conf", 500 * time.Millisecond},
		{"Parse recipes", 2 * time.Second},
		{"Build image", 10 * time.Second},
		{"Generate artifacts", 1 * time.Second},
	}

	for _, step := range steps {
		select {
		case <-ctx.Done():
			return &job.JobResponse{
				Status:       "cancelled",
				Message:      "Build cancelled by context",
				ErrorMessage: ctx.Err().Error(),
			}, ctx.Err()
		default:
			logger.Info("Executing Yocto step", "step", step.name)
			time.Sleep(step.duration)
			logger.Info("Yocto step completed", "step", step.name)
		}
	}

	buildTime := time.Since(startTime)

	// Generate artifact paths
	artifactPath := filepath.Join(workDir, "tmp", "deploy", "images", machine, fmt.Sprintf("%s-%s.wic", image, machine))

	logger.Info("Yocto build completed successfully", "build_time", buildTime)

	return &job.JobResponse{
		Status:    "completed",
		Message:   fmt.Sprintf("Yocto build completed in %s", buildTime),
		Artifacts: []string{artifactPath},
		Metadata: map[string]string{
			"build_time_seconds": fmt.Sprintf("%.0f", buildTime.Seconds()),
			"image":              image,
			"machine":            machine,
			"cache_enabled":      fmt.Sprintf("%t", req.CacheEnabled),
		},
	}, nil
}

// getConfig retrieves the config YAML from one of the available sources
func (p *YoctoPlugin) getConfig(ctx context.Context, req plugins.YoctoBuildRequest) (string, error) {
	switch {
	case req.ConfigCompressed != "":
		return decodeAndDecompress(req.ConfigCompressed)
	case req.ConfigURL != "":
		return fetchConfigFromURL(ctx, req.ConfigURL)
	case req.ConfigRaw != "":
		return req.ConfigRaw, nil
	default:
		return "", fmt.Errorf("no config provided (must specify config_url, config_compressed, or config)")
	}
}

// decodeAndDecompress decodes base64 and decompresses gzipped config
func decodeAndDecompress(b64 string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gz.Close()

	decompressed, err := io.ReadAll(gz)
	if err != nil {
		return "", fmt.Errorf("failed to decompress: %w", err)
	}

	return string(decompressed), nil
}

// fetchConfigFromURL fetches config from a URL (e.g., GitHub raw URL)
func fetchConfigFromURL(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}
