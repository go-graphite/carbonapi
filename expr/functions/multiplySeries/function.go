package multiplySeries

import (
	"fmt"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	pb "github.com/go-graphite/carbonzipper/carbonzipperpb3"
	"math"
	"strings"
)

func init() {
	metadata.RegisterFunction("multiplySeries", &MultiplySeries{})
	metadata.RegisterFunction("multiplySeriesWithWildcards", &MultiplySeriesWithWildcards{})
}

type MultiplySeries struct {
	interfaces.FunctionBase
}

// multiplySeries(factorsSeriesList)
func (f *MultiplySeries) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	r := types.MetricData{
		FetchResponse: pb.FetchResponse{
			Name:      fmt.Sprintf("multiplySeries(%s)", e.RawArgs()),
			StartTime: from,
			StopTime:  until,
		},
	}
	for _, arg := range e.Args() {
		series, err := helper.GetSeriesArg(arg, from, until, values)
		if err != nil {
			return nil, err
		}

		if r.Values == nil {
			r.IsAbsent = make([]bool, len(series[0].IsAbsent))
			r.StepTime = series[0].StepTime
			r.Values = make([]float64, len(series[0].Values))
			copy(r.IsAbsent, series[0].IsAbsent)
			copy(r.Values, series[0].Values)
			series = series[1:]
		}

		for _, factor := range series {
			for i, v := range r.Values {
				if r.IsAbsent[i] || factor.IsAbsent[i] {
					r.IsAbsent[i] = true
					r.Values[i] = math.NaN()
					continue
				}

				r.Values[i] = v * factor.Values[i]
			}
		}
	}

	return []*types.MetricData{&r}, nil
}

type MultiplySeriesWithWildcards struct {
	interfaces.FunctionBase
}

// multiplySeriesWithWildcards(seriesList, *position)
func (f *MultiplySeriesWithWildcards) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	/* TODO(dgryski): make sure the arrays are all the same 'size'
	   (duplicated from sumSeriesWithWildcards because of similar logic but multiplication) */
	args, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	fields, err := e.GetIntArgs(1)
	if err != nil {
		return nil, err
	}

	var results []*types.MetricData

	nodeList := []string{}
	groups := make(map[string][]*types.MetricData)

	for _, a := range args {
		metric := helper.ExtractMetric(a.Name)
		nodes := strings.Split(metric, ".")
		var s []string
		// Yes, this is O(n^2), but len(nodes) < 10 and len(fields) < 3
		// Iterating an int slice is faster than a map for n ~ 30
		// http://www.antoine.im/posts/someone_is_wrong_on_the_internet
		for i, n := range nodes {
			if !helper.Contains(fields, i) {
				s = append(s, n)
			}
		}

		node := strings.Join(s, ".")

		if len(groups[node]) == 0 {
			nodeList = append(nodeList, node)
		}

		groups[node] = append(groups[node], a)
	}

	for _, series := range nodeList {
		args := groups[series]
		r := *args[0]
		r.Name = fmt.Sprintf("multiplySeriesWithWildcards(%s)", series)
		r.Values = make([]float64, len(args[0].Values))
		r.IsAbsent = make([]bool, len(args[0].Values))

		atLeastOne := make([]bool, len(args[0].Values))
		hasVal := make([]bool, len(args[0].Values))

		for _, arg := range args {
			for i, v := range arg.Values {
				if arg.IsAbsent[i] {
					continue
				}

				atLeastOne[i] = true
				if hasVal[i] == false {
					r.Values[i] = v
					hasVal[i] = true
				} else {
					r.Values[i] *= v
				}
			}
		}

		for i, v := range atLeastOne {
			if !v {
				r.IsAbsent[i] = true
			}
		}

		results = append(results, &r)
	}
	return results, nil
}
