package interfaces

import (
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

// FunctionBase is a set of base methods that partly satisfy Function interface and most probably nobody will modify
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

// Evaluator is an interface for any existing expression parser
type Evaluator interface {
	EvalExpr(e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error)
}

// Rewriter is an interface for any existing rewriters (main one and the one used in tests)
type Rewriter interface {
	RewriteExpr(e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) (bool, []string, error)
}

type Order int

const (
	Any Order = iota
	Last
)

type RewriteFunctionMetadata struct {
	Name  string
	Order Order
	F     RewriteFunction
}

type FunctionMetadata struct {
	Name  string
	Order Order
	F     Function
}

// Function is interface that all graphite functions should follow
type Function interface {
	SetEvaluator(evaluator Evaluator)
	GetEvaluator() Evaluator
	Do(e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error)
	Description() map[string]types.FunctionDescription
}

// Function is interface that all graphite functions should follow
type RewriteFunction interface {
	SetEvaluator(evaluator Evaluator)
	GetEvaluator() Evaluator
	Do(e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) (bool, []string, error)
	Description() map[string]types.FunctionDescription
}
