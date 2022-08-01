package functions

import (
	"sort"
	"strings"

	"github.com/grafana/carbonapi/expr/functions/absolute"
	"github.com/grafana/carbonapi/expr/functions/aggregate"
	"github.com/grafana/carbonapi/expr/functions/aggregateLine"
	"github.com/grafana/carbonapi/expr/functions/aggregateWithWildcards"
	"github.com/grafana/carbonapi/expr/functions/alias"
	"github.com/grafana/carbonapi/expr/functions/aliasByBase64"
	"github.com/grafana/carbonapi/expr/functions/aliasByMetric"
	"github.com/grafana/carbonapi/expr/functions/aliasByNode"
	"github.com/grafana/carbonapi/expr/functions/aliasByPostgres"
	"github.com/grafana/carbonapi/expr/functions/aliasByRedis"
	"github.com/grafana/carbonapi/expr/functions/aliasSub"
	"github.com/grafana/carbonapi/expr/functions/asPercent"
	"github.com/grafana/carbonapi/expr/functions/averageOutsidePercentile"
	"github.com/grafana/carbonapi/expr/functions/averageSeriesWithWildcards"
	"github.com/grafana/carbonapi/expr/functions/baselines"
	"github.com/grafana/carbonapi/expr/functions/below"
	"github.com/grafana/carbonapi/expr/functions/cactiStyle"
	"github.com/grafana/carbonapi/expr/functions/cairo"
	"github.com/grafana/carbonapi/expr/functions/changed"
	"github.com/grafana/carbonapi/expr/functions/consolidateBy"
	"github.com/grafana/carbonapi/expr/functions/constantLine"
	"github.com/grafana/carbonapi/expr/functions/cumulative"
	"github.com/grafana/carbonapi/expr/functions/delay"
	"github.com/grafana/carbonapi/expr/functions/derivative"
	"github.com/grafana/carbonapi/expr/functions/divideSeries"
	"github.com/grafana/carbonapi/expr/functions/ewma"
	"github.com/grafana/carbonapi/expr/functions/exclude"
	"github.com/grafana/carbonapi/expr/functions/exp"
	"github.com/grafana/carbonapi/expr/functions/fallbackSeries"
	"github.com/grafana/carbonapi/expr/functions/fft"
	"github.com/grafana/carbonapi/expr/functions/filter"
	"github.com/grafana/carbonapi/expr/functions/graphiteWeb"
	"github.com/grafana/carbonapi/expr/functions/grep"
	"github.com/grafana/carbonapi/expr/functions/group"
	"github.com/grafana/carbonapi/expr/functions/groupByNode"
	"github.com/grafana/carbonapi/expr/functions/groupByTags"
	"github.com/grafana/carbonapi/expr/functions/heatMap"
	"github.com/grafana/carbonapi/expr/functions/highestLowest"
	"github.com/grafana/carbonapi/expr/functions/hitcount"
	"github.com/grafana/carbonapi/expr/functions/holtWintersAberration"
	"github.com/grafana/carbonapi/expr/functions/holtWintersConfidenceBands"
	"github.com/grafana/carbonapi/expr/functions/holtWintersForecast"
	"github.com/grafana/carbonapi/expr/functions/ifft"
	"github.com/grafana/carbonapi/expr/functions/integral"
	"github.com/grafana/carbonapi/expr/functions/integralByInterval"
	"github.com/grafana/carbonapi/expr/functions/integralWithReset"
	"github.com/grafana/carbonapi/expr/functions/interpolate"
	"github.com/grafana/carbonapi/expr/functions/invert"
	"github.com/grafana/carbonapi/expr/functions/isNotNull"
	"github.com/grafana/carbonapi/expr/functions/keepLastValue"
	"github.com/grafana/carbonapi/expr/functions/kolmogorovSmirnovTest2"
	"github.com/grafana/carbonapi/expr/functions/legendValue"
	"github.com/grafana/carbonapi/expr/functions/limit"
	"github.com/grafana/carbonapi/expr/functions/linearRegression"
	"github.com/grafana/carbonapi/expr/functions/logarithm"
	"github.com/grafana/carbonapi/expr/functions/logit"
	"github.com/grafana/carbonapi/expr/functions/lowPass"
	"github.com/grafana/carbonapi/expr/functions/mapSeries"
	"github.com/grafana/carbonapi/expr/functions/mostDeviant"
	"github.com/grafana/carbonapi/expr/functions/moving"
	"github.com/grafana/carbonapi/expr/functions/movingMedian"
	"github.com/grafana/carbonapi/expr/functions/multiplySeriesWithWildcards"
	"github.com/grafana/carbonapi/expr/functions/nPercentile"
	"github.com/grafana/carbonapi/expr/functions/nonNegativeDerivative"
	"github.com/grafana/carbonapi/expr/functions/offset"
	"github.com/grafana/carbonapi/expr/functions/offsetToZero"
	"github.com/grafana/carbonapi/expr/functions/pearson"
	"github.com/grafana/carbonapi/expr/functions/pearsonClosest"
	"github.com/grafana/carbonapi/expr/functions/perSecond"
	"github.com/grafana/carbonapi/expr/functions/percentileOfSeries"
	"github.com/grafana/carbonapi/expr/functions/polyfit"
	"github.com/grafana/carbonapi/expr/functions/pow"
	"github.com/grafana/carbonapi/expr/functions/randomWalk"
	"github.com/grafana/carbonapi/expr/functions/rangeOfSeries"
	"github.com/grafana/carbonapi/expr/functions/reduce"
	"github.com/grafana/carbonapi/expr/functions/removeBelowSeries"
	"github.com/grafana/carbonapi/expr/functions/removeEmptySeries"
	"github.com/grafana/carbonapi/expr/functions/round"
	"github.com/grafana/carbonapi/expr/functions/scale"
	"github.com/grafana/carbonapi/expr/functions/scaleToSeconds"
	"github.com/grafana/carbonapi/expr/functions/seriesByTag"
	"github.com/grafana/carbonapi/expr/functions/seriesList"
	"github.com/grafana/carbonapi/expr/functions/sinFunction"
	"github.com/grafana/carbonapi/expr/functions/slo"
	"github.com/grafana/carbonapi/expr/functions/smartSummarize"
	"github.com/grafana/carbonapi/expr/functions/sortBy"
	"github.com/grafana/carbonapi/expr/functions/sortByName"
	"github.com/grafana/carbonapi/expr/functions/squareRoot"
	"github.com/grafana/carbonapi/expr/functions/stdev"
	"github.com/grafana/carbonapi/expr/functions/substr"
	"github.com/grafana/carbonapi/expr/functions/sumSeriesWithWildcards"
	"github.com/grafana/carbonapi/expr/functions/summarize"
	"github.com/grafana/carbonapi/expr/functions/timeFunction"
	"github.com/grafana/carbonapi/expr/functions/timeShift"
	"github.com/grafana/carbonapi/expr/functions/timeShiftByMetric"
	"github.com/grafana/carbonapi/expr/functions/timeSlice"
	"github.com/grafana/carbonapi/expr/functions/timeStack"
	"github.com/grafana/carbonapi/expr/functions/transformNull"
	"github.com/grafana/carbonapi/expr/functions/tukey"
	"github.com/grafana/carbonapi/expr/functions/unique"
	"github.com/grafana/carbonapi/expr/functions/weightedAverage"
	"github.com/grafana/carbonapi/expr/interfaces"
	"github.com/grafana/carbonapi/expr/metadata"
)

type initFunc struct {
	name     string
	filename string
	order    interfaces.Order
	f        func(configFile string) []interfaces.FunctionMetadata
}

func New(configs map[string]string) {
	funcs := []initFunc{
		{name: "absolute", filename: "absolute", order: absolute.GetOrder(), f: absolute.New},
		{name: "aggregate", filename: "aggregate", order: aggregate.GetOrder(), f: aggregate.New},
		{name: "aggregateLine", filename: "aggregateLine", order: aggregateLine.GetOrder(), f: aggregateLine.New},
		{name: "aggregateWithWildcards", filename: "aggregateWithWildcards", order: aggregateWithWildcards.GetOrder(), f: aggregateWithWildcards.New},
		{name: "alias", filename: "alias", order: alias.GetOrder(), f: alias.New},
		{name: "aliasByBase64", filename: "aliasByBase64", order: aliasByBase64.GetOrder(), f: aliasByBase64.New},
		{name: "aliasByMetric", filename: "aliasByMetric", order: aliasByMetric.GetOrder(), f: aliasByMetric.New},
		{name: "aliasByNode", filename: "aliasByNode", order: aliasByNode.GetOrder(), f: aliasByNode.New},
		{name: "aliasByPostgres", filename: "aliasByPostgres", order: aliasByPostgres.GetOrder(), f: aliasByPostgres.New},
		{name: "aliasByRedis", filename: "aliasByRedis", order: aliasByRedis.GetOrder(), f: aliasByRedis.New},
		{name: "aliasSub", filename: "aliasSub", order: aliasSub.GetOrder(), f: aliasSub.New},
		{name: "asPercent", filename: "asPercent", order: asPercent.GetOrder(), f: asPercent.New},
		{name: "averageOutsidePercentile", filename: "averageOutsidePercentile", order: averageOutsidePercentile.GetOrder(), f: averageOutsidePercentile.New},
		{name: "averageSeriesWithWildcards", filename: "averageSeriesWithWildcards", order: averageSeriesWithWildcards.GetOrder(), f: averageSeriesWithWildcards.New},
		{name: "baselines", filename: "baselines", order: baselines.GetOrder(), f: baselines.New},
		{name: "below", filename: "below", order: below.GetOrder(), f: below.New},
		{name: "cactiStyle", filename: "cactiStyle", order: cactiStyle.GetOrder(), f: cactiStyle.New},
		{name: "cairo", filename: "cairo", order: cairo.GetOrder(), f: cairo.New},
		{name: "changed", filename: "changed", order: changed.GetOrder(), f: changed.New},
		{name: "consolidateBy", filename: "consolidateBy", order: consolidateBy.GetOrder(), f: consolidateBy.New},
		{name: "constantLine", filename: "constantLine", order: constantLine.GetOrder(), f: constantLine.New},
		{name: "cumulative", filename: "cumulative", order: cumulative.GetOrder(), f: cumulative.New},
		{name: "delay", filename: "delay", order: delay.GetOrder(), f: delay.New},
		{name: "derivative", filename: "derivative", order: derivative.GetOrder(), f: derivative.New},
		{name: "divideSeries", filename: "divideSeries", order: divideSeries.GetOrder(), f: divideSeries.New},
		{name: "ewma", filename: "ewma", order: ewma.GetOrder(), f: ewma.New},
		{name: "exclude", filename: "exclude", order: exclude.GetOrder(), f: exclude.New},
		{name: "exp", filename: "exp", order: exp.GetOrder(), f: exp.New},
		{name: "fallbackSeries", filename: "fallbackSeries", order: fallbackSeries.GetOrder(), f: fallbackSeries.New},
		{name: "fft", filename: "fft", order: fft.GetOrder(), f: fft.New},
		{name: "filter", filename: "filter", order: filter.GetOrder(), f: filter.New},
		{name: "graphiteWeb", filename: "graphiteWeb", order: graphiteWeb.GetOrder(), f: graphiteWeb.New},
		{name: "grep", filename: "grep", order: grep.GetOrder(), f: grep.New},
		{name: "group", filename: "group", order: group.GetOrder(), f: group.New},
		{name: "groupByNode", filename: "groupByNode", order: groupByNode.GetOrder(), f: groupByNode.New},
		{name: "groupByTags", filename: "groupByTags", order: groupByTags.GetOrder(), f: groupByTags.New},
		{name: "heatMap", filename: "heatMap", order: heatMap.GetOrder(), f: heatMap.New},
		{name: "highestLowest", filename: "highestLowest", order: highestLowest.GetOrder(), f: highestLowest.New},
		{name: "hitcount", filename: "hitcount", order: hitcount.GetOrder(), f: hitcount.New},
		{name: "holtWintersAberration", filename: "holtWintersAberration", order: holtWintersAberration.GetOrder(), f: holtWintersAberration.New},
		{name: "holtWintersConfidenceBands", filename: "holtWintersConfidenceBands", order: holtWintersConfidenceBands.GetOrder(), f: holtWintersConfidenceBands.New},
		{name: "holtWintersForecast", filename: "holtWintersForecast", order: holtWintersForecast.GetOrder(), f: holtWintersForecast.New},
		{name: "ifft", filename: "ifft", order: ifft.GetOrder(), f: ifft.New},
		{name: "integral", filename: "integral", order: integral.GetOrder(), f: integral.New},
		{name: "integralByInterval", filename: "integralByInterval", order: integralByInterval.GetOrder(), f: integralByInterval.New},
		{name: "integralWithReset", filename: "integralWithReset", order: integralWithReset.GetOrder(), f: integralWithReset.New},
		{name: "interpolate", filename: "interpolate", order: interpolate.GetOrder(), f: interpolate.New},
		{name: "invert", filename: "invert", order: invert.GetOrder(), f: invert.New},
		{name: "isNotNull", filename: "isNotNull", order: isNotNull.GetOrder(), f: isNotNull.New},
		{name: "keepLastValue", filename: "keepLastValue", order: keepLastValue.GetOrder(), f: keepLastValue.New},
		{name: "kolmogorovSmirnovTest2", filename: "kolmogorovSmirnovTest2", order: kolmogorovSmirnovTest2.GetOrder(), f: kolmogorovSmirnovTest2.New},
		{name: "legendValue", filename: "legendValue", order: legendValue.GetOrder(), f: legendValue.New},
		{name: "limit", filename: "limit", order: limit.GetOrder(), f: limit.New},
		{name: "linearRegression", filename: "linearRegression", order: linearRegression.GetOrder(), f: linearRegression.New},
		{name: "logarithm", filename: "logarithm", order: logarithm.GetOrder(), f: logarithm.New},
		{name: "logit", filename: "logit", order: logit.GetOrder(), f: logit.New},
		{name: "lowPass", filename: "lowPass", order: lowPass.GetOrder(), f: lowPass.New},
		{name: "mapSeries", filename: "mapSeries", order: mapSeries.GetOrder(), f: mapSeries.New},
		{name: "mostDeviant", filename: "mostDeviant", order: mostDeviant.GetOrder(), f: mostDeviant.New},
		{name: "moving", filename: "moving", order: moving.GetOrder(), f: moving.New},
		{name: "movingMedian", filename: "movingMedian", order: movingMedian.GetOrder(), f: movingMedian.New},
		{name: "multiplySeriesWithWildcards", filename: "multiplySeriesWithWildcards", order: multiplySeriesWithWildcards.GetOrder(), f: multiplySeriesWithWildcards.New},
		{name: "nPercentile", filename: "nPercentile", order: nPercentile.GetOrder(), f: nPercentile.New},
		{name: "nonNegativeDerivative", filename: "nonNegativeDerivative", order: nonNegativeDerivative.GetOrder(), f: nonNegativeDerivative.New},
		{name: "offset", filename: "offset", order: offset.GetOrder(), f: offset.New},
		{name: "offsetToZero", filename: "offsetToZero", order: offsetToZero.GetOrder(), f: offsetToZero.New},
		{name: "pearson", filename: "pearson", order: pearson.GetOrder(), f: pearson.New},
		{name: "pearsonClosest", filename: "pearsonClosest", order: pearsonClosest.GetOrder(), f: pearsonClosest.New},
		{name: "perSecond", filename: "perSecond", order: perSecond.GetOrder(), f: perSecond.New},
		{name: "percentileOfSeries", filename: "percentileOfSeries", order: percentileOfSeries.GetOrder(), f: percentileOfSeries.New},
		{name: "polyfit", filename: "polyfit", order: polyfit.GetOrder(), f: polyfit.New},
		{name: "pow", filename: "pow", order: pow.GetOrder(), f: pow.New},
		{name: "randomWalk", filename: "randomWalk", order: randomWalk.GetOrder(), f: randomWalk.New},
		{name: "rangeOfSeries", filename: "rangeOfSeries", order: rangeOfSeries.GetOrder(), f: rangeOfSeries.New},
		{name: "reduce", filename: "reduce", order: reduce.GetOrder(), f: reduce.New},
		{name: "removeBelowSeries", filename: "removeBelowSeries", order: removeBelowSeries.GetOrder(), f: removeBelowSeries.New},
		{name: "removeEmptySeries", filename: "removeEmptySeries", order: removeEmptySeries.GetOrder(), f: removeEmptySeries.New},
		{name: "round", filename: "round", order: round.GetOrder(), f: round.New},
		{name: "scale", filename: "scale", order: scale.GetOrder(), f: scale.New},
		{name: "scaleToSeconds", filename: "scaleToSeconds", order: scaleToSeconds.GetOrder(), f: scaleToSeconds.New},
		{name: "seriesByTag", filename: "seriesByTag", order: seriesByTag.GetOrder(), f: seriesByTag.New},
		{name: "seriesList", filename: "seriesList", order: seriesList.GetOrder(), f: seriesList.New},
		{name: "sinFunction", filename: "sinFunction", order: sinFunction.GetOrder(), f: sinFunction.New},
		{name: "slo", filename: "slo", order: slo.GetOrder(), f: slo.New},
		{name: "smartSummarize", filename: "smartSummarize", order: smartSummarize.GetOrder(), f: smartSummarize.New},
		{name: "sortBy", filename: "sortBy", order: sortBy.GetOrder(), f: sortBy.New},
		{name: "sortByName", filename: "sortByName", order: sortByName.GetOrder(), f: sortByName.New},
		{name: "squareRoot", filename: "squareRoot", order: squareRoot.GetOrder(), f: squareRoot.New},
		{name: "stdev", filename: "stdev", order: stdev.GetOrder(), f: stdev.New},
		{name: "substr", filename: "substr", order: substr.GetOrder(), f: substr.New},
		{name: "sumSeriesWithWildcards", filename: "sumSeriesWithWildcards", order: sumSeriesWithWildcards.GetOrder(), f: sumSeriesWithWildcards.New},
		{name: "summarize", filename: "summarize", order: summarize.GetOrder(), f: summarize.New},
		{name: "timeFunction", filename: "timeFunction", order: timeFunction.GetOrder(), f: timeFunction.New},
		{name: "timeShift", filename: "timeShift", order: timeShift.GetOrder(), f: timeShift.New},
		{name: "timeShiftByMetric", filename: "timeShiftByMetric", order: timeShiftByMetric.GetOrder(), f: timeShiftByMetric.New},
		{name: "timeSlice", filename: "timeSlice", order: timeSlice.GetOrder(), f: timeSlice.New},
		{name: "timeStack", filename: "timeStack", order: timeStack.GetOrder(), f: timeStack.New},
		{name: "transformNull", filename: "transformNull", order: transformNull.GetOrder(), f: transformNull.New},
		{name: "tukey", filename: "tukey", order: tukey.GetOrder(), f: tukey.New},
		{name: "unique", filename: "unique", order: unique.GetOrder(), f: unique.New},
		{name: "weightedAverage", filename: "weightedAverage", order: weightedAverage.GetOrder(), f: weightedAverage.New},
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
			metadata.RegisterFunctionWithFilename(m.Name, f.filename, m.F)
		}
	}
}
