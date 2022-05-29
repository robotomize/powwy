package main

import (
	"context"
	"fmt"

	"github.com/robotomize/powwy/internal/logging"
	"github.com/robotomize/powwy/internal/setup"
	"github.com/robotomize/powwy/internal/shutdown"
)

func main() {
	ctx, cancel := shutdown.New()
	defer cancel()

	ctx = logging.WithLogger(ctx, logging.NewDevelopmentLogger())
	logger := logging.FromContext(ctx)

	if err := mainFunc(ctx); err != nil {
		logger.Errorf("mainFunc: %v", err)
	}
}

func mainFunc(ctx context.Context) error {
	environment, err := setup.Setup(ctx)
	if err != nil {
		return fmt.Errorf("setup.Setup: %w", err)
	}

	srv := environment.Server

	if err = srv.Serve(ctx); err != nil {
		return fmt.Errorf("srv.Serve: %w", err)
	}

	return nil
}
