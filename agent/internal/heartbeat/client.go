package heartbeat

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/schererja/smidr/agent/internal/logging"
	"github.com/schererja/smidr/agent/internal/signals"
)

// Config holds the heartbeat client configuration.
type Config struct {
	ControlPlaneURL string
	AgentID         string
	CertPath        string
	KeyPath         string
	CACertPath      string // optional: path to CA cert for server verification
	Interval        time.Duration
	Timeout         time.Duration
}

// Client sends periodic heartbeats to the control plane.
type Client struct {
	cfg        Config
	httpClient *http.Client
	log        *logging.Logger
}

// New creates a new heartbeat client with mTLS configuration.
func New(cfg Config, log *logging.Logger) (*Client, error) {
	if cfg.ControlPlaneURL == "" {
		return nil, errors.New("control_plane_url is required")
	}
	if cfg.AgentID == "" {
		return nil, errors.New("agent_id is required")
	}
	if cfg.CertPath == "" {
		return nil, errors.New("cert_path is required")
	}
	if cfg.KeyPath == "" {
		return nil, errors.New("key_path is required")
	}
	if cfg.Interval == 0 {
		cfg.Interval = 60 * time.Second
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}
	if log == nil {
		return nil, errors.New("logger is required")
	}

	tlsConfig, err := buildTLSConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("build tls config: %w", err)
	}

	httpClient := &http.Client{
		Timeout: cfg.Timeout,
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	return &Client{
		cfg:        cfg,
		httpClient: httpClient,
		log:        log,
	}, nil
}

// Run starts the heartbeat loop. Blocks until context is canceled.
func (c *Client) Run(ctx context.Context) error {
	c.log.Info("heartbeat loop started", "interval", c.cfg.Interval)

	ticker := time.NewTicker(c.cfg.Interval)
	defer ticker.Stop()

	// Send first heartbeat immediately
	if err := c.sendOnce(ctx); err != nil {
		c.log.Warn("initial heartbeat failed", "error", err)
	}

	for {
		select {
		case <-ctx.Done():
			c.log.Info("heartbeat loop stopped")
			return ctx.Err()
		case <-ticker.C:
			if err := c.sendOnce(ctx); err != nil {
				c.log.Warn("heartbeat failed", "error", err)
			}
		}
	}
}

func (c *Client) sendOnce(ctx context.Context) error {
	snap, err := signals.Collect()
	if err != nil {
		return fmt.Errorf("collect signals: %w", err)
	}

	payload := heartbeatRequest{
		AgentID:       c.cfg.AgentID,
		Timestamp:     snap.Timestamp,
		UptimeSeconds: snap.UptimeSeconds,
		LoadAverage1m: snap.LoadAverage1m,
		MemoryUsedPct: snap.MemoryUsedPct,
		DiskUsedPct:   snap.DiskUsedPct,
		ProcessCount:  snap.ProcessCount,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal heartbeat: %w", err)
	}

	url := strings.TrimRight(c.cfg.ControlPlaneURL, "/") + "/v0/agents/heartbeat"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send heartbeat: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		snippet := readResponseSnippet(resp.Body)
		return fmt.Errorf("heartbeat failed with status %d: %s", resp.StatusCode, snippet)
	}

	c.log.Debug("heartbeat sent",
		"uptime", snap.UptimeSeconds,
		"load", snap.LoadAverage1m,
		"memory_pct", snap.MemoryUsedPct,
		"disk_pct", snap.DiskUsedPct,
		"processes", snap.ProcessCount,
	)

	return nil
}

type heartbeatRequest struct {
	AgentID       string    `json:"agentId"`
	Timestamp     time.Time `json:"timestamp"`
	UptimeSeconds float64   `json:"uptimeSeconds"`
	LoadAverage1m float64   `json:"loadAverage1m"`
	MemoryUsedPct float64   `json:"memoryUsedPct"`
	DiskUsedPct   float64   `json:"diskUsedPct"`
	ProcessCount  int       `json:"processCount"`
}

func buildTLSConfig(cfg Config) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(cfg.CertPath, cfg.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("load client certificate: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	// Load CA cert if provided for server verification
	if cfg.CACertPath != "" {
		caCert, err := os.ReadFile(cfg.CACertPath)
		if err != nil {
			return nil, fmt.Errorf("read ca cert: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, errors.New("failed to parse ca certificate")
		}

		tlsConfig.RootCAs = caCertPool
	} else if strings.Contains(cfg.ControlPlaneURL, "localhost") || strings.Contains(cfg.ControlPlaneURL, "127.0.0.1") {
		// For localhost development without CA cert, skip verification
		tlsConfig.InsecureSkipVerify = true
	}

	return tlsConfig, nil
}

func readResponseSnippet(body io.Reader) string {
	data, err := io.ReadAll(io.LimitReader(body, 1024))
	if err != nil || len(data) == 0 {
		return "empty response"
	}
	return string(bytes.TrimSpace(data))
}
