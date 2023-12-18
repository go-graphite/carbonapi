package pearson

import (
	"context"
	"math"

	"github.com/dgryski/go-onlinestats"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type pearson struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &pearson{}
	functions := []string{"pearson"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// pearson(series, series, windowSize)
func (f *pearson) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg1, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}

	arg2, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(1), from, until, values)
	if err != nil {
		return nil, err
	}

	if len(arg1) != 1 || len(arg2) != 1 {
		return nil, types.ErrWildcardNotAllowed
	}

	a1 := arg1[0]
	a2 := arg2[0]

	windowSize, err := e.GetIntArg(2)
	if err != nil {
		return nil, err
	}

	w1 := &types.Windowed{Data: make([]float64, windowSize)}
	w2 := &types.Windowed{Data: make([]float64, windowSize)}

	r := a1.CopyLinkTags()
	r.Name = "pearson(" + a1.Name + "," + a2.Name + "," + e.Arg(2).StringValue() + ")"
	r.Values = make([]float64, len(a1.Values))
	r.StartTime = from
	r.StopTime = r.StartTime + int64(len(r.Values))*r.StepTime

	for i, v1 := range a1.Values {
		v2 := a2.Values[i]

		w1.Push(v1)
		w2.Push(v2)
		if i >= windowSize-1 {
			r.Values[i] = onlinestats.Pearson(w1.Data, w2.Data)
		} else {
			r.Values[i] = math.NaN()
		}
	}

	return []*types.MetricData{r}, nil
}

func (f *pearson) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"pearson": {
			Description: `
Implementation of Pearson product-moment correlation coefficient (PMCC) function(s)

.. code-block:: none

	pearson( seriesA, seriesB, windowSize )


Calculate Pearson score between seriesA and seriesB over windowSize.

Note:
Pearson will discard epochs where either series has a missing value.

Additionally there is a special case where a series (or window) containing only zeros leads to a division-by-zero
and will manifest as if the entire window/series had missing values.`,
			Function: "pearson(seriesList, seriesList, windowSize)",
			Group:    "Transform",
			Module:   "graphite.render.functions.custom",
			Name:     "pearsonClosest",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "windowSize",
					Required: true,
					Type:     types.Integer,
				},
			},
			NameChange:   true, // name changed
			ValuesChange: true, // values changed
		},
	}
}
