package fft

import (
	"fmt"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	realFFT "github.com/mjibson/go-dsp/fft"
	"math/cmplx"
)

func init() {
	metadata.RegisterFunction("fft", &fft{})
}

type fft struct {
	interfaces.FunctionBase
}

// fft(seriesList, mode)
// mode: "", abs, phase. Empty string means "both"
func (f *fft) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	mode, _ := e.GetStringArg(1)

	var results []*types.MetricData

	extractComponent := func(m *types.MetricData, values []complex128, t string, f func(x complex128) float64) *types.MetricData {
		name := fmt.Sprintf("fft(%s,'%s')", m.Name, t)
		r := *m
		r.Name = name
		r.Values = make([]float64, len(values))
		r.IsAbsent = make([]bool, len(values))
		for i, v := range values {
			r.Values[i] = f(v)
		}
		return &r
	}

	for _, a := range arg {
		values := realFFT.FFTReal(a.Values)

		switch mode {
		case "", "both", "all":
			results = append(results, extractComponent(a, values, "abs", cmplx.Abs))
			results = append(results, extractComponent(a, values, "phase", cmplx.Phase))
		case "abs":
			results = append(results, extractComponent(a, values, "abs", cmplx.Abs))
		case "phase":
			results = append(results, extractComponent(a, values, "phase", cmplx.Phase))

		}
	}
	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *fft) Description() map[string]*types.FunctionDescription {
	return map[string]*types.FunctionDescription{
		"fft": {
			Description: "An algorithm that samples a signal over a period of time (or space) and divides it into its frequency components. Computes discrete Fourier transform https://en.wikipedia.org/wiki/Fast_Fourier_transform \n\nExample:\n\n.. code-block:: none\n\n  &target=fft(server*.requests_per_second)\n\n  &target=fft(server*.requests_per_second, \"abs\")\n",
			Function:    "fft(seriesList, mode)",
			Group:       "Transform",
			Module:      "graphite.render.functions.custom",
			Name:        "fft",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "mode",
					Required: false,
					Type:     types.String,
					Options: []string{
						"abs",
						"phase",
						"both",
					},
				},
			},
		},
	}
}