package gearbox

import (
	"context"
	"log/slog"

	"github.com/jonbonazza/gearbox/http"
	"github.com/jonbonazza/gearbox/service"
)

type Listener interface {
	OnError(err error)
}

type ServiceRegistry interface {
	AddService(s service.Service)
}

func (c Config) AppConfig() Config {
	return c
}

type Environment struct {
	config         Config
	httpAPIs       []http.API
	serviceManager *service.Manager
	logger         *slog.Logger
}

func newEnvironment(config Config, logger *slog.Logger) *Environment {
	return &Environment{
		config:         config,
		logger:         logger,
		serviceManager: service.NewManager(logger),
	}
}

func (e *Environment) AddHTTPAPI(api http.API) {
	e.httpAPIs = append(e.httpAPIs, api)
}

func (e *Environment) Services() ServiceRegistry {
	return e.serviceManager
}

func (e *Environment) Logger() *slog.Logger {
	return e.logger
}

func (e *Environment) run(ctx context.Context) {
	httpServer := http.NewServer(e.config.HTTP, e.logger, e.httpAPIs...)
	e.serviceManager.AddService(httpServer)
	e.serviceManager.Start(ctx)
	<-ctx.Done()
}

func (e *Environment) shutdown(ctx context.Context) error {
	return e.serviceManager.Shutdown(ctx)
}
