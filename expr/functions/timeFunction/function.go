package timeFunction

import (
	"errors"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	pb "github.com/go-graphite/carbonzipper/carbonzipperpb3"
)

func init() {
	f := &function{}
	functions := []string{"timeFunction", "time"}
	for _, function := range functions {
		metadata.RegisterFunction(function, f)
	}
}

type function struct {
	interfaces.FunctionBase
}

func (f *function) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	name, err := e.GetStringArg(0)
	if err != nil {
		return nil, err
	}

	stepInt, err := e.GetIntArgDefault(1, 60)
	if err != nil {
		return nil, err
	}
	if stepInt <= 0 {
		return nil, errors.New("step can't be less than 0")
	}
	step := int32(stepInt)

	// emulate the behavior of this Python code:
	//   while when < requestContext["endTime"]:
	//     newValues.append(time.mktime(when.timetuple()))
	//     when += delta

	newValues := make([]float64, (until-from-1+step)/step)
	value := from
	for i := 0; i < len(newValues); i++ {
		newValues[i] = float64(value)
		value += step
	}

	p := types.MetricData{
		FetchResponse: pb.FetchResponse{
			Name:      name,
			StartTime: from,
			StopTime:  until,
			StepTime:  step,
			Values:    newValues,
			IsAbsent:  make([]bool, len(newValues)),
		},
	}

	return []*types.MetricData{&p}, nil

}
