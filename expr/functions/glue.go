package functions

import (
	"sort"

	"github.com/go-graphite/carbonapi/expr/functions/absolute"
	"github.com/go-graphite/carbonapi/expr/functions/alias"
	"github.com/go-graphite/carbonapi/expr/functions/aliasByMetric"
	"github.com/go-graphite/carbonapi/expr/functions/aliasByNode"
	"github.com/go-graphite/carbonapi/expr/functions/aliasSub"
	"github.com/go-graphite/carbonapi/expr/functions/asPercent"
	"github.com/go-graphite/carbonapi/expr/functions/averageSeries"
	"github.com/go-graphite/carbonapi/expr/functions/averageSeriesWithWildcards"
	"github.com/go-graphite/carbonapi/expr/functions/below"
	"github.com/go-graphite/carbonapi/expr/functions/cactiStyle"
	"github.com/go-graphite/carbonapi/expr/functions/changed"
	"github.com/go-graphite/carbonapi/expr/functions/consolidateBy"
	"github.com/go-graphite/carbonapi/expr/functions/constantLine"
	"github.com/go-graphite/carbonapi/expr/functions/countSeries"
	"github.com/go-graphite/carbonapi/expr/functions/cumulative"
	"github.com/go-graphite/carbonapi/expr/functions/delay"
	"github.com/go-graphite/carbonapi/expr/functions/derivative"
	"github.com/go-graphite/carbonapi/expr/functions/diffSeries"
	"github.com/go-graphite/carbonapi/expr/functions/divideSeries"
	"github.com/go-graphite/carbonapi/expr/functions/ewma"
	"github.com/go-graphite/carbonapi/expr/functions/exclude"
	"github.com/go-graphite/carbonapi/expr/functions/fallbackSeries"
	"github.com/go-graphite/carbonapi/expr/functions/fft"
	"github.com/go-graphite/carbonapi/expr/functions/grep"
	"github.com/go-graphite/carbonapi/expr/functions/group"
	"github.com/go-graphite/carbonapi/expr/functions/groupByNode"
	"github.com/go-graphite/carbonapi/expr/functions/highest"
	"github.com/go-graphite/carbonapi/expr/functions/hitcount"
	"github.com/go-graphite/carbonapi/expr/functions/holtWintersAberration"
	"github.com/go-graphite/carbonapi/expr/functions/holtWintersConfidenceBands"
	"github.com/go-graphite/carbonapi/expr/functions/holtWintersForecast"
	"github.com/go-graphite/carbonapi/expr/functions/ifft"
	"github.com/go-graphite/carbonapi/expr/functions/integral"
	"github.com/go-graphite/carbonapi/expr/functions/invert"
	"github.com/go-graphite/carbonapi/expr/functions/isNotNull"
	"github.com/go-graphite/carbonapi/expr/functions/keepLastValue"
	"github.com/go-graphite/carbonapi/expr/functions/kolmogorovSmirnovTest2"
	"github.com/go-graphite/carbonapi/expr/functions/legendValue"
	"github.com/go-graphite/carbonapi/expr/functions/limit"
	"github.com/go-graphite/carbonapi/expr/functions/linearRegression"
	"github.com/go-graphite/carbonapi/expr/functions/logarithm"
	"github.com/go-graphite/carbonapi/expr/functions/lowPass"
	"github.com/go-graphite/carbonapi/expr/functions/lowest"
	"github.com/go-graphite/carbonapi/expr/functions/mapSeries"
	"github.com/go-graphite/carbonapi/expr/functions/minMax"
	"github.com/go-graphite/carbonapi/expr/functions/mostDeviant"
	"github.com/go-graphite/carbonapi/expr/functions/moving"
	"github.com/go-graphite/carbonapi/expr/functions/movingMedian"
	"github.com/go-graphite/carbonapi/expr/functions/multiplySeries"
	"github.com/go-graphite/carbonapi/expr/functions/multiplySeriesWithWildcards"
	"github.com/go-graphite/carbonapi/expr/functions/nPercentile"
	"github.com/go-graphite/carbonapi/expr/functions/nonNegativeDerivative"
	"github.com/go-graphite/carbonapi/expr/functions/offset"
	"github.com/go-graphite/carbonapi/expr/functions/offsetToZero"
	"github.com/go-graphite/carbonapi/expr/functions/pearson"
	"github.com/go-graphite/carbonapi/expr/functions/pearsonClosest"
	"github.com/go-graphite/carbonapi/expr/functions/perSecond"
	"github.com/go-graphite/carbonapi/expr/functions/percentileOfSeries"
	"github.com/go-graphite/carbonapi/expr/functions/polyfit"
	"github.com/go-graphite/carbonapi/expr/functions/pow"
	"github.com/go-graphite/carbonapi/expr/functions/randomWalk"
	"github.com/go-graphite/carbonapi/expr/functions/rangeOfSeries"
	"github.com/go-graphite/carbonapi/expr/functions/reduce"
	"github.com/go-graphite/carbonapi/expr/functions/removeBelowSeries"
	"github.com/go-graphite/carbonapi/expr/functions/removeEmptySeries"
	"github.com/go-graphite/carbonapi/expr/functions/scale"
	"github.com/go-graphite/carbonapi/expr/functions/scaleToSeconds"
	"github.com/go-graphite/carbonapi/expr/functions/seriesList"
	"github.com/go-graphite/carbonapi/expr/functions/sortBy"
	"github.com/go-graphite/carbonapi/expr/functions/sortByName"
	"github.com/go-graphite/carbonapi/expr/functions/squareRoot"
	"github.com/go-graphite/carbonapi/expr/functions/stddevSeries"
	"github.com/go-graphite/carbonapi/expr/functions/stdev"
	"github.com/go-graphite/carbonapi/expr/functions/substr"
	"github.com/go-graphite/carbonapi/expr/functions/sum"
	"github.com/go-graphite/carbonapi/expr/functions/sumSeriesWithWildcards"
	"github.com/go-graphite/carbonapi/expr/functions/summarize"
	"github.com/go-graphite/carbonapi/expr/functions/timeFunction"
	"github.com/go-graphite/carbonapi/expr/functions/timeShift"
	"github.com/go-graphite/carbonapi/expr/functions/timeStack"
	"github.com/go-graphite/carbonapi/expr/functions/transformNull"
	"github.com/go-graphite/carbonapi/expr/functions/tukey"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/metadata"
)

type initFunc struct {
	name  string
	order interfaces.Order
	f     func(configFile string) []interfaces.FunctionMetadata
}

func New(configs map[string]string) {
	funcs := make([]initFunc, 0, 83)

	funcs = append(funcs, initFunc{name: "absolute", order: absolute.GetOrder(), f: absolute.New})

	funcs = append(funcs, initFunc{name: "alias", order: alias.GetOrder(), f: alias.New})

	funcs = append(funcs, initFunc{name: "aliasByMetric", order: aliasByMetric.GetOrder(), f: aliasByMetric.New})

	funcs = append(funcs, initFunc{name: "aliasByNode", order: aliasByNode.GetOrder(), f: aliasByNode.New})

	funcs = append(funcs, initFunc{name: "aliasSub", order: aliasSub.GetOrder(), f: aliasSub.New})

	funcs = append(funcs, initFunc{name: "asPercent", order: asPercent.GetOrder(), f: asPercent.New})

	funcs = append(funcs, initFunc{name: "averageSeries", order: averageSeries.GetOrder(), f: averageSeries.New})

	funcs = append(funcs, initFunc{name: "averageSeriesWithWildcards", order: averageSeriesWithWildcards.GetOrder(), f: averageSeriesWithWildcards.New})

	funcs = append(funcs, initFunc{name: "below", order: below.GetOrder(), f: below.New})

	funcs = append(funcs, initFunc{name: "cactiStyle", order: cactiStyle.GetOrder(), f: cactiStyle.New})

	funcs = append(funcs, initFunc{name: "changed", order: changed.GetOrder(), f: changed.New})

	funcs = append(funcs, initFunc{name: "consolidateBy", order: consolidateBy.GetOrder(), f: consolidateBy.New})

	funcs = append(funcs, initFunc{name: "constantLine", order: constantLine.GetOrder(), f: constantLine.New})

	funcs = append(funcs, initFunc{name: "countSeries", order: countSeries.GetOrder(), f: countSeries.New})

	funcs = append(funcs, initFunc{name: "cumulative", order: cumulative.GetOrder(), f: cumulative.New})

	funcs = append(funcs, initFunc{name: "delay", order: delay.GetOrder(), f: delay.New})

	funcs = append(funcs, initFunc{name: "derivative", order: derivative.GetOrder(), f: derivative.New})

	funcs = append(funcs, initFunc{name: "diffSeries", order: diffSeries.GetOrder(), f: diffSeries.New})

	funcs = append(funcs, initFunc{name: "divideSeries", order: divideSeries.GetOrder(), f: divideSeries.New})

	funcs = append(funcs, initFunc{name: "ewma", order: ewma.GetOrder(), f: ewma.New})

	funcs = append(funcs, initFunc{name: "exclude", order: exclude.GetOrder(), f: exclude.New})

	funcs = append(funcs, initFunc{name: "fallbackSeries", order: fallbackSeries.GetOrder(), f: fallbackSeries.New})

	funcs = append(funcs, initFunc{name: "fft", order: fft.GetOrder(), f: fft.New})

	funcs = append(funcs, initFunc{name: "graphiteWeb", order: graphiteWeb.GetOrder(), f: graphiteWeb.New})

	funcs = append(funcs, initFunc{name: "grep", order: grep.GetOrder(), f: grep.New})

	funcs = append(funcs, initFunc{name: "group", order: group.GetOrder(), f: group.New})

	funcs = append(funcs, initFunc{name: "groupByNode", order: groupByNode.GetOrder(), f: groupByNode.New})

	funcs = append(funcs, initFunc{name: "highest", order: highest.GetOrder(), f: highest.New})

	funcs = append(funcs, initFunc{name: "hitcount", order: hitcount.GetOrder(), f: hitcount.New})

	funcs = append(funcs, initFunc{name: "holtWintersAberration", order: holtWintersAberration.GetOrder(), f: holtWintersAberration.New})

	funcs = append(funcs, initFunc{name: "holtWintersConfidenceBands", order: holtWintersConfidenceBands.GetOrder(), f: holtWintersConfidenceBands.New})

	funcs = append(funcs, initFunc{name: "holtWintersForecast", order: holtWintersForecast.GetOrder(), f: holtWintersForecast.New})

	funcs = append(funcs, initFunc{name: "ifft", order: ifft.GetOrder(), f: ifft.New})

	funcs = append(funcs, initFunc{name: "integral", order: integral.GetOrder(), f: integral.New})

	funcs = append(funcs, initFunc{name: "invert", order: invert.GetOrder(), f: invert.New})

	funcs = append(funcs, initFunc{name: "isNotNull", order: isNotNull.GetOrder(), f: isNotNull.New})

	funcs = append(funcs, initFunc{name: "keepLastValue", order: keepLastValue.GetOrder(), f: keepLastValue.New})

	funcs = append(funcs, initFunc{name: "kolmogorovSmirnovTest2", order: kolmogorovSmirnovTest2.GetOrder(), f: kolmogorovSmirnovTest2.New})

	funcs = append(funcs, initFunc{name: "legendValue", order: legendValue.GetOrder(), f: legendValue.New})

	funcs = append(funcs, initFunc{name: "limit", order: limit.GetOrder(), f: limit.New})

	funcs = append(funcs, initFunc{name: "linearRegression", order: linearRegression.GetOrder(), f: linearRegression.New})

	funcs = append(funcs, initFunc{name: "logarithm", order: logarithm.GetOrder(), f: logarithm.New})

	funcs = append(funcs, initFunc{name: "lowPass", order: lowPass.GetOrder(), f: lowPass.New})

	funcs = append(funcs, initFunc{name: "lowest", order: lowest.GetOrder(), f: lowest.New})

	funcs = append(funcs, initFunc{name: "mapSeries", order: mapSeries.GetOrder(), f: mapSeries.New})

	funcs = append(funcs, initFunc{name: "minMax", order: minMax.GetOrder(), f: minMax.New})

	funcs = append(funcs, initFunc{name: "mostDeviant", order: mostDeviant.GetOrder(), f: mostDeviant.New})

	funcs = append(funcs, initFunc{name: "moving", order: moving.GetOrder(), f: moving.New})

	funcs = append(funcs, initFunc{name: "movingMedian", order: movingMedian.GetOrder(), f: movingMedian.New})

	funcs = append(funcs, initFunc{name: "multiplySeries", order: multiplySeries.GetOrder(), f: multiplySeries.New})

	funcs = append(funcs, initFunc{name: "multiplySeriesWithWildcards", order: multiplySeriesWithWildcards.GetOrder(), f: multiplySeriesWithWildcards.New})

	funcs = append(funcs, initFunc{name: "nPercentile", order: nPercentile.GetOrder(), f: nPercentile.New})

	funcs = append(funcs, initFunc{name: "nonNegativeDerivative", order: nonNegativeDerivative.GetOrder(), f: nonNegativeDerivative.New})

	funcs = append(funcs, initFunc{name: "offset", order: offset.GetOrder(), f: offset.New})

	funcs = append(funcs, initFunc{name: "offsetToZero", order: offsetToZero.GetOrder(), f: offsetToZero.New})

	funcs = append(funcs, initFunc{name: "pearson", order: pearson.GetOrder(), f: pearson.New})

	funcs = append(funcs, initFunc{name: "pearsonClosest", order: pearsonClosest.GetOrder(), f: pearsonClosest.New})

	funcs = append(funcs, initFunc{name: "perSecond", order: perSecond.GetOrder(), f: perSecond.New})

	funcs = append(funcs, initFunc{name: "percentileOfSeries", order: percentileOfSeries.GetOrder(), f: percentileOfSeries.New})

	funcs = append(funcs, initFunc{name: "polyfit", order: polyfit.GetOrder(), f: polyfit.New})

	funcs = append(funcs, initFunc{name: "pow", order: pow.GetOrder(), f: pow.New})

	funcs = append(funcs, initFunc{name: "randomWalk", order: randomWalk.GetOrder(), f: randomWalk.New})

	funcs = append(funcs, initFunc{name: "rangeOfSeries", order: rangeOfSeries.GetOrder(), f: rangeOfSeries.New})

	funcs = append(funcs, initFunc{name: "reduce", order: reduce.GetOrder(), f: reduce.New})

	funcs = append(funcs, initFunc{name: "removeBelowSeries", order: removeBelowSeries.GetOrder(), f: removeBelowSeries.New})

	funcs = append(funcs, initFunc{name: "removeEmptySeries", order: removeEmptySeries.GetOrder(), f: removeEmptySeries.New})

	funcs = append(funcs, initFunc{name: "scale", order: scale.GetOrder(), f: scale.New})

	funcs = append(funcs, initFunc{name: "scaleToSeconds", order: scaleToSeconds.GetOrder(), f: scaleToSeconds.New})

	funcs = append(funcs, initFunc{name: "seriesList", order: seriesList.GetOrder(), f: seriesList.New})

	funcs = append(funcs, initFunc{name: "sortBy", order: sortBy.GetOrder(), f: sortBy.New})

	funcs = append(funcs, initFunc{name: "sortByName", order: sortByName.GetOrder(), f: sortByName.New})

	funcs = append(funcs, initFunc{name: "squareRoot", order: squareRoot.GetOrder(), f: squareRoot.New})

	funcs = append(funcs, initFunc{name: "stddevSeries", order: stddevSeries.GetOrder(), f: stddevSeries.New})

	funcs = append(funcs, initFunc{name: "stdev", order: stdev.GetOrder(), f: stdev.New})

	funcs = append(funcs, initFunc{name: "substr", order: substr.GetOrder(), f: substr.New})

	funcs = append(funcs, initFunc{name: "sum", order: sum.GetOrder(), f: sum.New})

	funcs = append(funcs, initFunc{name: "sumSeriesWithWildcards", order: sumSeriesWithWildcards.GetOrder(), f: sumSeriesWithWildcards.New})

	funcs = append(funcs, initFunc{name: "summarize", order: summarize.GetOrder(), f: summarize.New})

	funcs = append(funcs, initFunc{name: "timeFunction", order: timeFunction.GetOrder(), f: timeFunction.New})

	funcs = append(funcs, initFunc{name: "timeShift", order: timeShift.GetOrder(), f: timeShift.New})

	funcs = append(funcs, initFunc{name: "timeStack", order: timeStack.GetOrder(), f: timeStack.New})

	funcs = append(funcs, initFunc{name: "transformNull", order: transformNull.GetOrder(), f: transformNull.New})

	funcs = append(funcs, initFunc{name: "tukey", order: tukey.GetOrder(), f: tukey.New})

	sort.Slice(funcs, func(i, j int) bool {
		if funcs[i].order == interfaces.Any && funcs[j].order == interfaces.Last {
			return true
		}
		if funcs[i].order == interfaces.Last && funcs[j].order == interfaces.Any {
			return false
		}
		return funcs[i].name > funcs[j].name
	})

	for _, f := range funcs {
		md := f.f(configs[f.name])
		for _, m := range md {
			metadata.RegisterFunction(m.Name, m.F)
		}
	}
}
