package interfaces

import (
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type FunctionBase struct {
	Evaluator Evaluator
}

func (b *FunctionBase) SetEvaluator(evaluator Evaluator) {
	b.Evaluator = evaluator
}

func (b *FunctionBase) GetEvaluator() Evaluator {
	return b.Evaluator
}

type Evaluator interface {
	EvalExpr(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error)
}

type Function interface {
	SetEvaluator(evaluator Evaluator)
	GetEvaluator() Evaluator
	Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error)
}
