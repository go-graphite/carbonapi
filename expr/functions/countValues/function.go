package countValues

import (
	"context"
	"math"
	"strconv"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type countValues struct {
	interfaces.Function
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(_ string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &countValues{}
	functions := []string{"countValues"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

const (
	defaultValuesLimit      = 32
	LimitExceededMetricName = "error.too.many.values.limit.reached"
)

// countValues(seriesList)
func (f *countValues) Do(ctx context.Context, eval interfaces.Evaluator, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	// TODO(civil): Check that series have equal length
	args, err := helper.GetSeriesArg(ctx, eval, e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	valuesLimit, err := e.GetIntNamedOrPosArgDefault("valuesLimit", 1, defaultValuesLimit)
	if err != nil {
		return nil, err
	}

	var results []*types.MetricData

	data := map[int][]float64{}

	for _, arg := range args {
		for bucket, value := range arg.Values {

			if math.IsNaN(value) {
				continue
			}

			key := int(value)
			v, ok := data[key]
			if !ok {
				if len(data) >= valuesLimit {
					m := *args[0]
					m.Name = LimitExceededMetricName
					m.Values = make([]float64, len(m.Values))
					return []*types.MetricData{&m}, nil
				}

				v = make([]float64, len(arg.Values))
				data[key] = v
			}
			v[bucket]++

		}
	}

	for key, value := range data {
		mName := strconv.FormatInt(int64(key), 10)
		m := *args[0]
		m.Name = mName
		m.Values = value
		results = append(results, &m)
	}

	return results, nil
}

const functionDescription = `Draws line for each unique value in the seriesList. Each line displays count of the value in current bucket.

.. code-block:: none

&target=countValues(carbon.agents.*.*)`

func (f *countValues) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"countValues": {
			Description: functionDescription,
			Function:    "countValues(*seriesLists)",
			Group:       "Combine",
			Module:      "graphite.render.functions",
			Name:        "countValues",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Default: types.NewSuggestion(defaultValuesLimit),
					Name:    "valuesLimit",
					Type:    types.Integer,
				},
			},
		},
	}
}
