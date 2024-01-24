package interfaces

import (
	"context"

	"github.com/ansel1/merry"

	"github.com/go-graphite/carbonapi/expr/types"
	zipperTypes "github.com/go-graphite/carbonapi/zipper/types"
	pb "github.com/go-graphite/protocol/carbonapi_v3_pb"
)

// The CarbonZipper interface exposes access to realZipper
// Exposes the functionality to find, get info or render metrics.
type CarbonZipper interface {
	Find(ctx context.Context, request pb.MultiGlobRequest) (*pb.MultiGlobResponse, *zipperTypes.Stats, merry.Error)
	Info(ctx context.Context, metrics []string) (*pb.ZipperInfoResponse, *zipperTypes.Stats, merry.Error)
	RenderCompat(ctx context.Context, metrics []string, from, until int64) ([]*types.MetricData, *zipperTypes.Stats, merry.Error)
	Render(ctx context.Context, request pb.MultiFetchRequest) ([]*types.MetricData, *zipperTypes.Stats, merry.Error)
	TagNames(ctx context.Context, query string, limit int64) ([]string, merry.Error)
	TagValues(ctx context.Context, query string, limit int64) ([]string, merry.Error)
	ScaleToCommonStep() bool
}
