package interfaces

import (
	"context"

	"github.com/go-graphite/carbonapi/expr/types"
	zipperTypes "github.com/go-graphite/carbonapi/zipper/types"
	pb "github.com/go-graphite/protocol/carbonapi_v3_pb"
)

// The CarbonZipper interface exposes access to realZipper
// Exposes the functionality to find, get info or render metrics.
type CarbonZipper interface {
	Find(ctx context.Context, metrics []string) (*pb.MultiGlobResponse, *zipperTypes.Stats, error)
	Info(ctx context.Context, metrics []string) (*pb.ZipperInfoResponse, *zipperTypes.Stats, error)
	RenderCompat(ctx context.Context, metrics []string, from, until int64) ([]*types.MetricData, *zipperTypes.Stats, error)
	Render(ctx context.Context, request pb.MultiFetchRequest) ([]*types.MetricData, *zipperTypes.Stats, error)
}
