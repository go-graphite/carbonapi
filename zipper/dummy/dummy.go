package dummy

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/go-graphite/carbonapi/zipper/errors"
	"github.com/go-graphite/carbonapi/zipper/types"
	protov3 "github.com/go-graphite/protocol/carbonapi_v3_pb"
)

type FetchResponse struct {
	Response *protov3.MultiFetchResponse
	Stats    *types.Stats
	Errors   *errors.Errors
}

type FindResponse struct {
	Response *protov3.MultiGlobResponse
	Stats    *types.Stats
	Errors   *errors.Errors
}

type InfoResponse struct {
	Response *protov3.ZipperInfoResponse
	Stats    *types.Stats
	Errors   *errors.Errors
}

type ListResponse struct {
	Response *protov3.ListMetricsResponse
	Stats    *types.Stats
	Errors   *errors.Errors
}

type StatsResponse struct {
	Response *protov3.MetricDetailsResponse
	Stats    *types.Stats
	Errors   *errors.Errors
}

type ProbeResponse struct {
	Response []string
	Errors   *errors.Errors
}

type DummyClient struct {
	name                 string
	backends             []string
	maxMetricsPerRequest int

	fetchResponses map[string]FetchResponse
	findResponses  map[string]FindResponse
	infoResponses  map[string]InfoResponse
	statsResponses map[string]StatsResponse
	probeResponses ProbeResponse
	alwaysTimeout  time.Duration
}

func NewDummyClient(name string, backends []string, maxMetricsPerRequest int) *DummyClient {
	return &DummyClient{
		name:                 name,
		backends:             backends,
		maxMetricsPerRequest: maxMetricsPerRequest,

		fetchResponses: make(map[string]FetchResponse),
		findResponses:  make(map[string]FindResponse),
		infoResponses:  make(map[string]InfoResponse),
		statsResponses: make(map[string]StatsResponse),
		alwaysTimeout:  0,
	}
}

func NewDummyClientWithTimeout(name string, backends []string, maxMetricsPerRequest int, alwaysTimeout time.Duration) *DummyClient {
	return &DummyClient{
		name:                 name,
		backends:             backends,
		maxMetricsPerRequest: maxMetricsPerRequest,

		fetchResponses: make(map[string]FetchResponse),
		findResponses:  make(map[string]FindResponse),
		infoResponses:  make(map[string]InfoResponse),
		statsResponses: make(map[string]StatsResponse),
		alwaysTimeout:  alwaysTimeout,
	}
}

func (c *DummyClient) Name() string {
	return c.name
}

func (c *DummyClient) Backends() []string {
	return c.backends
}

func (c *DummyClient) MaxMetricsPerRequest() int {
	return c.maxMetricsPerRequest
}

func fetchRequestToKey(request *protov3.MultiFetchRequest) string {
	var key []byte
	for _, r := range request.Metrics {
		key = append(key, []byte("&"+r.Name+"&start="+strconv.FormatUint(uint64(r.StartTime), 10)+"&stop="+strconv.FormatUint(uint64(r.StopTime), 10)+"\n")...)
	}

	return string(key)
}

func (c *DummyClient) AddFetchResponse(request *protov3.MultiFetchRequest, response *protov3.MultiFetchResponse, stats *types.Stats, errors *errors.Errors) {
	key := fetchRequestToKey(request)
	c.fetchResponses[key] = FetchResponse{response, stats, errors}

	for _, m := range request.Metrics {
		findRequest := protov3.MultiGlobRequest{
			Metrics: []string{
				m.Name,
			},
		}
		findResponse := protov3.MultiGlobResponse{
			Metrics: []protov3.GlobResponse{
				{
					Name: m.Name,
					Matches: []protov3.GlobMatch{
						{
							Path:   m.Name,
							IsLeaf: true,
						},
					},
				},
			},
		}
		c.AddFindResponse(&findRequest, &findResponse, &types.Stats{}, errors)
	}
}

func (c *DummyClient) Fetch(ctx context.Context, request *protov3.MultiFetchRequest) (*protov3.MultiFetchResponse, *types.Stats, *errors.Errors) {
	if c.alwaysTimeout > 0 {
		time.Sleep(c.alwaysTimeout)
		return nil, nil, errors.Fatalf("timeout fetching response")
	}
	select {
	case <-ctx.Done():
		return nil, nil, errors.Fatalf("timeout fetching response")
	default:
	}

	key := fetchRequestToKey(request)
	r, ok := c.fetchResponses[key]
	if ok {
		return r.Response, r.Stats, r.Errors
	}

	return nil, nil, nil
}

func findRequestToKey(request *protov3.MultiGlobRequest) string {
	return strings.Join(request.Metrics, "&")
}

func (c *DummyClient) AddFindResponse(request *protov3.MultiGlobRequest, response *protov3.MultiGlobResponse, stats *types.Stats, errors *errors.Errors) {
	key := findRequestToKey(request)
	c.findResponses[key] = FindResponse{response, stats, errors}
}

func (c *DummyClient) Find(ctx context.Context, request *protov3.MultiGlobRequest) (*protov3.MultiGlobResponse, *types.Stats, *errors.Errors) {
	if c.alwaysTimeout > 0 {
		time.Sleep(c.alwaysTimeout)
		return nil, nil, errors.Fatalf("timeout fetching response")
	}
	select {
	case <-ctx.Done():
		return nil, nil, errors.Fatalf("timeout fetching response")
	default:
	}

	r, ok := c.findResponses[findRequestToKey(request)]
	if ok {
		return r.Response, r.Stats, r.Errors
	}
	return nil, nil, nil
}

func (c *DummyClient) Info(ctx context.Context, request *protov3.MultiMetricsInfoRequest) (*protov3.ZipperInfoResponse, *types.Stats, *errors.Errors) {
	return nil, nil, errors.Fatalf("not implemented")
}

func (c *DummyClient) List(ctx context.Context) (*protov3.ListMetricsResponse, *types.Stats, *errors.Errors) {
	return nil, nil, errors.Fatalf("not implemented")
}

func (c *DummyClient) Stats(ctx context.Context) (*protov3.MetricDetailsResponse, *types.Stats, *errors.Errors) {
	return nil, nil, errors.Fatalf("not implemented")
}

func (c *DummyClient) SetTLDResponse(response ProbeResponse) {
	c.probeResponses = response
}

func (c *DummyClient) ProbeTLDs(ctx context.Context) ([]string, *errors.Errors) {
	return c.probeResponses.Response, c.probeResponses.Errors
}
