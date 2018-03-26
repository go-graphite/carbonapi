package lowPass

import (
	"fmt"
	"math"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type lowPass struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &lowPass{}
	functions := []string{"lowPass", "lpf"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// lowPass(seriesList, cutPercent)
func (f *lowPass) Do(e parser.Expr, from, until uint32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	cutPercent, err := e.GetFloatArg(1)
	if err != nil {
		return nil, err
	}

	var results []*types.MetricData
	for _, a := range arg {
		name := fmt.Sprintf("lowPass(%s,%v)", a.Name, cutPercent)
		r := *a
		r.Name = name
		r.Values = make([]float64, len(a.Values))
		lowCut := int((cutPercent / 200) * float64(len(a.Values)))
		highCut := len(a.Values) - lowCut
		for i, v := range a.Values {
			if i < lowCut || i >= highCut {
				r.Values[i] = v
			} else {
				r.Values[i] = math.NaN()
			}
		}

		results = append(results, &r)
	}
	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *lowPass) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"lpf": {
			Description: "Low-pass filters provide a smoother form of a signal, removing the short-term fluctuations, and leaving the longer-term trend. https://en.wikipedia.org/wiki/Low-pass_filter",
			Function:    "lpf(seriesList, cutPercent)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "lpf",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "cutPercent",
					Required: true,
					Type:     types.Float,
				},
			},
		},
		"lowPass": {
			Description: "Low-pass filters provide a smoother form of a signal, removing the short-term fluctuations, and leaving the longer-term trend. https://en.wikipedia.org/wiki/Low-pass_filter",
			Function:    "lowPass(seriesList, cutPercent)",
			Group:       "Transform",
			Module:      "graphite.render.functions.custom",
			Name:        "lowPass",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "cutPercent",
					Required: true,
					Type:     types.Float,
				},
			},
		},
	}
}
