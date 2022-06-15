package aliasByBase64

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/grafana/carbonapi/expr/helper"
	"github.com/grafana/carbonapi/expr/interfaces"
	"github.com/grafana/carbonapi/expr/types"
	"github.com/grafana/carbonapi/pkg/parser"
)

type aliasByBase64 struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &aliasByBase64{}
	for _, n := range []string{"aliasByBase64"} {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

func (f *aliasByBase64) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArg(ctx, e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	field, err := e.GetIntArg(1)
	field--
	withoutFieldArg := err != nil

	results := make([]*types.MetricData, 0, len(args))

	for _, a := range args {
		r := *a
		if withoutFieldArg {
			decoded, err := base64.StdEncoding.DecodeString(r.Name)
			if err == nil {
				r.Name = string(decoded)
			}
		} else {
			metric := helper.ExtractMetric(r.Name)
			var name []string
			for i, n := range strings.Split(metric, ".") {
				if i == field {
					decoded, err := base64.StdEncoding.DecodeString(n)
					if err == nil {
						n = string(decoded)
					}
				}
				name = append(name, n)
			}
			r.Name = strings.Join(name, ".")
		}

		results = append(results, &r)
	}

	return results, nil
}

func (f *aliasByBase64) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"aliasByBase64": {
			Description: "Takes a seriesList and decodes its name with base64\n" +
				"if node not specified, whole metric name will be decoded as bas64, otherwise only specific node will be\n\n" +
				".. code-block:: none\n\n" +
				"  &target=aliasByBase64(bWV0cmljLm5hbWU=)",
			Function: "aliasByBase64(seriesList)",
			Group:    "Alias",
			Module:   "graphite.render.functions",
			Name:     "aliasByBase64",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "nodeNum",
					Required: false,
					Type:     types.NodeOrTag,
				},
			},
		},
	}
}
