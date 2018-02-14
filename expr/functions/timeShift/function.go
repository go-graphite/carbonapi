package timeShift

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
	functions := []string{"timeShift"}
	for _, function := range functions {
		metadata.RegisterFunction(function, f)
	}
}

type function struct {
	interfaces.FunctionBase
}

// timeShift(seriesList, timeShift, resetEnd=True)
func (f *function) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	// FIXME(dgryski): support resetEnd=true
	offs, err := e.GetIntervalArg(1, -1)
	if err != nil {
		return nil, err
	}

	arg, err := helper.GetSeriesArg(e.Args()[0], from+offs, until+offs, values)
	if err != nil {
		return nil, err
	}

	var results []*types.MetricData

	for _, a := range arg {
		r := *a
		r.Name = fmt.Sprintf("timeShift(%s,'%d')", a.Name, offs)
		r.StartTime = a.StartTime - offs
		r.StopTime = a.StopTime - offs
		results = append(results, &r)
	}

	return results, nil
}
