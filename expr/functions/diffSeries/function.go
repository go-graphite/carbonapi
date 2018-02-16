package diffSeries

import (
	"fmt"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
	"strings"
)

func init() {
	metadata.RegisterFunction("diffSeries", &diffSeries{})
}

type diffSeries struct {
	interfaces.FunctionBase
}

// diffSeries(*seriesLists)
func (f *diffSeries) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	minuends, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	subtrahends, err := helper.GetSeriesArgs(e.Args()[1:], from, until, values)
	if err != nil {
		if len(minuends) < 2 {
			return nil, err
		}
		subtrahends = minuends[1:]
		err = nil
	}

	// We need to rewrite name if there are some missing metrics
	if len(subtrahends)+len(minuends) < len(e.Args()) {
		args := []string{
			helper.RemoveEmptySeriesFromName(minuends),
			helper.RemoveEmptySeriesFromName(subtrahends),
		}
		e.SetRawArgs(strings.Join(args, ","))
	}

	minuend := minuends[0]

	// FIXME: need more error checking on minuend, subtrahends here
	r := *minuend
	r.Name = fmt.Sprintf("diffSeries(%s)", e.RawArgs())
	r.Values = make([]float64, len(minuend.Values))
	r.IsAbsent = make([]bool, len(minuend.Values))

	for i, v := range minuend.Values {

		if minuend.IsAbsent[i] {
			r.IsAbsent[i] = true
			continue
		}

		var sub float64
		for _, s := range subtrahends {
			iSubtrahend := (int32(i) * minuend.StepTime) / s.StepTime

			if s.IsAbsent[iSubtrahend] {
				continue
			}
			sub += s.Values[iSubtrahend]
		}

		r.Values[i] = v - sub
	}
	return []*types.MetricData{&r}, err
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *diffSeries) Description() map[string]*types.FunctionDescription {
	return map[string]*types.FunctionDescription{
		"diffSeries": {
			Description: "Subtracts series 2 through n from series 1.\n\nExample:\n\n.. code-block:: none\n\n  &target=diffSeries(service.connections.total,service.connections.failed)\n\nTo diff a series and a constant, one should use offset instead of (or in\naddition to) diffSeries\n\nExample:\n\n.. code-block:: none\n\n  &target=offset(service.connections.total,-5)\n\n  &target=offset(diffSeries(service.connections.total,service.connections.failed),-4)\n\nThis is an alias for :py:func:`aggregate <aggregate>` with aggregation ``diff``.",
			Function:    "diffSeries(*seriesLists)",
			Group:       "Combine",
			Module:      "graphite.render.functions",
			Name:        "diffSeries",
			Params: []types.FunctionParam{
				{
					Multiple: true,
					Name:     "seriesLists",
					Required: true,
					Type:     types.SeriesList,
				},
			},
		},
	}
}