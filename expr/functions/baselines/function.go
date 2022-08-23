package baselines

import (
	"context"
	"math"

	"github.com/go-graphite/carbonapi/expr/consolidations"
	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type baselines struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &baselines{}
	functions := []string{"baseline", "baselineAberration"}
	for _, n := range functions {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

func (f *baselines) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	unit, err := e.GetIntervalArg(1, -1)
	if err != nil {
		return nil, err
	}
	start, err := e.GetIntArg(2)
	if err != nil {
		return nil, err
	}
	end, err := e.GetIntArg(3)
	if err != nil {
		return nil, err
	}
	maxAbsentPercent, err := e.GetFloatArgDefault(4, math.NaN())
	if err != nil {
		return nil, err
	}
	minAvgLimit, err := e.GetFloatArgDefault(5, math.NaN())
	if err != nil {
		return nil, err
	}

	isAberration := false
	if e.Target() == "baselineAberration" {
		isAberration = true
	}

	current := make(map[string]*types.MetricData)
	arg, _ := helper.GetSeriesArg(ctx, e.Arg(0), from, until, values)
	for _, a := range arg {
		current[a.Name] = a
	}

	groups := make(map[string][]*types.MetricData)
	for i := int32(start); i < int32(end); i++ {
		if i == 0 {
			continue
		}
		offs := int64(i * unit)
		arg, _ := helper.GetSeriesArg(ctx, e.Arg(0), from+offs, until+offs, values)
		for _, a := range arg {
			r := a.CopyLinkTags()
			if _, ok := current[r.Name]; ok || !isAberration {
				r.StartTime = a.StartTime - offs
				r.StopTime = a.StopTime - offs
				groups[r.Name] = append(groups[r.Name], r)
			}
		}
	}

	results := make([]*types.MetricData, 0, len(groups))
	for name, args := range groups {
		var newName string
		if isAberration {
			newName = "baselineAberration(" + name + ")"
		} else {
			newName = "baseline(" + name + ")"
		}
		r := args[0].CopyName(newName)
		r.Values = make([]float64, len(args[0].Values))

		tmp := make([][]float64, len(args[0].Values)) // number of points
		lengths := make([]int, len(args[0].Values))   // number of points with data
		atLeastOne := make([]bool, len(args[0].Values))
		for _, arg := range args {
			for i, v := range arg.Values {
				if math.IsNaN(arg.Values[i]) {
					continue
				}
				atLeastOne[i] = true
				tmp[i] = append(tmp[i], v)
				lengths[i]++
			}
		}

		totalSum := 0.0
		totalNotAbsent := 0
		totalCnt := len(r.Values)

		for i, v := range atLeastOne {
			if v {
				r.Values[i] = consolidations.Percentile(tmp[i][0:lengths[i]], 50, true)
				totalSum += r.Values[i]
				totalNotAbsent++
				if isAberration {
					if math.IsNaN(current[name].Values[i]) {
						r.Values[i] = math.NaN()
					} else if r.Values[i] != 0 {
						r.Values[i] = current[name].Values[i] / r.Values[i]
					}
				}
			} else {
				r.Values[i] = math.NaN()
			}
		}

		if !math.IsNaN(maxAbsentPercent) {
			absentPercent := float64(100*(totalCnt-totalNotAbsent)) / float64(totalCnt)
			if absentPercent > maxAbsentPercent {
				continue
			}
		}

		if !math.IsNaN(minAvgLimit) && (totalNotAbsent != 0) {
			avg := totalSum / float64(totalNotAbsent)
			if avg < minAvgLimit {
				continue
			}
		}

		results = append(results, r)
	}

	return results, nil
}

const baselineDescription = `Produce a baseline for the seriesList. Arguments are similar to timestack function.

For each series takes an array of shifted points and computes a median for that.

Example:
.. code-block:: none

  baseline(metric, "1w", 1, 4)

This would take 4 points of a metric, with 1-week interval and for each point will compute a median.

Optional arguments:
  * maxAbsentPercent - do not compute a baseline is percentage of absent points is higher than this value.
  * minAvg - do not compute a baseline if average is lower than this value
`

const baselineAberrationDescription = `Deviation from baseline, in fractions. E.x. if value is over baseline by 10% result will be 1.1`

func (f *baselines) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"baseline": {
			Description: baselineDescription,
			Function:    "baseline(seriesList, timeShiftUnit, timeShiftStart, timeShiftEnd, [maxAbsentPercent, minAvg])",
			Group:       "Calculate",
			Module:      "graphite.render.functions",
			Name:        "baseline",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Default: types.NewSuggestion("1d"),
					Name:    "timeShiftUnit",
					Suggestions: types.NewSuggestions(
						"1h",
						"6h",
						"12h",
						"1d",
						"2d",
						"7d",
						"14d",
						"30d",
					),
					Type: types.Interval,
				},
				{
					Default: types.NewSuggestion(0),
					Name:    "timeShiftStart",
					Type:    types.Integer,
				},
				{
					Default: types.NewSuggestion(7),
					Name:    "timeShiftEnd",
					Type:    types.Integer,
				},
			},
			SeriesChange: true, // function aggregate metrics or change series items count
			NameChange:   true, // name changed
			TagsChange:   true, // name tag changed
			ValuesChange: true, // values changed
		},
		"baselineAberration": {
			Description: baselineAberrationDescription,
			Function:    "baselineAberration(seriesList, timeShiftUnit, timeShiftStart, timeShiftEnd, [maxAbsentPercent, minAvg])",
			Group:       "Calculate",
			Module:      "graphite.render.functions",
			Name:        "baselineAberration",
			Params: []types.FunctionParam{
				{
					Default: types.NewSuggestion("1d"),
					Name:    "timeShiftUnit",
					Suggestions: types.NewSuggestions(
						"1h",
						"6h",
						"12h",
						"1d",
						"2d",
						"7d",
						"14d",
						"30d",
					),
					Type: types.Interval,
				},
				{
					Default: types.NewSuggestion(0),
					Name:    "timeShiftStart",
					Type:    types.Integer,
				},
				{
					Default: types.NewSuggestion(7),
					Name:    "timeShiftEnd",
					Type:    types.Integer,
				},
				{
					Default: types.NewSuggestion(0.0),
					Name:    "maxAbsentPercent",
					Type:    types.Float,
				},
				{
					Default: types.NewSuggestion(0.0),
					Name:    "minAvg",
					Type:    types.Float,
				},
			},
			SeriesChange: true, // function aggregate metrics or change series items count
			NameChange:   true, // name changed
			TagsChange:   true, // name tag changed
			ValuesChange: true, // values changed
		},
	}
}
