package below

import (
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"strings"
)

func init() {
	functions := []string{"averageAbove", "averageBelow", "currentAbove", "currentBelow", "maximumAbove", "maximumBelow", "minimumAbove", "minimumBelow"}
	for _, f := range functions {
		metadata.RegisterFunction(f, &Below{})
	}
}

type Below struct {
	interfaces.FunctionBase
}

// averageAbove(seriesList, n), averageBelow(seriesList, n), currentAbove(seriesList, n), currentBelow(seriesList, n), maximumAbove(seriesList, n), maximumBelow(seriesList, n), minimumAbove(seriesList, n), minimumBelow
func (f *Below) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	n, err := e.GetFloatArg(1)
	if err != nil {
		return nil, err
	}

	isAbove := strings.HasSuffix(e.Target(), "Above")
	isInclusive := true
	var compute func([]float64, []bool) float64
	switch {
	case strings.HasPrefix(e.Target(), "average"):
		compute = helper.AvgValue
	case strings.HasPrefix(e.Target(), "current"):
		compute = helper.CurrentValue
	case strings.HasPrefix(e.Target(), "maximum"):
		compute = helper.MaxValue
		isInclusive = false
	case strings.HasPrefix(e.Target(), "minimum"):
		compute = helper.MinValue
		isInclusive = false
	}
	var results []*types.MetricData
	for _, a := range args {
		value := compute(a.Values, a.IsAbsent)
		if isAbove {
			if isInclusive {
				if value >= n {
					results = append(results, a)
				}
			} else {
				if value > n {
					results = append(results, a)
				}
			}
		} else {
			if value <= n {
				results = append(results, a)
			}
		}
	}

	return results, err
}
