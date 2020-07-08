package expr

import (
	"context"

	utilctx "github.com/go-graphite/carbonapi/util/ctx"

	"github.com/ansel1/merry"
	"github.com/go-graphite/carbonapi/cmd/carbonapi/config"
	_ "github.com/go-graphite/carbonapi/expr/functions"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	pb "github.com/go-graphite/protocol/carbonapi_v3_pb"
)

type evaluator struct{}

// FetchAndEvalExp fetch data and evalualtes expressions
func (eval evaluator) FetchAndEvalExp(ctx context.Context, exp parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	config.Config.Limiter.Enter()
	defer config.Config.Limiter.Leave()

	multiFetchRequest := pb.MultiFetchRequest{}
	metricRequestCache := make(map[string]parser.MetricRequest)
	maxDataPoints := utilctx.GetMaxDatapoints(ctx)

	for _, m := range exp.Metrics() {
		fetchRequest := pb.FetchRequest{
			Name:           m.Metric,
			PathExpression: m.Metric,
			StartTime:      m.From + from,
			StopTime:       m.Until + until,
			MaxDataPoints:  maxDataPoints,
		}
		metricRequest := parser.MetricRequest{
			Metric: fetchRequest.PathExpression,
			From:   fetchRequest.StartTime,
			Until:  fetchRequest.StopTime,
		}

		// avoid multiple requests in a function, E.g divideSeries(a.b, a.b)
		if cachedMetricRequest, ok := metricRequestCache[m.Metric]; ok &&
			cachedMetricRequest.From == metricRequest.From &&
			cachedMetricRequest.Until == metricRequest.Until {
			continue
		}

		// avoid multiple requests in a http request, E.g render?target=a.b&target=a.b
		if _, ok := values[metricRequest]; ok {
			continue
		}

		metricRequestCache[m.Metric] = metricRequest
		multiFetchRequest.Metrics = append(multiFetchRequest.Metrics, fetchRequest)
	}

	if len(multiFetchRequest.Metrics) > 0 {
		metrics, _, err := config.Config.ZipperInstance.Render(ctx, multiFetchRequest)
		// If we had only partial result, we want to do our best to actually do our job
		if err != nil && merry.HTTPCode(err) >= 400 {
			return nil, err
		}
		for _, metric := range metrics {
			metricRequest := metricRequestCache[metric.PathExpression]
			if metric.RequestStartTime != 0 && metric.RequestStopTime != 0 {
				metricRequest.From = metric.RequestStartTime
				metricRequest.Until = metric.RequestStopTime
			}
			data, ok := values[metricRequest]
			if !ok {
				data = make([]*types.MetricData, 0, 1)
			}
			values[metricRequest] = append(data, metric)
		}
	}

	return eval.Eval(ctx, exp, from, until, values)
}

// Eval evalualtes expressions
func (eval evaluator) Eval(ctx context.Context, exp parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) (results []*types.MetricData, err error) {
	rewritten, targets, err := RewriteExpr(ctx, exp, from, until, values)
	if err != nil {
		return nil, err
	}
	if rewritten {
		for _, target := range targets {
			exp, _, err = parser.ParseExpr(target)
			if err != nil {
				return nil, err
			}
			result, err := eval.FetchAndEvalExp(ctx, exp, from, until, values)
			if err != nil {
				return nil, err
			}
			results = append(results, result...)
		}
		return results, nil
	}
	return EvalExpr(ctx, exp, from, until, values)
}

var _evaluator = evaluator{}

func init() {
	helper.SetEvaluator(_evaluator)
	metadata.SetEvaluator(_evaluator)
}

// FetchAndEvalExp fetch data and evalualtes expressions
func FetchAndEvalExp(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	return _evaluator.FetchAndEvalExp(ctx, e, from, until, values)
}

// Eval is the main expression evaluator
func EvalExpr(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
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
		v, err := f.Do(ctx, e, from, until, values)
		if err != nil {
			err = merry.WithMessagef(err, "function=%s", e.Target())
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
func RewriteExpr(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) (bool, []string, error) {
	if e.IsFunc() {
		metadata.FunctionMD.RLock()
		f, ok := metadata.FunctionMD.RewriteFunctions[e.Target()]
		metadata.FunctionMD.RUnlock()
		if ok {
			return f.Do(ctx, e, from, until, values)
		}
	}
	return false, nil, nil
}
