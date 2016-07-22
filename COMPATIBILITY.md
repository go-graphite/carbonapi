# CarbonAPI compatibility with Graphite

Topics:
* [URI Parameters](#uri-params)
* [Functions](#functions)

<a name="uri-params"></a>
## URI Parameters

### /render/?...

* `target` : graphite series, seriesList or function (likely containing series or seriesList)
* `from`, `until` : time specifiers. Eg. "1d", "10min", "04:37_20150822", "now", "today", ... (**NOTE** does not handle timezones the same as graphite)
* `format` : support graphite values of { json, raw, pickle, csv, png } adds { protobuf } and does not support { svg, pdf }
* `jsonp` : (...)
* `noCache` : prevent query-response caching (which is 60s if enabled)
* `cacheTimeout` : override default result cache (60s)
* `rawdata` -or- `rawData` : true for `format=raw`

**Explicitly NOT supported**
* `_salt`
* `_ts`
* `_t`

_When `format=png`_ (default if not specified)
* `width`, `height` : number of pixels (default: width=330 , height=250)
* `margin` : (10)
* `logBase` : Y-scale should use. Recognizes "e" or a floating point ( >= 1 )
* `fgcolor` : foreground color
* `bgcolor` : background color
* `majorLine` : major line color
* `minorLine` : minor line color
* `fontName` : ("Sans")
* `fontSize` : (10.0)
* `fontBold` : (false)
* `fontItalic` : (false)
* `graphOnly` : (false)
* `hideLegend` : (false) (**NOTE** if not defined and >10 result metrics this becomes true)
* `hideGrid` : (false)
* `hideAxes` : (false)
* `hideYAxis` : (false)
* `hideXAxis` : (false)
* `yAxisSide` : ("left")
* `connectedLimit` : number of missing points to bridge when `linemode` is not one of { "slope", "staircase" } likely "connected" (4294967296)
* `lineMode` : ("slope")
* `areaMode` : ("none") also recognizes { "first", "all", "stacked" }
* `areaAlpha` : ( <not defined> ) float value for area alpha
* `pieMode` : ("average") also recognizes { "maximum", "minimum" } (**NOTE** pie graph support is explicitly unplanned)
* `lineWidth` : (1.2) float value for line width
* `dashed` : (false) dashed lines
* `rightWidth` : (1.2) ...
* `rightDashed` : (false)
* `rightColor` : ...
* `leftWidth` : (1.2)
* `leftDashed` : (false)
* `leftColor` : ...
* `title` : ("") graph title
* `vtitle` : ("") ...
* `vtitleRight` : ("") ...
* `colorList` : ("blue,green,red,purple,yellow,aqua,grey,magenta,pink,gold,rose")
* `majorGridLineColor` : ("rose")
* `minorGridLineColor` : ("grey")
* `uniqueLegend` : (false)
* `drawNullAsZero` : (false) (**NOTE** affects display only - does not translate missing values to zero in functions. For that use ...)
* `drawAsInfinite` : (false) ...
* `yMin` : <undefined>
* `yMax` : <undefined>
* `yStep` : <undefined>
* `xMin` : <undefined>
* `xMax` : <undefined>
* `xStep` : <undefined>
* `xFormat` : ("") ...
* `minorY` : (1) ...
* `yMinLeft` : <undefined>
* `yMinRight` : <undefined>
* `yMaxLeft` : <undefined>
* `yMaxRight` : <undefined>
* `yStepL` : <undefined>
* `ySTepR` : <undefined>
* `yLimitLeft` : <undefined>
* `yLimitRight` : <undefined>
* `yUnitSystem` : ("si") also recognizes { "binary" }
* `yDivisors` : (4,5,6) ...

### /metrics/find/?

* `format` : ("treejson") also recognizes { "json" (same as "treejson"), "completer", "raw" }
* `jsonp` : ...
* `query` : the metric or glob-pattern to find

---

<a name="functions"></a>
## Functions

**Note:** _Version_ listed in the table below represents the earliest graphite version where the function appeared with the current signature. In **most** cases this was when the function was introduced.

Graphite Function                                                         | Version | Carbon API
:------------------------------------------------------------------------ | :------ | :---------
absolute(seriesList)                                                      |  0.9.10 | Supported
aggregateLine(seriesList, func='avg')                                     |  0.9.14 |
alias(seriesList, newName)                                                |  0.9.9  | Supported
aliasByMetric(seriesList)                                                 |  0.9.10 | Supported
aliasByNode(seriesList, *nodes)                                           |  0.9.14 | Supported
aliasSub(seriesList, search, replace)                                     |  0.9.10 | Supported
alpha(seriesList, alpha)                                                  |  0.9.10 |
areaBetween(seriesList)                                                   |  0.9.14 |
asPercent(seriesList, total=None)                                         |  0.9.10 | Supported
averageAbove(seriesList, n)                                               |  0.9.9  | Supported
averageBelow(seriesList, n)                                               |  0.9.9  | Supported
averageOutsidePercentile(seriesList, n)                                   |  0.9.11 |
averageSeries(*seriesLists), Short Alias: avg()                           |  0.9.9  | Supported
averageSeriesWithWildcards(seriesList, *position)                         |  0.9.10 |
cactiStyle(seriesList, system=None)                                       |  0.9.14 |
changed(seriesList)                                                       |  0.9.14 | Supported
color(seriesList, theColor)                                               |  0.9.9  | Supported
consolidateBy(seriesList, consolidationFunc)                              |  0.9.14 | Supported
cumulative(seriesList, consolidationFunc='sum')                           |  0.9.14 |
constantLine(value)                                                       |  0.9.9  |
countSeries(*seriesLists)                                                 |  0.9.14 |
currentAbove(seriesList, n)                                               |  0.9.9  | Supported
currentBelow(seriesList, n)                                               |  0.9.9  | Supported
dashed(*seriesList)                                                       |  0.9.9  | Supported
derivative(seriesList)                                                    |  0.9.9  | Supported
diffSeries(*seriesLists)                                                  |  0.9.9  | Supported
divideSeries(dividendSeriesList, divisorSeries)                           |  0.9.14 | Supported
drawAsInfinite(seriesList)                                                |  0.9.9  | Supported
events(*tags)                                                             |  0.9.9  |
exclude(seriesList, pattern)                                              |  0.9.9  | Supported
fallbackSeries( seriesList, fallback )                                    |  0.9.14 |
grep(seriesList, pattern)                                                 |  0.9.14 | Supported
group(*seriesLists)                                                       |  0.9.10 | Supported
groupByNode(seriesList, nodeNum, callback)                                |  0.9.9  | Supported
highestAverage(seriesList, n)                                             |  0.9.9  | Supported
highestCurrent(seriesList, n)                                             |  0.9.9  | Supported
highestMax(seriesList, n)                                                 |  0.9.9  | Supported
hitcount(seriesList, intervalString, alignToInterval=False)               |  0.9.10 | Supported
holtWintersAberration(seriesList, delta=3)                                |  0.9.10 | [#66](/dgryski/carbonapi/issues/66)
holtWintersConfidenceArea(seriesList, delta=3)                            |  0.9.10 | [#66](/dgryski/carbonapi/issues/66)
holtWintersConfidenceBands(seriesList, delta=3)                           |  0.9.10 | [#66](/dgryski/carbonapi/issues/66)
holtWintersForecast(seriesList)                                           |  0.9.10 | Supported - but see: [#66](/dgryski/carbonapi/issues/66)
identity(name)                                                            |  0.9.14 |
integral(seriesList)                                                      |  0.9.9  | Supported
invert(seriesList)                                                        |  0.9.14 | Supported
isNonNull(seriesList)                                                     |  0.9.11 | Supported (also isNotNull alias)
keepLastValue(seriesList, limit=inf)                                      |  0.9.14 | Supported
kolmogorovSmirnovTest2(series, series, windowSize) alias ksTest2()        |  not in graphite | Experimental
legendValue(seriesList, *valueTypes)                                      |  0.9.10 |
limit(seriesList, n)                                                      |  0.9.9  | Supported
lineWidth(seriesList, width)                                              |  0.9.9  |
logarithm(seriesList, base=10), alias log()                               |  0.9.10 | Supported
lowestAverage(seriesList, n)                                              |  0.9.9  | Supported
lowestCurrent(seriesList, n)                                              |  0.9.9  | Supported
mapSeries(seriesList, mapNode), Short form: map()                         |  0.9.14 |
maxSeries(*seriesLists)                                                   |  0.9.9  | Supported
maximumAbove(seriesList, n)                                               |  0.9.9  | Supported
maximumBelow(seriesList, n)                                               |  0.9.9  | Supported
minSeries(*seriesLists)                                                   |  0.9.9  | Supported
minimumAbove(seriesList, n)                                               |  0.9.10 | Supported
minimumBelow(seriesList, n)                                               |  0.9.14 | Supported
mostDeviant(seriesList, n)                                                |  0.9.9  | Supported
movingAverage(seriesList, windowSize)                                     |  0.9.14 | Supported
movingMedian(seriesList, windowSize)                                      |  0.9.14 | Supported
multiplySeries(*seriesLists)                                              |  0.9.10 | Supported
multiplySeriesWithWildcards(seriesList, *position)                        |  0.9.14 |
nPercentile(seriesList, n)                                                |  0.9.9  | Supported
nonNegativeDerivative(seriesList, maxValue=None)                          |  0.9.9  | Supported
offset(seriesList, factor)                                                |  0.9.9  | Supported
offsetToZero(seriesList)                                                  |  0.9.11 | Supported
pearson(series, series, n)                                                |  not in graphite | Experimental
pearsonClosest(series, seriesList, windowSize, direction="abs")           |  not in graphite | Experimental
perSecond(seriesList, maxValue=None)                                      |  0.9.14 | Supported
percentileOfSeries(seriesList, n, interpolate=False)                      |  0.9.10 | Supported
pow(seriesList, factor)                                                   |  0.9.14 | Supported
randomWalkFunction(name, step=60), Short Alias: randomWalk()              |  0.9.9  | Supported
rangeOfSeries(*seriesLists)                                               |  0.9.10 | Supported
reduceSeries(seriesLists, reduceFunction, reduceNode, *reduceMatchers)    |  0.9.14 |
+ reduce() Short form of reduceSeries()                                   |  - - -  |
removeAbovePercentile(seriesList, n)                                      |  0.9.10 | Supported
removeAboveValue(seriesList, n)                                           |  0.9.10 | Supported
removeBelowPercentile(seriesList, n)                                      |  0.9.10 | Supported
removeBelowValue(seriesList, n)                                           |  0.9.10 | Supported
removeBetweenPercentile(seriesList, n)                                    |  0.9.11 |
removeEmptySeries(seriesList)                                             |  0.9.14 | Supported
removeZeroSeries(seriesList)                                              |  0.9.14 | Supported
scale(seriesList, factor)                                                 |  0.9.9  | Supported
scaleToSeconds(seriesList, seconds)                                       |  0.9.10 | Supported
secondYAxis(seriesList)                                                   |  0.9.10 | Supported
sinFunction(name, amplitude=1, step=60), Short Alias: sin()               |  0.9.9  |
smartSummarize(seriesList, intervalString, func='sum', alignToFrom=False) |  0.9.10 |
sortByMaxima(seriesList)                                                  |  0.9.9  | Supported
sortByMinima(seriesList)                                                  |  0.9.9  | Supported
sortByName(seriesList)                                                    |  0.9.15 | Supported
sortByTotal(seriesList)                                                   |  0.9.11 | Supported
squareRoot(seriesList)                                                    |  0.9.14 | Supported
stacked(seriesLists, stackName='__DEFAULT__')                             |  0.9.10 | [#74](/dgryski/carbonapi/issues/74)
stddevSeries(*seriesLists)                                                |  0.9.14 |
stdev(seriesList, points, windowTolerance=0.1)                            |  0.9.10 | Supported + alias stddev()
substr(seriesList, start=0, stop=0)                                       |  0.9.9  |
sumSeries(*seriesLists), Short form: sum()                                |  0.9.9  | Supported
sumSeriesWithWildcards(seriesList, *position)                             |  0.9.10 | Supported
summarize(seriesList, intervalString, func='sum', alignToFrom=False)      |  0.9.9  | Supported
threshold(value, label=None, color=None)                                  |  0.9.9  | Supported
timeFunction(name, step=60), Short Alias: time()                          |  0.9.9  | Supported
timeShift(seriesList, timeShift, resetEnd=True)                           |  0.9.11 | Supported
timeSlice(seriesList, startSliceAt, endSliceAt='now')                     |  0.9.14 |
timeStack(seriesList, timeShiftUnit, timeShiftStart, timeShiftEnd)        |  0.9.14 | Supported
tukeyAbove(seriesList, basis, n, interval=0)                              |  not in graphite | Experimental
tukeyBelow(seriesList, basis, n, interval=0)                              |  not in graphite | Experimental
transformNull(seriesList, default=0)                                      |  0.9.10 | Supported
useSeriesAbove(seriesList, value, search, replace)                        |  0.9.10 |
weightedAverage(seriesListAvg, seriesListWeight, node)                    |  0.9.14 |

-----

