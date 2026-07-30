package handler

import (
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/scouser-122/go-metrics/internal/config"
	"github.com/scouser-122/go-metrics/internal/logger"
)

// Server entity for processing http requests
type Server struct {
	httpServer *http.Server
	config     *config.ServerConfig
	wg         sync.WaitGroup
}

// NewServer creates new server entity
func NewServer(config *config.ServerConfig) *Server {
	return &Server{
		config: config,
	}
}

func (s *Server) Init(handlers []Handler) error {
	r := chi.NewRouter()
	err := AddHandlersForRouter(r, &handlers, s.config)
	if err != nil {
		return err
	}
	s.httpServer = &http.Server{
		Addr:    s.config.RunAddr,
		Handler: r,
	}
	return nil
}

// Start starts server
func (s *Server) Start() error {
	logger.Sugar.Infof("starting server on %s", s.config.RunAddr)
	return s.httpServer.ListenAndServe()
}

// Shutdown shuts down server
func (s *Server) Shutdown() {
	logger.Sugar.Infof("shutting down server on %s", s.config.RunAddr)

	ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}

	log.Println("Server gracefully stopped")
}
