package app

import (
	"context"
	"fmt"
	"io"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/victorzhuk/homyak/internal/logger"
)

const (
	EnvLocal = "local"
	EnvDev   = "development"
	EnvProd  = "production"
)

var (
	Version = "undefined"
	Commit  = "undefined"
	BuildAt = time.Now().UTC().Format(time.RFC3339)
)

func PrintBuildInfo(w io.Writer) {
	_, err := fmt.Fprintf(w, "ver:\t%s\ncommit:\t%s\ndt:\t%s\n\n", Version, Commit, BuildAt)
	if err != nil {
		panic(err)
	}
}

type worker interface {
	Start(ctx context.Context) error
}

type initializer interface {
	Initialize(ctx context.Context) error
}

type finalizer interface {
	Finalize(ctx context.Context) error
}

type app struct {
	workers      []worker
	initializers []initializer
	finalizers   []finalizer
}

func New(cfg *config) *app {
	logger.Debug("creating new app")

	workers, initializers, finalizers := bootstrap(cfg)

	return &app{
		workers:      workers,
		initializers: initializers,
		finalizers:   finalizers,
	}
}

func (a *app) Run(ctx context.Context) error {
	logger.Info("starting app")

	for _, i := range a.initializers {
		if err := i.Initialize(ctx); err != nil {
			return fmt.Errorf("error initializing worker: %w", err)
		}
	}

	defer func() {
		for _, i := range a.finalizers {
			if err := i.Finalize(ctx); err != nil {
				logger.Error("error finalizing worker", zap.Error(err))
			}
		}
	}()

	g, ctx := errgroup.WithContext(ctx)
	for _, w := range a.workers {
		g.Go(func() error {
			if err := w.Start(ctx); err != nil {
				return fmt.Errorf("error starting worker: %w", err)
			}
			return nil
		})
	}

	err := g.Wait()
	if err != nil {
		return fmt.Errorf("app run exited with error: %w", err)
	}

	return nil
}
