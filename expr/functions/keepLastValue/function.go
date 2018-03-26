package keepLastValue

import (
	"fmt"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"math"
)

type keepLastValue struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &keepLastValue{}
	functions := []string{"keepLastValue"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// keepLastValue(seriesList, limit=inf)
func (f *keepLastValue) Do(e parser.Expr, from, until uint32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	keep, err := e.GetIntNamedOrPosArgDefault("limit", 1, -1)
	if err != nil {
		return nil, err
	}
	_, ok := e.NamedArgs()["limit"]
	if !ok {
		ok = len(e.Args()) > 1
	}

	var results []*types.MetricData

	for _, a := range arg {
		var name string
		if ok {
			name = fmt.Sprintf("keepLastValue(%s,%d)", a.Name, keep)
		} else {
			name = fmt.Sprintf("keepLastValue(%s)", a.Name)
		}

		r := *a
		r.Name = name
		r.Values = make([]float64, len(a.Values))

		prev := math.NaN()
		missing := 0

		for i, v := range a.Values {
			if math.IsNaN(v) {

				if (keep < 0 || missing < keep) && !math.IsNaN(prev) {
					r.Values[i] = prev
					missing++
				} else {
					r.Values[i] = math.NaN()
				}

				continue
			}
			missing = 0
			prev = v
			r.Values[i] = v
		}
		results = append(results, &r)
	}
	return results, err
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *keepLastValue) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"keepLastValue": {
			Description: "Takes one metric or a wildcard seriesList, and optionally a limit to the number of 'None' values to skip over.\nContinues the line with the last received value when gaps ('None' values) appear in your data, rather than breaking your line.\n\nExample:\n\n.. code-block:: none\n\n  &target=keepLastValue(Server01.connections.handled)\n  &target=keepLastValue(Server01.connections.handled, 10)",
			Function:    "keepLastValue(seriesList, limit=inf)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "keepLastValue",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Default: types.NewSuggestion("INF"),
					Name:    "limit",
					Type:    types.Integer,
				},
			},
		},
	}
}
