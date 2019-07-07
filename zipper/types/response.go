package types

import (
	"context"
	"math"

	"github.com/go-graphite/carbonapi/zipper/errors"
	protov3 "github.com/go-graphite/protocol/carbonapi_v3_pb"
	"github.com/lomik/zapwriter"
	"go.uber.org/zap"
)

// type Fetcher func(ctx context.Context, logger *zap.Logger, client types.BackendServer, reqs interface{}, resCh chan<- types.ServerFetchResponse) {
//type Fetcher func(ctx context.Context, logger *zap.Logger, client BackendServer, reqs interface{}, resCh chan ServerFetchResponse) {
type Fetcher func(ctx context.Context, logger *zap.Logger, client BackendServer, reqs interface{}, resCh chan ServerFetcherResponse)

type ServerFetcherResponse interface {
	Self() interface{}
	MergeI(second ServerFetcherResponse) *errors.Errors
	Errors() *errors.Errors
	GetServer() string
}

func NoAnswerBackends(backends []BackendServer, answered map[string]struct{}) []string {
	noAnswer := make([]string, 0)
	for _, s := range backends {
		if _, ok := answered[s.Name()]; !ok {
			noAnswer = append(noAnswer, s.Name())
		}
	}

	return noAnswer
}

// Helper function
func DoRequest(ctx context.Context, logger *zap.Logger, clients []BackendServer, result ServerFetcherResponse, requests interface{}, fetcher Fetcher) (ServerFetcherResponse, int) {
	resCh := make(chan ServerFetcherResponse, len(clients))

	for _, client := range clients {
		logger.Debug("single fetch",
			zap.Any("client", client),
		)
		go fetcher(ctx, logger, client, requests, resCh)
	}

	answeredServers := make(map[string]struct{})
	responseCount := 0
GATHER:
	for responseCount < len(clients) {
		select {
		case res := <-resCh:
			answeredServers[res.GetServer()] = struct{}{}
			result.MergeI(res)
			responseCount++

		case <-ctx.Done():
			logger.Warn("timeout waiting for more responses",
				zap.Strings("no_answers_from", NoAnswerBackends(clients, answeredServers)),
			)
			result.Errors().Add(ErrTimeoutExceeded)

			break GATHER
		}
	}
	return result, responseCount
}

type ServerTagResponse struct {
	Server   string
	Response []string
	Err      *errors.Errors
}

func NewServerTagResponse() *ServerTagResponse {
	return &ServerTagResponse{
		Response: []string{},
		Err:      new(errors.Errors),
	}
}

func (s *ServerTagResponse) Self() interface{} {
	return s
}

func (s ServerTagResponse) GetServer() string {
	return s.Server
}

func (first *ServerTagResponse) MergeI(second ServerFetcherResponse) *errors.Errors {
	secondSelf := second.Self()
	s, ok := secondSelf.(*ServerTagResponse)
	if !ok {
		return errors.Fatalf("got '%T', expected '%T'", secondSelf, first)
	}
	return first.Merge(s)
}

func (first *ServerTagResponse) Errors() *errors.Errors {
	return first.Err
}

func (first *ServerTagResponse) Merge(second *ServerTagResponse) *errors.Errors {
	if first.Err == nil {
		first.Err = new(errors.Errors)
	}
	first.Err.Merge(second.Err)

	if second.Response == nil {
		return first.Err
	}

	// We cannot assume in general that results are sorted
	firstMap := make(map[string]struct{}, len(first.Response))
	for _, v := range first.Response {
		firstMap[v] = struct{}{}
	}

	for _, v := range second.Response {
		if _, ok := firstMap[v]; !ok {
			first.Response = append(first.Response, v)
		}
	}

	return nil
}

type ServerInfoResponse struct {
	Server   string
	Response *protov3.ZipperInfoResponse
	Stats    *Stats
	Err      *errors.Errors
}

func NewServerInfoResponse() *ServerInfoResponse {
	return &ServerInfoResponse{
		Response: &protov3.ZipperInfoResponse{Info: make(map[string]protov3.MultiMetricsInfoResponse)},
		Stats:    new(Stats),
		Err:      new(errors.Errors),
	}
}

func (s *ServerInfoResponse) Self() interface{} {
	return s
}

func (s ServerInfoResponse) GetServer() string {
	return s.Server
}

func (first *ServerInfoResponse) MergeI(second ServerFetcherResponse) *errors.Errors {
	secondSelf := second.Self()
	s, ok := secondSelf.(*ServerInfoResponse)
	if !ok {
		return errors.Fatalf("got '%T', expected '%T'", secondSelf, first)
	}
	return first.Merge(s)
}

func (first *ServerInfoResponse) Errors() *errors.Errors {
	return first.Err
}

func (first *ServerInfoResponse) Merge(second *ServerInfoResponse) *errors.Errors {
	if second.Stats != nil {
		first.Stats.Merge(second.Stats)
	}

	if first.Err == nil {
		first.Err = new(errors.Errors)
	}
	first.Err.Merge(second.Err)

	if second.Response == nil {
		return first.Err
	}

	for k, v := range second.Response.Info {
		first.Response.Info[k] = v
	}

	return nil
}

type ServerFindResponse struct {
	Server   string
	Response *protov3.MultiGlobResponse
	Stats    *Stats
	Err      *errors.Errors
}

func NewServerFindResponse() *ServerFindResponse {
	return &ServerFindResponse{
		Response: new(protov3.MultiGlobResponse),
		Stats:    new(Stats),
		Err:      new(errors.Errors),
	}
}

func (s *ServerFindResponse) Self() interface{} {
	return s
}

func (s ServerFindResponse) GetServer() string {
	return s.Server
}

func (first *ServerFindResponse) MergeI(second ServerFetcherResponse) *errors.Errors {
	secondSelf := second.Self()
	s, ok := secondSelf.(*ServerFindResponse)
	if !ok {
		return errors.Fatalf("got '%T', expected '%T'", secondSelf, first)
	}
	return first.Merge(s)
}

func (first *ServerFindResponse) Errors() *errors.Errors {
	return first.Err
}

func (first *ServerFindResponse) Merge(second *ServerFindResponse) *errors.Errors {
	if second.Stats != nil {
		first.Stats.Merge(second.Stats)
	}

	if first.Err == nil {
		first.Err = new(errors.Errors)
	}
	first.Err.Merge(second.Err)

	if second.Response == nil {
		return first.Err
	}

	seenMetrics := make(map[string]int)
	seenMatches := make(map[string]struct{})
	for i, m := range first.Response.Metrics {
		seenMetrics[m.Name] = i
		for _, mm := range m.Matches {
			seenMatches[m.Name+"."+mm.Path] = struct{}{}
		}
	}

	var i int
	var ok bool
	for _, m := range second.Response.Metrics {
		if i, ok = seenMetrics[m.Name]; !ok {
			first.Response.Metrics = append(first.Response.Metrics, m)
			continue
		}

		for _, mm := range m.Matches {
			key := first.Response.Metrics[i].Name + "." + mm.Path
			if _, ok := seenMatches[key]; !ok {
				seenMatches[key] = struct{}{}
				first.Response.Metrics[i].Matches = append(first.Response.Metrics[i].Matches, mm)
			}
		}
	}

	return nil
}

type ServerFetchResponse struct {
	Server   string
	Response *protov3.MultiFetchResponse
	Stats    *Stats
	Err      *errors.Errors
}

func NewServerFetchResponse() *ServerFetchResponse {
	return &ServerFetchResponse{
		Response: new(protov3.MultiFetchResponse),
		Stats:    new(Stats),
		Err:      new(errors.Errors),
	}
}

func (s *ServerFetchResponse) Self() interface{} {
	return s
}

func (s ServerFetchResponse) GetServer() string {
	return s.Server
}

func (first *ServerFetchResponse) MergeI(second ServerFetcherResponse) *errors.Errors {
	secondSelf := second.Self()
	s, ok := secondSelf.(*ServerFetchResponse)
	if !ok {
		return errors.Fatalf("got '%T', expected '%T'", secondSelf, first)
	}
	return first.Merge(s)
}

func (first *ServerFetchResponse) Errors() *errors.Errors {
	return first.Err
}

func (s *ServerFetchResponse) NonFatalError(err error) *ServerFetchResponse {
	s.Err.Add(err)
	return s
}

func swapFetchResponses(m1, m2 *protov3.FetchResponse) {
	m1.Name, m2.Name = m2.Name, m1.Name
	m1.StartTime, m2.StartTime = m2.StartTime, m1.StartTime
	m1.StepTime, m2.StepTime = m2.StepTime, m1.StepTime
	m1.ConsolidationFunc, m2.ConsolidationFunc = m2.ConsolidationFunc, m1.ConsolidationFunc
	m1.XFilesFactor, m2.XFilesFactor = m2.XFilesFactor, m1.XFilesFactor
	m1.Values, m2.Values = m2.Values, m1.Values
	m1.AppliedFunctions, m2.AppliedFunctions = m2.AppliedFunctions, m1.AppliedFunctions
	m1.StopTime, m2.StopTime = m2.StopTime, m1.StopTime
}

func mergeFetchResponsesWithEqualStepTimes(m1, m2 *protov3.FetchResponse) error {
	if m1.StartTime != m2.StartTime {
		return ErrResponseStartTimeMismatch
	}

	if len(m1.Values) < len(m2.Values) {
		swapFetchResponses(m1, m2)
	}

	for i := 0; i < len(m2.Values); i++ {
		if math.IsNaN(m1.Values[i]) {
			m1.Values[i] = m2.Values[i]
		}
	}

	return nil
}

func mergeFetchResponsesWithUnequalStepTimes(m1, m2 *protov3.FetchResponse) error {
	if m1.StepTime > m2.StepTime {
		swapFetchResponses(m1, m2)
	}

	zapwriter.Logger("zipper").Warn("Fetch responses had different step times",
		zap.Int64("m1_request_start_time", m1.RequestStartTime),
		zap.Int64("m1_start_time", m1.StartTime),
		zap.Int64("m1_stop_time", m1.StopTime),
		zap.Int64("m1_step_time", m1.StepTime),
		zap.Int64("m2_request_start_time", m2.RequestStartTime),
		zap.Int64("m2_start_time", m2.StartTime),
		zap.Int64("m2_stop_time", m2.StopTime),
		zap.Int64("m2_step_time", m2.StepTime),
	)

	return nil
}

func MergeFetchResponses(m1, m2 *protov3.FetchResponse) *errors.Errors {
	var err error
	if m1.RequestStartTime != m2.RequestStartTime {
		err = ErrResponseStartTimeMismatch
	} else if m1.StepTime == m2.StepTime {
		err = mergeFetchResponsesWithEqualStepTimes(m1, m2)
	} else {
		err = mergeFetchResponsesWithUnequalStepTimes(m1, m2)
	}

	if err != nil {
		zapwriter.Logger("zipper").Error("Unable to merge fetch responses",
			zap.Error(err),
			zap.Int64("m1_request_start_time", m1.RequestStartTime),
			zap.Int64("m1_start_time", m1.StartTime),
			zap.Int64("m1_stop_time", m1.StopTime),
			zap.Int64("m1_step_time", m1.StepTime),
			zap.Int64("m2_request_start_time", m2.RequestStartTime),
			zap.Int64("m2_start_time", m2.StartTime),
			zap.Int64("m2_stop_time", m2.StopTime),
			zap.Int64("m2_step_time", m2.StepTime),
		)
	}

	return errors.FromErr(err)
}

func (first *ServerFetchResponse) Merge(second *ServerFetchResponse) *errors.Errors {
	if first.Server == "" && second.Server != "" {
		first.Server = second.Server
	}

	if second.Stats != nil {
		first.Stats.Merge(second.Stats)
	}

	if first.Err == nil {
		first.Err = &errors.Errors{}
	}
	first.Err.Merge(second.Err)

	if first.Err.HaveFatalErrors {
		return first.Err
	}

	if second.Response == nil {
		return errors.Fatalf("no response to merge")
	}

	metrics := make(map[fetchResponseCoordinates]int)
	for i := range first.Response.Metrics {
		metrics[coordinates(&first.Response.Metrics[i])] = i
	}

	for i := range second.Response.Metrics {
		if j, ok := metrics[coordinates(&second.Response.Metrics[i])]; ok {
			err := MergeFetchResponses(&first.Response.Metrics[j], &second.Response.Metrics[i])
			if err != nil {
				// TODO: Normal error handling
				continue
			}
		} else {
			first.Response.Metrics = append(first.Response.Metrics, second.Response.Metrics[i])
		}
	}
	return nil
}

type fetchResponseCoordinates struct {
	name  string
	from  int64
	until int64
}

func coordinates(r *protov3.FetchResponse) fetchResponseCoordinates {
	return fetchResponseCoordinates{
		name:  r.Name,
		from:  r.RequestStartTime,
		until: r.RequestStopTime,
	}
}

type ServerResponse struct {
	Server   string
	Response []byte
}
