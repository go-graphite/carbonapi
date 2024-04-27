package types

import (
	"context"

	protov3 "github.com/go-graphite/protocol/carbonapi_v3_pb"
)

type Request interface {
	Marshal() ([]byte, error)
	LogInfo() interface{}
}

type BackendServer interface {
	Name() string
	Backends() []string
	MaxMetricsPerRequest() int

	Fetch(ctx context.Context, request *protov3.MultiFetchRequest) (*protov3.MultiFetchResponse, *Stats, error)
	Find(ctx context.Context, request *protov3.MultiGlobRequest) (*protov3.MultiGlobResponse, *Stats, error)
	Info(ctx context.Context, request *protov3.MultiMetricsInfoRequest) (*protov3.ZipperInfoResponse, *Stats, error)

	List(ctx context.Context) (*protov3.ListMetricsResponse, *Stats, error)
	Stats(ctx context.Context) (*protov3.MetricDetailsResponse, *Stats, error)

	ProbeTLDs(ctx context.Context) ([]string, error)

	TagNames(ctx context.Context, query string, limit int64) ([]string, error)
	TagValues(ctx context.Context, query string, limit int64) ([]string, error)

	Children() []BackendServer
}
