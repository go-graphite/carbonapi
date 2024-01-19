package auto

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/ansel1/merry"
	protov3 "github.com/go-graphite/protocol/carbonapi_v3_pb"
	"go.uber.org/zap"

	"github.com/go-graphite/carbonapi/internal/dns"
	"github.com/go-graphite/carbonapi/limiter"
	"github.com/go-graphite/carbonapi/zipper/broadcast"
	"github.com/go-graphite/carbonapi/zipper/helper"
	"github.com/go-graphite/carbonapi/zipper/httpHeaders"
	"github.com/go-graphite/carbonapi/zipper/metadata"
	"github.com/go-graphite/carbonapi/zipper/types"
)

func init() {
	aliases := []string{"auto"}
	metadata.Metadata.Lock()
	for _, name := range aliases {
		metadata.Metadata.SupportedProtocols[name] = struct{}{}
		metadata.Metadata.ProtocolInits[name] = New
		metadata.Metadata.ProtocolInitsWithLimiter[name] = NewWithLimiter
	}
	defer metadata.Metadata.Unlock()
}

type capabilityResponse struct {
	server   string
	protocol string
}

// _internal/capabilities/
func doQuery(ctx context.Context, logger *zap.Logger, groupName string, httpClient *http.Client, limiter limiter.ServerLimiter, server string, request types.Request, resChan chan<- capabilityResponse) {
	httpQuery := helper.NewHttpQuery(groupName, []string{server}, 1, limiter, httpClient, httpHeaders.ContentTypeCarbonAPIv3PB)
	rewrite, _ := url.Parse("http://127.0.0.1/_internal/capabilities/")

	res, e := httpQuery.DoQuery(ctx, logger, rewrite.RequestURI(), request)
	if e != nil || res == nil || res.Response == nil || len(res.Response) == 0 {
		logger.Info("will assume old protocol")
		resChan <- capabilityResponse{
			server:   server,
			protocol: "protobuf",
		}
		return
	}

	response := protov3.CapabilityResponse{}
	logger.Debug("response",
		zap.String("server", res.Server),
		zap.String("response", string(res.Response)),
	)
	err := response.Unmarshal(res.Response)

	if err != nil {
		resChan <- capabilityResponse{
			server:   server,
			protocol: "protobuf",
		}
		return
	}

	resChan <- capabilityResponse{
		server:   server,
		protocol: response.SupportedProtocols[0],
	}

}

type CapabilityResponse struct {
	ProtoToServers map[string][]string
}

func getBestSupportedProtocol(logger *zap.Logger, servers []string) *CapabilityResponse {
	response := &CapabilityResponse{
		ProtoToServers: make(map[string][]string),
	}
	groupName := "capability query"
	l := limiter.NoopLimiter{}

	httpClient := &http.Client{
		Transport: &http.Transport{
			DialContext: dns.GetDialContextWithTimeout(200*time.Millisecond, 30*time.Second),
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	request := types.CapabilityRequestV3{}
	resCh := make(chan capabilityResponse, len(servers))

	for _, srv := range servers {
		go doQuery(ctx, logger, groupName, httpClient, l, srv, request, resCh)
	}

	answeredServers := make(map[string]struct{})
	responseCounts := 0
GATHER:
	for {
		if responseCounts == len(servers) && len(resCh) == 0 {
			break GATHER
		}
		select {
		case res := <-resCh:
			responseCounts++
			answeredServers[res.server] = struct{}{}
			if res.protocol == "" {
				return nil
			}
			p := response.ProtoToServers[res.protocol]
			response.ProtoToServers[res.protocol] = append(p, res.server)
		case <-ctx.Done():
			noAnswer := make([]string, 0)
			for _, s := range servers {
				if _, ok := answeredServers[s]; !ok {
					noAnswer = append(noAnswer, s)
				}
			}
			logger.Warn("timeout waiting for more responses",
				zap.Strings("no_answers_from", noAnswer),
			)
			break GATHER
		}
	}

	return response
}

type AutoGroup struct {
	groupName string
}

func NewWithLimiter(logger *zap.Logger, config types.BackendV2, tldCacheDisabled, requireSuccessAll bool, limiter limiter.ServerLimiter) (types.BackendServer, merry.Error) {
	return nil, merry.New("auto group doesn't support anything useful except for New")
}

func New(logger *zap.Logger, config types.BackendV2, tldCacheDisabled, requireSuccessAll bool) (types.BackendServer, merry.Error) {
	logger = logger.With(zap.String("type", "autoGroup"), zap.String("name", config.GroupName))

	if config.ConcurrencyLimit == nil {
		logger.Error("this behavior changes in 0.14.0, before that there was an implied concurrencyLimit of 100 for the backend. Currently it's required to specify this limit for auto backend as well.")
		return nil, types.ErrConcurrencyLimitNotSet
	}

	res := getBestSupportedProtocol(logger, config.Servers)
	if res == nil {
		return nil, merry.New("can't query all backend")
	}

	var backends []types.BackendServer
	for proto, servers := range res.ProtoToServers {
		metadata.Metadata.RLock()
		backendInit, ok := metadata.Metadata.ProtocolInits[proto]
		metadata.Metadata.RUnlock()
		if !ok {
			var protocols []string
			metadata.Metadata.RLock()
			for p := range metadata.Metadata.SupportedProtocols {
				protocols = append(protocols, p)
			}
			metadata.Metadata.RUnlock()
			logger.Error("unknown backend protocol",
				zap.Any("backend", config),
				zap.String("requested_protocol", proto),
				zap.Strings("supported_backends", protocols),
			)
			return nil, merry.New("unknown backend protocol").WithValue("protocol", proto)
		}

		cfg := config
		cfg.GroupName = config.GroupName + "_" + proto
		cfg.Servers = servers
		c, ePtr := backendInit(logger, cfg, tldCacheDisabled, requireSuccessAll)
		if ePtr != nil {
			return nil, ePtr
		}

		backends = append(backends, c)
	}

	return broadcast.NewBroadcastGroup(logger, config.GroupName+"_broadcast", config.DoMultipleRequestsIfSplit, backends,
		600, *config.ConcurrencyLimit, *config.MaxBatchSize, *config.Timeouts, tldCacheDisabled, requireSuccessAll)
}

func (c AutoGroup) MaxMetricsPerRequest() int {
	return -1
}

func (c AutoGroup) Name() string {
	return c.groupName
}

func (c AutoGroup) Backends() []string {
	return nil
}

func (c *AutoGroup) Fetch(ctx context.Context, request *protov3.MultiFetchRequest) (*protov3.MultiFetchResponse, *types.Stats, merry.Error) {
	return nil, nil, merry.New("auto group doesn't support fetch")
}

func (c *AutoGroup) Find(ctx context.Context, request *protov3.MultiGlobRequest) (*protov3.MultiGlobResponse, *types.Stats, merry.Error) {
	return nil, nil, merry.New("auto group doesn't support find")
}

func (c *AutoGroup) Info(ctx context.Context, request *protov3.MultiMetricsInfoRequest) (*protov3.ZipperInfoResponse, *types.Stats, merry.Error) {
	return nil, nil, merry.New("auto group doesn't support info")
}

func (c *AutoGroup) List(ctx context.Context) (*protov3.ListMetricsResponse, *types.Stats, merry.Error) {
	return nil, nil, merry.New("auto group doesn't support list")
}
func (c *AutoGroup) Stats(ctx context.Context) (*protov3.MetricDetailsResponse, *types.Stats, merry.Error) {
	return nil, nil, merry.New("auto group doesn't support stats")
}

func (bg *AutoGroup) TagNames(ctx context.Context, prefix string, exprs []string, limit int64) ([]string, merry.Error) {
	return nil, merry.New("auto group doesn't support tag names")
}

func (bg *AutoGroup) TagValues(ctx context.Context, tagName, prefix string, exprs []string, limit int64) ([]string, merry.Error) {
	return nil, merry.New("auto group doesn't support tag values")
}

func (c *AutoGroup) ProbeTLDs(ctx context.Context) ([]string, merry.Error) {
	return nil, merry.New("auto group doesn't support probing")
}
