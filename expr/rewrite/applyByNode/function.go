package applyByNode

import (
	"context"
	"strings"

	"github.com/ansel1/merry"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

func GetOrder() interfaces.Order {
	return interfaces.Any
}

type applyByNode struct {
	interfaces.FunctionBase
}

func New(configFile string) []interfaces.RewriteFunctionMetadata {
	res := make([]interfaces.RewriteFunctionMetadata, 0)
	f := &applyByNode{}
	for _, n := range []string{"applyByNode"} {
		res = append(res, interfaces.RewriteFunctionMetadata{Name: n, F: f})
	}
	return res
}

func (f *applyByNode) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) (bool, []string, error) {
	args, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return false, nil, err
	}

	nodeNum, err := e.GetIntArg(1)
	if err != nil {
		return false, nil, err
	}

	callback, err := e.GetStringArg(2)
	if err != nil {
		return false, nil, err
	}

	var newName string
	if e.ArgsLen() == 4 {
		newName, err = e.GetStringArg(3)
		if err != nil {
			return false, nil, err
		}
	}

	rv := make([]string, 0, len(args))
	for _, a := range args {
		var node string
		metric := a.Tags["name"]
		nodes := strings.Split(metric, ".")
		if nodeNum >= len(nodes) {
			// field overflow
			err := merry.WithMessagef(parser.ErrInvalidArg, "name=%s: nodeNum must be less than %d", metric, len(nodes))
			return false, nil, err
		} else {
			node = strings.Join(nodes[0:nodeNum+1], ".")
		}
		newTarget := strings.ReplaceAll(callback, "%", node)

		if newName != "" {
			newTarget = "alias(" + newTarget + ",\"" + strings.ReplaceAll(newName, "%", node) + "\")"
		}
		rv = append(rv, newTarget)
	}
	return true, rv, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *applyByNode) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"applyByNode": {
			Name: "applyByNode",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "nodeNum",
					Required: true,
					Type:     types.Node,
				},
				{
					Name:     "templateFunction",
					Required: true,
					Type:     types.String,
				},
				{
					Name: "newName",
					Type: types.String,
				},
			},
			Module:      "graphite.render.functions",
			Description: "Takes a seriesList and applies some complicated function (described by a string), replacing templates with unique\nprefixes of keys from the seriesList (the key is all nodes up to the index given as `nodeNum`).\n\nIf the `newName` parameter is provided, the name of the resulting series will be given by that parameter, with any\n\"%\" characters replaced by the unique prefix.\n\nExample:\n\n.. code-block:: none\n\n  &target=applyByNode(servers.*.disk.bytes_free,1,\"divideSeries(%.disk.bytes_free,sumSeries(%.disk.bytes_*))\")\n\nWould find all series which match `servers.*.disk.bytes_free`, then trim them down to unique series up to the node\ngiven by nodeNum, then fill them into the template function provided (replacing % by the prefixes).\n\nAdditional Examples:\n\nGiven keys of\n\n- `stats.counts.haproxy.web.2XX`\n- `stats.counts.haproxy.web.3XX`\n- `stats.counts.haproxy.web.5XX`\n- `stats.counts.haproxy.microservice.2XX`\n- `stats.counts.haproxy.microservice.3XX`\n- `stats.counts.haproxy.microservice.5XX`\n\nThe following will return the rate of 5XX's per service:\n\n.. code-block:: none\n\n  applyByNode(stats.counts.haproxy.*.*XX, 3, \"asPercent(%.5XX, sumSeries(%.*XX))\", \"%.pct_5XX\")\n\nThe output series would have keys `stats.counts.haproxy.web.pct_5XX` and `stats.counts.haproxy.microservice.pct_5XX`.",
			Function:    "applyByNode(seriesList, nodeNum, templateFunction, newName=None)",
			Group:       "Combine",
		},
	}
}
