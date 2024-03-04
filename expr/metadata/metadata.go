package metadata

import (
	"sync"

	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/lomik/zapwriter"
	"go.uber.org/zap"
)

// RegisterRewriteFunctionWithFilename registers function for a rewrite phase in metadata and fills out all Description structs
func RegisterRewriteFunctionWithFilename(name, filename string, function interfaces.RewriteFunction) {
	FunctionMD.Lock()
	defer FunctionMD.Unlock()

	if _, ok := FunctionMD.RewriteFunctions[name]; ok {
		n := FunctionMD.RewriteFunctionsFilenames[name]
		logger := zapwriter.Logger("registerRewriteFunction")
		logger.Warn("function already registered, will register new anyway",
			zap.String("name", name),
			zap.String("current_filename", filename),
			zap.Strings("previous_filenames", n),
			zap.Stack("stack"),
		)
	} else {
		FunctionMD.RewriteFunctionsFilenames[name] = make([]string, 0)
	}
	// Check if we are colliding with non-rewrite Functions
	if _, ok := FunctionMD.Functions[name]; ok {
		n := FunctionMD.FunctionsFilenames[name]
		logger := zapwriter.Logger("registerRewriteFunction")
		logger.Warn("non-rewrite function with the same name already registered",
			zap.String("name", name),
			zap.String("current_filename", filename),
			zap.Strings("previous_filenames", n),
			zap.Stack("stack"),
		)
	}
	FunctionMD.RewriteFunctionsFilenames[name] = append(FunctionMD.RewriteFunctionsFilenames[name], filename)
	FunctionMD.RewriteFunctions[name] = function

	for k, v := range function.Description() {
		FunctionMD.Descriptions[k] = v
		if _, ok := FunctionMD.DescriptionsGrouped[v.Group]; !ok {
			FunctionMD.DescriptionsGrouped[v.Group] = make(map[string]types.FunctionDescription)
		}
		FunctionMD.DescriptionsGrouped[v.Group][k] = v
	}
}

// RegisterRewriteFunction registers function for a rewrite phase in metadata and fills out all Description structs
func RegisterRewriteFunction(name string, function interfaces.RewriteFunction) {
	RegisterRewriteFunctionWithFilename(name, "", function)
}

// RegisterFunctionWithFilename registers function in metadata and fills out all Description structs
func RegisterFunctionWithFilename(name, filename string, function interfaces.Function) {
	FunctionMD.Lock()
	defer FunctionMD.Unlock()

	if _, ok := FunctionMD.Functions[name]; ok {
		n := FunctionMD.FunctionsFilenames[name]
		logger := zapwriter.Logger("registerFunction")
		logger.Warn("function already registered, will register new anyway",
			zap.String("name", name),
			zap.String("current_filename", filename),
			zap.Strings("previous_filenames", n),
			zap.Stack("stack"),
		)
	} else {
		FunctionMD.FunctionsFilenames[name] = make([]string, 0)
	}
	// Check if we are colliding with non-rewrite Functions
	if _, ok := FunctionMD.RewriteFunctions[name]; ok {
		n := FunctionMD.RewriteFunctionsFilenames[name]
		logger := zapwriter.Logger("registerRewriteFunction")
		logger.Warn("rewrite function with the same name already registered",
			zap.String("name", name),
			zap.String("current_filename", filename),
			zap.Strings("previous_filenames", n),
			zap.Stack("stack"),
		)
	}
	FunctionMD.Functions[name] = function
	FunctionMD.FunctionsFilenames[name] = append(FunctionMD.FunctionsFilenames[name], filename)

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
	RegisterFunctionWithFilename(name, "", function)
}

// Metadata is a type to store global function metadata
type Metadata struct {
	sync.RWMutex

	Functions                 map[string]interfaces.Function
	RewriteFunctions          map[string]interfaces.RewriteFunction
	Descriptions              map[string]types.FunctionDescription
	DescriptionsGrouped       map[string]map[string]types.FunctionDescription
	FunctionConfigFiles       map[string]string
	FunctionsFilenames        map[string][]string
	RewriteFunctionsFilenames map[string][]string

	evaluator interfaces.Evaluator
}

// FunctionMD is actual global variable that stores metadata
var FunctionMD = Metadata{
	RewriteFunctions:          make(map[string]interfaces.RewriteFunction),
	Functions:                 make(map[string]interfaces.Function),
	Descriptions:              make(map[string]types.FunctionDescription),
	DescriptionsGrouped:       make(map[string]map[string]types.FunctionDescription),
	FunctionConfigFiles:       make(map[string]string),
	FunctionsFilenames:        make(map[string][]string),
	RewriteFunctionsFilenames: make(map[string][]string),
}
