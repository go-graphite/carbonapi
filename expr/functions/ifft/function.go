package ifft

import (
	"fmt"
	"math"
	"math/cmplx"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	realFFT "github.com/mjibson/go-dsp/fft"
)

type ifft struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &ifft{}
	functions := []string{"ifft"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// ifft(absSeriesList, phaseSeriesList)
func (f *ifft) Do(e parser.Expr, from, until uint32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
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
		if len(phaseSeriesList) > j {
			p := phaseSeriesList[j]
			name := fmt.Sprintf("ifft(%s, %s)", a.Name, p.Name)
			r.Name = name
			values := make([]complex128, len(a.Values))
			for i, v := range a.Values {
				if math.IsNaN(v) {
					v = 0
				}

				values[i] = cmplx.Rect(v, p.Values[i])
			}

			values = realFFT.IFFT(values)
			for i, v := range values {
				r.Values[i] = cmplx.Abs(v)
			}
		} else {
			name := fmt.Sprintf("ifft(%s)", a.Name)
			r.Name = name
			values := realFFT.IFFTReal(a.Values)
			for i, v := range values {
				r.Values[i] = cmplx.Abs(v)
			}
		}

		results = append(results, &r)
	}
	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *ifft) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"ifft": {
			Description: "An algorithm that samples a signal over a period of time (or space) and divides it into its frequency components. Computes discrete Fourier transform https://en.wikipedia.org/wiki/Fast_Fourier_transform \n\nExample:\n\n.. code-block:: none\n\n  &target=fft(server*.requests_per_second)\n\n  &target=fft(server*.requests_per_second, \"abs\")\n",
			Function:    "ifft(seriesList, phaseSeriesList)",
			Group:       "Transform",
			Module:      "graphite.render.functions.custom",
			Name:        "ifft",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "phaseSeriesList",
					Required: true,
					Type:     types.SeriesList,
				},
			},
		},
	}
}
