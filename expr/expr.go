package expr

import (
	"fmt"
	"strings"

	// Import all known functions
	_ "github.com/go-graphite/carbonapi/expr/functions"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/png"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	pb "github.com/go-graphite/carbonzipper/carbonzipperpb3"
)

type evaluator struct{}

// EvalExpr evalualtes expressions
func (eval evaluator) EvalExpr(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	return EvalExpr(e, from, until, values)
}

var _evaluator = evaluator{}

func init() {
	helper.SetEvaluator(_evaluator)
	metadata.SetEvaluator(_evaluator)
}

func EvalExpr(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {

	if e.IsName() {
		return values[parser.MetricRequest{Metric: e.Target(), From: from, Until: until}], nil
	} else if e.IsConst() {
		p := types.MetricData{FetchResponse: pb.FetchResponse{Name: e.Target(), Values: []float64{e.FloatValue()}}}
		return []*types.MetricData{&p}, nil
	}

	// evaluate the function

	// all functions have arguments -- check we do too
	if len(e.Args()) == 0 {
		return nil, parser.ErrMissingArgument
	}

	metadata.FunctionMD.RLock()
	f, ok := metadata.FunctionMD.Functions[e.Target()]
	metadata.FunctionMD.RUnlock()
	if ok {
		return f.Do(e, from, until, values)
	}

	if png.HaveGraphSupport {
		return png.EvalExprGraph(e, from, until, values)
	}

	return nil, helper.ErrUnknownFunction(e.Target())
}

// RewriteExpr expands targets that use applyByNode into a new list of targets.
// eg:
// applyByNode(foo*, 1, "%") -> (true, ["foo1", "foo2"], nil)
// sumSeries(foo) -> (false, nil, nil)
// Assumes that applyByNode only appears as the outermost function.
func RewriteExpr(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) (bool, []string, error) {
	if e.IsFunc() && e.Target() == "applyByNode" {
		args, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
		if err != nil {
			return false, nil, err
		}

		field, err := e.GetIntArg(1)
		if err != nil {
			return false, nil, err
		}

		callback, err := e.GetStringArg(2)
		if err != nil {
			return false, nil, err
		}

		var newName string
		if len(e.Args()) == 4 {
			newName, err = e.GetStringArg(3)
			if err != nil {
				return false, nil, err
			}
		}

		var rv []string
		for _, a := range args {
			metric := helper.ExtractMetric(a.Name)
			nodes := strings.Split(metric, ".")
			node := strings.Join(nodes[0:field], ".")
			newTarget := strings.Replace(callback, "%", node, -1)

			if newName != "" {
				newTarget = fmt.Sprintf("alias(%s,\"%s\")", newTarget, strings.Replace(newName, "%", node, -1))
			}
			rv = append(rv, newTarget)
		}
		return true, rv, nil
	}
	return false, nil, nil
}
