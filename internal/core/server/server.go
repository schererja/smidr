package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/intrik8-labs/smidr/internal/logging"
)

// Server represents the core server
type Server struct {
	router     *chi.Mux
	httpServer *http.Server
	address    string
	port       int
	log        *logging.Logger
}

// HealthCheckResponse represents the response from the healthcheck endpoint
type HealthCheckResponse struct {
	Status  string    `json:"status"`
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}

func NewServer(address string, port int, log *logging.Logger) *Server {
	return &Server{
		address: address,
		port:    port,
		log:     log,
		router:  chi.NewRouter(),
	}
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	// Middleware
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Timeout(60 * time.Second))

	// Routes
	s.router.Get("/healthcheck", s.healthcheck)

	// 404 handler
	s.router.NotFound(s.notFound)
}

// healthcheck handles the /healthcheck endpoint
func (s *Server) healthcheck(w http.ResponseWriter, r *http.Request) {
	s.log.Debug("Healthcheck request received")

	response := HealthCheckResponse{
		Status:  "healthy",
		Message: "Core server is running",
		Time:    time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// notFound handles undefined routes
func (s *Server) notFound(w http.ResponseWriter, r *http.Request) {
	s.log.Debug("Not found request",
		logging.String("path", r.RequestURI),
		logging.String("method", r.Method))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{
		"error": "endpoint not found",
	})
}

func (s *Server) Start() error {
	s.setupRoutes()

	addr := fmt.Sprintf("%s:%d", s.address, s.port)
	s.log.Info("Starting core server",
		logging.String("address", s.address),
		logging.Int("port", s.port))

	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.log.Error("Server error", logging.Err(err))
		}
	}()

	s.log.Info("Core server started successfully",
		logging.String("address", addr))

	return nil
}

func (s *Server) Stop() error {
	if s.httpServer == nil {
		return nil
	}

	s.log.Info("Stopping core server")

	// Graceful shutdown with 30 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.log.Error("Error stopping server", logging.Err(err))
		return err
	}

	s.log.Info("Core server stopped successfully")
	return nil
}
