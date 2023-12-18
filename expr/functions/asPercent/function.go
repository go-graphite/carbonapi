package asPercent

import (
	"context"
	"errors"
	"math"
	"sort"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

type asPercent struct {
	interfaces.FunctionBase
}

func GetOrder() interfaces.Order {
	return interfaces.Any
}

func New(configFile string) []interfaces.FunctionMetadata {
	res := make([]interfaces.FunctionMetadata, 0)
	f := &asPercent{}
	for _, n := range []string{"asPercent", "pct"} {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

func getTotal(arg []*types.MetricData, i int) float64 {
	var t float64
	var atLeastOne bool
	for _, a := range arg {
		if math.IsNaN(a.Values[i]) {
			continue
		}
		atLeastOne = true
		t += a.Values[i]
	}
	if atLeastOne {
		return t
	}
	return math.NaN()
}

// sum aligned series
func sumSeries(seriesList []*types.MetricData) *types.MetricData {
	result := &types.MetricData{}
	result.Values = make([]float64, len(seriesList[0].Values))

	for _, s := range seriesList {
		for i := range result.Values {
			if !math.IsNaN(s.Values[i]) {
				result.Values[i] += s.Values[i]
			}
		}
	}

	return result
}

func calculatePercentage(seriesValue, totalValue float64) float64 {
	var value float64
	if math.IsNaN(seriesValue) || math.IsNaN(totalValue) || totalValue == 0 {
		value = math.NaN()
	} else {
		value = seriesValue * (100 / totalValue)
	}
	return value
}

// GetPercentages contains the logic to apply to the series in order to properly
// calculate percentages. If the length of the values in series and totalSeries are
// not equal, special handling is required. If the number of values in seriesList is
// greater than the number of values in totalSeries, math.NaN() needs to be set to the
// indices in series starting at the length of totalSeries.Values. If the number of values
// in totalSeries is greater than the number of values in series, then math.NaN() needs
// to be appended to series until its values have the same length as totalSeries.Values
func getPercentages(series, totalSeries *types.MetricData) {
	// If there are more series values than totalSeries values, set series value to math.NaN() for those indices
	if len(series.Values) > len(totalSeries.Values) {
		for i := 0; i < len(totalSeries.Values); i++ {
			series.Values[i] = calculatePercentage(series.Values[i], totalSeries.Values[i])
		}
		for i := len(totalSeries.Values); i < len(series.Values); i++ {
			series.Values[i] = math.NaN()
		}
	} else {
		for i := range series.Values {
			series.Values[i] = calculatePercentage(series.Values[i], totalSeries.Values[i])
		}

		// If there are more totalSeries values than series values, append math.NaN() to the series values
		if lengthDiff := len(totalSeries.Values) - len(series.Values); lengthDiff > 0 {
			for i := 0; i < lengthDiff; i++ {
				series.Values = append(series.Values, math.NaN())
			}
		}
	}
}

func groupByNodes(seriesList []*types.MetricData, nodesOrTags []parser.NodeOrTag) map[string][]*types.MetricData {
	groups := make(map[string][]*types.MetricData)

	for _, series := range seriesList {
		key := helper.AggKey(series, nodesOrTags)
		groups[key] = append(groups[key], series)
	}

	return groups
}

func seriesAsPercent(arg, total []*types.MetricData) []*types.MetricData {
	if len(total) == 0 {
		// asPercent(seriesList, MISSING)
		for _, a := range arg {
			for i := range a.Values {
				a.Values[i] = math.NaN()
			}

			a.Name = "asPercent(" + a.Name + ",MISSING)"
		}
	} else if len(total) == 1 {
		// asPercent(seriesList, totalSeries)
		for _, a := range arg {
			getPercentages(a, total[0])

			a.Name = "asPercent(" + a.Name + "," + total[0].Name + ")"
		}
	} else {
		// asPercent(seriesList, totalSeriesList)
		sort.Sort(helper.ByName(arg))
		sort.Sort(helper.ByName(total))
		if len(arg) <= len(total) {
			// asPercent(seriesList, totalSeriesList) for series with len(seriesList) <= len(totalSeriesList)
			for n, a := range arg {
				getPercentages(a, total[n])

				a.Name = "asPercent(" + a.Name + "," + total[n].Name + ")"
			}
			if len(arg) < len(total) {
				total = total[len(arg):]
				for _, tot := range total {
					for i := range tot.Values {
						tot.Values[i] = math.NaN()
					}
					tot.Name = "asPercent(MISSING," + tot.Name + ")"
					tot.Tags = map[string]string{"name": "MISSING"}
				}
				arg = append(arg, total...)
			}
		} else {
			// asPercent(seriesList, totalSeriesList) for series with unaligned length
			// len(seriesList) > len(totalSeriesList)
			for n := range total {
				a := arg[n]
				getPercentages(a, total[n])

				a.Name = "asPercent(" + a.Name + "," + total[n].Name + ")"
			}
			for n := len(total); n < len(arg); n++ {
				a := arg[n]
				for i := range a.Values {
					a.Values[i] = math.NaN()
				}

				a.Name = "asPercent(" + a.Name + ",MISSING)"
			}
		}

	}
	return arg
}

func seriesGroupAsPercent(arg []*types.MetricData, nodesOrTags []parser.NodeOrTag) []*types.MetricData {
	// asPercent(seriesList, None, *nodes)
	argGroups := groupByNodes(arg, nodesOrTags)

	keys := make([]string, len(argGroups))
	i := 0
	for key := range argGroups {
		keys[i] = key
		i++
	}
	sort.Strings(keys)

	arg = make([]*types.MetricData, 0, len(arg))
	for _, k := range keys {
		argGroup := argGroups[k]
		sum := sumSeries(argGroup)
		start := len(arg)
		for _, a := range argGroup {
			for i := range sum.Values {
				if math.IsNaN(sum.Values[i]) || sum.Values[i] == 0 {
					if !math.IsNaN(a.Values[i]) {
						a.Values[i] = math.NaN()
					}
				} else {
					a.Values[i] *= 100 / sum.Values[i]
				}
			}
			a.Name = "asPercent(" + a.Name + ",None)"
			arg = append(arg, a)
		}
		end := len(arg)
		sort.Sort(helper.ByName(arg[start:end]))
	}
	return arg
}

func seriesGroup2AsPercent(arg, total []*types.MetricData, nodesOrTags []parser.NodeOrTag) []*types.MetricData {
	argGroups := groupByNodes(arg, nodesOrTags)
	totalGroups := groupByNodes(total, nodesOrTags)

	keys := make([]string, len(argGroups), len(argGroups)+4)
	i := 0
	for key := range argGroups {
		keys[i] = key
		i++
	}
	for key := range totalGroups {
		if _, exist := argGroups[key]; !exist {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)

	arg = make([]*types.MetricData, 0, len(arg))
	for _, key := range keys {
		if argGroup, exists := argGroups[key]; exists {
			if totalGroup, exist := totalGroups[key]; exist {
				if len(totalGroup) == 1 {
					// asPercent(seriesList, totalSeries, *nodes)
					start := len(arg)
					for _, a := range argGroup {
						getPercentages(a, totalGroup[0])

						a.Name = "asPercent(" + a.Name + "," + totalGroup[0].Name + ")"
						arg = append(arg, a)
					}
					end := len(arg)
					sort.Sort(helper.ByName(arg[start:end]))
				} else if len(argGroup) <= len(totalGroup) {
					// asPercent(seriesList, totalSeriesList, *nodes)
					// len(seriesGroupList) <= len(totalSeriesGroupList)

					start := len(arg)
					for n, a := range argGroup {
						for i := range a.Values {
							t := totalGroup[n].Values[i]
							if math.IsNaN(a.Values[i]) || math.IsNaN(t) || t == 0 {
								a.Values[i] = math.NaN()
							} else {
								a.Values[i] *= 100 / t
							}
						}
						a.Name = "asPercent(" + a.Name + "," + totalGroup[n].Name + ")"
						arg = append(arg, a)
					}
					if len(argGroup) < len(totalGroup) {
						totalGroup = totalGroup[len(argGroup):]
						for _, tot := range totalGroup {
							for i := range tot.Values {
								tot.Values[i] = math.NaN()
							}
							tot.Name = "asPercent(MISSING," + tot.Name + ")"
							tot.Tags = map[string]string{"name": "MISSING"}
						}
						arg = append(arg, totalGroup...)
					}
					end := len(arg)
					sort.Sort(helper.ByName(arg[start:end]))
				} else {
					// asPercent(seriesList, totalSeriesList, *nodes) for series with unaligned length
					// len(seriesGroupList) > len(totalSeriesGroupList)

					start := len(arg)
					for n := range totalGroup {
						a := argGroup[n]
						for i := range a.Values {
							t := total[n].Values[i]
							if math.IsNaN(a.Values[i]) || math.IsNaN(t) || t == 0 {
								a.Values[i] = math.NaN()
							} else {
								a.Values[i] *= 100 / t
							}
						}
						a.Name = "asPercent(" + a.Name + "," + total[n].Name + ")"
					}
					for n := len(total); n < len(arg); n++ {
						a := arg[n]
						for i := range a.Values {
							a.Values[i] = math.NaN()
						}

						a.Name = "asPercent(" + a.Name + ",MISSING)"
						arg = append(arg, a)
					}
					end := len(arg)
					sort.Sort(helper.ByName(arg[start:end]))
				}
			} else {
				start := len(arg)
				for _, a := range argGroup {
					for i := range a.Values {
						a.Values[i] = math.NaN()
					}
					a.Name = "asPercent(" + a.Name + ",MISSING)"
					arg = append(arg, a)
				}
				end := len(arg)
				sort.Sort(helper.ByName(arg[start:end]))
			}
		} else {
			totalGroup := totalGroups[key]
			if _, exist := argGroups[key]; !exist {
				start := len(arg)
				for _, t := range totalGroup {
					for i := range t.Values {
						t.Values[i] = math.NaN()
					}
					t.Name = "asPercent(MISSING," + t.Name + ")"
					t.Tags = map[string]string{"name": "MISSING"}
					arg = append(arg, t)
				}
				end := len(arg)
				sort.Sort(helper.ByName(arg[start:end]))
			}
		}
	}
	return arg
}

// asPercent(seriesList, total=None, *nodes)
func (f *asPercent) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(0), from, until, values)
	if err != nil {
		return nil, err
	}
	if len(arg) == 0 {
		return nil, nil
	}

	if e.ArgsLen() == 1 {
		// asPercent(seriesList)

		// TODO (msaf1980): may be copy in before start eval (based on function pipeline descritptions (ValueChange field)) and avoid copy metrics in functions
		arg = helper.AlignSeries(types.CopyMetricDataSlice(arg))
		for i := range arg[0].Values {
			total := getTotal(arg, i)

			for _, a := range arg {
				if math.IsNaN(a.Values[i]) || math.IsNaN(total) || total == 0 {
					a.Values[i] = math.NaN()
				} else {
					a.Values[i] *= 100 / total
				}
			}
		}

		for _, a := range arg {
			a.Name = "asPercent(" + a.Name + ")"
		}
		return arg, nil
	} else if e.ArgsLen() == 2 && (e.Arg(1).IsConst() || e.Arg(1).IsString()) {
		// asPercent(seriesList, N)

		total, err := e.GetFloatArg(1)

		if err != nil {
			return nil, err
		}

		// TODO (msaf1980): may be copy in before start eval (based on function pipeline descritptions (ValueChange field)) and avoid copy metrics in functions
		arg = helper.AlignSeries(types.CopyMetricDataSlice(arg))

		for _, a := range arg {
			for i := range a.Values {
				if math.IsNaN(a.Values[i]) || math.IsNaN(total) || total == 0 {
					a.Values[i] = math.NaN()
				} else {
					a.Values[i] *= 100 / total
				}
			}
			a.Name = "asPercent(" + a.Name + "," + e.Arg(1).StringValue() + ")"
		}
		return arg, nil
	} else if e.ArgsLen() == 2 && (e.Arg(1).IsName() || e.Arg(1).IsFunc()) {
		// asPercent(seriesList, totalList)
		total, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(1), from, until, values)
		if err != nil {
			return nil, err
		}

		alignedSeries := helper.AlignSeries(types.CopyMetricDataSlice(append(arg, total...)))
		arg = alignedSeries[0:len(arg)]
		total = alignedSeries[len(arg):]

		return seriesAsPercent(arg, total), nil

	} else if e.ArgsLen() >= 3 && e.Arg(1).IsName() || e.Arg(1).IsFunc() {
		// Group by
		nodesOrTags, err := e.GetNodeOrTagArgs(2, false)
		if err != nil {
			return nil, err
		}

		if e.Arg(1).Target() == "None" {
			// asPercent(seriesList, None, *nodes)
			arg = helper.AlignSeries(types.CopyMetricDataSlice(arg))

			return seriesGroupAsPercent(arg, nodesOrTags), nil
		} else {
			// asPercent(seriesList, totalSeriesList, *nodes)
			total, err := helper.GetSeriesArg(ctx, f.GetEvaluator(), e.Arg(1), from, until, values)
			if err != nil {
				return nil, err
			}

			alignedSeries := helper.AlignSeries(types.CopyMetricDataSlice(append(arg, total...)))
			arg = alignedSeries[0:len(arg)]
			total = alignedSeries[len(arg):]

			return seriesGroup2AsPercent(arg, total, nodesOrTags), nil
		}
	}

	return nil, errors.New("total must be either a constant or a series")
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *asPercent) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"asPercent": {
			Description: "Calculates a percentage of the total of a wildcard series. If `total` is specified,\neach series will be calculated as a percentage of that total. If `total` is not specified,\nthe sum of all points in the wildcard series will be used instead.\n\nA list of nodes can optionally be provided, if so they will be used to match series with their\ncorresponding totals following the same logic as :py:func:`groupByNodes <groupByNodes>`.\n\nWhen passing `nodes` the `total` parameter may be a series list or `None`.  If it is `None` then\nfor each series in `seriesList` the percentage of the sum of series in that group will be returned.\n\nWhen not passing `nodes`, the `total` parameter may be a single series, reference the same number\nof series as `seriesList` or be a numeric value.\n\nExample:\n\n.. code-block:: none\n\n  # Server01 connections failed and succeeded as a percentage of Server01 connections attempted\n  &target=asPercent(Server01.connections.{failed,succeeded}, Server01.connections.attempted)\n\n  # For each server, its connections failed as a percentage of its connections attempted\n  &target=asPercent(Server*.connections.failed, Server*.connections.attempted)\n\n  # For each server, its connections failed and succeeded as a percentage of its connections attemped\n  &target=asPercent(Server*.connections.{failed,succeeded}, Server*.connections.attempted, 0)\n\n  # apache01.threads.busy as a percentage of 1500\n  &target=asPercent(apache01.threads.busy,1500)\n\n  # Server01 cpu stats as a percentage of its total\n  &target=asPercent(Server01.cpu.*.jiffies)\n\n  # cpu stats for each server as a percentage of its total\n  &target=asPercent(Server*.cpu.*.jiffies, None, 0)\n\nWhen using `nodes`, any series or totals that can't be matched will create output series with\nnames like ``asPercent(someSeries,MISSING)`` or ``asPercent(MISSING,someTotalSeries)`` and all\nvalues set to None. If desired these series can be filtered out by piping the result through\n``|exclude(\"MISSING\")`` as shown below:\n\n.. code-block:: none\n\n  &target=asPercent(Server{1,2}.memory.used,Server{1,3}.memory.total,0)\n\n  # will produce 3 output series:\n  # asPercent(Server1.memory.used,Server1.memory.total) [values will be as expected}\n  # asPercent(Server2.memory.used,MISSING) [all values will be None}\n  # asPercent(MISSING,Server3.memory.total) [all values will be None}\n\n  &target=asPercent(Server{1,2}.memory.used,Server{1,3}.memory.total,0)|exclude(\"MISSING\")\n\n  # will produce 1 output series:\n  # asPercent(Server1.memory.used,Server1.memory.total) [values will be as expected}\n\nEach node may be an integer referencing a node in the series name or a string identifying a tag.\n\n.. note::\n\n  When `total` is a seriesList, specifying `nodes` to match series with the corresponding total\n  series will increase reliability.",
			Function:    "asPercent(seriesList, total=None, *nodes)",
			Group:       "Combine",
			Module:      "graphite.render.functions",
			Name:        "asPercent",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name: "total",
					Type: types.SeriesList,
				},
				{
					Multiple: true,
					Name:     "nodes",
					Type:     types.NodeOrTag,
				},
			},
			SeriesChange: true, // function aggregate metrics or change series items count
			NameChange:   true, // name changed
			TagsChange:   true, // name tag changed
			ValuesChange: true, // values changed
		},
		"pct": {
			Description: "Calculates a percentage of the total of a wildcard series. If `total` is specified,\neach series will be calculated as a percentage of that total. If `total` is not specified,\nthe sum of all points in the wildcard series will be used instead.\n\nA list of nodes can optionally be provided, if so they will be used to match series with their\ncorresponding totals following the same logic as :py:func:`groupByNodes <groupByNodes>`.\n\nWhen passing `nodes` the `total` parameter may be a series list or `None`.  If it is `None` then\nfor each series in `seriesList` the percentage of the sum of series in that group will be returned.\n\nWhen not passing `nodes`, the `total` parameter may be a single series, reference the same number\nof series as `seriesList` or be a numeric value.\n\nExample:\n\n.. code-block:: none\n\n  # Server01 connections failed and succeeded as a percentage of Server01 connections attempted\n  &target=asPercent(Server01.connections.{failed,succeeded}, Server01.connections.attempted)\n\n  # For each server, its connections failed as a percentage of its connections attempted\n  &target=asPercent(Server*.connections.failed, Server*.connections.attempted)\n\n  # For each server, its connections failed and succeeded as a percentage of its connections attemped\n  &target=asPercent(Server*.connections.{failed,succeeded}, Server*.connections.attempted, 0)\n\n  # apache01.threads.busy as a percentage of 1500\n  &target=asPercent(apache01.threads.busy,1500)\n\n  # Server01 cpu stats as a percentage of its total\n  &target=asPercent(Server01.cpu.*.jiffies)\n\n  # cpu stats for each server as a percentage of its total\n  &target=asPercent(Server*.cpu.*.jiffies, None, 0)\n\nWhen using `nodes`, any series or totals that can't be matched will create output series with\nnames like ``asPercent(someSeries,MISSING)`` or ``asPercent(MISSING,someTotalSeries)`` and all\nvalues set to None. If desired these series can be filtered out by piping the result through\n``|exclude(\"MISSING\")`` as shown below:\n\n.. code-block:: none\n\n  &target=asPercent(Server{1,2}.memory.used,Server{1,3}.memory.total,0)\n\n  # will produce 3 output series:\n  # asPercent(Server1.memory.used,Server1.memory.total) [values will be as expected}\n  # asPercent(Server2.memory.used,MISSING) [all values will be None}\n  # asPercent(MISSING,Server3.memory.total) [all values will be None}\n\n  &target=asPercent(Server{1,2}.memory.used,Server{1,3}.memory.total,0)|exclude(\"MISSING\")\n\n  # will produce 1 output series:\n  # asPercent(Server1.memory.used,Server1.memory.total) [values will be as expected}\n\nEach node may be an integer referencing a node in the series name or a string identifying a tag.\n\n.. note::\n\n  When `total` is a seriesList, specifying `nodes` to match series with the corresponding total\n  series will increase reliability.",
			Function:    "pct(seriesList, total=None, *nodes)",
			Group:       "Combine",
			Module:      "graphite.render.functions",
			Name:        "pct",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name: "total",
					Type: types.SeriesList,
				},
				{
					Multiple: true,
					Name:     "nodes",
					Type:     types.NodeOrTag,
				},
			},
			SeriesChange: true, // function aggregate metrics or change series items count
			NameChange:   true, // name changed
			TagsChange:   true, // name tag changed
			ValuesChange: true, // values changed
		},
	}
}
