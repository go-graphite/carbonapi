# CarbonAPI compatibility with Graphite

Topics:
* [Default settings](#default-settings)
* [URI Parameters](#uri-params)
* [Functions](#functions)
* [Features of configuration functions](#functions-features)

<a name="default-settings"></a>
## Default Settings

### Default Line Colors
Default colors for png or svg rendering intentionally specified like it is in graphite-web 1.1.0

You can redefine that in config to be more more precise. In default config example they are defined in the same way as in [original graphite PR to make them right](https://github.com/graphite-project/graphite-web/pull/2239)

Reason behind that change is that on dark background it's much nicer to read old colors than new one

<a name="uri-params"></a>
## URI Parameters

### /render/?...

* `target` : graphite series, seriesList or function (likely containing series or seriesList)
* `from`, `until` : time specifiers. Eg. "1d", "10min", "04:37_20150822", "now", "today", ... (**NOTE** does not handle timezones the same as graphite)
* `format` : support graphite values of { json, raw, pickle, csv, png, svg } adds { protobuf } and does not support { pdf }
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

Missing function: "applyByNode", "aliasQuery", "filterSeries", "unique", "integralByInterval", "xFilesFactor", "lowest"

Graphite Function                                                         | Version | Carbon API
:------------------------------------------------------------------------ | :------ | :---------
absolute(seriesList)                                                      |  0.9.10 | Supported
aggregate                                                                 |  1.1.0  |
aggregateLine(seriesList, func='avg')                                     |  1.0.0  |
aggregateWithWildcards                                                    |  1.1.0  |
alias(seriesList, newName)                                                |  0.9.9  | Supported
aliasByMetric(seriesList)                                                 |  0.9.10 | Supported
aliasByNode(seriesList, *nodes)                                           |  0.9.14 | Supported
aliasByPostgre(seriesList, database, key-string, node[i])                 |  not in graphite  | Experimental
aliasByTags                                                               |  1.1.0  |
aliasSub(seriesList, search, replace)                                     |  0.9.10 | Supported
alpha(seriesList, alpha)                                                  |  0.9.10 | Supported
applyByNode(seriesList, nodeNum, templateFunction, newName=None)          |  1.0.0  | Supported
areaBetween(seriesList)                                                   |  0.9.14 | Supported
asPercent(seriesList, total=None, *nodes)                                 |  1.1.1  | Supported
averageAbove(seriesList, n)                                               |  0.9.9  | Supported
averageBelow(seriesList, n)                                               |  0.9.9  | Supported
averageOutsidePercentile(seriesList, n)                                   |  1.0.0  |
averageSeries(*seriesLists), Short Alias: avg()                           |  0.9.9  | Supported
averageSeriesWithWildcards(seriesList, *position)                         |  0.9.10 | Supported
cactiStyle(seriesList, system=None)                                       |  latest | Supported
changed(seriesList)                                                       |  0.9.14 | Supported
color(seriesList, theColor)                                               |  0.9.9  | Supported
consolidateBy(seriesList, consolidationFunc)                              |  0.9.14 | Supported
constantLine(value)                                                       |  0.9.9  | Supported
countSeries(*seriesLists)                                                 |  0.9.14 | Supported
cumulative(seriesList)                                                    |  0.9.14 | Supported
currentAbove(seriesList, n)                                               |  0.9.9  | Supported
currentBelow(seriesList, n)                                               |  0.9.9  | Supported
dashed(*seriesList)                                                       |  0.9.9  | Supported
delay(seriesList, steps)                                                  |  1.0.0  | Supported
derivative(seriesList)                                                    |  0.9.9  | Supported
diffSeries(*seriesLists)                                                  |  0.9.9  | Supported
divideSeries(dividendSeriesList, divisorSeries)                           |  0.9.14 | Supported
divideSeriesLists(dividendSeriesList, divisorSeriesList)                  |  1.0.2  | Supported
diffSeriesLists(leftSeriesList, rightSeriesList)                          |  not in graphite  | Experimental
multiplySeriesLists(leftSeriesList, rightSeriesList)                      |  not in graphite  | Experimental
drawAsInfinite(seriesList)                                                |  0.9.9  | Supported
events(*tags)                                                             |  0.9.9  |
exclude(seriesList, pattern)                                              |  0.9.9  | Supported
exponentialMovingAverage(seriesList, windowSize)                          |  1.0.0  |
exponentialWeightedMovingAverage(seriesList, alpha)                       | not in graphite | Experimental
ewma(seriesList, alpha)                                                   | - - -   | Short form of exponentialWeightedMovingAverage
fallbackSeries( seriesList, fallback )                                    |  1.0.0  |
[fft](https://en.wikipedia.org/wiki/Fast_Fourier_transform)(absSeriesList, phaseSeriesList)                                       |  not in graphite | Experimental
grep(seriesList, pattern)                                                 |  1.0.0  | Supported
group(*seriesLists)                                                       |  0.9.10 | Supported
groupByNode(seriesList, nodeNum, callback)                                |  0.9.9  | Supported
groupByNodes(seriesList, callback, *nodes)                                |  1.0.0  | Supported
groupByTags                                                               |  1.1.0  |
highest                                                                   |  1.1.0  |
highestAverage(seriesList, n)                                             |  0.9.9  | Supported
highestCurrent(seriesList, n)                                             |  0.9.9  | Supported
highestMax(seriesList, n)                                                 |  0.9.9  | Supported
hitcount(seriesList, intervalString, alignToInterval=False)               |  0.9.10 | Supported
holtWintersAberration(seriesList, delta=3)                                |  0.9.10 | Supported
holtWintersConfidenceArea(seriesList, delta=3)                            |  0.9.10 | [#66](https://github.com/go-graphite/carbonapi/issues/66)
holtWintersConfidenceBands(seriesList, delta=3)                           |  0.9.10 | Supported
holtWintersForecast(seriesList)                                           |  0.9.10 | Supported
identity(name)                                                            |  0.9.14 |
[ifft](https://en.wikipedia.org/wiki/Fast_Fourier_transform)(absSeriesList, phaseSeriesList)                                      |  not in graphite | Experimental
integral(seriesList)                                                      |  0.9.9  | Supported
integralByInterval(seriesList, intervalUnit)                              |  1.0.0  |
interpolate(seriesList, limit=inf)                                        |  1.0.0  |
invert(seriesList)                                                        |  1.0.0  | Supported
isNonNull(seriesList)                                                     |  1.0.0  | Supported (also isNotNull alias)
keepLastValue(seriesList, limit=inf)                                      |  0.9.14 | Supported
[kolmogorovSmirnovTest2](https://en.wikipedia.org/wiki/Kolmogorov%E2%80%93Smirnov_test)(series, series, windowSize) alias ksTest2()        |  not in graphite | Experimental
legendValue(seriesList, *valueTypes)                                      |  0.9.10 | Supported
limit(seriesList, n)                                                      |  0.9.9  | Supported
lineWidth(seriesList, width)                                              |  0.9.9  | Supported
linearRegression(seriesList, startSourceAt=None, endSourceAt=None)        |  1.0.0 | Supported (based on polyfit)
linearRegressionAnalysis(series)                                          |  1.0.0 |
logarithm(seriesList, base=10), alias log()                               |  0.9.10 | Supported
lowestAverage(seriesList, n)                                              |  0.9.9  | Supported
lowestCurrent(seriesList, n)                                              |  0.9.9  | Supported
[lowPass](https://en.wikipedia.org/wiki/Low-pass_filter)(seriesList, cutPercent)                                           |  not in graphite | Experimental
mapSeries(seriesList, mapNode), Short form: map()                         |  1.0.0  | Supported
maxSeries(*seriesLists)                                                   |  0.9.9  | Supported
maximumAbove(seriesList, n)                                               |  0.9.9  | Supported
maximumBelow(seriesList, n)                                               |  0.9.9  | Supported
minSeries(*seriesLists)                                                   |  0.9.9  | Supported
minMax                                                                    |  1.1.0  |
minimumAbove(seriesList, n)                                               |  0.9.10 | Supported
minimumBelow(seriesList, n)                                               |  0.9.14 | Supported
mostDeviant(seriesList, n)                                                |  0.9.9  | Supported
movingAverage(seriesList, windowSize)                                     |  0.9.14 | Supported
movingMax(seriesList, windowSize)                                         |  1.0.0  | Supported
movingMedian(seriesList, windowSize)                                      |  0.9.14 | Supported
movingMin(seriesList, windowSize)                                         |  1.0.0  | Supported
movingSum(seriesList, windowSize)                                         |  1.0.0  | Supported
movingWindow                                                              |  1.1.0  |
multiplySeries(*seriesLists)                                              |  0.9.10 | Supported
multiplySeriesWithWildcards(seriesList, *position)                        |  1.0.0  | Supported
nPercentile(seriesList, n)                                                |  0.9.9  | Supported
nonNegativeDerivative(seriesList, maxValue=None)                          |  0.9.9  | Supported
offset(seriesList, factor)                                                |  0.9.9  | Supported
offsetToZero(seriesList)                                                  |  1.0.0  | Supported
pct                                                                       |  1.1.0  |
[pearson](https://en.wikipedia.org/wiki/Pearson_product-moment_correlation_coefficient)(series, series, n)                                                |  not in graphite | Experimental
pearsonClosest(series, seriesList, windowSize, direction="abs")           |  not in graphite | Experimental
perSecond(seriesList, maxValue=None)                                      |  0.9.14 | Supported
percentileOfSeries(seriesList, n, interpolate=False)                      |  0.9.10 | Supported
[polyfit](https://en.wikipedia.org/wiki/Polynomial_regression)(seriesList, degree=1, offset='0d')                                |  not in graphite | Experimental
pow(seriesList, factor)                                                   |  0.9.14 | Supported
powSeries(*seriesLists)                                                   |  1.0.0  |
randomWalkFunction(name, step=60), Short Alias: randomWalk()              |  0.9.9  | Supported
rangeOfSeries(*seriesLists)                                               |  0.9.10 | Supported
reduceSeries(seriesLists, reduceFunction, reduceNode, *reduceMatchers)    |  0.9.14 | Supported
reduce()                                                                  |  - - -  | Short form of reduceSeries()
removeAbovePercentile(seriesList, n)                                      |  0.9.10 | Supported
removeAboveValue(seriesList, n)                                           |  0.9.10 | Supported
removeBelowPercentile(seriesList, n)                                      |  0.9.10 | Supported
removeBelowValue(seriesList, n)                                           |  0.9.10 | Supported
removeBetweenPercentile(seriesList, n)                                    |  1.0.0  |
removeEmptySeries(seriesList)                                             |  1.0.0  | Supported
removeZeroSeries(seriesList)                                              |  0.9.14 | Supported
round                                                                     |  1.1.0  |
scale(seriesList, factor)                                                 |  0.9.9  | Supported
scaleToSeconds(seriesList, seconds)                                       |  0.9.10 | Supported
secondYAxis(seriesList)                                                   |  0.9.10 | Supported
seriesByTag                                                               |  1.1.0  |
setXFilesFactor                                                           |  1.1.0  |
sinFunction(name, amplitude=1, step=60), Short Alias: sin()               |  0.9.9  |
smartSummarize(seriesList, intervalString, func='sum', alignToFrom=False) |  0.9.10 |
sortBy                                                                    |  1.1.0  |
sortByMaxima(seriesList)                                                  |  0.9.9  | Supported
sortByMinima(seriesList)                                                  |  0.9.9  | Supported
sortByName(seriesList)                                                    |  0.9.15 | Supported
sortByTotal(seriesList)                                                   |  0.9.11 | Supported
squareRoot(seriesList)                                                    |  1.0.0  | Supported
stacked(seriesLists, stackName='__DEFAULT__')                             |  0.9.10 | [#74](https://github.com/go-graphite/carbonapi/issues/74)
stddevSeries(*seriesLists)                                                |  0.9.14 | Supported
stdev(seriesList, points, windowTolerance=0.1)                            |  0.9.10 | Supported + alias stddev()
substr(seriesList, start=0, stop=0)                                       |  0.9.9  | Supported
sumSeries(*seriesLists), Short form: sum()                                |  0.9.9  | Supported
sumSeriesWithWildcards(seriesList, *position)                             |  0.9.10 | Supported
summarize(seriesList, intervalString, func='sum', alignToFrom=False)      |  0.9.9  | Supported
threshold(value, label=None, color=None)                                  |  0.9.9  | Supported
timeFunction(name, step=60), Short Alias: time()                          |  0.9.9  | Supported
timeShift(seriesList, timeShift, resetEnd=True)                           |  0.9.11 | Supported
timeSlice(seriesList, startSliceAt, endSliceAt='now')                     |  1.0.0  |
timeStack(seriesList, timeShiftUnit, timeShiftStart, timeShiftEnd)        |  0.9.14 | Supported
[tukeyAbove](https://en.wikipedia.org/wiki/Tukey%27s_range_test)(seriesList, basis, n, interval=0)                              |  not in graphite | Experimental
[tukeyBelow](https://en.wikipedia.org/wiki/Tukey%27s_range_test)(seriesList, basis, n, interval=0)                              |  not in graphite | Experimental
transformNull(seriesList, default=0)                                      |  0.9.10 | Supported
useSeriesAbove(seriesList, value, search, replace)                        |  0.9.10 |
verticalLine(ts, label=None, color=None)                                  |  1.0.0  |
weightedAverage(seriesListAvg, seriesListWeight, node)                    |  1.0.0  |

<a name="functions-features"></a>
## Features of configuration functions
### aliasByPostgre
1. Make config for function with pairs key-string - request
```yaml
enabled: true
database:
  "databaseAlias":
    urlDB: "localhost:5432"
    username: "portgres_user"
    password: "postgres_password"
    nameDB: "database_name"
    keyString:
      "resolve_switch_name_byId":
        varName: "var"
        queryString: "SELECT field_with_switch_name FROM some_table_with_switch_names_id_and_other WHERE field_with_switchID like 'var0';"
        matchString: ".*"
      "resolve_interface_description_from_table":
        varName: "var"
        queryString: "SELECT interface_desc FROM some_table_with_switch_data WHERE field_with_hostname like 'var0' AND field_with_interface_id like 'var1';"
        matchString: ".*"
```

Examples:
We have data series:
```
switches.switchId.CPU1Min
```
We need to get CPU load resolved by switchname, aliasByPostgre( switches.*.CPU1Min, databaseAlias, resolve_switch_name_byId, 1 ) will return series like this:
```
switchnameA
switchnameB
switchnameC
switchnameD
```
We have data series:
```
switches.hostname.interfaceID.scope_of_interface_metrics
```
We want to see interfaces stats sticked to their descriptions, aliasByPostgre(switches.hostname.*.ifOctets.rx, databaseAlias, resolve_interface_description_from_table, 1, 2 )
will return series:
```
InterfaceADesc
InterfaceBDesc
InterfaceCDesc
InterfaceDDesc
```

2. Add to main config path to configuration file
```yaml
functionsConfigs:
        aliasByPostgre: /path/to/funcConfig.yaml
```
-----
