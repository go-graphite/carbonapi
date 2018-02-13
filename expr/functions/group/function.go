package group

import (
	"fmt"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"strings"
)

func init() {
	metadata.RegisterFunction("group", &Group{})
	metadata.RegisterFunction("groupByNode", &GroupByNode{})
	metadata.RegisterFunction("groupByNodes", &GroupByNode{})
}

type Group struct {
	interfaces.FunctionBase
}

// group(*seriesLists)
func (f *Group) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArgsAndRemoveNonExisting(e, from, until, values)
	if err != nil {
		return nil, err
	}

	return args, nil
}

type GroupByNode struct {
	interfaces.FunctionBase
}

// groupByNode(seriesList, nodeNum, callback)
// groupByNodes(seriesList, callback, *nodes)
func (f *GroupByNode) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}
	var callback string
	var fields []int

	if e.Target() == "groupByNode" {
		field, err := e.GetIntArg(1)
		if err != nil {
			return nil, err
		}

		callback, err = e.GetStringArg(2)
		if err != nil {
			return nil, err
		}
		fields = []int{field}
	} else {
		callback, err = e.GetStringArg(1)
		if err != nil {
			return nil, err
		}

		fields, err = e.GetIntArgs(2)
		if err != nil {
			return nil, err
		}
	}

	var results []*types.MetricData

	groups := make(map[string][]*types.MetricData)
	nodeList := []string{}

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

	for _, k := range nodeList {
		k := k // k's reference is used later, so it's important to make it unique per loop
		v := groups[k]

		// Ensure that names won't be parsed as consts, appending stub to them
		expr := fmt.Sprintf("%s(stub_%s)", callback, k)

		// create a stub context to evaluate the callback in
		nexpr, _, err := parser.ParseExpr(expr)
		// remove all stub_ prefixes we've prepended before
		nexpr.SetRawArgs(strings.Replace(nexpr.RawArgs(), "stub_", "", 1))
		for argIdx := range nexpr.Args() {
			nexpr.Args()[argIdx].SetTarget(strings.Replace(nexpr.Args()[0].Target(), "stub_", "", 1))
		}
		if err != nil {
			return nil, err
		}

		nvalues := values
		if e.Target() == "groupByNode" || e.Target() == "groupByNodes" {
			nvalues = map[parser.MetricRequest][]*types.MetricData{
				parser.MetricRequest{k, from, until}: v,
			}
		}

		r, _ := f.Evaluator.EvalExpr(nexpr, from, until, nvalues)
		if r != nil {
			r[0].Name = k
			results = append(results, r...)
		}
	}

	return results, nil
}
