package ifft

import (
	"context"
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
func (f *ifft) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	absSeriesList, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}

	var phaseSeriesList []*types.MetricData
	if e.ArgsLen() > 1 {
		phaseSeriesList, err = helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(1), from, until, values)
		if err != nil {
			return nil, err
		}
	}

	results := make([]*types.MetricData, len(absSeriesList))
	for j, a := range absSeriesList {
		r := a.CopyLinkTags()
		r.Values = make([]float64, len(a.Values))
		if len(phaseSeriesList) > j {
			p := phaseSeriesList[j]
			r.Name = "ifft(" + a.Name + "," + p.Name + ")"
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
			r.Name = "ifft(" + a.Name + ")"
			values := realFFT.IFFTReal(a.Values)
			for i, v := range values {
				r.Values[i] = cmplx.Abs(v)
			}
		}

		results[j] = r
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
			NameChange:   true, // name changed
			ValuesChange: true, // values changed
		},
	}
}
