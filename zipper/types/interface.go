package types

import (
	"context"

	protov3 "github.com/go-graphite/protocol/carbonapi_v3_pb"
)

type ServerClient interface {
	Name() string
	Backends() []string

	Fetch(ctx context.Context, request *protov3.MultiFetchRequest) (*protov3.MultiFetchResponse, *Stats, error)
	Find(ctx context.Context, request *protov3.MultiGlobRequest) (*protov3.MultiGlobResponse, *Stats, error)
	Info(ctx context.Context, request *protov3.MultiMetricsInfoRequest) (*protov3.ZipperInfoResponse, *Stats, error)

	List(ctx context.Context) (*protov3.ListMetricsResponse, *Stats, error)
	Stats(ctx context.Context) (*protov3.MetricDetailsResponse, *Stats, error)

	ProbeTLDs(ctx context.Context) ([]string, error)
}

/*
type Fetcher interface {
	// PB-compatible methods
	FetchProtoV2(ctx context.Context, query []string, startTime, stopTime int32) (*protov2.MultiFetchResponse, *Stats, error)
	FindProtoV2(ctx context.Context, query []string) (*protov2.GlobResponse, *Stats, error)

	InfoProtoV2(ctx context.Context, targets []string) (*protov2.ZipperInfoResponse, *Stats, error)
	ListProtoV2(ctx context.Context) (*protov2.ListMetricsResponse, *Stats, error)
	StatsProtoV2(ctx context.Context) (*protov2.MetricDetailsResponse, *Stats, error)

	// GRPC-compatible methods
	FetchProtoV3(ctx context.Context, request *protov3.MultiFetchRequest) (*protov3.MultiFetchResponse, *Stats, error)
	FindProtoV3(ctx context.Context, request *protov3.MultiGlobRequest) ([]*protov3.MultiGlobResponse, *Stats, error)

	InfoProtoV3(ctx context.Context, request *protov3.MultiMetricsInfoRequest) (*protov3.ZipperInfoResponse, *Stats, error)
	ListProtoV3(ctx context.Context) (*protov3.ListMetricsResponse, *Stats, error)
	StatsProtoV3(ctx context.Context) (*protov3.MetricDetailsResponse, *Stats, error)
}
*/
