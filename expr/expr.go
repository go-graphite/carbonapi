package expr

import (
	// Import all known functions
	_ "github.com/go-graphite/carbonapi/expr/functions"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	pb "github.com/go-graphite/protocol/carbonapi_v3_pb"
)

type evaluator struct{}

// EvalExpr evalualtes expressions
func (eval evaluator) EvalExpr(e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	return EvalExpr(e, from, until, values)
}

var _evaluator = evaluator{}

type rewriter struct{}

// EvalExpr evalualtes expressions
func (eval rewriter) RewriteExpr(e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) (bool, []string, error) {
	return RewriteExpr(e, from, until, values)
}

var _rewriter = rewriter{}

func init() {
	helper.SetEvaluator(_evaluator)
	helper.SetRewriter(_rewriter)
	metadata.SetEvaluator(_evaluator)
	metadata.SetRewriter(_rewriter)
}

// EvalExpr is the main expression evaluator
func EvalExpr(e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
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

	return nil, helper.ErrUnknownFunction(e.Target())
}

// RewriteExpr expands targets that use applyByNode into a new list of targets.
// eg:
// applyByNode(foo*, 1, "%") -> (true, ["foo1", "foo2"], nil)
// sumSeries(foo) -> (false, nil, nil)
// Assumes that applyByNode only appears as the outermost function.
func RewriteExpr(e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) (bool, []string, error) {
	if e.IsFunc() {
		metadata.FunctionMD.RLock()
		f, ok := metadata.FunctionMD.RewriteFunctions[e.Target()]
		metadata.FunctionMD.RUnlock()
		if ok {
			return f.Do(e, from, until, values)
		}
	}
	return false, nil, nil
}
