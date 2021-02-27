package types

import (
	"context"

	"github.com/ansel1/merry"
	protov3 "github.com/go-graphite/protocol/carbonapi_v3_pb"
)

type Request interface {
	Marshal() ([]byte, merry.Error)
	LogInfo() interface{}
}

type BackendServer interface {
	Name() string
	Backends() []string
	MaxMetricsPerRequest() int

	Fetch(ctx context.Context, request *protov3.MultiFetchRequest) (*protov3.MultiFetchResponse, *Stats, merry.Error)
	Find(ctx context.Context, request *protov3.MultiGlobRequest) (*protov3.MultiGlobResponse, *Stats, merry.Error)
	Info(ctx context.Context, request *protov3.MultiMetricsInfoRequest) (*protov3.ZipperInfoResponse, *Stats, merry.Error)

	List(ctx context.Context) (*protov3.ListMetricsResponse, *Stats, merry.Error)
	Stats(ctx context.Context) (*protov3.MetricDetailsResponse, *Stats, merry.Error)

	ProbeTLDs(ctx context.Context) ([]string, merry.Error)

	TagNames(ctx context.Context, query string, limit int64) ([]string, merry.Error)
	TagValues(ctx context.Context, query string, limit int64) ([]string, merry.Error)

	Children() []BackendServer
}
