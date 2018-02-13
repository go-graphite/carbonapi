package asPercent

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"
)

func init() {
	metadata.RegisterFunction("asPercent", &AsPercent{})
}

type AsPercent struct {
	interfaces.FunctionBase
}

// asPercent(seriesList, total=None, *nodes)
func (f *AsPercent) Do(e parser.Expr, from, until int32, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
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
		getTotal = func(i int) float64 {
			var t float64
			var atLeastOne bool
			for _, a := range arg {
				if a.IsAbsent[i] {
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
		if len(total) == 1 {
			getTotal = func(i int) float64 {
				if total[0].IsAbsent[i] {
					return math.NaN()
				}
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

			result, err := f.Evaluator.EvalExpr(parser.NewExprTyped("sumSeries", seriesNameExprs), from, until, values)

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
				result.IsAbsent = make([]bool, len(totalSeries.Values))
				for i := range result.Values {
					result.Values[i] = 0
					result.IsAbsent[i] = true
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
					result.IsAbsent = make([]bool, len(metaSeries.Values))
					for i := range result.Values {
						result.Values[i] = 0
						result.IsAbsent[i] = true
					}
				} else {
					result.Name = fmt.Sprintf("asPercent(%s,%s)", metaSeries.Name, totalSeries.Name)
					result.Values = make([]float64, len(metaSeries.Values))
					result.IsAbsent = make([]bool, len(metaSeries.Values))
					for i := range metaSeries.Values {
						if metaSeries.IsAbsent[i] || totalSeries.IsAbsent[i] {
							result.Values[i] = 0
							result.IsAbsent[i] = true
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
		fmt.Printf("%v %v\n", len(e.Args()), len(arg))
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
			r.IsAbsent = make([]bool, len(a.Values))
			for k := range a.Values {
				if a.IsAbsent[k] || b.IsAbsent[k] {
					r.Values[k] = 0
					r.IsAbsent[k] = true
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
			r.IsAbsent = make([]bool, len(a.Values))
			results = append(results, &r)
		}

		for i := range results[0].Values {

			total := getTotal(i)

			for j := range results {
				r := results[j]
				a := arg[j]

				if a.IsAbsent[i] || math.IsNaN(total) || total == 0 {
					r.Values[i] = 0
					r.IsAbsent[i] = true
					continue
				}

				r.Values[i] = (a.Values[i] / total) * 100
			}
		}
	}
	return results, nil

}
