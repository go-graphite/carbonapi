package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-graphite/carbonapi/expr"
	pb "github.com/go-graphite/carbonzipper/carbonzipperpb3"
	realZipper "github.com/go-graphite/carbonzipper/zipper"
	"go.uber.org/zap"
)

var errNoMetrics = errors.New("no metrics")

type zipper struct {
	z *realZipper.Zipper

	logger      *zap.Logger
	statsSender func(*realZipper.Stats)
	ignoreClientTimeout bool
}

// The CarbonZipper interface exposes access to realZipper
// Exposes the functionality to find, get info or render metrics.
type CarbonZipper interface {
	Find(ctx context.Context, metric string) (pb.GlobResponse, error)
	Info(ctx context.Context, metric string) (map[string]pb.InfoResponse, error)
	Render(ctx context.Context, metric string, from, until int32) ([]*expr.MetricData, error)
}

func newZipper(sender func(*realZipper.Stats), config *realZipper.Config, ignoreClientTimeout bool, logger *zap.Logger) *zipper {
	z := &zipper{
		z:           realZipper.NewZipper(sender, config, logger),
		logger:      logger,
		statsSender: sender,
		ignoreClientTimeout: ignoreClientTimeout,
	}

	return z
}

func (z zipper) Find(ctx context.Context, metric string) (pb.GlobResponse, error) {
	var pbresp pb.GlobResponse
	newCtx := ctx
	if z.ignoreClientTimeout {
		newCtx = context.Background()
	}

	res, stats, err := z.z.Find(newCtx, z.logger, metric)
	if err != nil {
		return pbresp, err
	}

	pbresp.Name = metric
	pbresp.Matches = res

	z.statsSender(stats)

	return pbresp, err
}

func (z zipper) Info(ctx context.Context, metric string) (map[string]pb.InfoResponse, error) {
	newCtx := ctx
	if z.ignoreClientTimeout {
		newCtx = context.Background()
	}
	resp, stats, err := z.z.Info(newCtx, z.logger, metric)
	if err != nil {
		return nil, fmt.Errorf("http.Get: %+v", err)
	}

	z.statsSender(stats)

	return resp, nil
}

func (z zipper) Render(ctx context.Context, metric string, from, until int32) ([]*expr.MetricData, error) {
	var result []*expr.MetricData
	newCtx := ctx
	if z.ignoreClientTimeout {
		newCtx = context.Background()
	}
	pbresp, stats, err := z.z.Render(newCtx, z.logger, metric, from, until)
	if err != nil {
		return result, err
	}

	z.statsSender(stats)

	if m := pbresp.Metrics; len(m) == 0 {
		return result, errNoMetrics
	}

	for i := range pbresp.Metrics {
		result = append(result, &expr.MetricData{FetchResponse: pbresp.Metrics[i]})
	}

	return result, nil
}
