package exclude

import (
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"regexp"
)

func init() {
	metadata.RegisterFunction("exclude", &Function{})
}

type Function struct {
	interfaces.FunctionBase
}

// exclude(seriesList, pattern)
func (f *Function) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	pat, err := e.GetStringArg(1)
	if err != nil {
		return nil, err
	}

	patre, err := regexp.Compile(pat)
	if err != nil {
		return nil, err
	}

	var results []*types.MetricData

	for _, a := range arg {
		if !patre.MatchString(a.Name) {
			results = append(results, a)
		}
	}

	return results, nil
}
