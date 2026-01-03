package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/intrik8-labs/smidr/internal/core/registry"
	"github.com/intrik8-labs/smidr/internal/logging"
)

// WebServer handles web-facing API endpoints (for UI, CLI, etc.)
type WebServer struct {
	router *chi.Mux
	server *http.Server
	store  registry.Store
	logger *logging.Logger
}

// NewWebServer creates a new web API server
func NewWebServer(addr string, store registry.Store, logger *logging.Logger) *WebServer {
	s := &WebServer{
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

func (s *WebServer) setupRoutes() {
	s.router.Get("/health", s.healthCheck)

	// Agent management endpoints
	s.router.Get("/api/v1/agents", s.listAgents)
	s.router.Get("/api/v1/agents/{agentID}", s.getAgent)

	// Job management endpoints
	s.router.Get("/api/v1/jobs", s.listJobs)
	s.router.Post("/api/v1/jobs", s.createJob)
	s.router.Get("/api/v1/jobs/{jobID}", s.getJob)
}

func (s *WebServer) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"server": "web-api",
	})
}

func (s *WebServer) listAgents(w http.ResponseWriter, r *http.Request) {
	agents := s.store.List(r.Context())

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agents": agents,
		"count":  len(agents),
	})
}

func (s *WebServer) getAgent(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	if agentID == "" {
		http.Error(w, "agent_id required", http.StatusBadRequest)
		return
	}

	agent, err := s.store.Get(r.Context(), agentID)
	if err != nil {
		s.logger.Error("failed to get agent", "agent_id", agentID, "error", err)
		http.Error(w, "agent not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agent)
}

func (s *WebServer) listJobs(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement job listing logic
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"jobs":  []interface{}{},
		"count": 0,
	})
}

func (s *WebServer) createJob(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement job creation logic
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "created",
		"message": "job creation not yet implemented",
	})
}

func (s *WebServer) getJob(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "jobID")
	if jobID == "" {
		http.Error(w, "job_id required", http.StatusBadRequest)
		return
	}

	// TODO: Implement job retrieval logic
	w.Header().Set("Content-Type", "application/json")
	http.Error(w, "job retrieval not yet implemented", http.StatusNotImplemented)
}

// Start begins listening for HTTP requests
func (s *WebServer) Start() error {
	s.logger.Info("starting web API server", "address", s.server.Addr)
	return s.server.ListenAndServe()
}

// Stop gracefully shuts down the server
func (s *WebServer) Stop(ctx context.Context) error {
	s.logger.Info("stopping web API server")
	return s.server.Shutdown(ctx)
}
