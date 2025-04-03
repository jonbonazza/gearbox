package gearbox

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Configurator interface {
	AppConfig() Config
}

type Finalizer interface {
	Finalize(ctx context.Context)
}

type App[C Configurator] interface {
	Init(env *Environment, config C) error
}

func Run[C Configurator](ctx context.Context, name, version string, app App[C], config C) error {
	logger := newLogger(config.AppConfig().Logging)

	logger = logger.With(
		slog.String("appName", name),
		slog.String("appVersion", version),
	)

	env := newEnvironment(config.AppConfig(), logger)

	logger.Info("Initializing application")

	if err := app.Init(env, config); err != nil {
		logger.With(slog.Any("err", err)).Error("Failed to initialize application")
		return fmt.Errorf("failed to initialize application: %w", err)
	}

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	logger.Info("Application is running")
	env.run(ctx)

	if ctx.Err() != nil {
		logger.With(slog.Any("err", ctx.Err())).Error("Context error")
		return ctx.Err()
	}

	if err := shutdown(env, app, config.AppConfig().ShutdownTimeout); err != nil {
		logger.With(slog.Any("err", err)).Error("Failed to shutdown application environment")
		return err
	}

	logger.Info("Application shut down")
	return nil
}

func shutdown[C Configurator](env *Environment, app App[C], timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := env.shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown application environment: %w", err)
	}

	if finalizer, ok := app.(Finalizer); ok {
		finalizer.Finalize(ctx)
	}

	return nil
}
