package interpolate

import (
	"context"
	"math"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type interpolate struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	return []interfaces.FunctionMetadata{{
		F:    &interpolate{},
		Name: "interpolate",
	}}
}

func (f *interpolate) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) (resultData []*types.MetricData, resultError error) {
	seriesList, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}

	limit, err := e.GetIntOrInfArgDefault(1, parser.IntOrInf{IsInf: true})
	if err != nil {
		return nil, err
	}

	resultSeriesList := make([]*types.MetricData, 0, len(seriesList))
	for _, series := range seriesList {
		pointsQty := len(series.Values)
		resultSeries := series.CopyLinkTags()
		resultSeries.Name = "interpolate(" + series.Name + ")"

		resultSeries.Values = make([]float64, pointsQty)
		copy(resultSeries.Values, series.Values)

		consecutiveNulls := 0
		for i := 0; i < pointsQty; i++ {
			if i == 0 {
				// no "keeping" can be done on the first value
				// because we have no idea what came before it
				continue
			}

			value := resultSeries.Values[i]

			if math.IsNaN(value) {
				consecutiveNulls += 1
			} else if consecutiveNulls == 0 {
				// have a value but no need to interpolate
				continue
			} else if math.IsNaN(resultSeries.Values[i-consecutiveNulls-1]) {
				// # have a value but can't interpolate: reset counter
				consecutiveNulls = 0
				continue
			} else {
				// have a value and can interpolate
				// if a non-null value is seen before the limit is hit
				// backfill all the missing datapoints with the last known value
				if consecutiveNulls > 0 && (limit.IsInf || consecutiveNulls <= limit.IntVal) {
					lastNotNullIndex := i - consecutiveNulls - 1
					lastNotNullValue := resultSeries.Values[lastNotNullIndex]

					for j := 0; j < consecutiveNulls; j++ {
						coefficient := float64(j+1) / float64(consecutiveNulls+1)
						index := i - consecutiveNulls + j

						resultSeries.Values[index] = lastNotNullValue + coefficient*(value-lastNotNullValue)
					}
				}

				// reset counter
				consecutiveNulls = 0
			}
		}

		resultSeriesList = append(resultSeriesList, resultSeries)
	}

	return resultSeriesList, nil
}

func (f *interpolate) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"interpolate": {
			Description: "Takes one metric or a wildcard seriesList, and optionally a limit to the number of 'None' values to skip over." +
				"\nContinues the line with the last received value when gaps ('None' values) appear in your data, rather than breaking your line." +
				"\n\n.. code-block:: none\n\n  &target=interpolate(Server01.connections.handled)\n  &target=interpolate(Server01.connections.handled, 10)",
			Function: "interpolate(seriesList, limit)",
			Group:    "Transform",
			Module:   "graphite.render.functions",
			Name:     "interpolate",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "limit",
					Required: false,
					Type:     types.IntOrInf,
					Default:  types.NewSuggestion(math.Inf(1)),
				},
			},
			NameChange:   true, // name changed
			ValuesChange: true, // values changed
		},
	}
}
