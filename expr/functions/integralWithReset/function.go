package integralWithReset

import (
	"context"
	"math"

	"github.com/ansel1/merry"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type integralWithReset struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &integralWithReset{}
	functions := []string{"integralWithReset"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// integralWithReset(seriesList, resettingSeries)
func (f *integralWithReset) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}
	resettingSeriesList, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(1), from, until, values)
	if err != nil {
		return nil, err
	}
	if len(resettingSeriesList) != 1 {
		return nil, types.ErrWildcardNotAllowed
	}
	resettingSeries := resettingSeriesList[0]

	for _, a := range arg {
		if a.StepTime != resettingSeries.StepTime || len(a.Values) != len(resettingSeries.Values) {
			return nil, merry.Errorf("series %s must have the same length as %s", a.Name, resettingSeries.Name)
		}
	}

	results := make([]*types.MetricData, len(arg))
	for i, a := range arg {
		r := a.CopyLinkTags()
		r.Name = "integralWithReset(" + a.Name + "," + resettingSeries.Name + ")"
		r.Values = make([]float64, len(a.Values))

		current := 0.0
		for i, v := range a.Values {
			if math.IsNaN(v) {
				r.Values[i] = v
				continue
			}
			if !math.IsNaN(resettingSeries.Values[i]) && resettingSeries.Values[i] != 0 {
				current = 0
			} else {
				current += v
			}
			r.Values[i] = current
		}
		results[i] = r
	}
	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *integralWithReset) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"integralWithReset": {
			Description: "Just like integral(seriesList) but with resets: every time resettingSeries is not 0, the integral resets to 0.",
			Function:    "integralWithReset(seriesList, resettingSeries)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "integralWithReset",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "resettingSeries",
					Required: true,
					Type:     types.SeriesList,
				},
			},
			NameChange:   true, // name changed
			ValuesChange: true, // values changed
		},
	}
}
