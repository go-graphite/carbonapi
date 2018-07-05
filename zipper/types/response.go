package types

import (
	"math"

	"github.com/go-graphite/carbonapi/zipper/errors"
	protov3 "github.com/go-graphite/protocol/carbonapi_v3_pb"
	"github.com/lomik/zapwriter"
	"go.uber.org/zap"
)

type ServerResponse struct {
	Server   string
	Response []byte
}

type ServerInfoResponse struct {
	Server   string
	Response *protov3.ZipperInfoResponse
	Stats    *Stats
	Err      *errors.Errors
}

type ServerFindResponse struct {
	Server   string
	Response *protov3.MultiGlobResponse
	Stats    *Stats
	Err      *errors.Errors
}

/*
func mergeFindRequests(f1, f2 []protov3.GlobMatch) []protov3.GlobMatch {
	uniqList := make(map[string]protov3.GlobMatch)

	for _, v := range f1 {
		uniqList[v.Path] = v
	}
	for _, v := range f2 {
		uniqList[v.Path] = v
	}

	res := make([]protov3.GlobMatch, 0, len(uniqList))
	for _, v := range uniqList {
		res = append(res, v)
	}

	return res
}
*/

func (first *ServerFindResponse) Merge(second *ServerFindResponse) *errors.Errors {
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

	if first.Err != nil && second.Err == nil {
		first.Err = nil
	}

	return nil
}

type ServerFetchResponse struct {
	Server       string
	ResponsesMap map[string][]protov3.FetchResponse
	Response     *protov3.MultiFetchResponse
	Stats        *Stats
	Err          *errors.Errors
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

func MergeFetchResponses(m1, m2 *protov3.FetchResponse) *errors.Errors {
	logger := zapwriter.Logger("zipper_render")

	if len(m1.Values) != len(m2.Values) {
		interpolate := false
		if len(m1.Values) < len(m2.Values) {
			swapFetchResponses(m1, m2)
		}
		if m1.StepTime < m2.StepTime {
			interpolate = true
		} else {
			if m1.StartTime == m2.StartTime {
				for i := 0; i < len(m1.Values)-len(m2.Values); i++ {
					m2.Values = append(m2.Values, math.NaN())
				}

				goto out
			}
		}

		// TODO(Civil): we must fix the case of m1.StopTime != m2.StopTime
		// We should check if m1.StopTime and m2.StopTime actually the same
		// Also we need to append nans in case StopTimes dramatically differs

		if !interpolate || m1.StopTime-m1.StopTime%m2.StepTime != m2.StopTime {
			// m1.Step < m2.Step and len(m1) < len(m2) - most probably garbage data
			logger.Error("unable to merge ovalues",
				zap.Int("metric_values", len(m2.Values)),
				zap.Int("response_values", len(m1.Values)),
			)

			return errors.FromErr(ErrResponseLengthMismatch)
		}

		// len(m1) > len(m2)
		values := make([]float64, 0, len(m1.Values))
		for ts := m1.StartTime; ts < m1.StopTime; ts += m1.StepTime {
			idx := (ts - m1.StartTime) / m2.StepTime
			values = append(values, m2.Values[idx])
		}
		m2.Values = values
		m2.StepTime = m1.StepTime
		m2.StartTime = m1.StartTime
		m2.StopTime = m1.StopTime
	}
out:

	if m1.StartTime != m2.StartTime {
		return errors.FromErr(ErrResponseStartTimeMismatch)
	}

	for i := range m1.Values {
		if !math.IsNaN(m1.Values[i]) {
			continue
		}

		// found one
		if !math.IsNaN(m2.Values[i]) {
			m1.Values[i] = m2.Values[i]
		}
	}
	return nil
}

func (first *ServerFetchResponse) Merge(second *ServerFetchResponse) {
	if first.Server == "" {
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
		return
	}

	if second.Response == nil {
		return
	}

	metrics := make(map[string]int)
	for i := range first.Response.Metrics {
		metrics[first.Response.Metrics[i].Name] = i
	}

	for i := range second.Response.Metrics {
		if j, ok := metrics[second.Response.Metrics[i].Name]; ok {
			err := MergeFetchResponses(&first.Response.Metrics[j], &second.Response.Metrics[i])
			if err != nil {
				// TODO: Normal error handling
				continue
			}
		} else {
			first.Response.Metrics = append(first.Response.Metrics, second.Response.Metrics[i])
		}
	}

	if first.Err != nil && second.Err == nil {
		first.Err = nil
	}

	return
}
