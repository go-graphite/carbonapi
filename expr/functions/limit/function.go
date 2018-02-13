package limit

import (
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

func init() {
	metadata.RegisterFunction("limit", &Function{})
}

type Function struct {
	interfaces.FunctionBase
}

// limit(seriesList, n)
func (f *Function) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	limit, err := e.GetIntArg(1) // get limit
	if err != nil {
		return nil, err
	}

	if limit >= len(arg) {
		return arg, nil
	}

	return arg[:limit], nil
}
