package aliasByBase64

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"github.com/msaf1980/go-stringutils"
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
	args, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}

	field, err := e.GetIntArg(1)
	field--
	withoutFieldArg := err != nil

	results := make([]*types.MetricData, len(args))

	for k, a := range args {
		var r *types.MetricData
		if withoutFieldArg {
			decoded, err := base64.StdEncoding.DecodeString(a.Name)
			if err == nil {
				r = a.CopyName(string(decoded))
			} else {
				r = a
			}
		} else {
			var changed bool
			metric := a.Tags["name"]
			nodeList := strings.Split(metric, ".")
			if field < len(nodeList) {
				decoded, err := base64.StdEncoding.DecodeString(nodeList[field])
				if err == nil {
					n := stringutils.UnsafeString(decoded)
					nodeList[field] = n
					changed = true
				}
			}
			if changed {
				r = a.CopyName(strings.Join(nodeList, "."))
			} else {
				r = a
			}
		}

		results[k] = r
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
			NameChange: true, // name changed
			TagsChange: true, // name tag changed
		},
	}
}
