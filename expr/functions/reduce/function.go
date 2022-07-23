package reduce

import (
	"context"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/helper/metric"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"

	"strings"
)

type reduce struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &reduce{}
	functions := []string{"reduceSeries", "reduce"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

func (f *reduce) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	const matchersStartIndex = 3

	if len(e.Args()) < matchersStartIndex+1 {
		return nil, parser.ErrMissingArgument
	}

	seriesList, err := helper.GetSeriesArg(ctx, e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	reduceFunction, err := e.GetStringArg(1)
	if err != nil {
		return nil, err
	}

	reduceNode, err := e.GetIntArg(2)
	if err != nil {
		return nil, err
	}

	argsCount := len(e.Args())
	matchersCount := argsCount - matchersStartIndex
	reduceMatchers := make([]string, matchersCount)
	for i := matchersStartIndex; i < argsCount; i++ {
		reduceMatcher, err := e.GetStringArg(i)
		if err != nil {
			return nil, err
		}

		reduceMatchers[i-matchersStartIndex] = reduceMatcher
	}

	var results []*types.MetricData

	reduceGroups := make(map[string]map[string]*types.MetricData)
	reducedValues := values
	var aliasNames []string

	for _, series := range seriesList {
		metric := metric.ExtractMetric(series.Name)
		nodes := strings.Split(metric, ".")
		reduceNodeKey := nodes[reduceNode]
		nodes[reduceNode] = "reduce." + reduceFunction
		aliasName := strings.Join(nodes, ".")
		_, exist := reduceGroups[aliasName]
		if !exist {
			reduceGroups[aliasName] = make(map[string]*types.MetricData)
			aliasNames = append(aliasNames, aliasName)
		}

		reduceGroups[aliasName][reduceNodeKey] = series
		valueKey := parser.MetricRequest{series.Name, from, until}
		reducedValues[valueKey] = append(reducedValues[valueKey], series)
	}
AliasLoop:
	for _, aliasName := range aliasNames {

		reducedNodes := make([]parser.Expr, len(reduceMatchers))
		for i, reduceMatcher := range reduceMatchers {
			matched, ok := reduceGroups[aliasName][reduceMatcher]
			if !ok {
				continue AliasLoop
			}
			reducedNodes[i] = parser.NewTargetExpr(matched.Name)
		}

		result, err := f.Evaluator.Eval(ctx, parser.NewExprTyped("alias", []parser.Expr{
			parser.NewExprTyped(reduceFunction, reducedNodes),
			parser.NewValueExpr(aliasName),
		}), from, until, reducedValues)

		if err != nil {
			return nil, err
		}

		results = append(results, result...)
	}

	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *reduce) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"reduceSeries": {
			Description: "Short form: ``reduce()``\n\nTakes a list of seriesLists and reduces it to a list of series by means of the reduceFunction.\n\nReduction is performed by matching the reduceNode in each series against the list of\nreduceMatchers. Then each series is passed to the reduceFunction as arguments in the order\ngiven by reduceMatchers. The reduceFunction should yield a single series.\n\nThe resulting list of series are aliased so that they can easily be nested in other functions.\n\n**Example**: Map/Reduce asPercent(bytes_used,total_bytes) for each server\n\nAssume that metrics in the form below exist:\n\n.. code-block:: none\n\n     servers.server1.disk.bytes_used\n     servers.server1.disk.total_bytes\n     servers.server2.disk.bytes_used\n     servers.server2.disk.total_bytes\n     servers.server3.disk.bytes_used\n     servers.server3.disk.total_bytes\n     ...\n     servers.serverN.disk.bytes_used\n     servers.serverN.disk.total_bytes\n\nTo get the percentage of disk used for each server:\n\n.. code-block:: none\n\n    reduceSeries(mapSeries(servers.*.disk.*,1),\"asPercent\",3,\"bytes_used\",\"total_bytes\") =>\n\n      alias(asPercent(servers.server1.disk.bytes_used,servers.server1.disk.total_bytes),\"servers.server1.disk.reduce.asPercent\"),\n      alias(asPercent(servers.server2.disk.bytes_used,servers.server2.disk.total_bytes),\"servers.server2.disk.reduce.asPercent\"),\n      alias(asPercent(servers.server3.disk.bytes_used,servers.server3.disk.total_bytes),\"servers.server3.disk.reduce.asPercent\"),\n      ...\n      alias(asPercent(servers.serverN.disk.bytes_used,servers.serverN.disk.total_bytes),\"servers.serverN.disk.reduce.asPercent\")\n\nIn other words, we will get back the following metrics::\n\n    servers.server1.disk.reduce.asPercent\n    servers.server2.disk.reduce.asPercent\n    servers.server3.disk.reduce.asPercent\n    ...\n    servers.serverN.disk.reduce.asPercent\n\n.. seealso:: :py:func:`mapSeries`",
			Function:    "reduceSeries(seriesLists, reduceFunction, reduceNode, *reduceMatchers)",
			Group:       "Combine",
			Module:      "graphite.render.functions",
			Name:        "reduceSeries",
			Params: []types.FunctionParam{
				{
					Name:     "seriesLists",
					Required: true,
					Type:     types.SeriesLists,
				},
				{
					Name:     "reduceFunction",
					Required: true,
					Type:     types.String,
				},
				{
					Name:     "reduceNode",
					Required: true,
					Type:     types.Node,
				},
				{
					Multiple: true,
					Name:     "reduceMatchers",
					Required: true,
					Type:     types.String,
				},
			},
		},
		"reduce": {
			Description: "Short form: ``reduce()``\n\nTakes a list of seriesLists and reduces it to a list of series by means of the reduceFunction.\n\nReduction is performed by matching the reduceNode in each series against the list of\nreduceMatchers. Then each series is passed to the reduceFunction as arguments in the order\ngiven by reduceMatchers. The reduceFunction should yield a single series.\n\nThe resulting list of series are aliased so that they can easily be nested in other functions.\n\n**Example**: Map/Reduce asPercent(bytes_used,total_bytes) for each server\n\nAssume that metrics in the form below exist:\n\n.. code-block:: none\n\n     servers.server1.disk.bytes_used\n     servers.server1.disk.total_bytes\n     servers.server2.disk.bytes_used\n     servers.server2.disk.total_bytes\n     servers.server3.disk.bytes_used\n     servers.server3.disk.total_bytes\n     ...\n     servers.serverN.disk.bytes_used\n     servers.serverN.disk.total_bytes\n\nTo get the percentage of disk used for each server:\n\n.. code-block:: none\n\n    reduceSeries(mapSeries(servers.*.disk.*,1),\"asPercent\",3,\"bytes_used\",\"total_bytes\") =>\n\n      alias(asPercent(servers.server1.disk.bytes_used,servers.server1.disk.total_bytes),\"servers.server1.disk.reduce.asPercent\"),\n      alias(asPercent(servers.server2.disk.bytes_used,servers.server2.disk.total_bytes),\"servers.server2.disk.reduce.asPercent\"),\n      alias(asPercent(servers.server3.disk.bytes_used,servers.server3.disk.total_bytes),\"servers.server3.disk.reduce.asPercent\"),\n      ...\n      alias(asPercent(servers.serverN.disk.bytes_used,servers.serverN.disk.total_bytes),\"servers.serverN.disk.reduce.asPercent\")\n\nIn other words, we will get back the following metrics::\n\n    servers.server1.disk.reduce.asPercent\n    servers.server2.disk.reduce.asPercent\n    servers.server3.disk.reduce.asPercent\n    ...\n    servers.serverN.disk.reduce.asPercent\n\n.. seealso:: :py:func:`mapSeries`",
			Function:    "reduce(seriesLists, reduceFunction, reduceNode, *reduceMatchers)",
			Group:       "Combine",
			Module:      "graphite.render.functions",
			Name:        "reduce",
			Params: []types.FunctionParam{
				{
					Name:     "seriesLists",
					Required: true,
					Type:     types.SeriesLists,
				},
				{
					Name:     "reduceFunction",
					Required: true,
					Type:     types.String,
				},
				{
					Name:     "reduceNode",
					Required: true,
					Type:     types.Node,
				},
				{
					Multiple: true,
					Name:     "reduceMatchers",
					Required: true,
					Type:     types.String,
				},
			},
		},
	}
}
