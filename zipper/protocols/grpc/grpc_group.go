package grpc

import (
	"context"
	"github.com/ansel1/merry"
	"math"

	"github.com/go-graphite/carbonapi/limiter"
	"github.com/go-graphite/carbonapi/zipper/metadata"
	"github.com/go-graphite/carbonapi/zipper/types"
	protov3grpc "github.com/go-graphite/protocol/carbonapi_v3_grpc"
	protov3 "github.com/go-graphite/protocol/carbonapi_v3_pb"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/resolver/manual"

	"go.uber.org/zap"
)

func init() {
	aliases := []string{"carbonapi_v3_grpc", "proto_v3_grpc", "v3_grpc"}
	metadata.Metadata.Lock()
	for _, name := range aliases {
		metadata.Metadata.SupportedProtocols[name] = struct{}{}
		metadata.Metadata.ProtocolInits[name] = NewClientGRPCGroup
		metadata.Metadata.ProtocolInitsWithLimiter[name] = NewClientGRPCGroupWithLimiter
	}
	defer metadata.Metadata.Unlock()
}

// RoundRobin is used to connect to backends inside clientGRPCGroups, implements BackendServer interface
type ClientGRPCGroup struct {
	groupName string
	servers   []string

	r                    *manual.Resolver
	conn                 *grpc.ClientConn
	cleanup              func()
	timeout              types.Timeouts
	maxMetricsPerRequest int

	client protov3grpc.CarbonV1Client
	logger *zap.Logger
}

func (c *ClientGRPCGroup) Children() []types.BackendServer {
	return []types.BackendServer{c}
}

func NewClientGRPCGroupWithLimiter(logger *zap.Logger, config types.BackendV2, limiter limiter.ServerLimiter) (types.BackendServer, merry.Error) {
	return NewClientGRPCGroup(logger, config)
}

func NewClientGRPCGroup(logger *zap.Logger, config types.BackendV2) (types.BackendServer, merry.Error) {
	logger = logger.With(zap.String("type", "grpcGroup"), zap.String("name", config.GroupName))
	// TODO: Implement normal resolver
	if len(config.Servers) == 0 {
		return nil, types.ErrNoServersSpecified
	}
	r, cleanup := manual.GenerateAndRegisterManualResolver()
	var resolvedAddrs []resolver.Address
	for _, addr := range config.Servers {
		resolvedAddrs = append(resolvedAddrs, resolver.Address{Addr: addr})
	}

	r.UpdateState(resolver.State{
		Addresses: resolvedAddrs,
	})

	opts := []grpc.DialOption{
		grpc.WithUserAgent("carbonzipper"),
		grpc.WithCompressor(grpc.NewGZIPCompressor()),
		grpc.WithDecompressor(grpc.NewGZIPDecompressor()),
		grpc.WithBalancerName("round_robin"), // TODO: Make that configurable
		grpc.WithMaxMsgSize(math.MaxInt32),  // TODO: make that configurable
		grpc.WithInsecure(),                  // TODO: Make configurable
	}

	conn, err := grpc.Dial(r.Scheme()+":///server", opts...)
	if err != nil {
		cleanup()
		return nil, merry.Wrap(err)
	}

	client := &ClientGRPCGroup{
		groupName:            config.GroupName,
		servers:              config.Servers,
		maxMetricsPerRequest: config.MaxBatchSize,

		r:       r,
		cleanup: cleanup,
		conn:    conn,
		client:  protov3grpc.NewCarbonV1Client(conn),
		timeout: *config.Timeouts,
		logger:  logger,
	}

	return client, nil
}

func (c ClientGRPCGroup) MaxMetricsPerRequest() int {
	return c.maxMetricsPerRequest
}

func (c ClientGRPCGroup) Name() string {
	return c.groupName
}

func (c ClientGRPCGroup) Backends() []string {
	return c.servers
}

func (c *ClientGRPCGroup) Fetch(ctx context.Context, request *protov3.MultiFetchRequest) (*protov3.MultiFetchResponse, *types.Stats, merry.Error) {
	stats := &types.Stats{
		Servers: []string{c.Name()},
	}
	ctx, cancel := context.WithTimeout(ctx, c.timeout.Render)
	defer cancel()

	res, err := c.client.FetchMetrics(ctx, request)
	if err != nil {
		stats.RenderErrors++
		stats.FailedServers = stats.Servers
		stats.Servers = []string{}
	}
	stats.MemoryUsage = int64(res.Size())

	return res, stats, merry.Wrap(err)
}

func (c *ClientGRPCGroup) Find(ctx context.Context, request *protov3.MultiGlobRequest) (*protov3.MultiGlobResponse, *types.Stats, merry.Error) {
	stats := &types.Stats{
		Servers: []string{c.Name()},
	}
	ctx, cancel := context.WithTimeout(ctx, c.timeout.Find)
	defer cancel()

	res, err := c.client.FindMetrics(ctx, request)
	if err != nil {
		stats.RenderErrors++
		stats.FailedServers = stats.Servers
		stats.Servers = []string{}
	}
	stats.MemoryUsage = int64(res.Size())

	return res, stats, merry.Wrap(err)
}
func (c *ClientGRPCGroup) Info(ctx context.Context, request *protov3.MultiMetricsInfoRequest) (*protov3.ZipperInfoResponse, *types.Stats, merry.Error) {
	stats := &types.Stats{
		Servers: []string{c.Name()},
	}
	ctx, cancel := context.WithTimeout(ctx, c.timeout.Render)
	defer cancel()

	res, err := c.client.MetricsInfo(ctx, request)
	if err != nil {
		stats.RenderErrors++
		stats.FailedServers = stats.Servers
		stats.Servers = []string{}
	}
	stats.MemoryUsage = int64(res.Size())

	r := &protov3.ZipperInfoResponse{
		Info: map[string]protov3.MultiMetricsInfoResponse{
			c.Name(): *res,
		},
	}

	return r, stats, merry.Wrap(err)
}

func (c *ClientGRPCGroup) List(ctx context.Context) (*protov3.ListMetricsResponse, *types.Stats, merry.Error) {
	stats := &types.Stats{
		Servers: []string{c.Name()},
	}
	ctx, cancel := context.WithTimeout(ctx, c.timeout.Render)
	defer cancel()

	res, err := c.client.ListMetrics(ctx, types.EmptyMsg)
	if err != nil {
		stats.RenderErrors++
		stats.FailedServers = stats.Servers
		stats.Servers = []string{}
	}
	stats.MemoryUsage = int64(res.Size())

	return res, stats, merry.Wrap(err)
}
func (c *ClientGRPCGroup) Stats(ctx context.Context) (*protov3.MetricDetailsResponse, *types.Stats, merry.Error) {
	stats := &types.Stats{
		Servers: []string{c.Name()},
	}
	ctx, cancel := context.WithTimeout(ctx, c.timeout.Render)
	defer cancel()

	res, err := c.client.Stats(ctx, types.EmptyMsg)
	if err != nil {
		stats.RenderErrors++
		stats.FailedServers = stats.Servers
		stats.Servers = []string{}
	}
	stats.MemoryUsage = int64(res.Size())

	return res, stats, merry.Wrap(err)
}

func (c *ClientGRPCGroup) TagNames(ctx context.Context, query string, limit int64) ([]string, merry.Error) {
	return nil, types.ErrNotImplementedYet
}

func (c *ClientGRPCGroup) TagValues(ctx context.Context, query string, limit int64) ([]string, merry.Error) {
	return nil, types.ErrNotImplementedYet
}

func (c *ClientGRPCGroup) ProbeTLDs(ctx context.Context) ([]string, merry.Error) {
	logger := c.logger.With(zap.String("type", "probe"))

	ctx, cancel := context.WithTimeout(ctx, c.timeout.Find)
	defer cancel()

	req := &protov3.MultiGlobRequest{
		Metrics: []string{"*"},
	}

	logger.Debug("doing request",
		zap.Any("request", req),
	)

	res, _, err := c.Find(ctx, req)
	if err != nil {
		return nil, err
	}
	var tlds []string
	for _, m := range res.Metrics {
		for _, v := range m.Matches {
			tlds = append(tlds, v.Path)
		}
	}
	return tlds, nil
}
