package metadata

import (
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/lomik/zapwriter"
	"go.uber.org/zap"
	"sync"
)

// RegisterRewriteFunction registers function for a rewrite phase in metadata and fills out all Description structs
func RegisterRewriteFunction(name string, function interfaces.RewriteFunction) {
	FunctionMD.Lock()
	defer FunctionMD.Unlock()
	function.SetEvaluator(FunctionMD.evaluator)
	_, ok := FunctionMD.Functions[name]
	if ok {
		logger := zapwriter.Logger("registerFunction")
		logger.Warn("function already registered, will register new anyway",
			zap.String("name", name),
			zap.Stack("stack"),
		)
	}
	FunctionMD.RewriteFunctions[name] = function

	for k, v := range function.Description() {
		FunctionMD.Descriptions[k] = v
		if _, ok := FunctionMD.DescriptionsGrouped[v.Group]; !ok {
			FunctionMD.DescriptionsGrouped[v.Group] = make(map[string]types.FunctionDescription)
		}
		FunctionMD.DescriptionsGrouped[v.Group][k] = v
	}
}

// RegisterFunction registers function in metadata and fills out all Description structs
func RegisterFunction(name string, function interfaces.Function) {
	FunctionMD.Lock()
	defer FunctionMD.Unlock()
	function.SetEvaluator(FunctionMD.evaluator)
	_, ok := FunctionMD.Functions[name]
	if ok {
		logger := zapwriter.Logger("registerFunction")
		logger.Warn("function already registered, will register new anyway",
			zap.String("name", name),
			zap.Stack("stack"),
		)
	}
	FunctionMD.Functions[name] = function

	for k, v := range function.Description() {
		FunctionMD.Descriptions[k] = v
		if _, ok := FunctionMD.DescriptionsGrouped[v.Group]; !ok {
			FunctionMD.DescriptionsGrouped[v.Group] = make(map[string]types.FunctionDescription)
		}
		FunctionMD.DescriptionsGrouped[v.Group][k] = v
	}
}

// SetEvaluator sets new evaluator function to be default for everything that needs it
func SetEvaluator(evaluator interfaces.Evaluator) {
	FunctionMD.Lock()
	defer FunctionMD.Unlock()

	FunctionMD.evaluator = evaluator
	for _, v := range FunctionMD.Functions {
		v.SetEvaluator(evaluator)
	}

	for _, v := range FunctionMD.RewriteFunctions {
		v.SetEvaluator(evaluator)
	}
}

// GetEvaluator returns evaluator
func GetEvaluator() interfaces.Evaluator {
	FunctionMD.RLock()
	defer FunctionMD.RUnlock()

	return FunctionMD.evaluator
}

// SetRewriter sets new rewriter function to be default for everything that needs it
func SetRewriter(rewriter interfaces.Rewriter) {
	FunctionMD.Lock()
	defer FunctionMD.Unlock()

	FunctionMD.rewriter = rewriter
	// For now no rewrite function needs custom rewriter later.
}

// GetEvaluator returns evaluator
func GetRewriter() interfaces.Rewriter {
	FunctionMD.RLock()
	defer FunctionMD.RUnlock()

	return FunctionMD.rewriter
}

// Metadata is a type to store global function metadata
type Metadata struct {
	sync.RWMutex

	Functions           map[string]interfaces.Function
	RewriteFunctions    map[string]interfaces.RewriteFunction
	Descriptions        map[string]types.FunctionDescription
	DescriptionsGrouped map[string]map[string]types.FunctionDescription
	FunctionConfigFiles map[string]string

	evaluator interfaces.Evaluator
	rewriter  interfaces.Rewriter
}

// FunctionMD is actual global variable that stores metadata
var FunctionMD = Metadata{
	RewriteFunctions:    make(map[string]interfaces.RewriteFunction),
	Functions:           make(map[string]interfaces.Function),
	Descriptions:        make(map[string]types.FunctionDescription),
	DescriptionsGrouped: make(map[string]map[string]types.FunctionDescription),
	FunctionConfigFiles: make(map[string]string),
}
