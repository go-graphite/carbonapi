package scaleToSeconds

import (
	"context"
	"strconv"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type scaleToSeconds struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &scaleToSeconds{}
	functions := []string{"scaleToSeconds"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// scaleToSeconds(seriesList, seconds)
func (f *scaleToSeconds) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	if e.ArgsLen() < 2 {
		return nil, parser.ErrMissingArgument
	}

	arg, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}
	seconds, err := e.GetFloatArg(1)
	if err != nil {
		return nil, err
	}
	secondsStr := strconv.Itoa(int(seconds))

	results := make([]*types.MetricData, len(arg))

	for j, a := range arg {
		r := a.CopyLink()
		r.Name = "scaleToSeconds(" + a.Name + "," + secondsStr + ")"
		r.Values = make([]float64, len(a.Values))
		r.Tags["scaleToSeconds"] = secondsStr

		factor := seconds / float64(a.StepTime)

		for i, v := range a.Values {
			r.Values[i] = v * factor
		}

		results[j] = r
	}
	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *scaleToSeconds) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"scaleToSeconds": {
			Description: "Takes one metric or a wildcard seriesList and returns \"value per seconds\" where\nseconds is a last argument to this functions.\n\nUseful in conjunction with derivative or integral function if you want\nto normalize its result to a known resolution for arbitrary retentions",
			Function:    "scaleToSeconds(seriesList, seconds)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "scaleToSeconds",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "seconds",
					Required: true,
					Type:     types.Integer,
				},
			},
			NameChange:   true, // name changed
			ValuesChange: true, // values changed
		},
	}
}
