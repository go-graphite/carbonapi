package linearRegression

import (
	"context"
	"math"
	"strconv"

	"github.com/go-graphite/carbonapi/expr/consolidations"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"gonum.org/v1/gonum/mat"
)

type linearRegression struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &linearRegression{}
	functions := []string{"linearRegression"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// linearRegression(seriesList, startSourceAt=None, endSourceAt=None)
func (f *linearRegression) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}

	degree := 1

	results := make([]*types.MetricData, 0, len(arg))

	for _, a := range arg {
		r := a.CopyLink()
		if e.ArgsLen() > 2 {
			r.Name = "linearRegression(" + a.GetName() + ",'" + e.Arg(1).StringValue() + "','" + e.Arg(2).StringValue() + "')"
		} else if e.ArgsLen() > 1 {
			r.Name = "linearRegression(" + a.GetName() + ",'" + e.Arg(1).StringValue() + "')"
		} else {
			r.Name = "linearRegression(" + a.Name + ")"
		}

		r.Values = make([]float64, len(a.Values))
		r.StopTime = a.GetStopTime()

		// Removing absent values from original dataset
		nonNulls := make([]float64, 0, len(a.Values))
		for i, v := range a.Values {
			if !math.IsNaN(v) {
				nonNulls = append(nonNulls, a.Values[i])
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
		r.Tags["linearRegressions"] = strconv.FormatInt(a.GetStartTime(), 10) + ", " + strconv.FormatInt(a.GetStopTime(), 10)

		results = append(results, r)
	}
	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *linearRegression) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"linearRegression": {
			Description: "Graphs the liner regression function by least squares method.\n\nTakes one metric or a wildcard seriesList, followed by a quoted string with the\ntime to start the line and another quoted string with the time to end the line.\nThe start and end times are inclusive (default range is from to until). See\n``from / until`` in the render\\_api_ for examples of time formats. Datapoints\nin the range is used to regression.\n\nExample:\n\n.. code-block:: none\n\n  &target=linearRegression(Server.instance01.threads.busy, '-1d')\n  &target=linearRegression(Server.instance*.threads.busy, \"00:00 20140101\",\"11:59 20140630\")",
			Function:    "linearRegression(seriesList, startSourceAt=None, endSourceAt=None)",
			Group:       "Calculate",
			Module:      "graphite.render.functions",
			Name:        "linearRegression",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name: "startSourceAt",
					Type: types.Date,
				},
				{
					Name: "endSourceAt",
					Type: types.Date,
				},
			},
			SeriesChange: true, // function aggregate metrics or change series items count
			NameChange:   true, // name changed
			ValuesChange: true, // values changed
		},
	}
}
