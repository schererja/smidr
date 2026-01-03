package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/intrik8-labs/smidr/internal/core/registry"
	"github.com/intrik8-labs/smidr/internal/logging"
)

// AgentServer handles agent-facing API endpoints
type AgentServer struct {
	router *chi.Mux
	server *http.Server
	store  registry.Store
	logger *logging.Logger
}

// RegisterRequest is the payload for agent registration
type RegisterRequest struct {
	AgentID      string            `json:"agent_id"`
	Name         string            `json:"name"`
	Capabilities []string          `json:"capabilities"`
	Metadata     map[string]string `json:"metadata"`
}

// HeartbeatRequest is the payload for agent heartbeat
type HeartbeatRequest struct {
	AgentID string `json:"agent_id"`
	Status  string `json:"status"`
}

// NewAgentServer creates a new agent API server
func NewAgentServer(addr string, store registry.Store, logger *logging.Logger) *AgentServer {
	s := &AgentServer{
		router: chi.NewRouter(),
		store:  store,
		logger: logger,
	}

	// Middleware
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)

	// Routes
	s.setupRoutes()

	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s
}

func (s *AgentServer) setupRoutes() {
	s.router.Get("/health", s.healthCheck)
	s.router.Post("/api/v1/register", s.register)
	s.router.Post("/api/v1/heartbeat", s.heartbeat)
	s.router.Get("/api/v1/jobs/poll", s.pollJobs)
}

func (s *AgentServer) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"server": "agent-api",
	})
}

func (s *AgentServer) register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Error("failed to decode register request", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := s.store.Register(r.Context(), req.AgentID, req.Name, req.Capabilities, req.Metadata); err != nil {
		s.logger.Error("failed to register agent", "agent_id", req.AgentID, "error", err)
		http.Error(w, fmt.Sprintf("registration failed: %v", err), http.StatusInternalServerError)
		return
	}

	s.logger.Info("agent registered", "agent_id", req.AgentID, "name", req.Name)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "registered",
		"message": fmt.Sprintf("agent %s registered successfully", req.AgentID),
	})
}

func (s *AgentServer) heartbeat(w http.ResponseWriter, r *http.Request) {
	var req HeartbeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Error("failed to decode heartbeat request", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := s.store.Heartbeat(r.Context(), req.AgentID); err != nil {
		s.logger.Error("heartbeat failed", "agent_id", req.AgentID, "error", err)
		http.Error(w, fmt.Sprintf("heartbeat failed: %v", err), http.StatusNotFound)
		return
	}

	// Update status if provided
	if req.Status != "" {
		if err := s.store.UpdateStatus(r.Context(), req.AgentID, registry.AgentStatus(req.Status)); err != nil {
			s.logger.Warn("failed to update agent status", "agent_id", req.AgentID, "error", err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

func (s *AgentServer) pollJobs(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Query().Get("agent_id")
	if agentID == "" {
		http.Error(w, "agent_id required", http.StatusBadRequest)
		return
	}

	// TODO: Implement job polling logic
	// For now, return empty job list
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"jobs": []interface{}{},
	})
}

// Start begins listening for HTTP requests
func (s *AgentServer) Start() error {
	s.logger.Info("starting agent API server", "address", s.server.Addr)
	return s.server.ListenAndServe()
}

// Stop gracefully shuts down the server
func (s *AgentServer) Stop(ctx context.Context) error {
	s.logger.Info("stopping agent API server")
	return s.server.Shutdown(ctx)
}
