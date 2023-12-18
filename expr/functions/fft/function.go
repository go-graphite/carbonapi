package fft

import (
	"context"
	"math/cmplx"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	realFFT "github.com/mjibson/go-dsp/fft"
)

type fft struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &fft{}
	functions := []string{"fft"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

func extractComponent(m *types.MetricData, values []complex128, t string, f func(x complex128) float64) *types.MetricData {
	r := m.CopyTag("fft("+m.Name+","+t+")", m.Tags)
	r.Values = make([]float64, len(values))
	for i, v := range values {
		r.Values[i] = f(v)
	}
	return r
}

// fft(seriesList, mode)
// mode: "", abs, phase. Empty string means "both"
func (f *fft) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}

	mode, _ := e.GetStringArg(1)

	var results []*types.MetricData
	if mode == "abs" || mode == "phase" {
		results = make([]*types.MetricData, 0, len(arg))
	} else {
		results = make([]*types.MetricData, 0, 2*len(arg))
	}
	for _, a := range arg {
		values := realFFT.FFTReal(a.Values)

		switch mode {
		case "", "both", "all":
			results = append(results, extractComponent(a, values, "abs", cmplx.Abs), extractComponent(a, values, "phase", cmplx.Phase))
		case "abs":
			results = append(results, extractComponent(a, values, "abs", cmplx.Abs))
		case "phase":
			results = append(results, extractComponent(a, values, "phase", cmplx.Phase))
		}
	}
	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *fft) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
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
					Options: types.StringsToSuggestionList([]string{
						"abs",
						"phase",
						"both",
					}),
				},
			},
			SeriesChange: true, // function aggregate metrics or change series items count
			NameChange:   true, // name changed
			ValuesChange: true, // values changed
		},
	}
}
