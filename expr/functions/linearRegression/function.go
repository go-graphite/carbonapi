package linearRegression

import (
	"fmt"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"github.com/gonum/matrix/mat64"
)

func init() {
	f := &Function{}
	functions := []string{"linearRegression"}
	for _, function := range functions {
		metadata.RegisterFunction(function, f)
	}
}

type Function struct {
	interfaces.FunctionBase
}

// linearRegression(seriesList, startSourceAt=None, endSourceAt=None)
func (f *Function) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	degree := 1

	var results []*types.MetricData

	for _, a := range arg {
		r := *a
		if len(e.Args()) > 2 {
			r.Name = fmt.Sprintf("linearRegression(%s,'%s','%s')", a.GetName(), e.Args()[1].StringValue(), e.Args()[2].StringValue())
		} else if len(e.Args()) > 1 {
			r.Name = fmt.Sprintf("linearRegression(%s,'%s')", a.GetName(), e.Args()[2].StringValue())
		} else {
			r.Name = fmt.Sprintf("linearRegression(%s)", a.GetName())
		}

		r.Values = make([]float64, len(a.Values))
		r.IsAbsent = make([]bool, len(r.Values))
		r.StopTime = a.GetStopTime()

		// Removing absent values from original dataset
		nonNulls := make([]float64, 0)
		for i := range a.Values {
			if !a.IsAbsent[i] {
				nonNulls = append(nonNulls, a.Values[i])
			}
		}
		if len(nonNulls) < 2 {
			for i := range r.IsAbsent {
				r.IsAbsent[i] = true
			}
			results = append(results, &r)
			continue
		}

		// STEP 1: Creating Vandermonde (X)
		v := helper.Vandermonde(a.IsAbsent, degree)
		// STEP 2: Creating (X^T * X)**-1
		var t mat64.Dense
		t.Mul(v.T(), v)
		var i mat64.Dense
		err := i.Inverse(&t)
		if err != nil {
			continue
		}
		// STEP 3: Creating I * X^T * y
		var c mat64.Dense
		c.Product(&i, v.T(), mat64.NewDense(len(nonNulls), 1, nonNulls))
		// END OF STEPS

		for i := range r.Values {
			r.Values[i] = helper.Poly(float64(i), c.RawMatrix().Data...)
		}
		results = append(results, &r)
	}
	return results, nil
}
