package mapSeries

import (
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"strings"
)

func init() {
	metadata.RegisterFunction("mapSeries", &Function{})
	metadata.RegisterFunction("map", &Function{})
}

type Function struct {
	interfaces.FunctionBase
}

// mapSeries(seriesList, *mapNodes)
// Alias: map
func (f *Function) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	var fields []int

	fields, err = e.GetIntArgs(1)
	if err != nil {
		return nil, err
	}

	var results []*types.MetricData

	groups := make(map[string][]*types.MetricData)
	var nodeList []string

	for _, a := range args {
		metric := helper.ExtractMetric(a.Name)
		nodes := strings.Split(metric, ".")
		nodeKey := make([]string, 0, len(fields))
		for _, f := range fields {
			nodeKey = append(nodeKey, nodes[f])
		}
		node := strings.Join(nodeKey, ".")
		if len(groups[node]) == 0 {
			nodeList = append(nodeList, node)
		}

		groups[node] = append(groups[node], a)
	}

	for _, node := range nodeList {
		results = append(results, groups[node]...)
	}

	return results, nil
}
