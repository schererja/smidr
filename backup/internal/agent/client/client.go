package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client handles communication with smidr-core control plane
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new control plane client
func NewClient(controlPlaneURI string) *Client {
	return &Client{
		baseURL: controlPlaneURI,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// RegisterAgent registers the agent with the control plane
func (c *Client) RegisterAgent(ctx context.Context, agentID, name string, capabilities []string, metadata map[string]string) error {
	payload := map[string]interface{}{
		"agent_id":     agentID,
		"name":         name,
		"capabilities": capabilities,
		"metadata":     metadata,
	}

	return c.post(ctx, "/api/v1/register", payload)
}

// Heartbeat sends a heartbeat to the control plane
func (c *Client) Heartbeat(ctx context.Context, agentID, status string) error {
	payload := map[string]interface{}{
		"agent_id": agentID,
		"status":   status,
	}

	return c.post(ctx, "/api/v1/heartbeat", payload)
}

// PollJobs requests available jobs from the control plane
func (c *Client) PollJobs(ctx context.Context, agentID string) ([]interface{}, error) {
	url := fmt.Sprintf("%s/api/v1/jobs/poll?agent_id=%s", c.baseURL, agentID)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("poll request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("poll failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode poll response: %w", err)
	}

	jobs, ok := result["jobs"].([]interface{})
	if !ok {
		return []interface{}{}, nil
	}

	return jobs, nil
}

// post makes a POST request to the control plane
func (c *Client) post(ctx context.Context, endpoint string, payload interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
