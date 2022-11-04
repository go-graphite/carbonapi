package main

import (
	"context"
	"fmt"
	"os/exec"
	"sync"

	"github.com/tevino/abool"
	"go.uber.org/zap"
)

type runner struct {
	App

	out string

	isRunning *abool.AtomicBool
	ctx       context.Context
	cancel    context.CancelFunc
	logger    *zap.Logger
	wg        sync.WaitGroup
}

func new(config *App, logger *zap.Logger) *runner {
	r := &runner{
		App: *config,
		logger: logger.With(
			zap.String("name", config.Name),
		),
		isRunning: abool.New(),
	}

	return r
}

func (r *runner) Run() {
	r.wg.Add(1)
	defer r.wg.Done()

	if !r.isRunning.SetToIf(false, true) {
		r.logger.Fatal("app is already running")
	}
	r.logger.Debug("will start application",
		zap.Any("config", r.App),
	)
	r.ctx, r.cancel = context.WithCancel(context.Background())

	cmd := exec.CommandContext(r.ctx, r.Binary, r.Args...)
	out, err := cmd.CombinedOutput()
	r.out = string(out)
	if err != nil {
		r.logger.Error("error running program",
			zap.Any("config", r.App),
			zap.String("output", "will follow next"),
			zap.Error(err),
		)
		fmt.Print(r.out)
	}

	r.isRunning.UnSet()
}

func (r *runner) Finish() {
	r.cancel()
	r.wg.Wait()
}

func (r *runner) IsRunning() bool {
	return r.isRunning.IsSet()
}

func (r *runner) Out() string {
	return r.out
}
