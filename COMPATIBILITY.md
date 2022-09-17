# CarbonAPI compatibility with Graphite

Topics:
* [Default settings](#default-settings)
* [URI Parameters](#uri-params)
* [Graphite-web 1.1 Compatibility](#graphite-web-11-compatibility)
* [Supported Functions](#supported-functions)
* [Features of configuration functions](#functions-features)

<a name="default-settings"></a>
## Default Settings

### Default Line Colors
Default colors for png or svg rendering intentionally specified like it is in graphite-web 1.1.7

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
* `pixelRatio` : (1.0)
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




## Graphite-web 1.1.7 compatibility
### Unsupported functions
| Function                                                                  |
| :------------------------------------------------------------------------ |
| aliasQuery |
| events |
| holtWintersConfidenceArea |
| pct |

### Partly supported functions
| Function                 | Incompatibilities                              |
| :------------------------|:---------------------------------------------- |
| aggregate | parameter not supported: xFilesFactor |
| asPercent | total: type mismatch: got seriesList, should be any |
| averageAbove | n: type mismatch: got integer, should be float |
| averageBelow | n: type mismatch: got integer, should be float |
| currentAbove | n: type mismatch: got integer, should be float |
| currentBelow | n: type mismatch: got integer, should be float |
| groupByNode | callback: different amount of parameters, `[averageSeries averageSeriesWithWildcards countSeries current diffSeries maxSeries minSeries multiplySeries multiplySeriesWithWildcards powSeries rangeOf rangeOfSeries stddevSeries sumSeries sumSeriesWithWildcards]` are missing
callback: type mismatch: got aggFunc, should be aggOrSeriesFunc |
| groupByNodes | callback: different amount of parameters, `[averageSeries averageSeriesWithWildcards countSeries current diffSeries maxSeries minSeries multiplySeries multiplySeriesWithWildcards powSeries rangeOf rangeOfSeries stddevSeries sumSeries sumSeriesWithWildcards]` are missing
callback: type mismatch: got aggFunc, should be aggOrSeriesFunc |
| groupByTags | callback: different amount of parameters, `[averageSeries averageSeriesWithWildcards countSeries current diffSeries maxSeries minSeries multiplySeries multiplySeriesWithWildcards powSeries rangeOf rangeOfSeries stddevSeries sumSeries sumSeriesWithWildcards]` are missing
callback: type mismatch: got aggFunc, should be aggOrSeriesFunc |
| highest | func: type mismatch: got string, should be aggFunc |
| holtWintersAberration | parameter not supported: seasonality |
| holtWintersConfidenceBands | parameter not supported: seasonality |
| holtWintersForecast | parameter not supported: seasonality |
| integralByInterval | parameter not supported: intervalUnit |
| interpolate | limit: type mismatch: got float, should be intOrInf
limit: default value mismatch: got (empty), should be "Infinity" |
| keepLastValue | limit: type mismatch: got integer, should be intOrInf
limit: default value mismatch: got "INF", should be "Infinity" |
| legendValue | valuesTypes: different amount of parameters, `[averageSeries avgSeries avg_zeroSeries binary countSeries current currentSeries diffSeries lastSeries maxSeries medianSeries minSeries multiplySeries rangeOf rangeOfSeries rangeSeries si stddevSeries sumSeries totalSeries]` are missing |
| lowest | func: type mismatch: got string, should be aggFunc |
| maximumAbove | n: type mismatch: got integer, should be float |
| maximumBelow | n: type mismatch: got integer, should be float |
| minimumAbove | n: type mismatch: got integer, should be float |
| minimumBelow | n: type mismatch: got integer, should be float |
| nPercentile | n: type mismatch: got integer, should be float |
| percentileOfSeries | n: type mismatch: got integer, should be float |
| removeAbovePercentile | n: type mismatch: got integer, should be float |
| removeAboveValue | n: type mismatch: got integer, should be float |
| removeBelowPercentile | n: type mismatch: got integer, should be float |
| removeBelowValue | n: type mismatch: got integer, should be float |
| round | precision: default value mismatch: got (empty), should be 0 |
| scaleToSeconds | seconds: type mismatch: got integer, should be float |
| smartSummarize | func: different amount of parameters, `[current rangeOf]` are missing
alignTo: different amount of parameters, `[<nil> days hours minutes months seconds weeks years]` are missing
alignTo: type mismatch: got interval, should be string |
| sortBy | func: different amount of parameters, `[average avg avg_zero count current diff last max median min multiply range rangeOf stddev sum total]` are missing
func: default value mismatch: got (empty), should be "average"
reverse: default value mismatch: got (empty), should be false |
| summarize | func: different amount of parameters, `[current rangeOf]` are missing |
| timeShift | parameter not supported: alignDst |
| timeSlice | endSliceAt: type mismatch: got interval, should be date |

## Supported functions
| Function                                                                                                | Carbonapi-only |
|:--------------------------------------------------------------------------------------------------------|:---------------|
| absolute(seriesList)                                                                                    | no             |
| add(seriesList, constant)                                                                               | no             |
| aggregate(seriesList, func, xFilesFactor=None)                                                          | no             |
| aggregateLine(seriesList, func='average', keepStep=False)                                               | no             |
| aggregateWithWildcards(seriesList, func, *positions)                                                    | no             |
| alias(seriesList, newName)                                                                              | no             |
| aliasByMetric(seriesList)                                                                               | no             |
| aliasByNode(seriesList, *nodes)                                                                         | no             |
| aliasByTags(seriesList, *tags)                                                                          | no             |
| aliasSub(seriesList, search, replace)                                                                   | no             |
| alpha(seriesList, alpha)                                                                                | no             |
| applyByNode(seriesList, nodeNum, templateFunction, newName=None)                                        | no             |
| areaBetween(seriesList)                                                                                 | no             |
| asPercent(seriesList, total=None, *nodes)                                                               | no             |
| averageAbove(seriesList, n)                                                                             | no             |
| averageBelow(seriesList, n)                                                                             | no             |
| averageOutsidePercentile(seriesList, n)                                                                 | no             |
| averageSeries(*seriesLists)                                                                             | no             |
| averageSeriesWithWildcards(seriesList, *position)                                                       | no             |
| avg(*seriesLists)                                                                                       | no             |
| cactiStyle(seriesList, system=None, units=None)                                                         | no             |
| changed(seriesList)                                                                                     | no             |
| color(seriesList, theColor)                                                                             | no             |
| consolidateBy(seriesList, consolidationFunc)                                                            | no             |
| constantLine(value)                                                                                     | no             |
| countSeries(*seriesLists)                                                                               | no             |
| cumulative(seriesList)                                                                                  | no             |
| currentAbove(seriesList, n)                                                                             | no             |
| currentBelow(seriesList, n)                                                                             | no             |
| dashed(seriesList, dashLength=5)                                                                        | no             |
| delay(seriesList, steps)                                                                                | no             |
| derivative(seriesList)                                                                                  | no             |
| diffSeries(*seriesLists)                                                                                | no             |
| divideSeries(dividendSeriesList, divisorSeries)                                                         | no             |
| divideSeriesLists(dividendSeriesList, divisorSeriesList)                                                | no             |
| drawAsInfinite(seriesList)                                                                              | no             |
| exclude(seriesList, pattern)                                                                            | no             |
| exp(seriesList)                                                                                         | no             |
| exponentialMovingAverage(seriesList, windowSize)                                                        | no             |
| fallbackSeries(seriesList, fallback)                                                                    | no             |
| filterSeries(seriesList, func, operator, threshold)                                                     | no             |
| grep(seriesList, pattern)                                                                               | no             |
| group(*seriesLists)                                                                                     | no             |
| groupByNode(seriesList, nodeNum, callback='average')                                                    | no             |
| groupByNodes(seriesList, callback, *nodes)                                                              | no             |
| groupByTags(seriesList, callback, *tags)                                                                | no             |
| highest(seriesList, n=1, func='average')                                                                | no             |
| highestAverage(seriesList, n)                                                                           | no             |
| highestCurrent(seriesList, n)                                                                           | no             |
| highestMax(seriesList, n)                                                                               | no             |
| hitcount(seriesList, intervalString, alignToInterval=False)                                             | no             |
| holtWintersAberration(seriesList, delta=3, bootstrapInterval='7d')                                      | no             |
| holtWintersConfidenceBands(seriesList, delta=3, bootstrapInterval='7d')                                 | no             |
| holtWintersForecast(seriesList, bootstrapInterval='7d')                                                 | no             |
| logit(seriesList)                                                                                       | no             |
| identity(name)                                                                                          | no             |
| integral(seriesList)                                                                                    | no             |
| integralByInterval(seriesList, intervalString)                                                          | no             |
| interpolate(seriesList, limit)                                                                          | no             |
| invert(seriesList)                                                                                      | no             |
| isNonNull(seriesList)                                                                                   | no             |
| keepLastValue(seriesList, limit=inf)                                                                    | no             |
| legendValue(seriesList, *valueTypes)                                                                    | no             |
| limit(seriesList, n)                                                                                    | no             |
| lineWidth(seriesList, width)                                                                            | no             |
| linearRegression(seriesList, startSourceAt=None, endSourceAt=None)                                      | no             |
| log(seriesList, base=10)                                                                                | no             |
| lowest(seriesList, n=1, func='average')                                                                 | no             |
| lowestAverage(seriesList, n)                                                                            | no             |
| lowestCurrent(seriesList, n)                                                                            | no             |
| map(seriesList, *mapNodes)                                                                              | no             |
| mapSeries(seriesList, *mapNodes)                                                                        | no             |
| maxSeries(*seriesLists)                                                                                 | no             |
| maximumAbove(seriesList, n)                                                                             | no             |
| maximumBelow(seriesList, n)                                                                             | no             |
| minSeries(*seriesLists)                                                                                 | no             |
| minimumAbove(seriesList, n)                                                                             | no             |
| minimumBelow(seriesList, n)                                                                             | no             |
| minMax(seriesList)                                                                                      | no             |
| mostDeviant(seriesList, n)                                                                              | no             |
| movingAverage(seriesList, windowSize, xFilesFactor=None)                                                | no             |
| movingMax(seriesList, windowSize, xFilesFactor=None)                                                    | no             |
| movingMedian(seriesList, windowSize, xFilesFactor=None)                                                 | no             |
| movingMin(seriesList, windowSize, xFilesFactor=None)                                                    | no             |
| movingSum(seriesList, windowSize, xFilesFactor=None)                                                    | no             |
| movingWindow(seriesList, windowSize, func='average', xFilesFactor=None)                                 | no             |
| multiplySeries(*seriesLists)                                                                            | no             |
| multiplySeriesWithWildcards(seriesList, *position)                                                      | no             |
| nPercentile(seriesList, n)                                                                              | no             |
| nonNegativeDerivative(seriesList, maxValue=None)                                                        | no             |
| offset(seriesList, factor)                                                                              | no             |
| offsetToZero(seriesList)                                                                                | no             |
| perSecond(seriesList, maxValue=None)                                                                    | no             |
| percentileOfSeries(seriesList, n, interpolate=False)                                                    | no             |
| pow(seriesList, factor)                                                                                 | no             |
| powSeries(*seriesLists)                                                                                 | no             |
| randomWalk(name, step=60)                                                                               | no             |
| randomWalkFunction(name, step=60)                                                                       | no             |
| rangeOfSeries(*seriesLists)                                                                             | no             |
| reduce(seriesLists, reduceFunction, reduceNode, *reduceMatchers)                                        | no             |
| reduceSeries(seriesLists, reduceFunction, reduceNode, *reduceMatchers)                                  | no             |
| removeAbovePercentile(seriesList, n)                                                                    | no             |
| removeAboveValue(seriesList, n)                                                                         | no             |
| removeBelowPercentile(seriesList, n)                                                                    | no             |
| removeBelowValue(seriesList, n)                                                                         | no             |
| removeBelowValue(seriesList, n)                                                                         | no             |
| removeEmptySeries(seriesList, xFilesFactor=None)                                                        | no             |
| round(seriesList, precision)                                                                            | no             |
| scale(seriesList, factor)                                                                               | no             |
| scaleToSeconds(seriesList, seconds)                                                                     | no             |
| secondYAxis(seriesList)                                                                                 | no             |
| seriesByTag(*tagExpressions)                                                                            | no             |
| setXFilesFactor(seriesList, xFilesFactor)                                                               | no             |
| sigmoid(seriesList)                                                                                     | no             |
| sinFunction(seriesList, amplitude=1, step=60)                                                           | no             |
| smartSummarize(seriesList, intervalString, func='sum', alignTo=None)                                    | no             |
| sortBy(seriesList, func='average', reverse=False)                                                       | no             |
| sortByMaxima(seriesList)                                                                                | no             |
| sortByMinima(seriesList)                                                                                | no             |
| sortByName(seriesList, natural=False, reverse=False)                                                    | no             |
| sortByTotal(seriesList)                                                                                 | no             |
| squareRoot(seriesList)                                                                                  | no             |
| stacked(seriesLists, stackName='__DEFAULT__')                                                           | no             |
| stddevSeries(*seriesLists)                                                                              | no             |
| stdev(seriesList, points, windowTolerance=0.1)                                                          | no             |
| substr(seriesList, start=0, stop=0)                                                                     | no             |
| sum(*seriesLists)                                                                                       | no             |
| sumSeries(*seriesLists)                                                                                 | no             |
| sumSeriesWithWildcards(seriesList, *position)                                                           | no             |
| summarize(seriesList, intervalString, func='sum', alignToFrom=False)                                    | no             |
| threshold(value, label=None, color=None)                                                                | no             |
| time(name, step=60)                                                                                     | no             |
| timeFunction(name, step=60)                                                                             | no             |
| timeShift(seriesList, timeShift, resetEnd=True, alignDST=False)                                         | no             |
| timeSlice(seriesList, startSliceAt, endSliceAt='now')                                                   | no             |
| timeStack(seriesList, timeShiftUnit='1d', timeShiftStart=0, timeShiftEnd=7)                             | no             |
| toLowerCase(seruesList)                                                                                 | no             |
| transformNull(seriesList, default=0, referenceSeries=None)                                              | no             |
| unique(*seriesLists)                                                                                    | no             |
| useSeriesAbove(seriesList, value, search, replace)                                                      | no             |
| weightedAverage(seriesListAvg, seriesListWeight, *nodes)                                                | no             |
| verticalLine(ts, label=None, color=None)                                                                | no             |
| aliasByBase64(seriesList)                                                                               | yes            |
| aliasByPostgres(seriesList, *nodes)                                                                     | yes            |
| aliasByRedis(seriesList. keyName)                                                                       | yes            |
| baseline(seriesList, timeShiftUnit, timeShiftStart, timeShiftEnd, [maxAbsentPercent, minAvg])           | yes            |
| baselineAberration(seriesList, timeShiftUnit, timeShiftStart, timeShiftEnd, [maxAbsentPercent, minAvg]) | yes            |
| count(*seriesLists)                                                                                     | yes            |
| diff(*seriesLists)                                                                                      | yes            |
| diffSeriesLists(firstSeriesList, secondSeriesList)                                                      | yes            |
| exponentialWeightedMovingAverage(seriesList, alpha)                                                     | yes            |
| exponentialWeightedMovingAverage(seriesList, alpha)                                                     | yes            |
| fft(seriesList, mode)                                                                                   | yes            |
| heatMap(seriesList)                                                                                     | yes            |
| highestMin(seriesList, n)                                                                               | yes            |
| ifft(seriesList, phaseSeriesList)                                                                       | yes            |
| integralWithReset(seriesList, resettingSeries)                                                          | yes            |
| isNotNull(seriesList)                                                                                   | yes            |
| kolmogorovSmirnovTest2(seriesList, seriesList, windowSize)                                              | yes            |
| ksTest2(seriesList, seriesList, windowSize)                                                             | yes            |
| log(seriesList, base=10)                                                                                | yes            |
| lowPass(seriesList, cutPercent)                                                                         | yes            |
| lowestMax(seriesList, n)                                                                                | yes            |
| lowestMin(seriesList, n)                                                                                | yes            |
| lpf(seriesList, cutPercent)                                                                             | yes            |
| maxSeries(*seriesLists)                                                                                 | yes            |
| minSeries(*seriesLists)                                                                                 | yes            |
| multiply(*seriesLists)                                                                                  | yes            |
| multiplySeriesLists(sourceSeriesList, factorSeriesList)                                                 | yes            |
| pearson(seriesList, seriesList, windowSize)                                                             | yes            |
| pearsonClosest(seriesList, seriesList, n, direction)                                                    | yes            |
| polyfit(seriesList, degree=1, offset="0d")                                                              | yes            |
| powSeriesLists(sourceSeriesList, factorSeriesList)                                                      | yes            |
| removeZeroSeries(seriesList, xFilesFactor=None)                                                         | yes            |
| scale(seriesList, factor)                                                                               | yes            |
| slo(seriesList, interval, method, value)                                                                | yes            |
| sloErrorBudget(seriesList, interval, method, value, objective)                                          | yes            |
| stddev(*seriesLists)                                                                                    | yes            |
| timeShiftByMetric(seriesList, markSource, versionRankIndex)                                             | yes            |
| stddev(*seriesLists)                                                                                    | yes            |
| tukeyAbove(seriesList, basis, n, interval=0)                                                            | yes            |
| tukeyBelow(seriesList, basis, n, interval=0)                                                            | yes            |
<a name="functions-features"></a>
## Features of configuration functions
### aliasByPostgres
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

#### Examples

We have data series:
```
switches.switchId.CPU1Min
```
We need to get CPU load resolved by switchname, aliasByPostgres( switches.*.CPU1Min, databaseAlias, resolve_switch_name_byId, 1 ) will return series like this:
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
We want to see interfaces stats sticked to their descriptions, aliasByPostgres(switches.hostname.*.ifOctets.rx, databaseAlias, resolve_interface_description_from_table, 1, 2 )
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
        aliasByPostgres: /path/to/funcConfig.yaml
```
-----
