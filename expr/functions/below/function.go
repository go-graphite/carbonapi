package below

import (
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"strings"
)

type below struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &below{}
	functions := []string{"averageAbove", "averageBelow", "currentAbove", "currentBelow", "maximumAbove", "maximumBelow", "minimumAbove", "minimumBelow"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// averageAbove(seriesList, n), averageBelow(seriesList, n), currentAbove(seriesList, n), currentBelow(seriesList, n), maximumAbove(seriesList, n), maximumBelow(seriesList, n), minimumAbove(seriesList, n), minimumBelow
func (f *below) Do(e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	n, err := e.GetFloatArg(1)
	if err != nil {
		return nil, err
	}

	isAbove := strings.HasSuffix(e.Target(), "Above")
	isInclusive := true
	var compute func([]float64) float64
	switch {
	case strings.HasPrefix(e.Target(), "average"):
		compute = helper.AvgValue
	case strings.HasPrefix(e.Target(), "current"):
		compute = helper.CurrentValue
	case strings.HasPrefix(e.Target(), "maximum"):
		compute = helper.MaxValue
		isInclusive = false
	case strings.HasPrefix(e.Target(), "minimum"):
		compute = helper.MinValue
		isInclusive = false
	}
	var results []*types.MetricData
	for _, a := range args {
		value := compute(a.Values)
		if isAbove {
			if isInclusive {
				if value >= n {
					results = append(results, a)
				}
			} else {
				if value > n {
					results = append(results, a)
				}
			}
		} else {
			if value <= n {
				results = append(results, a)
			}
		}
	}

	return results, err
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *below) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"averageAbove": {
			Description: "Takes one metric or a wildcard seriesList followed by an integer N.\nOut of all metrics passed, draws only the metrics with an average value\nabove N for the time period specified.\n\nExample:\n\n.. code-block:: none\n\n  &target=averageAbove(server*.instance*.threads.busy,25)\n\nDraws the servers with average values above 25.",
			Function:    "averageAbove(seriesList, n)",
			Group:       "Filter Series",
			Module:      "graphite.render.functions",
			Name:        "averageAbove",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "n",
					Required: true,
					Type:     types.Integer,
				},
			},
		},
		"averageBelow": {
			Description: "Takes one metric or a wildcard seriesList followed by an integer N.\nOut of all metrics passed, draws only the metrics with an average value\nbelow N for the time period specified.\n\nExample:\n\n.. code-block:: none\n\n  &target=averageBelow(server*.instance*.threads.busy,25)\n\nDraws the servers with average values below 25.",
			Function:    "averageBelow(seriesList, n)",
			Group:       "Filter Series",
			Module:      "graphite.render.functions",
			Name:        "averageBelow",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "n",
					Required: true,
					Type:     types.Integer,
				},
			},
		},
		"currentAbove": {
			Description: "Takes one metric or a wildcard seriesList followed by an integer N.\nOut of all metrics passed, draws only the  metrics whose value is above N\nat the end of the time period specified.\n\nExample:\n\n.. code-block:: none\n\n  &target=currentAbove(server*.instance*.threads.busy,50)\n\nDraws the servers with more than 50 busy threads.",
			Function:    "currentAbove(seriesList, n)",
			Group:       "Filter Series",
			Module:      "graphite.render.functions",
			Name:        "currentAbove",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "n",
					Required: true,
					Type:     types.Integer,
				},
			},
		},
		"currentBelow": {
			Description: "Takes one metric or a wildcard seriesList followed by an integer N.\nOut of all metrics passed, draws only the  metrics whose value is below N\nat the end of the time period specified.\n\nExample:\n\n.. code-block:: none\n\n  &target=currentBelow(server*.instance*.threads.busy,3)\n\nDraws the servers with less than 3 busy threads.",
			Function:    "currentBelow(seriesList, n)",
			Group:       "Filter Series",
			Module:      "graphite.render.functions",
			Name:        "currentBelow",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "n",
					Required: true,
					Type:     types.Integer,
				},
			},
		},
		"maximumAbove": {
			Description: "Takes one metric or a wildcard seriesList followed by a constant n.\nDraws only the metrics with a maximum value above n.\n\nExample:\n\n.. code-block:: none\n\n  &target=maximumAbove(system.interface.eth*.packetsSent,1000)\n\nThis would only display interfaces which sent more than 1000 packets/min.",
			Function:    "maximumAbove(seriesList, n)",
			Group:       "Filter Series",
			Module:      "graphite.render.functions",
			Name:        "maximumAbove",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "n",
					Required: true,
					Type:     types.Integer,
				},
			},
		},
		"maximumBelow": {
			Description: "Takes one metric or a wildcard seriesList followed by a constant n.\nDraws only the metrics with a maximum value below n.\n\nExample:\n\n.. code-block:: none\n\n  &target=maximumBelow(system.interface.eth*.packetsSent,1000)\n\nThis would only display interfaces which sent less than 1000 packets/min.",
			Function:    "maximumBelow(seriesList, n)",
			Group:       "Filter Series",
			Module:      "graphite.render.functions",
			Name:        "maximumBelow",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "n",
					Required: true,
					Type:     types.Integer,
				},
			},
		},
		"minimumAbove": {
			Description: "Takes one metric or a wildcard seriesList followed by a constant n.\nDraws only the metrics with a minimum value above n.\n\nExample:\n\n.. code-block:: none\n\n  &target=minimumAbove(system.interface.eth*.packetsSent,1000)\n\nThis would only display interfaces which sent more than 1000 packets/min.",
			Function:    "minimumAbove(seriesList, n)",
			Group:       "Filter Series",
			Module:      "graphite.render.functions",
			Name:        "minimumAbove",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "n",
					Required: true,
					Type:     types.Integer,
				},
			},
		},
		"minimumBelow": {
			Description: "Takes one metric or a wildcard seriesList followed by a constant n.\nDraws only the metrics with a minimum value below n.\n\nExample:\n\n.. code-block:: none\n\n  &target=minimumBelow(system.interface.eth*.packetsSent,1000)\n\nThis would only display interfaces which at one point sent less than 1000 packets/min.",
			Function:    "minimumBelow(seriesList, n)",
			Group:       "Filter Series",
			Module:      "graphite.render.functions",
			Name:        "minimumBelow",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "n",
					Required: true,
					Type:     types.Integer,
				},
			},
		},
	}
}
