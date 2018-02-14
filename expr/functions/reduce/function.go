package reduce

import (
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"

	"strings"
)

func init() {
	f := &function{}
	metadata.RegisterFunction("reduceSeries", f)
	metadata.RegisterFunction("reduce", f)
}

type function struct {
	interfaces.FunctionBase
}

func (f *function) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	const matchersStartIndex = 3

	if len(e.Args()) < matchersStartIndex+1 {
		return nil, parser.ErrMissingArgument
	}

	seriesList, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
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
		metric := helper.ExtractMetric(series.Name)
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

	for _, aliasName := range aliasNames {

		reducedNodes := make([]parser.Expr, len(reduceMatchers))
		for i, reduceMatcher := range reduceMatchers {
			reducedNodes[i] = parser.NewTargetExpr(reduceGroups[aliasName][reduceMatcher].Name)
		}

		result, err := f.Evaluator.EvalExpr(parser.NewExprTyped("alias", []parser.Expr{
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
