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
		FunctionMD.Descriptions[k] = v
		if _, ok := FunctionMD.DescriptionsGrouped[v.Group]; !ok {
			FunctionMD.DescriptionsGrouped[v.Group] = make(map[string]*types.FunctionDescription)
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
}

// Metadata is a type to store global function metadata
type Metadata struct {
	sync.RWMutex

	Functions map[string]interfaces.Function
	Descriptions map[string]*types.FunctionDescription
	DescriptionsGrouped map[string]map[string]*types.FunctionDescription
	FunctionConfigFiles map[string]string

	evaluator interfaces.Evaluator
}

// FunctionMD is actual global variable that stores metadata
var FunctionMD = Metadata{
	Functions: make(map[string]interfaces.Function),
	Descriptions: make(map[string]*types.FunctionDescription),
	DescriptionsGrouped: make(map[string]map[string]*types.FunctionDescription),
	FunctionConfigFiles: make(map[string]string),
}