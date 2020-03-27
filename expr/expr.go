package expr

import (
	"context"
	"fmt"

	"github.com/go-graphite/carbonapi/cmd/carbonapi/config"
	_ "github.com/go-graphite/carbonapi/expr/functions"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	pb "github.com/go-graphite/protocol/carbonapi_v3_pb"
)

type evaluator struct{}

// FetchTargetExp evalualtes expressions
func (eval evaluator) FetchTargetExp(exp parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	config.Config.Limiter.Enter()

	var req pb.MultiFetchRequest

	for _, m := range exp.Metrics() {
		req.Metrics = append(req.Metrics, pb.FetchRequest{
			Name:           m.Metric,
			PathExpression: m.Metric,
			StartTime:      m.From + from,
			StopTime:       m.Until + until,
		})
	}

	metrics, _, err := config.Config.ZipperInstance.Render(context.TODO(), req)
	if err != nil {
		return nil, err
	}
	for _, metric := range metrics {
		mFetch := parser.MetricRequest{
			Metric: metric.PathExpression,
			From:   from,
			Until:  until,
		}
		data, ok := values[mFetch]
		if !ok {
			data = make([]*types.MetricData, 0, 1)
		}
		values[mFetch] = append(data, metric)
	}

	config.Config.Limiter.Leave()

	return eval.Eval(exp, from, until, values)
}

// Eval evalualtes expressions
func (eval evaluator) Eval(exp parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) (results []*types.MetricData, err error) {
	rewritten, targets, err := RewriteExpr(exp, from, until, values)
	if err != nil {
		return nil, err
	}
	if rewritten {
		for _, target := range targets {
			exp, _, err = parser.ParseExpr(target)
			if err != nil {
				return nil, err
			}
			result, err := eval.FetchTargetExp(exp, from, until, values)
			if err != nil {
				return nil, err
			}
			results = append(results, result...)
		}
		return results, nil
	}
	return EvalExpr(exp, from, until, values)
}

var _evaluator = evaluator{}

func init() {
	helper.SetEvaluator(_evaluator)
	metadata.SetEvaluator(_evaluator)
}

func FetchTargetExp(e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	return _evaluator.FetchTargetExp(e, from, until, values)
}

// Eval is the main expression evaluator
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
		v, err := f.Do(e, from, until, values)
		if err != nil {
			err = fmt.Errorf("function=%s, err=%v", e.Target(), err)
		}
		return v, err
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
