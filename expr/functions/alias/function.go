package alias

import (
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"

	"regexp"
	"strings"
)

func init() {
	metadata.RegisterFunction("alias", &Alias{})
	metadata.RegisterFunction("aliasByMetric", &AliasByMetric{})
	metadata.RegisterFunction("aliasByNode", &AliasByNode{})
	metadata.RegisterFunction("aliasSub", &AliasSub{})
}

type Alias struct {
	interfaces.FunctionBase
}

func (f *Alias) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}
	alias, err := e.GetStringArg(1)
	if err != nil {
		return nil, err
	}

	var results []*types.MetricData

	for _, a := range arg {
		r := *a
		r.Name = alias
		results = append(results, &r)
	}
	return results, nil
}

type AliasByMetric struct {
	interfaces.FunctionBase
}

func (f *AliasByMetric) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	return helper.ForEachSeriesDo(e, from, until, values, func(a *types.MetricData, r *types.MetricData) *types.MetricData {
		metric := helper.ExtractMetric(a.Name)
		part := strings.Split(metric, ".")
		r.Name = part[len(part)-1]
		r.Values = a.Values
		r.IsAbsent = a.IsAbsent
		return r
	})
}

type AliasByNode struct {
	interfaces.FunctionBase
}

func (f *AliasByNode) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	fields, err := e.GetIntArgs(1)
	if err != nil {
		return nil, err
	}

	var results []*types.MetricData

	for _, a := range args {

		metric := helper.ExtractMetric(a.Name)
		nodes := strings.Split(metric, ".")

		var name []string
		for _, f := range fields {
			if f < 0 {
				f += len(nodes)
			}
			if f >= len(nodes) || f < 0 {
				continue
			}
			name = append(name, nodes[f])
		}

		r := *a
		r.Name = strings.Join(name, ".")
		results = append(results, &r)
	}

	return results, nil
}

type AliasSub struct {
	interfaces.FunctionBase
}

func (f *AliasSub) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	search, err := e.GetStringArg(1)
	if err != nil {
		return nil, err
	}

	replace, err := e.GetStringArg(2)
	if err != nil {
		return nil, err
	}

	re, err := regexp.Compile(search)
	if err != nil {
		return nil, err
	}

	replace = helper.Backref.ReplaceAllString(replace, "$${$1}")

	var results []*types.MetricData

	for _, a := range args {
		metric := helper.ExtractMetric(a.Name)

		r := *a
		r.Name = re.ReplaceAllString(metric, replace)
		results = append(results, &r)
	}

	return results, nil

}
