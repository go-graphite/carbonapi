package interfaces

import (
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

// FunctionBase is a set of base methods that partly statisfy Function interface and most probably nobody will modify
type FunctionBase struct {
	Evaluator Evaluator
}

// SetEvaluator sets evaluator
func (b *FunctionBase) SetEvaluator(evaluator Evaluator) {
	b.Evaluator = evaluator
}

// GetEvaluator returns evaluator
func (b *FunctionBase) GetEvaluator() Evaluator {
	return b.Evaluator
}

// Evaluator is a interface for any existing expression parser
type Evaluator interface {
	EvalExpr(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error)
}

// Function is interface that all graphite functions should follow
type Function interface {
	SetEvaluator(evaluator Evaluator)
	GetEvaluator() Evaluator
	Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error)
	Description() map[string]*types.FunctionDescription
}
