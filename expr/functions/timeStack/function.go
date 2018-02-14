package timeStack

import (
	"fmt"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

func init() {
	f := &function{}
	functions := []string{"timeStack"}
	for _, function := range functions {
		metadata.RegisterFunction(function, f)
	}
}

type function struct {
	interfaces.FunctionBase
}

// timeStack(seriesList, timeShiftUnit, timeShiftStart, timeShiftEnd)
func (f *function) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	unit, err := e.GetIntervalArg(1, -1)
	if err != nil {
		return nil, err
	}

	start, err := e.GetIntArg(2)
	if err != nil {
		return nil, err
	}

	end, err := e.GetIntArg(3)
	if err != nil {
		return nil, err
	}

	var results []*types.MetricData
	for i := int32(start); i < int32(end); i++ {
		offs := i * unit
		arg, err := helper.GetSeriesArg(e.Args()[0], from+offs, until+offs, values)
		if err != nil {
			return nil, err
		}

		for _, a := range arg {
			r := *a
			r.Name = fmt.Sprintf("timeShift(%s,%d)", a.Name, offs)
			r.StartTime = a.StartTime - offs
			r.StopTime = a.StopTime - offs
			results = append(results, &r)
		}
	}

	return results, nil
}
