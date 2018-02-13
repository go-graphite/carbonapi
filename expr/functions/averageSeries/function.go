package averageSeries

import (
	"fmt"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"strings"
)

func init() {
	metadata.RegisterFunction("avg", &AverageSeries{})
	metadata.RegisterFunction("averageSeries", &AverageSeries{})
	metadata.RegisterFunction("averageSeriesWithWildcards", &AverageSeriesWithWildcards{})
}

type AverageSeries struct {
	interfaces.FunctionBase
}

// averageSeries(*seriesLists)
func (f *AverageSeries) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArgsAndRemoveNonExisting(e, from, until, values)
	if err != nil {
		return nil, err
	}

	e.SetTarget("averageSeries")
	return helper.AggregateSeries(e, args, func(values []float64) float64 {
		sum := 0.0
		for _, value := range values {
			sum += value
		}
		return sum / float64(len(values))
	})
}

type AverageSeriesWithWildcards struct {
	interfaces.FunctionBase
}

// averageSeriesWithWildcards(seriesLIst, *position)
func (f *AverageSeriesWithWildcards) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	/* TODO(dgryski): make sure the arrays are all the same 'size'
	   (duplicated from sumSeriesWithWildcards because of similar logic but aggregation) */
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
		r.Name = fmt.Sprintf("averageSeriesWithWildcards(%s)", series)
		r.Values = make([]float64, len(args[0].Values))
		r.IsAbsent = make([]bool, len(args[0].Values))

		length := make([]float64, len(args[0].Values))
		atLeastOne := make([]bool, len(args[0].Values))
		for _, arg := range args {
			for i, v := range arg.Values {
				if arg.IsAbsent[i] {
					continue
				}
				atLeastOne[i] = true
				length[i]++
				r.Values[i] += v
			}
		}

		for i, v := range atLeastOne {
			if v {
				r.Values[i] = r.Values[i] / length[i]
			} else {
				r.IsAbsent[i] = true
			}
		}

		results = append(results, &r)
	}
	return results, nil
}
