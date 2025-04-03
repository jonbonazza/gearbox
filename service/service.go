package service

import (
	"context"
	"errors"
	"log/slog"
	"sync"
)

type Initializer interface {
	Init(ctx context.Context) error
}

type Shutdowner interface {
	Shutdown(ctx context.Context) error
}

type Service interface {
	Run(ctx context.Context) error
}

type Listener interface {
	OnError(s Service, err error)
}

type serviceError struct {
	s   Service
	err error
}

type Manager struct {
	logger    *slog.Logger
	services  []Service
	listeners []Listener
	wg        sync.WaitGroup
}

func NewManager(logger *slog.Logger) *Manager {
	return &Manager{logger: logger}
}

func (m *Manager) AddService(s Service) {
	m.services = append(m.services, s)
}

func (m *Manager) AddListener(l Listener) {
	m.listeners = append(m.listeners, l)
}

func (m *Manager) Start(ctx context.Context) {
	m.logger.Info("Started services")

	errCh := make(chan serviceError)
	m.wg.Add(len(m.services))
	for _, s := range m.services {
		if initializer, ok := s.(Initializer); ok {
			if err := initializer.Init(ctx); err != nil {
				m.notifyError(serviceError{s, err})
				continue
			}
		}
		go m.runService(ctx, s, errCh)
	}

	go func() {
		defer m.wg.Done()
		for err := range errCh {
			m.notifyError(err)
		}
	}()

	go func() {
		<-ctx.Done()
		close(errCh)
	}()
}

func (m *Manager) Shutdown(ctx context.Context) (err error) {
	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		slog.With(slog.Any("err", ctx.Err())).Error("Context expired while waiting for service manager to shut down")
		return ctx.Err()
	case <-done:
	}

	for _, s := range m.services {
		if shutdowner, ok := s.(Shutdowner); ok {
			if e := shutdowner.Shutdown(ctx); e != nil {
				err = errors.Join(err, e)
			}
		}
	}

	if err != nil {
		m.logger.With(slog.Any("err", err)).Error("Error(s) while shutting down services")
	} else {
		m.logger.Info("Services shut down successfully")
	}

	return err
}

func (m *Manager) runService(ctx context.Context, s Service, errCh chan<- serviceError) {
	defer m.wg.Done()
	if err := s.Run(ctx); err != nil {
		errCh <- serviceError{s, err}
	}
}

func (m *Manager) notifyError(err serviceError) {
	for _, l := range m.listeners {
		l.OnError(err.s, err.err)
	}
}
