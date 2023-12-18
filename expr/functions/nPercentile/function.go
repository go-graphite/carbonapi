package nPercentile

import (
	"context"
	"math"
	"strconv"

	"github.com/go-graphite/carbonapi/expr/consolidations"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type nPercentile struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &nPercentile{}
	functions := []string{"nPercentile"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// nPercentile(seriesList, n)
func (f *nPercentile) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	if e.ArgsLen() < 2 {
		return nil, parser.ErrMissingArgument
	}

	arg, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}
	percent, err := e.GetFloatArg(1)
	if err != nil {
		return nil, err
	}
	percentStr := strconv.FormatFloat(percent, 'g', -1, 64)

	results := make([]*types.MetricData, len(arg))
	for i, a := range arg {
		r := a.CopyLink()
		r.Name = "nPercentile(" + a.Name + "," + percentStr + ")"
		r.Values = make([]float64, len(a.Values))
		r.Tags["nPercentile"] = percentStr
		var values []float64
		for _, v := range a.Values {
			if !math.IsNaN(v) {
				values = append(values, v)
			}
		}

		value := consolidations.Percentile(values, percent, true)
		for i := range r.Values {
			r.Values[i] = value
		}

		results[i] = r
	}
	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *nPercentile) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"nPercentile": {
			Description: "Returns n-percent of each series in the seriesList.",
			Function:    "nPercentile(seriesList, n)",
			Group:       "Calculate",
			Module:      "graphite.render.functions",
			Name:        "nPercentile",
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
