package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/util"
	realZipper "github.com/go-graphite/carbonapi/zipper"
	zipperCfg "github.com/go-graphite/carbonapi/zipper/config"
	zipperTypes "github.com/go-graphite/carbonapi/zipper/types"
	pb "github.com/go-graphite/protocol/carbonapi_v3_pb"
	"go.uber.org/zap"
)

var errNoMetrics = errors.New("no metrics")

type zipper struct {
	z *realZipper.Zipper

	logger              *zap.Logger
	statsSender         func(*zipperTypes.Stats)
	ignoreClientTimeout bool
}

// The CarbonZipper interface exposes access to realZipper
// Exposes the functionality to find, get info or render metrics.
type CarbonZipper interface {
	Find(ctx context.Context, metrics []string) (*pb.MultiGlobResponse, error)
	Info(ctx context.Context, metrics []string) (*pb.ZipperInfoResponse, error)
	RenderCompat(ctx context.Context, metrics []string, from, until int64) ([]*types.MetricData, error)
	Render(ctx context.Context, request pb.MultiFetchRequest) ([]*types.MetricData, error)
}

func newZipper(sender func(*zipperTypes.Stats), config *zipperCfg.Config, ignoreClientTimeout bool, logger *zap.Logger) *zipper {
	logger.Debug("initializing zipper")
	zz, err := realZipper.NewZipper(sender, config, logger)
	if err != nil {
		logger.Fatal("failed to initialize zipper",
			zap.Error(err),
		)
		return nil
	}
	z := &zipper{
		z:                   zz,
		logger:              logger,
		statsSender:         sender,
		ignoreClientTimeout: ignoreClientTimeout,
	}

	return z
}

func (z zipper) Find(ctx context.Context, metrics []string) (*pb.MultiGlobResponse, error) {
	newCtx := ctx
	if z.ignoreClientTimeout {
		uuid := util.GetUUID(ctx)
		newCtx = util.SetUUID(context.Background(), uuid)
	}

	req := pb.MultiGlobRequest{
		Metrics: metrics,
	}

	res, stats, err := z.z.FindProtoV3(newCtx, &req)
	if err != nil {
		return nil, err
	}

	z.statsSender(stats)

	return res, err
}

func (z zipper) Info(ctx context.Context, metrics []string) (*pb.ZipperInfoResponse, error) {
	newCtx := ctx
	if z.ignoreClientTimeout {
		uuid := util.GetUUID(ctx)
		newCtx = util.SetUUID(context.Background(), uuid)
	}

	req := pb.MultiGlobRequest{
		Metrics: metrics,
	}

	resp, stats, err := z.z.InfoProtoV3(newCtx, &req)
	if err != nil {
		return nil, fmt.Errorf("http.Get: %+v", err)
	}

	z.statsSender(stats)

	return resp, nil
}

func (z zipper) Render(ctx context.Context, request pb.MultiFetchRequest) ([]*types.MetricData, error) {
	var result []*types.MetricData
	newCtx := ctx
	if z.ignoreClientTimeout {
		uuid := util.GetUUID(ctx)
		newCtx = util.SetUUID(context.Background(), uuid)
	}

	pbresp, stats, err := z.z.FetchProtoV3(newCtx, &request)
	if err != nil {
		return result, err
	}

	z.statsSender(stats)

	for i := range pbresp.Metrics {
		result = append(result, &types.MetricData{FetchResponse: pbresp.Metrics[i]})
	}

	if len(result) == 0 {
		return result, errNoMetrics
	}

	return result, nil
}

func (z zipper) RenderCompat(ctx context.Context, metrics []string, from, until int64) ([]*types.MetricData, error) {
	var result []*types.MetricData
	newCtx := ctx
	if z.ignoreClientTimeout {
		uuid := util.GetUUID(ctx)
		newCtx = util.SetUUID(context.Background(), uuid)
	}

	req := pb.MultiFetchRequest{}
	for _, metric := range metrics {
		req.Metrics = append(req.Metrics, pb.FetchRequest{
			Name:      metric,
			StartTime: from,
			StopTime:  until,
		})
	}

	pbresp, stats, err := z.z.FetchProtoV3(newCtx, &req)
	if err != nil {
		return result, err
	}

	z.statsSender(stats)

	if m := pbresp.Metrics; len(m) == 0 {
		return result, errNoMetrics
	}

	for i := range pbresp.Metrics {
		result = append(result, &types.MetricData{FetchResponse: pbresp.Metrics[i]})
	}

	return result, nil
}
