package types

import (
	"context"
	"math"

	"github.com/ansel1/merry"

	protov3 "github.com/go-graphite/protocol/carbonapi_v3_pb"
	"github.com/lomik/zapwriter"
	"go.uber.org/zap"
)

// type Fetcher func(ctx context.Context, logger *zap.Logger, client types.BackendServer, reqs interface{}, resCh chan<- types.ServerFetchResponse) {
// type Fetcher func(ctx context.Context, logger *zap.Logger, client BackendServer, reqs interface{}, resCh chan ServerFetchResponse) {
type Fetcher func(ctx context.Context, logger *zap.Logger, client BackendServer, reqs interface{}, resCh chan ServerFetcherResponse)

type ServerFetcherResponse interface {
	Self() interface{}
	MergeI(second ServerFetcherResponse) merry.Error
	AddError(err merry.Error)
	Errors() []merry.Error
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
func DoRequest(ctx context.Context, logger *zap.Logger, clients []BackendServer, result ServerFetcherResponse, request interface{}, fetcher Fetcher) (ServerFetcherResponse, int) {
	resCh := make(chan ServerFetcherResponse, len(clients))

	for _, client := range clients {
		logger.Debug("single fetch",
			zap.Any("client", client),
		)
		go fetcher(ctx, logger, client, request, resCh)
	}

	answeredServers := make(map[string]struct{})
	responseCount := 0
GATHER:
	for responseCount < len(clients) {
		select {
		case res := <-resCh:
			answeredServers[res.GetServer()] = struct{}{}
			if err := result.MergeI(res); err == nil {
				responseCount++
			} else {
				result.AddError(err)
			}
		case <-ctx.Done():
			err := ErrTimeoutExceeded.WithValue("timedout_backends", NoAnswerBackends(clients, answeredServers))
			result.AddError(err)

			break GATHER
		}
	}
	return result, responseCount
}

type ServerTagResponse struct {
	Server   string
	Response []string
	Err      []merry.Error
}

func NewServerTagResponse() *ServerTagResponse {
	return &ServerTagResponse{
		Response: []string{},
	}
}

func (s *ServerTagResponse) Self() interface{} {
	return s
}

func (s ServerTagResponse) GetServer() string {
	return s.Server
}

func (first *ServerTagResponse) MergeI(second ServerFetcherResponse) merry.Error {
	secondSelf := second.Self()
	s, ok := secondSelf.(*ServerTagResponse)
	if !ok {
		return ErrResponseTypeMismatch.Here().WithMessagef("got '%T', expected '%T'", secondSelf, first)
	}
	return first.Merge(s)
}

func (s *ServerTagResponse) AddError(err merry.Error) {
	if err == nil {
		return
	}
	if s.Err == nil {
		s.Err = []merry.Error{err}
	} else {
		s.Err = append(s.Err, err)
	}
}

func (first *ServerTagResponse) Errors() []merry.Error {
	return first.Err
}

func (first *ServerTagResponse) Merge(second *ServerTagResponse) merry.Error {
	if first.Err == nil {
		if second.Err != nil {
			first.Err = second.Err
		}
	} else {
		if second.Err != nil {
			first.Err = append(first.Err, second.Err...)
		}
	}

	if second.Response == nil {
		return nil
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
	Err      []merry.Error
}

func NewServerInfoResponse() *ServerInfoResponse {
	return &ServerInfoResponse{
		Response: &protov3.ZipperInfoResponse{Info: make(map[string]protov3.MultiMetricsInfoResponse)},
		Stats:    new(Stats),
	}
}

func (s *ServerInfoResponse) Self() interface{} {
	return s
}

func (s ServerInfoResponse) GetServer() string {
	return s.Server
}

func (first *ServerInfoResponse) MergeI(second ServerFetcherResponse) merry.Error {
	secondSelf := second.Self()
	s, ok := secondSelf.(*ServerInfoResponse)
	if !ok {
		return ErrResponseTypeMismatch.Here().WithMessagef("got '%T', expected '%T'", secondSelf, first)
	}
	return first.Merge(s)
}

func (s *ServerInfoResponse) AddError(err merry.Error) {
	if err == nil {
		return
	}
	if s.Err == nil {
		s.Err = []merry.Error{err}
	} else {
		s.Err = append(s.Err, err)
	}
}

func (first *ServerInfoResponse) Errors() []merry.Error {
	return first.Err
}

func (first *ServerInfoResponse) Merge(second *ServerInfoResponse) merry.Error {
	if second.Stats != nil {
		first.Stats.Merge(second.Stats)
	}

	if first.Err == nil {
		if second.Err != nil {
			first.Err = second.Err
		}
	} else {
		if second.Err != nil {
			first.Err = append(first.Err, second.Err...)
		}
	}

	if second.Response == nil {
		return nil
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
	Err      []merry.Error
}

func NewServerFindResponse() *ServerFindResponse {
	return &ServerFindResponse{
		Response: new(protov3.MultiGlobResponse),
		Stats:    new(Stats),
	}
}

func (s *ServerFindResponse) Self() interface{} {
	return s
}

func (s ServerFindResponse) GetServer() string {
	return s.Server
}

func (first *ServerFindResponse) MergeI(second ServerFetcherResponse) merry.Error {
	secondSelf := second.Self()
	s, ok := secondSelf.(*ServerFindResponse)
	if !ok {
		return ErrResponseTypeMismatch.Here().WithMessagef("got '%T', expected '%T'", secondSelf, first)
	}
	return first.Merge(s)
}

func (s *ServerFindResponse) AddError(err merry.Error) {
	if err == nil {
		return
	}
	if s.Err == nil {
		s.Err = []merry.Error{err}
	} else {
		s.Err = append(s.Err, err)
	}
}

func (first *ServerFindResponse) Errors() []merry.Error {
	return first.Err
}

func (first *ServerFindResponse) Merge(second *ServerFindResponse) merry.Error {
	if second.Stats != nil {
		first.Stats.Merge(second.Stats)
	}

	if first.Err == nil {
		if second.Err != nil {
			first.Err = second.Err
		}
	} else {
		if second.Err != nil {
			first.Err = append(first.Err, second.Err...)
		}
	}

	if second.Response == nil {
		return nil
	}

	var ok bool
	seenMetrics := make(map[string]int)
	seenMatches := make(map[string]map[bool]struct{})
	for i, m := range first.Response.Metrics {
		seenMetrics[m.Name] = i
		for _, mm := range m.Matches {
			lkey := m.Name + "." + mm.Path
			if _, ok = seenMatches[lkey]; !ok {
				seenMatches[lkey] = map[bool]struct{}{}
			}

			seenMatches[lkey][mm.IsLeaf] = struct{}{}
		}
	}

	var i int
	for _, m := range second.Response.Metrics {
		if i, ok = seenMetrics[m.Name]; !ok {
			first.Response.Metrics = append(first.Response.Metrics, m)
			continue
		}

		for _, mm := range m.Matches {
			key := first.Response.Metrics[i].Name + "." + mm.Path
			lisLeaf := seenMatches[key]
			if lisLeaf == nil {
				lisLeaf = map[bool]struct{}{}
				seenMatches[key] = lisLeaf
			}

			if _, ok = lisLeaf[mm.IsLeaf]; !ok {
				lisLeaf[mm.IsLeaf] = struct{}{}
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
	Err      []merry.Error
}

func NewServerFetchResponse() *ServerFetchResponse {
	return &ServerFetchResponse{
		Response: new(protov3.MultiFetchResponse),
		Stats:    new(Stats),
	}
}

func (s *ServerFetchResponse) Self() interface{} {
	return s
}

func (s ServerFetchResponse) GetServer() string {
	return s.Server
}

func (first *ServerFetchResponse) Merge(second *ServerFetchResponse) merry.Error {
	if second.Stats != nil {
		first.Stats.Merge(second.Stats)
	}

	if first.Err == nil {
		if second.Err != nil {
			first.Err = second.Err
		}
	} else {
		if second.Err != nil {
			first.Err = append(first.Err, second.Err...)
		}
	}

	if second.Response == nil {
		return nil
	}

	metrics := make(map[fetchResponseCoordinates]int)
	for i := range first.Response.Metrics {
		metrics[coordinates(&first.Response.Metrics[i])] = i
	}

	for i := range second.Response.Metrics {
		if j, ok := metrics[coordinates(&second.Response.Metrics[i])]; ok {
			err := MergeFetchResponses(&first.Response.Metrics[j], &second.Response.Metrics[i])
			if err != nil {
				// TODO: Normal merry.Error handling
				continue
			}
		} else {
			first.Response.Metrics = append(first.Response.Metrics, second.Response.Metrics[i])
		}
	}
	return nil
}

func (first *ServerFetchResponse) MergeI(second ServerFetcherResponse) merry.Error {
	secondSelf := second.Self()
	s, ok := secondSelf.(*ServerFetchResponse)
	if !ok {
		return ErrResponseTypeMismatch.Here().WithMessagef("got '%T', expected '%T'", secondSelf, first)
	}
	return first.Merge(s)
}

func (s *ServerFetchResponse) AddError(err merry.Error) {
	if err == nil {
		return
	}
	if s.Err == nil {
		s.Err = []merry.Error{err}
	} else {
		s.Err = append(s.Err, err)
	}
}

func (first *ServerFetchResponse) Errors() []merry.Error {
	return first.Err
}

func (s *ServerFetchResponse) NonFatalError(err merry.Error) *ServerFetchResponse {
	s.AddError(err)
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

func mergeFetchResponsesWithEqualStepTimes(m1, m2 *protov3.FetchResponse) merry.Error {
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

func mergeFetchResponsesWithUnequalStepTimes(m1, m2 *protov3.FetchResponse) merry.Error {
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

func MergeFetchResponses(m1, m2 *protov3.FetchResponse) merry.Error {
	var err merry.Error
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

	return err
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
