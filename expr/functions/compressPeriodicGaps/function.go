package compressPeriodicGaps

import (
	"context"
	"math"

	"github.com/go-graphite/carbonapi/expr/consolidations"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type compressPeriodicGaps struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &compressPeriodicGaps{}
	functions := []string{"compressPeriodicGaps"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// compressPeriodicGaps(seriesList)
func (f *compressPeriodicGaps) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	args, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}
	var results []*types.MetricData

	for _, a := range args {
		firstSeen := -1
		secondSeen := -1
		interval := math.NaN()
		name := "compressPeriodicGaps(" + a.Name + ")"

		for i, v := range a.Values {
			if !math.IsNaN(v) {
				if firstSeen >= 0 {
					secondSeen = i
					break
				} else {
					firstSeen = i
				}
			}
		}
		stepGuess := secondSeen - firstSeen
		thirdSeen := secondSeen + stepGuess
		if stepGuess > 1 && thirdSeen <= len(a.Values)-2 {
			if !math.IsNaN(a.Values[thirdSeen]) {
				if math.IsNaN(a.Values[thirdSeen-1]) && math.IsNaN(a.Values[thirdSeen+1]) {
					interval = float64(int64(stepGuess) * a.StepTime)
				}
			}
		}

		if math.IsNaN(interval) {
			r := a.CopyLink()
			r.Name = name
			results = append(results, r)
		} else {
			newStart := a.StartTime + int64(firstSeen)*a.StepTime
			newValues := make([]float64, 0, int64(interval)/a.StepTime)
			ridx := 0
			intervalItems := 0
			intervalEnd := float64(newStart) + interval
			t := a.StartTime // unadjusted
			buckets := helper.GetBuckets(newStart, a.StopTime, int64(interval))

			r := a.CopyLink()
			r.Name = name
			r.StepTime = int64(interval)
			r.StartTime = newStart
			r.Values = make([]float64, buckets)

			for _, v := range a.Values {
				intervalItems++
				if !math.IsNaN(v) {
					newValues = append(newValues, v)
				}

				t += a.StepTime

				if t >= a.StopTime {
					break
				}

				if t >= int64(intervalEnd) {
					rv := consolidations.SummarizeValues("last", newValues, a.XFilesFactor)

					r.Values[ridx] = rv
					ridx++
					intervalEnd += interval
					intervalItems = 0
					newValues = newValues[:0]
				}
			}

			// last partial bucket
			if intervalItems > 0 {
				rv := consolidations.SummarizeValues("last", newValues, a.XFilesFactor)
				r.Values[ridx] = rv
			}

			r.StopTime = r.StartTime + int64(len(r.Values))*r.StepTime
			results = append(results, r)
		}

	}

	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *compressPeriodicGaps) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"compressPeriodicGaps": {
			Description: "Tries to intelligently remove periodic None’s from series, recalculating start, stop and step values.\n You can use summarize(seriesList, ‘<desired step>’, ‘last’) function for that also, but this function trying to guess desired step automatically.\n Can be used in case of fix metric with improper resolution. Especially useful for derivative functions, which are not working with series with regular gaps.\n\n.. code-block:: none\n\n &target=compressPeriodicGaps(Server.instance01.threads.busy)\n\n",
			Function:    "compressPeriodicGaps(seriesList)",
			Group:       "Transform",
			Module:      "graphite.render.functions",
			Name:        "compressPeriodicGaps",
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
