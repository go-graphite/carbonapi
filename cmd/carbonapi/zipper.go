package main

import (
	"context"
	"sort"

	"github.com/ansel1/merry"
	"github.com/go-graphite/carbonapi/expr/helper"
	tags2 "github.com/go-graphite/carbonapi/expr/tags"
	"github.com/go-graphite/carbonapi/expr/types"
	util "github.com/go-graphite/carbonapi/util/ctx"
	realZipper "github.com/go-graphite/carbonapi/zipper"
	zipperCfg "github.com/go-graphite/carbonapi/zipper/config"
	zipperTypes "github.com/go-graphite/carbonapi/zipper/types"
	pb "github.com/go-graphite/protocol/carbonapi_v3_pb"
	"go.uber.org/zap"
)

type zipper struct {
	z *realZipper.Zipper

	logger              *zap.Logger
	statsSender         func(*zipperTypes.Stats)
	ignoreClientTimeout bool
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

func (z zipper) Find(ctx context.Context, req pb.MultiGlobRequest) (*pb.MultiGlobResponse, *zipperTypes.Stats, merry.Error) {
	newCtx := ctx
	if z.ignoreClientTimeout {
		uuid := util.GetUUID(ctx)
		hdrs := util.GetPassHeaders(ctx)
		newCtx = util.SetUUID(context.Background(), uuid)
		newCtx = util.SetPassHeaders(newCtx, hdrs)
	}

	res, stats, err := z.z.FindProtoV3(newCtx, &req)
	z.statsSender(stats)

	return res, stats, err
}

func (z zipper) Expand(ctx context.Context, req pb.MultiGlobRequest) (*pb.MultiGlobResponse, *zipperTypes.Stats, merry.Error) {
	newCtx := ctx
	if z.ignoreClientTimeout {
		uuid := util.GetUUID(ctx)
		hdrs := util.GetPassHeaders(ctx)
		newCtx = util.SetUUID(context.Background(), uuid)
		newCtx = util.SetPassHeaders(newCtx, hdrs)
	}

	res, stats, err := z.z.ExpandProtoV3(newCtx, &req)
	z.statsSender(stats)

	return res, stats, err
}

func (z zipper) Info(ctx context.Context, metrics []string) (*pb.ZipperInfoResponse, *zipperTypes.Stats, merry.Error) {
	newCtx := ctx
	if z.ignoreClientTimeout {
		uuid := util.GetUUID(ctx)
		hdrs := util.GetPassHeaders(ctx)
		newCtx = util.SetUUID(context.Background(), uuid)
		newCtx = util.SetPassHeaders(newCtx, hdrs)
	}

	req := pb.MultiGlobRequest{
		Metrics: metrics,
	}

	resp, stats, err := z.z.InfoProtoV3(newCtx, &req)
	z.statsSender(stats)

	return resp, stats, err
}

func (z zipper) Render(ctx context.Context, request pb.MultiFetchRequest) ([]*types.MetricData, *zipperTypes.Stats, merry.Error) {
	var result []*types.MetricData
	newCtx := ctx
	if z.ignoreClientTimeout {
		uuid := util.GetUUID(ctx)
		hdrs := util.GetPassHeaders(ctx)
		newCtx = util.SetUUID(context.Background(), uuid)
		newCtx = util.SetPassHeaders(newCtx, hdrs)
	}

	pbresp, stats, err := z.z.FetchProtoV3(newCtx, &request)
	z.statsSender(stats)

	if pbresp != nil {
		for i := range pbresp.Metrics {
			tags := tags2.ExtractTags(pbresp.Metrics[i].Name)
			result = append(result, &types.MetricData{
				FetchResponse: pbresp.Metrics[i],
				Tags:          tags,
			})
		}
	}

	sort.Sort(helper.ByNameNatural(result))

	return result, stats, err
}

func (z zipper) RenderCompat(ctx context.Context, metrics []string, from, until int64) ([]*types.MetricData, *zipperTypes.Stats, merry.Error) {
	req := pb.MultiFetchRequest{}
	for _, metric := range metrics {
		req.Metrics = append(req.Metrics, pb.FetchRequest{
			Name:      metric,
			StartTime: from,
			StopTime:  until,
		})
	}

	return z.Render(ctx, req)
}

func (z zipper) TagNames(ctx context.Context, query string, limit int64) ([]string, merry.Error) {
	return z.z.TagNames(ctx, query, limit)
}

func (z zipper) TagValues(ctx context.Context, query string, limit int64) ([]string, merry.Error) {
	return z.z.TagValues(ctx, query, limit)
}

func (z zipper) ScaleToCommonStep() bool {
	return z.z.ScaleToCommonStep
}
