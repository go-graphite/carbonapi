package functions

import (
	"sort"
	"strings"

	"github.com/go-graphite/carbonapi/expr/functions/aboveSeries"
	"github.com/go-graphite/carbonapi/expr/functions/absolute"
	"github.com/go-graphite/carbonapi/expr/functions/aggregate"
	"github.com/go-graphite/carbonapi/expr/functions/aggregateLine"
	"github.com/go-graphite/carbonapi/expr/functions/alias"
	"github.com/go-graphite/carbonapi/expr/functions/aliasByMetric"
	"github.com/go-graphite/carbonapi/expr/functions/aliasByNode"
	"github.com/go-graphite/carbonapi/expr/functions/aliasByPostgres"
	"github.com/go-graphite/carbonapi/expr/functions/aliasByTags"
	"github.com/go-graphite/carbonapi/expr/functions/aliasSub"
	"github.com/go-graphite/carbonapi/expr/functions/asPercent"
	"github.com/go-graphite/carbonapi/expr/functions/averageSeries"
	"github.com/go-graphite/carbonapi/expr/functions/averageSeriesWithWildcards"
	"github.com/go-graphite/carbonapi/expr/functions/below"
	"github.com/go-graphite/carbonapi/expr/functions/cactiStyle"
	"github.com/go-graphite/carbonapi/expr/functions/cairo"
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
	"github.com/go-graphite/carbonapi/expr/functions/filter"
	"github.com/go-graphite/carbonapi/expr/functions/graphiteWeb"
	"github.com/go-graphite/carbonapi/expr/functions/grep"
	"github.com/go-graphite/carbonapi/expr/functions/group"
	"github.com/go-graphite/carbonapi/expr/functions/groupByNode"
	"github.com/go-graphite/carbonapi/expr/functions/groupByTags"
	"github.com/go-graphite/carbonapi/expr/functions/highestLowest"
	"github.com/go-graphite/carbonapi/expr/functions/hitcount"
	"github.com/go-graphite/carbonapi/expr/functions/holtWintersAberration"
	"github.com/go-graphite/carbonapi/expr/functions/holtWintersConfidenceBands"
	"github.com/go-graphite/carbonapi/expr/functions/holtWintersForecast"
	"github.com/go-graphite/carbonapi/expr/functions/ifft"
	"github.com/go-graphite/carbonapi/expr/functions/integral"
	"github.com/go-graphite/carbonapi/expr/functions/integralByInterval"
	"github.com/go-graphite/carbonapi/expr/functions/invert"
	"github.com/go-graphite/carbonapi/expr/functions/isNotNull"
	"github.com/go-graphite/carbonapi/expr/functions/keepLastValue"
	"github.com/go-graphite/carbonapi/expr/functions/kolmogorovSmirnovTest2"
	"github.com/go-graphite/carbonapi/expr/functions/legendValue"
	"github.com/go-graphite/carbonapi/expr/functions/limit"
	"github.com/go-graphite/carbonapi/expr/functions/linearRegression"
	"github.com/go-graphite/carbonapi/expr/functions/logarithm"
	"github.com/go-graphite/carbonapi/expr/functions/lowPass"
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
	"github.com/go-graphite/carbonapi/expr/functions/round"
	"github.com/go-graphite/carbonapi/expr/functions/scale"
	"github.com/go-graphite/carbonapi/expr/functions/scaleToSeconds"
	"github.com/go-graphite/carbonapi/expr/functions/seriesByTag"
	"github.com/go-graphite/carbonapi/expr/functions/seriesList"
	"github.com/go-graphite/carbonapi/expr/functions/smartSummarize"
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
	funcs := []initFunc{
		{name: "aboveSeries", order: aboveSeries.GetOrder(), f: aboveSeries.New},
		{name: "absolute", order: absolute.GetOrder(), f: absolute.New},
		{name: "aggregate", order: aggregate.GetOrder(), f: aggregate.New},
		{name: "aggregateLine", order: aggregateLine.GetOrder(), f: aggregateLine.New},
		{name: "alias", order: alias.GetOrder(), f: alias.New},
		{name: "aliasByMetric", order: aliasByMetric.GetOrder(), f: aliasByMetric.New},
		{name: "aliasByNode", order: aliasByNode.GetOrder(), f: aliasByNode.New},
		{name: "aliasByPostgres", order: aliasByPostgres.GetOrder(), f: aliasByPostgres.New},
		{name: "aliasByTags", order: aliasByTags.GetOrder(), f: aliasByTags.New},
		{name: "aliasSub", order: aliasSub.GetOrder(), f: aliasSub.New},
		{name: "asPercent", order: asPercent.GetOrder(), f: asPercent.New},
		{name: "averageSeries", order: averageSeries.GetOrder(), f: averageSeries.New},
		{name: "averageSeriesWithWildcards", order: averageSeriesWithWildcards.GetOrder(), f: averageSeriesWithWildcards.New},
		{name: "below", order: below.GetOrder(), f: below.New},
		{name: "cactiStyle", order: cactiStyle.GetOrder(), f: cactiStyle.New},
		{name: "cairo", order: cairo.GetOrder(), f: cairo.New},
		{name: "changed", order: changed.GetOrder(), f: changed.New},
		{name: "consolidateBy", order: consolidateBy.GetOrder(), f: consolidateBy.New},
		{name: "constantLine", order: constantLine.GetOrder(), f: constantLine.New},
		{name: "countSeries", order: countSeries.GetOrder(), f: countSeries.New},
		{name: "cumulative", order: cumulative.GetOrder(), f: cumulative.New},
		{name: "delay", order: delay.GetOrder(), f: delay.New},
		{name: "derivative", order: derivative.GetOrder(), f: derivative.New},
		{name: "diffSeries", order: diffSeries.GetOrder(), f: diffSeries.New},
		{name: "divideSeries", order: divideSeries.GetOrder(), f: divideSeries.New},
		{name: "ewma", order: ewma.GetOrder(), f: ewma.New},
		{name: "exclude", order: exclude.GetOrder(), f: exclude.New},
		{name: "fallbackSeries", order: fallbackSeries.GetOrder(), f: fallbackSeries.New},
		{name: "fft", order: fft.GetOrder(), f: fft.New},
		{name: "filter", order: filter.GetOrder(), f: filter.New},
		{name: "graphiteWeb", order: graphiteWeb.GetOrder(), f: graphiteWeb.New},
		{name: "grep", order: grep.GetOrder(), f: grep.New},
		{name: "group", order: group.GetOrder(), f: group.New},
		{name: "groupByNode", order: groupByNode.GetOrder(), f: groupByNode.New},
		{name: "groupByTags", order: groupByTags.GetOrder(), f: groupByTags.New},
		{name: "highestLowest", order: highestLowest.GetOrder(), f: highestLowest.New},
		{name: "hitcount", order: hitcount.GetOrder(), f: hitcount.New},
		{name: "holtWintersAberration", order: holtWintersAberration.GetOrder(), f: holtWintersAberration.New},
		{name: "holtWintersConfidenceBands", order: holtWintersConfidenceBands.GetOrder(), f: holtWintersConfidenceBands.New},
		{name: "holtWintersForecast", order: holtWintersForecast.GetOrder(), f: holtWintersForecast.New},
		{name: "ifft", order: ifft.GetOrder(), f: ifft.New},
		{name: "integral", order: integral.GetOrder(), f: integral.New},
		{name: "integralByInterval", order: integralByInterval.GetOrder(), f: integralByInterval.New},
		{name: "invert", order: invert.GetOrder(), f: invert.New},
		{name: "isNotNull", order: isNotNull.GetOrder(), f: isNotNull.New},
		{name: "keepLastValue", order: keepLastValue.GetOrder(), f: keepLastValue.New},
		{name: "kolmogorovSmirnovTest2", order: kolmogorovSmirnovTest2.GetOrder(), f: kolmogorovSmirnovTest2.New},
		{name: "legendValue", order: legendValue.GetOrder(), f: legendValue.New},
		{name: "limit", order: limit.GetOrder(), f: limit.New},
		{name: "linearRegression", order: linearRegression.GetOrder(), f: linearRegression.New},
		{name: "logarithm", order: logarithm.GetOrder(), f: logarithm.New},
		{name: "lowPass", order: lowPass.GetOrder(), f: lowPass.New},
		{name: "mapSeries", order: mapSeries.GetOrder(), f: mapSeries.New},
		{name: "minMax", order: minMax.GetOrder(), f: minMax.New},
		{name: "mostDeviant", order: mostDeviant.GetOrder(), f: mostDeviant.New},
		{name: "moving", order: moving.GetOrder(), f: moving.New},
		{name: "movingMedian", order: movingMedian.GetOrder(), f: movingMedian.New},
		{name: "multiplySeries", order: multiplySeries.GetOrder(), f: multiplySeries.New},
		{name: "multiplySeriesWithWildcards", order: multiplySeriesWithWildcards.GetOrder(), f: multiplySeriesWithWildcards.New},
		{name: "nPercentile", order: nPercentile.GetOrder(), f: nPercentile.New},
		{name: "nonNegativeDerivative", order: nonNegativeDerivative.GetOrder(), f: nonNegativeDerivative.New},
		{name: "offset", order: offset.GetOrder(), f: offset.New},
		{name: "offsetToZero", order: offsetToZero.GetOrder(), f: offsetToZero.New},
		{name: "pearson", order: pearson.GetOrder(), f: pearson.New},
		{name: "pearsonClosest", order: pearsonClosest.GetOrder(), f: pearsonClosest.New},
		{name: "perSecond", order: perSecond.GetOrder(), f: perSecond.New},
		{name: "percentileOfSeries", order: percentileOfSeries.GetOrder(), f: percentileOfSeries.New},
		{name: "polyfit", order: polyfit.GetOrder(), f: polyfit.New},
		{name: "pow", order: pow.GetOrder(), f: pow.New},
		{name: "randomWalk", order: randomWalk.GetOrder(), f: randomWalk.New},
		{name: "rangeOfSeries", order: rangeOfSeries.GetOrder(), f: rangeOfSeries.New},
		{name: "reduce", order: reduce.GetOrder(), f: reduce.New},
		{name: "removeBelowSeries", order: removeBelowSeries.GetOrder(), f: removeBelowSeries.New},
		{name: "removeEmptySeries", order: removeEmptySeries.GetOrder(), f: removeEmptySeries.New},
		{name: "round", order: round.GetOrder(), f: round.New},
		{name: "scale", order: scale.GetOrder(), f: scale.New},
		{name: "scaleToSeconds", order: scaleToSeconds.GetOrder(), f: scaleToSeconds.New},
		{name: "seriesByTag", order: seriesByTag.GetOrder(), f: seriesByTag.New},
		{name: "seriesList", order: seriesList.GetOrder(), f: seriesList.New},
		{name: "smartSummarize", order: smartSummarize.GetOrder(), f: smartSummarize.New},
		{name: "sortBy", order: sortBy.GetOrder(), f: sortBy.New},
		{name: "sortByName", order: sortByName.GetOrder(), f: sortByName.New},
		{name: "squareRoot", order: squareRoot.GetOrder(), f: squareRoot.New},
		{name: "stddevSeries", order: stddevSeries.GetOrder(), f: stddevSeries.New},
		{name: "stdev", order: stdev.GetOrder(), f: stdev.New},
		{name: "substr", order: substr.GetOrder(), f: substr.New},
		{name: "sum", order: sum.GetOrder(), f: sum.New},
		{name: "sumSeriesWithWildcards", order: sumSeriesWithWildcards.GetOrder(), f: sumSeriesWithWildcards.New},
		{name: "summarize", order: summarize.GetOrder(), f: summarize.New},
		{name: "timeFunction", order: timeFunction.GetOrder(), f: timeFunction.New},
		{name: "timeShift", order: timeShift.GetOrder(), f: timeShift.New},
		{name: "timeStack", order: timeStack.GetOrder(), f: timeStack.New},
		{name: "transformNull", order: transformNull.GetOrder(), f: transformNull.New},
		{name: "tukey", order: tukey.GetOrder(), f: tukey.New},
	}

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
		md := f.f(configs[strings.ToLower(f.name)])
		for _, m := range md {
			metadata.RegisterFunction(m.Name, m.F)
		}
	}
}
