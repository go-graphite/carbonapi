package logarithm

import (
	"context"
	"math"
	"strconv"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type logarithm struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &logarithm{}
	functions := []string{"logarithm", "log"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// logarithm(seriesList, base=10)
// Alias: log
func (f *logarithm) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}
	base, err := e.GetIntNamedOrPosArgDefault("base", 1, 10)
	if err != nil {
		return nil, err
	}

	baseStr := strconv.Itoa(base)

	baseLog := math.Log(float64(base))

	results := make([]*types.MetricData, len(arg))

	for j, a := range arg {

		var name string
		if base != 10 {
			name = "logarithm(" + a.Name + "," + baseStr + ")"
		} else {
			name = "logarithm(" + a.Name + ")"
		}

		r := a.CopyLink()
		r.Name = name
		r.Values = make([]float64, len(a.Values))
		r.Tags["log"] = baseStr

		for i, v := range a.Values {
			r.Values[i] = math.Log(v) / baseLog
		}
		results[j] = r
	}
	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *logarithm) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"logarithm": {
			Description: "Takes one metric or a wildcard seriesList, a base, and draws the y-axis in logarithmic\nformat.  If base is omitted, the function defaults to base 10.\n\nExample:\n\n.. code-block:: none\n\n  &target=log(carbon.agents.hostname.avgUpdateTime,2)",
			Function:    "log(seriesList, base=10)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "log",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Default: types.NewSuggestion(10),
					Name:    "base",
					Type:    types.Integer,
				},
			},
		},
		"log": {
			Description: "Takes one metric or a wildcard seriesList, a base, and draws the y-axis in logarithmic\nformat.  If base is omitted, the function defaults to base 10.\n\nExample:\n\n.. code-block:: none\n\n  &target=log(carbon.agents.hostname.avgUpdateTime,2)",
			Function:    "log(seriesList, base=10)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "log",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Default: types.NewSuggestion(10),
					Name:    "base",
					Type:    types.Integer,
				},
			},
			NameChange:   true, // name changed
			ValuesChange: true, // values changed
		},
	}
}
