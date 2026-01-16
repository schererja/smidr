package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterAgent(t *testing.T) {
	tests := []struct {
		name          string
		agentID       string
		expectedError bool
		statusCode    int
	}{
		{
			name:          "successful registration",
			agentID:       "agent-1",
			expectedError: false,
			statusCode:    http.StatusCreated,
		},
		{
			name:          "server error",
			agentID:       "agent-2",
			expectedError: true,
			statusCode:    http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/register" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				if r.Method != "POST" {
					t.Errorf("unexpected method: %s", r.Method)
				}
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			client := NewClient(server.URL)
			err := client.RegisterAgent(context.Background(), tt.agentID, "test-agent", []string{"yocto"}, map[string]string{})

			if (err != nil) != tt.expectedError {
				t.Errorf("expected error: %v, got: %v", tt.expectedError, err != nil)
			}
		})
	}
}

func TestHeartbeat(t *testing.T) {
	tests := []struct {
		name          string
		agentID       string
		status        string
		expectedError bool
		statusCode    int
	}{
		{
			name:          "successful heartbeat",
			agentID:       "agent-1",
			status:        "idle",
			expectedError: false,
			statusCode:    http.StatusOK,
		},
		{
			name:          "agent not found",
			agentID:       "unknown-agent",
			status:        "idle",
			expectedError: true,
			statusCode:    http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/heartbeat" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				if r.Method != "POST" {
					t.Errorf("unexpected method: %s", r.Method)
				}
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			client := NewClient(server.URL)
			err := client.Heartbeat(context.Background(), tt.agentID, tt.status)

			if (err != nil) != tt.expectedError {
				t.Errorf("expected error: %v, got: %v", tt.expectedError, err != nil)
			}
		})
	}
}

func TestPollJobs(t *testing.T) {
	tests := []struct {
		name          string
		agentID       string
		responseBody  string
		expectedError bool
		statusCode    int
		expectedJobs  int
	}{
		{
			name:          "successful poll with jobs",
			agentID:       "agent-1",
			responseBody:  `{"jobs":[{"id":"job-1"},{"id":"job-2"}]}`,
			expectedError: false,
			statusCode:    http.StatusOK,
			expectedJobs:  2,
		},
		{
			name:          "successful poll with no jobs",
			agentID:       "agent-1",
			responseBody:  `{"jobs":[]}`,
			expectedError: false,
			statusCode:    http.StatusOK,
			expectedJobs:  0,
		},
		{
			name:          "server error",
			agentID:       "agent-1",
			responseBody:  "",
			expectedError: true,
			statusCode:    http.StatusInternalServerError,
			expectedJobs:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/jobs/poll" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				if r.Method != "GET" {
					t.Errorf("unexpected method: %s", r.Method)
				}
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := NewClient(server.URL)
			jobs, err := client.PollJobs(context.Background(), tt.agentID)

			if (err != nil) != tt.expectedError {
				t.Errorf("expected error: %v, got: %v", tt.expectedError, err != nil)
			}

			if len(jobs) != tt.expectedJobs {
				t.Errorf("expected %d jobs, got %d", tt.expectedJobs, len(jobs))
			}
		})
	}
}

func TestClientTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {}
	}))
	defer server.Close()

	client := NewClient(server.URL)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := client.RegisterAgent(ctx, "agent-1", "test", []string{}, map[string]string{})
	if err == nil {
		t.Error("expected timeout error")
	}
}
