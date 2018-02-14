package substr

import (
	"errors"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"strings"
)

func init() {
	f := &function{}
	functions := []string{"substr"}
	for _, function := range functions {
		metadata.RegisterFunction(function, f)
	}
}

type function struct {
	interfaces.FunctionBase
}

// aliasSub(seriesList, start, stop)
func (f *function) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	// BUG: affected by the same positional arg issue as 'threshold'.
	args, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	startField, err := e.GetIntNamedOrPosArgDefault("start", 1, 0)
	if err != nil {
		return nil, err
	}

	stopField, err := e.GetIntNamedOrPosArgDefault("stop", 2, 0)
	if err != nil {
		return nil, err
	}

	var results []*types.MetricData

	for _, a := range args {
		metric := helper.ExtractMetric(a.Name)
		nodes := strings.Split(metric, ".")
		if startField != 0 {
			if startField < 0 || startField > len(nodes)-1 {
				return nil, errors.New("start out of range")
			}
			nodes = nodes[startField:]
		}
		if stopField != 0 {
			if stopField <= startField || stopField-startField > len(nodes) {
				return nil, errors.New("stop out of range")
			}
			nodes = nodes[:stopField-startField]
		}

		r := *a
		r.Name = strings.Join(nodes, ".")
		results = append(results, &r)
	}

	return results, nil

}
