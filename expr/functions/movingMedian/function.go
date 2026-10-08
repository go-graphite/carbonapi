package movingMedian

import (
	"github.com/go-graphite/carbonapi/expr/functions/moving"
	"github.com/go-graphite/carbonapi/expr/interfaces"
)

func GetOrder() interfaces.Order {
	return moving.GetOrder()
}

// New registers movingMedian from the moving package. This package exists only so that
// movingMedian keeps reading its own functionsConfig.movingMedian config file, for backward compatibility.
func New(configFile string) []interfaces.FunctionMetadata {
	return moving.NewFunctions(configFile, "movingMedian")
}
