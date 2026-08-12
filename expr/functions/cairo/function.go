package cairo

import (
	"github.com/go-graphite/carbonapi/expr/interfaces"
)

type cairo struct{}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(_ string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &cairo{}
	functions := []string{"color", "stacked", "areaBetween", "alpha", "dashed", "drawAsInfinite", "secondYAxis", "lineWidth", "threshold"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}
