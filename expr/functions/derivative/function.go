package derivative

import (
	"math"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

func init() {
	metadata.RegisterFunction("derivative", &derivative{})
}

type derivative struct {
	interfaces.FunctionBase
}

// derivative(seriesList)
func (f *derivative) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	return helper.ForEachSeriesDo(e, from, until, values, func(a *types.MetricData, r *types.MetricData) *types.MetricData {
		prev := math.NaN()
		for i, v := range a.Values {
			if a.IsAbsent[i] {
				r.IsAbsent[i] = true
				continue
			} else if math.IsNaN(prev) {
				r.IsAbsent[i] = true
				prev = v
				continue
			}

			r.Values[i] = v - prev
			prev = v
		}
		return r
	})
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *derivative) Description() map[string]*types.FunctionDescription {
	return map[string]*types.FunctionDescription{
		"derivative": {
			Description: "This is the opposite of the integral function.  This is useful for taking a\nrunning total metric and calculating the delta between subsequent data points.\n\nThis function does not normalize for periods of time, as a true derivative would.\nInstead see the perSecond() function to calculate a rate of change over time.\n\nExample:\n\n.. code-block:: none\n\n  &target=derivative(company.server.application01.ifconfig.TXPackets)\n\nEach time you run ifconfig, the RX and TXPackets are higher (assuming there\nis network traffic.) By applying the derivative function, you can get an\nidea of the packets per minute sent or received, even though you're only\nrecording the total.",
			Function:    "derivative(seriesList)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "derivative",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
			},
		},
	}
}