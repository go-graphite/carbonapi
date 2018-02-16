package metadata

import (
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/lomik/zapwriter"
	"go.uber.org/zap"
	"sync"
)

// RegisterFunction registers function in metadata and fills out all Description structs
func RegisterFunction(name string, function interfaces.Function) {
	FunctionMD.Lock()
	defer FunctionMD.Unlock()
	function.SetEvaluator(FunctionMD.evaluator)
	_, ok := FunctionMD.Functions[name]
	if ok {
		logger := zapwriter.Logger("registerFunction")
		logger.Error("function already registered, will register new anyway",
			zap.String("name", name),
			zap.Stack("stack"),
		)
	}
	FunctionMD.Functions[name] = function

	for k, v := range function.Description() {
		FunctionDescriptions[k] = v
		if _, ok := FunctionDescriptionsGrouped[v.Group]; !ok {
			FunctionDescriptionsGrouped[v.Group] = make(map[string]*types.FunctionDescription)
		}
		FunctionDescriptionsGrouped[v.Group][k] = v
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
}

// Metadata is a type to store global function metadata
type Metadata struct {
	sync.RWMutex
	Functions map[string]interfaces.Function
	evaluator interfaces.Evaluator
}

// FunctionMD is actual global variable that stores metadata
var FunctionMD = Metadata{
	Functions: make(map[string]interfaces.Function),
}

// FunctionDescriptions is actual global variable that stores description of all functions we support
var FunctionDescriptions = make(map[string]*types.FunctionDescription)

// FunctionDescriptionsGrouped is actual global variable that stores description of all functions we support organised by group
var FunctionDescriptionsGrouped = make(map[string]map[string]*types.FunctionDescription)
