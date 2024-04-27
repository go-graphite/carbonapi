package types

import (
	"context"
	"math"

	"github.com/ansel1/merry/v2"

	protov3 "github.com/go-graphite/protocol/carbonapi_v3_pb"
	"github.com/lomik/zapwriter"
	"go.uber.org/zap"
)

type Fetcher func(ctx context.Context, logger *zap.Logger, client BackendServer, reqs interface{}, resCh chan ServerFetcherResponse)

type ServerFetcherResponse interface {
	Self() interface{}
	MergeI(second ServerFetcherResponse) error
	AddError(err error)
	Errors() []error
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
			err := merry.Wrap(ErrTimeoutExceeded, merry.WithValue("no_answer_backends", NoAnswerBackends(clients, answeredServers)))
			result.AddError(err)

			break GATHER
		}
	}
	return result, responseCount
}

type ServerTagResponse struct {
	Server   string
	Response []string
	Err      []error
}

func NewServerTagResponse() *ServerTagResponse {
	return &ServerTagResponse{
		Response: []string{},
	}
}

func (s *ServerTagResponse) Self() interface{} {
	return s
}

func (s *ServerTagResponse) GetServer() string {
	return s.Server
}

func (s *ServerTagResponse) MergeI(second ServerFetcherResponse) error {
	secondSelf := second.Self()
	secondConverted, ok := secondSelf.(*ServerTagResponse)
	if !ok {
		return merry.Wrap(ErrResponseTypeMismatch, merry.WithMessagef("got '%T', expected '%T'", secondSelf, s))
	}
	return s.Merge(secondConverted)
}

func (s *ServerTagResponse) AddError(err error) {
	if err == nil {
		return
	}
	if s.Err == nil {
		s.Err = []error{err}
	} else {
		s.Err = append(s.Err, err)
	}
}

func (s *ServerTagResponse) Errors() []error {
	return s.Err
}

func (s *ServerTagResponse) Merge(second *ServerTagResponse) error {
	if s.Err == nil {
		if second.Err != nil {
			s.Err = second.Err
		}
	} else {
		if second.Err != nil {
			s.Err = append(s.Err, second.Err...)
		}
	}

	if second.Response == nil {
		return nil
	}

	// We cannot assume in general that results are sorted
	firstMap := make(map[string]struct{}, len(s.Response))
	for _, v := range s.Response {
		firstMap[v] = struct{}{}
	}

	for _, v := range second.Response {
		if _, ok := firstMap[v]; !ok {
			s.Response = append(s.Response, v)
		}
	}

	return nil
}

type ServerInfoResponse struct {
	Server   string
	Response *protov3.ZipperInfoResponse
	Stats    *Stats
	Err      []error
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

func (s *ServerInfoResponse) GetServer() string {
	return s.Server
}

func (s *ServerInfoResponse) MergeI(second ServerFetcherResponse) error {
	secondSelf := second.Self()
	secondConverted, ok := secondSelf.(*ServerInfoResponse)
	if !ok {
		return merry.Wrap(ErrResponseTypeMismatch, merry.WithMessagef("got '%T', expected '%T'", secondSelf, s))
	}
	return s.Merge(secondConverted)
}

func (s *ServerInfoResponse) AddError(err error) {
	if err == nil {
		return
	}
	if s.Err == nil {
		s.Err = []error{err}
	} else {
		s.Err = append(s.Err, err)
	}
}

func (s *ServerInfoResponse) Errors() []error {
	return s.Err
}

func (s *ServerInfoResponse) Merge(second *ServerInfoResponse) error {
	if second.Stats != nil {
		s.Stats.Merge(second.Stats)
	}

	if s.Err == nil {
		if second.Err != nil {
			s.Err = second.Err
		}
	} else {
		if second.Err != nil {
			s.Err = append(s.Err, second.Err...)
		}
	}

	if second.Response == nil {
		return nil
	}

	for k, v := range second.Response.Info {
		s.Response.Info[k] = v
	}

	return nil
}

type ServerFindResponse struct {
	Server   string
	Response *protov3.MultiGlobResponse
	Stats    *Stats
	Err      []error
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

func (s *ServerFindResponse) GetServer() string {
	return s.Server
}

func (s *ServerFindResponse) MergeI(second ServerFetcherResponse) error {
	secondSelf := second.Self()
	secondConverted, ok := secondSelf.(*ServerFindResponse)
	if !ok {
		return merry.Wrap(ErrResponseTypeMismatch, merry.WithMessagef("got '%T', expected '%T'", secondSelf, s))
	}
	return s.Merge(secondConverted)
}

func (s *ServerFindResponse) AddError(err error) {
	if err == nil {
		return
	}
	if s.Err == nil {
		s.Err = []error{err}
	} else {
		s.Err = append(s.Err, err)
	}
}

func (s *ServerFindResponse) Errors() []error {
	return s.Err
}

func (s *ServerFindResponse) Merge(second *ServerFindResponse) error {
	if second.Stats != nil {
		s.Stats.Merge(second.Stats)
	}

	if s.Err == nil {
		if second.Err != nil {
			s.Err = second.Err
		}
	} else {
		if second.Err != nil {
			s.Err = append(s.Err, second.Err...)
		}
	}

	if second.Response == nil {
		return nil
	}

	var ok bool
	seenMetrics := make(map[string]int)
	seenMatches := make(map[string]map[bool]struct{})
	for i, m := range s.Response.Metrics {
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
			s.Response.Metrics = append(s.Response.Metrics, m)
			continue
		}

		for _, mm := range m.Matches {
			key := s.Response.Metrics[i].Name + "." + mm.Path
			lisLeaf := seenMatches[key]
			if lisLeaf == nil {
				lisLeaf = map[bool]struct{}{}
				seenMatches[key] = lisLeaf
			}

			if _, ok = lisLeaf[mm.IsLeaf]; !ok {
				lisLeaf[mm.IsLeaf] = struct{}{}
				s.Response.Metrics[i].Matches = append(s.Response.Metrics[i].Matches, mm)
			}
		}
	}

	return nil
}

type ServerFetchResponse struct {
	Server   string
	Response *protov3.MultiFetchResponse
	Stats    *Stats
	Err      []error
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

func (s *ServerFetchResponse) GetServer() string {
	return s.Server
}

func (s *ServerFetchResponse) Merge(second *ServerFetchResponse) error {
	if second.Stats != nil {
		s.Stats.Merge(second.Stats)
	}

	if s.Err == nil {
		if second.Err != nil {
			s.Err = second.Err
		}
	} else {
		if second.Err != nil {
			s.Err = append(s.Err, second.Err...)
		}
	}

	if second.Response == nil {
		return nil
	}

	metrics := make(map[fetchResponseCoordinates]int)
	for i := range s.Response.Metrics {
		metrics[coordinates(&s.Response.Metrics[i])] = i
	}

	for i := range second.Response.Metrics {
		if j, ok := metrics[coordinates(&second.Response.Metrics[i])]; ok {
			err := MergeFetchResponses(&s.Response.Metrics[j], &second.Response.Metrics[i])
			if err != nil {
				// TODO: Normal error handling
				continue
			}
		} else {
			s.Response.Metrics = append(s.Response.Metrics, second.Response.Metrics[i])
		}
	}
	return nil
}

func (s *ServerFetchResponse) MergeI(second ServerFetcherResponse) error {
	secondSelf := second.Self()
	secondConverted, ok := secondSelf.(*ServerFetchResponse)
	if !ok {
		return merry.Wrap(ErrResponseTypeMismatch, merry.WithMessagef("got '%T', expected '%T'", secondSelf, s))
	}
	return s.Merge(secondConverted)
}

func (s *ServerFetchResponse) AddError(err error) {
	if err == nil {
		return
	}
	if s.Err == nil {
		s.Err = []error{err}
	} else {
		s.Err = append(s.Err, err)
	}
}

func (s *ServerFetchResponse) Errors() []error {
	return s.Err
}

func (s *ServerFetchResponse) NonFatalError(err error) *ServerFetchResponse {
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

func MergeFetchResponses(m1, m2 *protov3.FetchResponse) error {
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
