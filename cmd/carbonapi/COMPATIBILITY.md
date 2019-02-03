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




## Unsupported functions
| Function                                                                  |
| :------------------------------------------------------------------------ |
| interpolate |
| useSeriesAbove |
| minMax |
| xFilesFactor |
| smartSummarize |
| averageOutsidePercentile |
| filterSeries |
| removeBetweenPercentile |
| identity |
| powSeries |
| aggregateWithWildcards |
| round |
| sin |
| setXFilesFactor |
| timeSlice |
| groupByTags |
| aggregateLine |
| sinFunction |
| pct |
| integralByInterval |
| unique |
| events |
| lowest |
| aliasQuery |
| exponentialMovingAverage |
| sortBy |
| verticalLine |
| holtWintersConfidenceArea |
| aggregate |
| weightedAverage |
| highest |
| movingWindow |


## Partly supported functions
| Function                 | Incompatibilities                              |
| :------------------------|:---------------------------------------------- |
| aliasByTags | nodes: parameter missing |

## Supported functions
| Function      | Carbonapi-only                                            |
| :-------------|:--------------------------------------------------------- |
| nPercentile(seriesList, n) | no |
| averageBelow(seriesList, n) | no |
| countSeries(*seriesLists) | no |
| alias(seriesList, newName) | no |
| cactiStyle(seriesList, system=None, units=None) | no |
| timeFunction(name, step=60) | no |
| groupByNode(seriesList, nodeNum, callback='average') | no |
| sum(*seriesLists) | no |
| log(seriesList, base=10) | no |
| offset(seriesList, factor) | no |
| alpha(seriesList, alpha) | no |
| lineWidth(seriesList, width) | no |
| groupByNodes(seriesList, callback, *nodes) | no |
| mapSeries(seriesList, *mapNodes) | no |
| invert(seriesList) | no |
| grep(seriesList, pattern) | no |
| highestMax(seriesList, n) | no |
| removeBelowPercentile(seriesList, n) | no |
| map(seriesList, *mapNodes) | no |
| percentileOfSeries(seriesList, n, interpolate=False) | no |
| reduceSeries(seriesLists, reduceFunction, reduceNode, *reduceMatchers) | no |
| nonNegativeDerivative(seriesList, maxValue=None, minValue=None) | no |
| diffSeries(*seriesLists) | no |
| derivative(seriesList) | no |
| drawAsInfinite(seriesList) | no |
| constantLine(value) | no |
| holtWintersConfidenceBands(seriesList, delta=3, bootstrapInterval='7d', seasonality='1d') | no |
| maximumBelow(seriesList, n) | no |
| minimumAbove(seriesList, n) | no |
| removeEmptySeries(seriesList, xFilesFactor=None) | no |
| color(seriesList, theColor) | no |
| isNonNull(seriesList) | no |
| keepLastValue(seriesList, limit=inf) | no |
| movingAverage(seriesList, windowSize, xFilesFactor=None) | no |
| lowestAverage(seriesList, n) | no |
| sortByTotal(seriesList) | no |
| aliasByNode(seriesList, *nodes) | no |
| substr(seriesList, start=0, stop=0) | no |
| minSeries(*seriesLists) | no |
| delay(seriesList, steps) | no |
| linearRegression(seriesList, startSourceAt=None, endSourceAt=None) | no |
| currentBelow(seriesList, n) | no |
| limit(seriesList, n) | no |
| rangeOfSeries(*seriesLists) | no |
| movingMax(seriesList, windowSize, xFilesFactor=None) | no |
| sortByName(seriesList, natural=False, reverse=False) | no |
| aliasByMetric(seriesList) | no |
| holtWintersAberration(seriesList, delta=3, bootstrapInterval='7d', seasonality='1d') | no |
| removeAboveValue(seriesList, n) | no |
| fallbackSeries(seriesList, fallback) | no |
| dashed(seriesList, dashLength=5) | no |
| maxSeries(*seriesLists) | no |
| sumSeries(*seriesLists) | no |
| highestAverage(seriesList, n) | no |
| mostDeviant(seriesList, n) | no |
| sortByMinima(seriesList) | no |
| randomWalk(name, step=60) | no |
| lowestCurrent(seriesList, n) | no |
| removeBelowValue(seriesList, n) | no |
| avg(*seriesLists) | no |
| consolidateBy(seriesList, consolidationFunc) | no |
| currentAbove(seriesList, n) | no |
| secondYAxis(seriesList) | no |
| threshold(value, label=None, color=None) | no |
| averageAbove(seriesList, n) | no |
| averageSeries(*seriesLists) | no |
| movingSum(seriesList, windowSize, xFilesFactor=None) | no |
| cumulative(seriesList) | no |
| hitcount(seriesList, intervalString, alignToInterval=False) | no |
| randomWalkFunction(name, step=60) | no |
| pow(seriesList, factor) | no |
| summarize(seriesList, intervalString, func='sum', alignToFrom=False) | no |
| stacked(seriesLists, stackName='__DEFAULT__') | no |
| asPercent(seriesList, total=None, *nodes) | no |
| multiplySeriesWithWildcards(seriesList, *position) | no |
| stddevSeries(*seriesLists) | no |
| transformNull(seriesList, default=0, referenceSeries=None) | no |
| areaBetween(seriesList) | no |
| scale(seriesList, factor) | no |
| highestCurrent(seriesList, n) | no |
| removeAbovePercentile(seriesList, n) | no |
| legendValue(seriesList, *valueTypes) | no |
| changed(seriesList) | no |
| absolute(seriesList) | no |
| divideSeriesLists(dividendSeriesList, divisorSeriesList) | no |
| scaleToSeconds(seriesList, seconds) | no |
| squareRoot(seriesList) | no |
| holtWintersForecast(seriesList, bootstrapInterval='7d', seasonality='1d') | no |
| sortByMaxima(seriesList) | no |
| multiplySeries(*seriesLists) | no |
| reduce(seriesLists, reduceFunction, reduceNode, *reduceMatchers) | no |
| time(name, step=60) | no |
| divideSeries(dividendSeriesList, divisorSeries) | no |
| integral(seriesList) | no |
| movingMedian(seriesList, windowSize, xFilesFactor=None) | no |
| movingMin(seriesList, windowSize, xFilesFactor=None) | no |
| stdev(seriesList, points, windowTolerance=0.1) | no |
| maximumAbove(seriesList, n) | no |
| seriesByTag(*tagExpressions) | no |
| applyByNode(seriesList, nodeNum, templateFunction, newName=None) | no |
| offsetToZero(seriesList) | no |
| perSecond(seriesList, maxValue=None, minValue=None) | no |
| timeShift(seriesList, timeShift, resetEnd=True, alignDST=False) | no |
| timeStack(seriesList, timeShiftUnit='1d', timeShiftStart=0, timeShiftEnd=7) | no |
| exclude(seriesList, pattern) | no |
| minimumBelow(seriesList, n) | no |
| aliasSub(seriesList, search, replace) | no |
| averageSeriesWithWildcards(seriesList, *position) | no |
| group(*seriesLists) | no |
| sumSeriesWithWildcards(seriesList, *position) | no |
| stdev(seriesList, points, windowTolerance=0.1) | yes |
| maxSeries(*seriesLists) | yes |
| pearsonClosest(seriesList, seriesList, n, direction) | yes |
| tukeyBelow(seriesList, basis, n, interval=0) | yes |
| polyfit(seriesList, degree=1, offset="0d") | yes |
| powSeriesLists(sourceSeriesList, factorSeriesList) | yes |
| lpf(seriesList, cutPercent) | yes |
| minSeries(*seriesLists) | yes |
| tukeyAbove(seriesList, basis, n, interval=0) | yes |
| exponentialWeightedMovingAverage(seriesList, alpha) | yes |
| multiplySeriesLists(sourceSeriesList, factorSeriesList) | yes |
| aboveSeries(seriesList, value, search, replace) | yes |
| exponentialWeightedMovingAverage(seriesList, alpha) | yes |
| fft(seriesList, mode) | yes |
| lowPass(seriesList, cutPercent) | yes |
| ksTest2(seriesList, seriesList, windowSize) | yes |
| diffSeriesLists(firstSeriesList, secondSeriesList) | yes |
| kolmogorovSmirnovTest2(seriesList, seriesList, windowSize) | yes |
| removeZeroSeries(seriesList, xFilesFactor=None) | yes |
| ifft(seriesList, phaseSeriesList) | yes |
| log(seriesList, base=10) | yes |
| isNotNull(seriesList) | yes |
| pearson(seriesList, seriesList, windowSize) | yes |
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
