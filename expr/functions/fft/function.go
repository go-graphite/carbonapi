package fft

import (
	"fmt"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"github.com/mjibson/go-dsp/fft"
	"math/cmplx"
)

func init() {
	metadata.RegisterFunction("fft", &FFT{})
	metadata.RegisterFunction("ifft", &IFFT{})
}

type FFT struct {
	interfaces.FunctionBase
}

// fft(seriesList, mode)
// mode: "", abs, phase. Empty string means "both"
func (f *FFT) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
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
		values := fft.FFTReal(a.Values)

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

type IFFT struct {
	interfaces.FunctionBase
}

// ifft(absSeriesList, phaseSeriesList)
func (f *IFFT) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	absSeriesList, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	var phaseSeriesList []*types.MetricData
	if len(e.Args()) > 1 {
		phaseSeriesList, err = helper.GetSeriesArg(e.Args()[1], from, until, values)
		if err != nil {
			return nil, err
		}
	}

	var results []*types.MetricData
	for j, a := range absSeriesList {
		r := *a
		r.Values = make([]float64, len(a.Values))
		r.IsAbsent = make([]bool, len(a.Values))
		if len(phaseSeriesList) > j {
			p := phaseSeriesList[j]
			name := fmt.Sprintf("ifft(%s, %s)", a.Name, p.Name)
			r.Name = name
			values := make([]complex128, len(a.Values))
			for i, v := range a.Values {
				if a.IsAbsent[i] {
					v = 0
				}

				values[i] = cmplx.Rect(v, p.Values[i])
			}

			values = fft.IFFT(values)
			for i, v := range values {
				r.Values[i] = cmplx.Abs(v)
			}
		} else {
			name := fmt.Sprintf("ifft(%s)", a.Name)
			r.Name = name
			values := fft.IFFTReal(a.Values)
			for i, v := range values {
				r.Values[i] = cmplx.Abs(v)
			}
		}

		results = append(results, &r)
	}
	return results, nil
}
