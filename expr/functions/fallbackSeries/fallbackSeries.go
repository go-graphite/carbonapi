package fallbackSeries

import (
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

func init() {
	metadata.RegisterFunction("fallbackSeries", &Function{})
}

type Function struct {
	interfaces.FunctionBase
}

// fallbackSeries( seriesList, fallback )
func (f *Function) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	/*
		Takes a wildcard seriesList, and a second fallback metric.
		If the wildcard does not match any series, draws the fallback metric.
	*/
	seriesList, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	fallback, errFallback := helper.GetSeriesArg(e.Args()[1], from, until, values)
	if errFallback != nil && err != nil {
		return nil, errFallback
	}

	if seriesList != nil && len(seriesList) > 0 {
		return seriesList, nil
	}
	return fallback, nil
}
