package rangeOfSeries

import (
	"context"
	"math"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type rangeOfSeries struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &rangeOfSeries{}
	functions := []string{"rangeOfSeries"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// rangeOfSeries(*seriesLists)
func (f *rangeOfSeries) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	series, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}
	if len(series) == 0 {
		return []*types.MetricData{}, nil
	}

	r := series[0].CopyLinkTags()
	r.Name = e.Target() + "(" + e.RawArgs() + ")"
	r.Values = make([]float64, len(series[0].Values))

	commonTags := helper.GetCommonTags(series)

	if _, ok := commonTags["name"]; !ok {
		commonTags["name"] = r.Name
	}
	r.Tags = commonTags

	for i := range series[0].Values {
		var min, max float64
		count := 0
		for _, s := range series {
			if math.IsNaN(s.Values[i]) {
				continue
			}

			if count == 0 {
				min = s.Values[i]
				max = s.Values[i]
			} else {
				min = math.Min(min, s.Values[i])
				max = math.Max(max, s.Values[i])
			}

			count++
		}

		if count >= 2 {
			r.Values[i] = max - min
		} else {
			r.Values[i] = math.NaN()
		}
	}
	return []*types.MetricData{r}, err
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *rangeOfSeries) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"rangeOfSeries": {
			Description: "Takes a wildcard seriesList.\nDistills down a set of inputs into the range of the series\n\nExample:\n\n.. code-block:: none\n\n    &target=rangeOfSeries(Server*.connections.total)\n\nThis is an alias for :py:func:`aggregate <aggregate>` with aggregation ``rangeOf``.",
			Function:    "rangeOfSeries(*seriesLists)",
			Group:       "Combine",
			Module:      "graphite.render.functions",
			Name:        "rangeOfSeries",
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
