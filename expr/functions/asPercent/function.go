package asPercent

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"

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
	for _, n := range []string{"asPercent"} {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

// asPercent(seriesList, total=None, *nodes)
func (f *asPercent) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	arg, err := helper.GetSeriesArg(e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	var getTotal func(i int) float64
	var formatName func(a, b string) string
	var totalString string
	var multipleSeries bool
	var numerators []*types.MetricData
	var denominators []*types.MetricData

	var results []*types.MetricData

	if len(e.Args()) == 1 {
		arg = helper.AlignSeries(types.CopyMetricDataSlice(arg))
		getTotal = func(i int) float64 {
			var t float64
			var atLeastOne bool
			for _, a := range arg {
				if math.IsNaN(a.Values[i]) {
					continue
				}
				atLeastOne = true
				t += a.Values[i]
			}
			if !atLeastOne {
				t = math.NaN()
			}

			return t
		}
		formatName = func(a, b string) string {
			return fmt.Sprintf("asPercent(%s)", a)
		}
	} else if len(e.Args()) == 2 && e.Args()[1].IsConst() {
		total, err := e.GetFloatArg(1)
		if err != nil {
			return nil, err
		}
		arg = helper.AlignSeries(types.CopyMetricDataSlice(arg))
		getTotal = func(i int) float64 { return total }
		totalString = fmt.Sprintf("%g", total)
		formatName = func(a, b string) string {
			return fmt.Sprintf("asPercent(%s,%s)", a, b)
		}
	} else if len(e.Args()) == 2 && (e.Args()[1].IsName() || e.Args()[1].IsFunc()) {
		total, err := helper.GetSeriesArg(e.Args()[1], from, until, values)
		if err != nil {
			return nil, err
		}

		if len(total) != 1 && len(total) != len(arg) {
			return nil, types.ErrWildcardNotAllowed
		}

		alignedSeries := helper.AlignSeries(types.CopyMetricDataSlice(append(arg, total...)))
		arg = alignedSeries[0:len(arg)]
		total = alignedSeries[len(arg):]

		if len(total) == 1 {
			getTotal = func(i int) float64 {
				return total[0].Values[i]
			}
			if e.Args()[1].IsName() {
				totalString = e.Args()[1].Target()
			} else {
				totalString = fmt.Sprintf("%s(%s)", e.Args()[1].Target(), e.Args()[1].RawArgs())
			}
		} else {
			multipleSeries = true
			numerators = arg
			denominators = total
			// Sort lists by name so that they match up.
			sort.Sort(helper.ByName(numerators))
			sort.Sort(helper.ByName(denominators))
		}
		formatName = func(a, b string) string {
			return fmt.Sprintf("asPercent(%s,%s)", a, b)
		}
	} else if len(e.Args()) >= 3 {
		total, err := helper.GetSeriesArg(e.Args()[1], from, until, values)
		if err != nil {
			return nil, err
		}

		if len(total) != 1 && len(total) != len(arg) {
			return nil, types.ErrWildcardNotAllowed
		}

		alignedSeries := helper.AlignSeries(types.CopyMetricDataSlice(append(arg, total...)))
		arg = alignedSeries[0:len(arg)]
		total = alignedSeries[len(arg):]

		nodeIndexes, err := e.GetIntArgs(2)
		if err != nil {
			return nil, err
		}

		sumSeries := func(seriesList []*types.MetricData) (*types.MetricData, error) {
			seriesNames := make([]string, len(seriesList))
			for i, series := range seriesList {
				seriesNames[i] = series.Name
			}

			seriesNameExprs := make([]parser.Expr, len(seriesList))
			for i, seriesName := range seriesNames {
				seriesNameExprs[i] = parser.NewTargetExpr(seriesName)
			}

			result, err := f.Evaluator.Eval(ctx, parser.NewExprTyped("sumSeries", seriesNameExprs), from, until, values)

			if err != nil {
				return nil, err
			}

			// sumSeries uses aggregateSeries function which returns only one series
			return result[0], nil
		}

		groupByNodes := func(seriesList []*types.MetricData, nodeIndexes []int) (map[string][]*types.MetricData, []string) {
			var nodeKeys []string
			groups := make(map[string][]*types.MetricData)

			for _, series := range seriesList {
				metric := helper.ExtractMetric(series.Name)
				nodes := strings.Split(metric, ".")
				nodeKey := make([]string, 0, len(nodeIndexes))
				for _, index := range nodeIndexes {
					nodeKey = append(nodeKey, nodes[index])
				}
				node := strings.Join(nodeKey, ".")
				_, exist := groups[node]

				if !exist {
					nodeKeys = append(nodeKeys, node)
				}

				groups[node] = append(groups[node], series)
			}

			return groups, nodeKeys
		}

		distinct := func(slice []string) []string {
			keys := make(map[string]bool)
			var list []string
			for _, entry := range slice {
				if _, value := keys[entry]; !value {
					keys[entry] = true
					list = append(list, entry)
				}
			}
			return list
		}

		metaSeriesGroup, metaKeys := groupByNodes(arg, nodeIndexes)

		totalSeriesGroup := make(map[string]*types.MetricData)
		var groups map[string][]*types.MetricData
		var groupKeys []string

		if len(total) == 0 {
			groups, groupKeys = metaSeriesGroup, metaKeys
		} else {
			groups, groupKeys = groupByNodes(total, nodeIndexes)
		}

		for _, nodeKey := range groupKeys {
			if len(groups[nodeKey]) == 1 {
				totalSeriesGroup[nodeKey] = groups[nodeKey][0]
			} else {
				totalSeriesGroup[nodeKey], err = sumSeries(groups[nodeKey])
				if err != nil {
					return nil, err
				}
			}
		}

		nodeKeys := distinct(append(metaKeys, groupKeys...))

		for _, nodeKey := range nodeKeys {
			metaSeriesList, existInMeta := metaSeriesGroup[nodeKey]
			if !existInMeta {
				totalSeries := totalSeriesGroup[nodeKey]
				result := *totalSeries
				result.Name = fmt.Sprintf("asPercent(MISSING,%s)", totalSeries.Name)
				result.Values = make([]float64, len(totalSeries.Values))
				for i := range result.Values {
					result.Values[i] = math.NaN()
				}

				results = append(results, &result)
				continue
			}

			for _, metaSeries := range metaSeriesList {
				result := *metaSeries
				totalSeries, existInTotal := totalSeriesGroup[nodeKey]
				if !existInTotal {
					result.Name = fmt.Sprintf("asPercent(%s,MISSING)", metaSeries.Name)
					result.Values = make([]float64, len(metaSeries.Values))
					for i := range result.Values {
						result.Values[i] = math.NaN()
					}
				} else {
					result.Name = fmt.Sprintf("asPercent(%s,%s)", metaSeries.Name, totalSeries.Name)
					result.Values = make([]float64, len(metaSeries.Values))
					for i := range metaSeries.Values {
						if math.IsNaN(metaSeries.Values[i]) || math.IsNaN(totalSeries.Values[i]) {
							result.Values[i] = math.NaN()
							continue
						}
						result.Values[i] = (metaSeries.Values[i] / totalSeries.Values[i]) * 100
					}
				}

				results = append(results, &result)
			}
		}

		return results, nil

	} else {
		return nil, errors.New("total must be either a constant or a series")
	}

	if multipleSeries {
		/* We should have two equal length lists of arguments
		   First one will be numerators
		   Second one - denominators

		   For each of them we will compute numerator/denominator
		*/
		for i := range numerators {
			a := numerators[i]
			b := denominators[i]

			r := *a
			r.Name = formatName(a.Name, b.Name)
			r.Values = make([]float64, len(a.Values))
			for k := range a.Values {
				if math.IsNaN(a.Values[k]) || math.IsNaN(b.Values[k]) {
					r.Values[k] = math.NaN()
					continue
				}
				r.Values[k] = (a.Values[k] / b.Values[k]) * 100
			}
			results = append(results, &r)
		}
	} else {
		for _, a := range arg {
			r := *a
			r.Name = formatName(a.Name, totalString)
			r.Values = make([]float64, len(a.Values))
			results = append(results, &r)
		}

		for i := range results[0].Values {

			total := getTotal(i)

			for j := range results {
				r := results[j]
				a := arg[j]

				if math.IsNaN(a.Values[i]) || math.IsNaN(total) || total == 0 {
					r.Values[i] = math.NaN()
					continue
				}

				r.Values[i] = (a.Values[i] / total) * 100
			}
		}
	}
	return results, nil
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
		},
	}
}
