package metadata

import (
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/lomik/zapwriter"
	"go.uber.org/zap"
	"sync"
)

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

func SetEvaluator(evaluator interfaces.Evaluator) {
	FunctionMD.Lock()
	defer FunctionMD.Unlock()

	FunctionMD.evaluator = evaluator
	for _, v := range FunctionMD.Functions {
		v.SetEvaluator(evaluator)
	}
}

type Metadata struct {
	sync.RWMutex
	Functions map[string]interfaces.Function
	evaluator interfaces.Evaluator
}

var FunctionMD = Metadata{
	Functions: make(map[string]interfaces.Function),
}

var FunctionDescriptions = make(map[string]*types.FunctionDescription)
var FunctionDescriptionsGrouped = make(map[string]map[string]*types.FunctionDescription)