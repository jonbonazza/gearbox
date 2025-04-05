package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sync/atomic"

	"github.com/gin-gonic/gin"
)

var (
	ErrAlreadyStarted = errors.New("already started")
	ErrNotStarted     = errors.New("not started")
)

type API interface {
	AddRoutes(router *gin.RouterGroup)
}

type Config struct {
	BindAddr string
}

type Server struct {
	config      Config
	middlewares []gin.HandlerFunc
	apis        []API
	logger      *slog.Logger
	router      *gin.Engine
	httpServer  *http.Server
	started     atomic.Bool
}

func NewServer(config Config, logger *slog.Logger, middlewares []gin.HandlerFunc, apis ...API) *Server {
	return &Server{
		config:      config,
		logger:      logger,
		middlewares: middlewares,
		apis:        apis,
	}
}

func (s *Server) Init(ctx context.Context) error {
	s.logger.Info("Initializing HTTP API")
	s.router = gin.Default()

	s.router.Use(s.middlewares...)

	for _, api := range s.apis {
		api.AddRoutes(&s.router.RouterGroup)
	}

	return nil
}

func (s *Server) Run(ctx context.Context) error {
	if !s.started.CompareAndSwap(false, true) {
		slog.Warn("Can't start HTTP Server that's already started")
		return ErrAlreadyStarted
	}

	s.httpServer = &http.Server{
		Addr:    s.config.BindAddr,
		Handler: s.router.Handler(),
	}

	s.logger.With(slog.String("bindAddr", s.config.BindAddr)).Info("HTTP server started")
	err := s.httpServer.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		s.logger.Info("HTTP server stopped")
		err = nil
	}

	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
	if !s.started.Load() {
		slog.Warn("Can't shut down HTTP server that's not started")
		return ErrNotStarted
	}

	if err := s.httpServer.Shutdown(ctx); err != nil {
		slog.With(slog.Any("err", err)).Error("Failed to shut down HTTP server")
		return err
	}

	s.httpServer = nil
	s.router = nil
	s.started.Store(false)

	return nil
}
