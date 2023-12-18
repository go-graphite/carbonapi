package polyfit

import (
	"context"
	"errors"
	"math"
	"strconv"

	"github.com/go-graphite/carbonapi/expr/consolidations"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"gonum.org/v1/gonum/mat"
)

type polyfit struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &polyfit{}
	functions := []string{"polyfit"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// polyfit(seriesList, degree=1, offset="0d")
func (f *polyfit) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	// Fitting Nth degree polynom to the dataset
	// https://en.wikipedia.org/wiki/Polynomial_regression#Matrix_form_and_calculation_of_estimates
	arg, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}

	degree, err := e.GetIntNamedOrPosArgDefault("degree", 1, 1)
	if err != nil {
		return nil, err
	} else if degree < 1 {
		return nil, errors.New("degree must be larger or equal to 1")
	}
	degreeStr := strconv.Itoa(degree)

	offsStr, err := e.GetStringNamedOrPosArgDefault("offset", 2, "0d")
	if err != nil {
		return nil, err
	}

	offs, err := parser.IntervalString(offsStr, 1)
	if err != nil {
		return nil, err
	}

	results := make([]*types.MetricData, 0, len(arg))
	for _, a := range arg {
		r := a.CopyLinkTags()
		if e.ArgsLen() > 2 {
			r.Name = "polyfit(" + a.Name + "," + degreeStr + ",'" + offsStr + "')"
		} else if e.ArgsLen() > 1 {
			r.Name = "polyfit(" + a.Name + "," + degreeStr + ")"
		} else {
			r.Name = "polyfit(" + a.Name + ")"
		}
		// Extending slice by "offset" so our graph slides into future!
		r.Values = make([]float64, len(a.Values)+int(offs)/int(r.StepTime))
		r.StopTime = a.StopTime + int64(offs)

		// Removing absent values from original dataset
		nonNulls := make([]float64, 0, len(a.Values))
		for _, v := range a.Values {
			if !math.IsNaN(v) {
				nonNulls = append(nonNulls, v)
			}
		}
		if len(nonNulls) < 2 {
			for i := range r.Values {
				r.Values[i] = math.NaN()
			}
			results = append(results, r)
			continue
		}

		// STEP 1: Creating Vandermonde (X)
		v := consolidations.Vandermonde(a.Values, degree)
		// STEP 2: Creating (X^T * X)**-1
		var t mat.Dense
		t.Mul(v.T(), v)
		var i mat.Dense
		err := i.Inverse(&t)
		if err != nil {
			continue
		}
		// STEP 3: Creating I * X^T * y
		var c mat.Dense
		c.Product(&i, v.T(), mat.NewDense(len(nonNulls), 1, nonNulls))
		// END OF STEPS

		for i := range r.Values {
			r.Values[i] = consolidations.Poly(float64(i), c.RawMatrix().Data...)
		}
		results = append(results, r)
	}
	return results, nil
}

func (f *polyfit) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"polyfit": {
			Description: "Fitting Nth degree polynom to the dataset. https://en.wikipedia.org/wiki/Polynomial_regression#Matrix_form_and_calculation_of_estimates",
			Function:    "polyfit(seriesList, degree=1, offset=\"0d\")",
			Group:       "Combine",
			Module:      "graphite.render.functions.custom",
			Name:        "polyfit",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "degree",
					Default:  types.NewSuggestion(1),
					Required: true,
					Type:     types.Integer,
				},
				{
					Default: types.NewSuggestion("0d"),
					Name:    "offset",
					Type:    types.Interval,
				},
			},
			SeriesChange: true, // function aggregate metrics or change series items count
			NameChange:   true, // name changed
			TagsChange:   true, // name tag changed
			ValuesChange: true, // values changed
		},
	}
}
