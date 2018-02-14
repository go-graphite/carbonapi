package polyfit

import (
	"errors"
	"fmt"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"github.com/gonum/matrix/mat64"
)

func init() {
	f := &function{}
	functions := []string{"polyfit"}
	for _, function := range functions {
		metadata.RegisterFunction(function, f)
	}
}

type function struct {
	interfaces.FunctionBase
}

// polyfit(seriesList, degree=1, offset="0d")
func (f *function) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	// Fitting Nth degree polynom to the dataset
	// https://en.wikipedia.org/wiki/Polynomial_regression#Matrix_form_and_calculation_of_estimates
	arg, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	degree, err := e.GetIntNamedOrPosArgDefault("degree", 1, 1)
	if err != nil {
		return nil, err
	} else if degree < 1 {
		return nil, errors.New("degree must be larger or equal to 1")
	}

	offsStr, err := e.GetStringNamedOrPosArgDefault("offset", 2, "0d")
	if err != nil {
		return nil, err
	}
	offs, err := parser.IntervalString(offsStr, 1)
	if err != nil {
		return nil, err
	}

	var results []*types.MetricData

	for _, a := range arg {
		r := *a
		if len(e.Args()) > 2 {
			r.Name = fmt.Sprintf("polyfit(%s,%d,'%s')", a.Name, degree, e.Args()[2].StringValue())
		} else if len(e.Args()) > 1 {
			r.Name = fmt.Sprintf("polyfit(%s,%d)", a.Name, degree)
		} else {
			r.Name = fmt.Sprintf("polyfit(%s)", a.Name)
		}
		// Extending slice by "offset" so our graph slides into future!
		r.Values = make([]float64, len(a.Values)+int(offs/r.StepTime))
		r.IsAbsent = make([]bool, len(r.Values))
		r.StopTime = a.StopTime + offs

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
