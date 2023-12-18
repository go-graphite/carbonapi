package powSeries

import (
	"context"
	"math"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type powSeries struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(_ string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)

	f := &powSeries{}
	functions := []string{"powSeries"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}

	return res
}

func (f *powSeries) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	series, err := helper.GetSeriesArgs(ctx, f.GetEvaluator(), e.Args(), from, until, values)
	if err != nil {
		return nil, err
	}

	overallLength := -1
	largestSeriesIdx := -1
	for i, s := range series {
		l := len(s.GetValues())
		if l > overallLength {
			overallLength = l
			largestSeriesIdx = i
		}
	}

	r := series[largestSeriesIdx].CopyName("powSeries(" + e.RawArgs() + ")")
	r.Values = make([]float64, 0)

	seriesValues := make([][]float64, 0, len(series))
	for _, s := range series {
		seriesValues = append(seriesValues, s.GetValues())
	}

	for i := 0; i < overallLength; i++ {
		first := true
		var result float64

		for _, vals := range seriesValues {
			var val float64
			if i < len(vals) {
				val = vals[i]
			} else {
				val = math.NaN()
			}

			if first {
				result = val
				first = false
			} else {
				result = math.Pow(result, val)
			}
		}

		if math.IsInf(result, 0) {
			result = math.NaN()
		}
		r.Values = append(r.Values, result)
	}

	return []*types.MetricData{r}, nil
}

func (f *powSeries) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"powSeries": {
			Description: "Takes two or more series and pows their points. A constant line may be used.",
			Function:    "powSeries(*seriesLists)",
			Group:       "Combine",
			Module:      "graphite.render.functions",
			Name:        "powSeries",
			Params: []types.FunctionParam{
				{
					Multiple: true,
					Name:     "seriesLists",
					Required: true,
					Type:     types.SeriesList,
				},
			},
		},
	}
}
