package zipper

import (
	"context"
	"math"
	_ "net/http/pprof"

	"github.com/ansel1/merry"
	"github.com/go-graphite/carbonapi/zipper/types"
	protov2 "github.com/go-graphite/protocol/carbonapi_v2_pb"
	protov3 "github.com/go-graphite/protocol/carbonapi_v3_pb"
	"go.uber.org/zap"

	_ "github.com/go-graphite/carbonapi/zipper/protocols/auto"
	_ "github.com/go-graphite/carbonapi/zipper/protocols/graphite"
	_ "github.com/go-graphite/carbonapi/zipper/protocols/prometheus"
	_ "github.com/go-graphite/carbonapi/zipper/protocols/v2"
	_ "github.com/go-graphite/carbonapi/zipper/protocols/v3"
)

// DEPRECATED
// This file contains legacy functions for carbonzipper old protocol versions
// Likely they will be removed in future

// PB3-compatible methods
func (z Zipper) FetchProtoV2(ctx context.Context, query []string, startTime, stopTime int32) (*protov2.MultiFetchResponse, *types.Stats, merry.Error) {
	logger := z.logger.With(zap.String("function", "FetchProtoV2"))
	request := &protov3.MultiFetchRequest{}
	for _, q := range query {
		request.Metrics = append(request.Metrics, protov3.FetchRequest{
			Name:      q,
			StartTime: int64(startTime),
			StopTime:  int64(stopTime),
		})
	}

	grpcRes, stats, err := z.FetchProtoV3(ctx, request)
	if err != nil {
		if grpcRes == nil || len(grpcRes.Metrics) == 0 {
			return nil, nil, err
		} else {
			logger.Debug("had errors while fetching result",
				zap.Any("errors", err),
			)
		}
	}

	var res protov2.MultiFetchResponse
	for i := range grpcRes.Metrics {
		vals := make([]float64, 0, len(grpcRes.Metrics[i].Values))
		isAbsent := make([]bool, 0, len(grpcRes.Metrics[i].Values))
		for _, v := range grpcRes.Metrics[i].Values {
			if math.IsNaN(v) {
				vals = append(vals, 0)
				isAbsent = append(isAbsent, true)
			} else {
				vals = append(vals, v)
				isAbsent = append(isAbsent, false)
			}
		}
		res.Metrics = append(res.Metrics,
			protov2.FetchResponse{
				Name:      grpcRes.Metrics[i].Name,
				StartTime: int32(grpcRes.Metrics[i].StartTime),
				StopTime:  int32(grpcRes.Metrics[i].StopTime),
				StepTime:  int32(grpcRes.Metrics[i].StepTime),
				Values:    vals,
				IsAbsent:  isAbsent,
			})
	}

	return &res, stats, nil
}

func (z Zipper) FindProtoV2(ctx context.Context, query []string) ([]*protov2.GlobResponse, *types.Stats, merry.Error) {
	logger := z.logger.With(zap.String("function", "FindProtoV2"))
	request := &protov3.MultiGlobRequest{
		Metrics: query,
	}
	grpcRes, stats, err := z.FindProtoV3(ctx, request)
	if err != nil {
		if grpcRes == nil || len(grpcRes.Metrics) == 0 {
			return nil, nil, err
		} else {
			logger.Debug("had errors while fetching result",
				zap.Any("errors", err),
			)
		}
	}

	reses := make([]*protov2.GlobResponse, 0, len(grpcRes.Metrics))
	for _, grpcRes := range grpcRes.Metrics {

		res := &protov2.GlobResponse{
			Name: grpcRes.Name,
		}

		for _, v := range grpcRes.Matches {
			match := protov2.GlobMatch{
				Path:   v.Path,
				IsLeaf: v.IsLeaf,
			}
			res.Matches = append(res.Matches, match)
		}
		reses = append(reses, res)
	}

	return reses, stats, nil
}

func (z Zipper) InfoProtoV2(ctx context.Context, targets []string) (*protov2.ZipperInfoResponse, *types.Stats, merry.Error) {
	logger := z.logger.With(zap.String("function", "InfoProtoV2"))
	request := &protov3.MultiGlobRequest{
		Metrics: targets,
	}
	grpcRes, stats, err := z.InfoProtoV3(ctx, request)
	if err != nil {
		if grpcRes == nil || len(grpcRes.Info) == 0 {
			return nil, nil, err
		} else {
			logger.Debug("had errors while fetching result",
				zap.Any("errors", err),
			)
		}
	}

	res := &protov2.ZipperInfoResponse{}

	for k, i := range grpcRes.Info {
		for _, v := range i.Metrics {
			rets := make([]protov2.Retention, 0, len(v.Retentions))
			for _, ret := range v.Retentions {
				rets = append(rets, protov2.Retention{
					SecondsPerPoint: int32(ret.SecondsPerPoint),
					NumberOfPoints:  int32(ret.NumberOfPoints),
				})
			}
			i := &protov2.InfoResponse{
				Name:              v.Name,
				AggregationMethod: v.ConsolidationFunc,
				MaxRetention:      int32(v.MaxRetention),
				XFilesFactor:      v.XFilesFactor,
				Retentions:        rets,
			}
			res.Responses = append(res.Responses, protov2.ServerInfoResponse{
				Server: k,
				Info:   i,
			})
		}
	}

	return res, stats, nil
}
func (z Zipper) ListProtoV2(ctx context.Context) (*protov2.ListMetricsResponse, *types.Stats, merry.Error) {
	logger := z.logger.With(zap.String("function", "ListProtoV2"))
	grpcRes, stats, err := z.ListProtoV3(ctx)
	if err != nil {
		if grpcRes == nil || len(grpcRes.Metrics) == 0 {
			return nil, nil, err
		} else {
			logger.Debug("had errors while fetching result",
				zap.Any("errors", err),
			)
		}
	}

	res := &protov2.ListMetricsResponse{
		Metrics: grpcRes.Metrics,
	}
	return res, stats, nil
}
func (z Zipper) StatsProtoV2(ctx context.Context) (*protov2.MetricDetailsResponse, *types.Stats, merry.Error) {
	logger := z.logger.With(zap.String("function", "StatsProtoV2"))
	grpcRes, stats, err := z.StatsProtoV3(ctx)
	if err != nil {
		if grpcRes == nil || len(grpcRes.Metrics) == 0 {
			return nil, nil, err
		} else {
			logger.Debug("had errors while fetching result",
				zap.Any("errors", err),
			)
		}
	}

	metrics := make(map[string]*protov2.MetricDetails, len(grpcRes.Metrics))
	for k, v := range grpcRes.Metrics {
		metrics[k] = &protov2.MetricDetails{
			Size_:   v.Size_,
			ModTime: v.ModTime,
			ATime:   v.ATime,
			RdTime:  v.RdTime,
		}
	}

	res := &protov2.MetricDetailsResponse{
		FreeSpace:  grpcRes.FreeSpace,
		TotalSpace: grpcRes.TotalSpace,
		Metrics:    metrics,
	}

	return res, stats, nil
}
