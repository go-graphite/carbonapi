package main

func getFn(fnName string) func(e *expr, from, until int32, values map[metricRequest][]*metricData) []*metricData {
	switch fnName {
	case "absolute":
		return absolute
	case "alias":
		return alias
	case "aliasByNode":
		return aliasByNode
	case "aliasByMetric":
		return aliasByMetric
	case "aliasSub":
		return aliasSub
	case "asPercent":
		return asPercent
	case "avg", "averageSeries":
		return avgSeries
	case "averageSeriesWithWildcards":
		return avgSeriesWithWildcards
	case "averageAbove":
		return filter(avgValue, true, true)
	case "averageBelow":
		return filter(avgValue, false, true)
	case "currentAbove":
		return filter(curValue, true, true)
	case "currentBelow":
		return filter(curValue, false, true)
	case "maximumAbove":
		return filter(maxValue, true, false)
	case "maximumBelow":
		return filter(maxValue, false, false)
	case "minimumAbove":
		return filter(minValue, true, false)
	case "minimumBelow":
		return filter(minValue, false, false)
	case "derivative":
		return derivative
	case "diffSeries":
		return diffSeries
	case "divideSeries":
		return divideSeries
	case "multiplySeries":
		return multiplySeries
	case "exclude":
		return exclude
	case "grep":
		return grep
	case "group":
		return group
	case "groupByNode":
		return groupByNode
	case "isNonNull", "isNotNull":
		return isNonNull
	case "lowestAverage":
		return lowest(avgValue)
	case "lowestCurrent":
		return lowest(curValue)
	case "highestAverage":
		return highest(avgValue)
	case "highestCurrent":
		return highest(curValue)
	case "highestMax":
		return highest(maxValue)
	case "hitcount":
		return hitcount
	case "integral":
		return integral
	case "invert":
		return invert
	case "keepLastValue":
		return keepLastValue
	case "changed":
		return changed
	case "kolmogorovSmirnovTest2", "ksTest2": // ksTest2(series, series, points|"interval")
		return ksTest
	case "limit":
		return limit
	case "logarithm", "log":
		return log
	case "maxSeries":
		return maxSeries
	case "minSeries":
		return minSeries
	case "mostDeviant":
		return mostDeviant
	case "movingAverage":
		return movingAverage
	case "movingMedian":
		return movingMedian
	case "nonNegativeDerivative":
		return nonNegativeDerivative
	case "perSecond":
		return perSecond
	case "nPercentile":
		return nPercentile
	case "pearson":
		return pearson
	case "pearsonClosest":
		return pearsonClosest
	case "offset":
		return offset
	case "offsetToZero":
		return offsetToZero
	case "scale":
		return scale
	case "scaleToSeconds":
		return scaleToSeconds
	case "pow":
		return pow
	case "sortByMaxima", "sortByMinima", "sortByTotal":
		return sortBy
	case "sortByName":
		return sortByName
	case "stdev", "stddev":
		return stdev
	case "sum", "sumSeries":
		return sum
	case "sumSeriesWithWildcards":
		return sumSeriesWithWildcards
	case "percentileOfSeries":
		return percentileOfSeries
	case "summarize":
		return summarize
	case "timeShift":
		return timeShift
	case "transformNull":
		return transformNull
	case "tukeyAbove", "tukeyBelow":
		return tukey
	case "color":
		return applyColor
	case "dashed", "drawAsInfinite", "secondYAxis":
		return draw
	case "constantLine":
		return constantLine
	case "holtWintersForecast":
		return holtWintersForecast
	case "squareRoot":
		return sqrt
	case "removeBelowValue":
		return removeBelowValue
	case "removeAboveValue":
		return removeAboveValue
	}
	return nil
}
