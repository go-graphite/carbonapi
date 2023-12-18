package removeBelowSeries

import (
	"context"
	"math"
	"strconv"
	"strings"

	"github.com/go-graphite/carbonapi/expr/consolidations"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type removeBelowSeries struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &removeBelowSeries{}
	functions := []string{"removeBelowValue", "removeAboveValue", "removeBelowPercentile", "removeAbovePercentile"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// removeBelowValue(seriesLists, n), removeAboveValue(seriesLists, n), removeBelowPercentile(seriesLists, percent), removeAbovePercentile(seriesLists, percent)
func (f *removeBelowSeries) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}

	number, err := e.GetFloatArg(1)
	if err != nil {
		return nil, err
	}
	numberStr := e.Arg(1).StringValue()

	condition := func(v float64, threshold float64) bool {
		return v < threshold
	}

	if strings.HasPrefix(e.Target(), "removeAbove") {
		condition = func(v float64, threshold float64) bool {
			return v > threshold
		}
	}

	results := make([]*types.MetricData, len(args))

	for n, a := range args {
		threshold := number
		if strings.HasSuffix(e.Target(), "Percentile") {
			var values []float64
			for _, v := range a.Values {
				if !math.IsNaN(v) {
					values = append(values, v)
				}
			}

			threshold = consolidations.Percentile(values, number, true)
		}

		r := a.CopyLink()
		r.Name = e.Target() + "(" + a.Name + ", " + numberStr + ")"
		r.Values = make([]float64, len(a.Values))
		r.Tags["removeBelowSeries"] = strconv.FormatFloat(threshold, 'f', -1, 64)

		for i, v := range a.Values {
			if math.IsNaN(v) || condition(v, threshold) {
				r.Values[i] = math.NaN()
			} else {
				r.Values[i] = v
			}
		}

		results[n] = r
	}

	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *removeBelowSeries) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"removeBelowValue": {
			Description: "Removes data below the given threshold from the series or list of series provided.\nValues below this threshold are assigned a value of None.",
			Function:    "removeBelowValue(seriesList, n)",
			Group:       "Filter Data",
			Module:      "graphite.render.functions",
			Name:        "removeBelowValue",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "n",
					Required: true,
					Type:     types.Integer,
				},
			},
			NameChange:   true, // name changed
			ValuesChange: true, // values changed
		},
		"removeAboveValue": {
			Description: "Removes data above the given threshold from the series or list of series provided.\nValues above this threshold are assigned a value of None.",
			Function:    "removeAboveValue(seriesList, n)",
			Group:       "Filter Data",
			Module:      "graphite.render.functions",
			Name:        "removeAboveValue",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "n",
					Required: true,
					Type:     types.Integer,
				},
			},
			NameChange:   true, // name changed
			ValuesChange: true, // values changed
		},
		"removeBelowPercentile": {
			Description: "Removes data below the nth percentile from the series or list of series provided.\nValues below this percentile are assigned a value of None.",
			Function:    "removeBelowPercentile(seriesList, n)",
			Group:       "Filter Data",
			Module:      "graphite.render.functions",
			Name:        "removeBelowPercentile",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "n",
					Required: true,
					Type:     types.Integer,
				},
			},
			NameChange:   true, // name changed
			ValuesChange: true, // values changed
		},
		"removeAbovePercentile": {
			Description: "Removes data above the nth percentile from the series or list of series provided.\nValues above this percentile are assigned a value of None.",
			Function:    "removeAbovePercentile(seriesList, n)",
			Group:       "Filter Data",
			Module:      "graphite.render.functions",
			Name:        "removeAbovePercentile",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "n",
					Required: true,
					Type:     types.Integer,
				},
			},
			NameChange:   true, // name changed
			ValuesChange: true, // values changed
		},
	}
}
