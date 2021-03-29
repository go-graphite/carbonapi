package main

import (
	"context"
	"fmt"
	"os/exec"

	"go.uber.org/zap"
)

type runner struct {
	App

	isRunning bool
	ctx       context.Context
	cancel    context.CancelFunc
	logger    *zap.Logger
}

func new(config *App, logger *zap.Logger) *runner {
	r := &runner{
		App: *config,
		logger: logger.With(
			zap.String("name", config.Name),
		),
	}

	return r
}

func (r *runner) Run() {
	if r.isRunning {
		r.logger.Fatal("app is already running")
	}
	r.logger.Debug("will start application",
		zap.Any("config", r.App),
	)
	r.ctx, r.cancel = context.WithCancel(context.Background())

	cmd := exec.CommandContext(r.ctx, r.Binary, r.Args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		r.logger.Error("error running program",
			zap.Any("config", r.App),
			zap.String("output", "will follow next"),
			zap.Error(err),
		)
		fmt.Print(string(out))
	}

	r.isRunning = true
}

func (r *runner) Finish() {
	r.cancel()

	r.isRunning = false
}
