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




## Graphite-web 1.1 compatibility
### Unsupported functions
| Function                                                                  |
| :------------------------------------------------------------------------ |
| aggregate |
| aggregateLine |
| aggregateWithWildcards |
| aliasQuery |
| averageOutsidePercentile |
| events |
| exponentialMovingAverage |
| filterSeries |
| highest |
| holtWintersConfidenceArea |
| identity |
| integralByInterval |
| interpolate |
| lowest |
| minMax |
| movingWindow |
| pct |
| powSeries |
| removeBetweenPercentile |
| round |
| setXFilesFactor |
| sin |
| sinFunction |
| smartSummarize |
| sortBy |
| timeSlice |
| unique |
| verticalLine |
| weightedAverage |
| xFilesFactor |


### Partly supported functions
| Function                 | Incompatibilities                              |
| :------------------------|:---------------------------------------------- |
| holtWintersAberration | parameter not supported: seasonality |
| holtWintersConfidenceBands | parameter not supported: seasonality |
| holtWintersForecast | parameter not supported: seasonality |
| nonNegativeDerivative | parameter not supported: minValue |
| perSecond | parameter not supported: minValue |
| timeShift | parameter not supported: alignDst |
| useSeriesAbove | value: type mismatch: got "integer", should be "string" |

## Supported functions
| Function      | Carbonapi-only                                            |
| :-------------|:--------------------------------------------------------- |
| absolute(seriesList) | no |
| alias(seriesList, newName) | no |
| aliasByMetric(seriesList) | no |
| aliasByNode(seriesList, *nodes) | no |
| aliasByTags(seriesList, *tags) | no |
| aliasSub(seriesList, search, replace) | no |
| alpha(seriesList, alpha) | no |
| applyByNode(seriesList, nodeNum, templateFunction, newName=None) | no |
| areaBetween(seriesList) | no |
| asPercent(seriesList, total=None, *nodes) | no |
| averageAbove(seriesList, n) | no |
| averageBelow(seriesList, n) | no |
| averageSeries(*seriesLists) | no |
| averageSeriesWithWildcards(seriesList, *position) | no |
| avg(*seriesLists) | no |
| cactiStyle(seriesList, system=None, units=None) | no |
| changed(seriesList) | no |
| color(seriesList, theColor) | no |
| consolidateBy(seriesList, consolidationFunc) | no |
| constantLine(value) | no |
| countSeries(*seriesLists) | no |
| cumulative(seriesList) | no |
| currentAbove(seriesList, n) | no |
| currentBelow(seriesList, n) | no |
| dashed(seriesList, dashLength=5) | no |
| delay(seriesList, steps) | no |
| derivative(seriesList) | no |
| diffSeries(*seriesLists) | no |
| divideSeries(dividendSeriesList, divisorSeries) | no |
| divideSeriesLists(dividendSeriesList, divisorSeriesList) | no |
| drawAsInfinite(seriesList) | no |
| exclude(seriesList, pattern) | no |
| fallbackSeries(seriesList, fallback) | no |
| grep(seriesList, pattern) | no |
| group(*seriesLists) | no |
| groupByNode(seriesList, nodeNum, callback='average') | no |
| groupByNodes(seriesList, callback, *nodes) | no |
| groupByTags(seriesList, callback, *tags) | no |
| highestAverage(seriesList, n) | no |
| highestCurrent(seriesList, n) | no |
| highestMax(seriesList, n) | no |
| hitcount(seriesList, intervalString, alignToInterval=False) | no |
| holtWintersAberration(seriesList, delta=3, bootstrapInterval='7d') | no |
| holtWintersConfidenceBands(seriesList, delta=3, bootstrapInterval='7d') | no |
| holtWintersForecast(seriesList, bootstrapInterval='7d') | no |
| integral(seriesList) | no |
| invert(seriesList) | no |
| isNonNull(seriesList) | no |
| keepLastValue(seriesList, limit=inf) | no |
| legendValue(seriesList, *valueTypes) | no |
| limit(seriesList, n) | no |
| lineWidth(seriesList, width) | no |
| linearRegression(seriesList, startSourceAt=None, endSourceAt=None) | no |
| log(seriesList, base=10) | no |
| lowestAverage(seriesList, n) | no |
| lowestCurrent(seriesList, n) | no |
| map(seriesList, *mapNodes) | no |
| mapSeries(seriesList, *mapNodes) | no |
| maxSeries(*seriesLists) | no |
| maximumAbove(seriesList, n) | no |
| maximumBelow(seriesList, n) | no |
| minSeries(*seriesLists) | no |
| minimumAbove(seriesList, n) | no |
| minimumBelow(seriesList, n) | no |
| mostDeviant(seriesList, n) | no |
| movingAverage(seriesList, windowSize, xFilesFactor=None) | no |
| movingMax(seriesList, windowSize, xFilesFactor=None) | no |
| movingMedian(seriesList, windowSize, xFilesFactor=None) | no |
| movingMin(seriesList, windowSize, xFilesFactor=None) | no |
| movingSum(seriesList, windowSize, xFilesFactor=None) | no |
| multiplySeries(*seriesLists) | no |
| multiplySeriesWithWildcards(seriesList, *position) | no |
| nPercentile(seriesList, n) | no |
| nonNegativeDerivative(seriesList, maxValue=None) | no |
| offset(seriesList, factor) | no |
| offsetToZero(seriesList) | no |
| perSecond(seriesList, maxValue=None) | no |
| percentileOfSeries(seriesList, n, interpolate=False) | no |
| pow(seriesList, factor) | no |
| randomWalk(name, step=60) | no |
| randomWalkFunction(name, step=60) | no |
| rangeOfSeries(*seriesLists) | no |
| reduce(seriesLists, reduceFunction, reduceNode, *reduceMatchers) | no |
| reduceSeries(seriesLists, reduceFunction, reduceNode, *reduceMatchers) | no |
| removeAbovePercentile(seriesList, n) | no |
| removeAboveValue(seriesList, n) | no |
| removeBelowPercentile(seriesList, n) | no |
| removeBelowValue(seriesList, n) | no |
| removeEmptySeries(seriesList, xFilesFactor=None) | no |
| scale(seriesList, factor) | no |
| scaleToSeconds(seriesList, seconds) | no |
| secondYAxis(seriesList) | no |
| seriesByTag(*tagExpressions) | no |
| sortByMaxima(seriesList) | no |
| sortByMinima(seriesList) | no |
| sortByName(seriesList, natural=False, reverse=False) | no |
| sortByTotal(seriesList) | no |
| squareRoot(seriesList) | no |
| stacked(seriesLists, stackName='__DEFAULT__') | no |
| stddevSeries(*seriesLists) | no |
| stdev(seriesList, points, windowTolerance=0.1) | no |
| substr(seriesList, start=0, stop=0) | no |
| sum(*seriesLists) | no |
| sumSeries(*seriesLists) | no |
| sumSeriesWithWildcards(seriesList, *position) | no |
| summarize(seriesList, intervalString, func='sum', alignToFrom=False) | no |
| threshold(value, label=None, color=None) | no |
| time(name, step=60) | no |
| timeFunction(name, step=60) | no |
| timeShift(seriesList, timeShift, resetEnd=True, alignDST=False) | no |
| timeStack(seriesList, timeShiftUnit='1d', timeShiftStart=0, timeShiftEnd=7) | no |
| transformNull(seriesList, default=0, referenceSeries=None) | no |
| useSeriesAbove(seriesList, value, search, replace) | no |
| diffSeriesLists(firstSeriesList, secondSeriesList) | yes |
| exponentialWeightedMovingAverage(seriesList, alpha) | yes |
| exponentialWeightedMovingAverage(seriesList, alpha) | yes |
| fft(seriesList, mode) | yes |
| ifft(seriesList, phaseSeriesList) | yes |
| isNotNull(seriesList) | yes |
| kolmogorovSmirnovTest2(seriesList, seriesList, windowSize) | yes |
| ksTest2(seriesList, seriesList, windowSize) | yes |
| log(seriesList, base=10) | yes |
| lowPass(seriesList, cutPercent) | yes |
| lpf(seriesList, cutPercent) | yes |
| maxSeries(*seriesLists) | yes |
| minSeries(*seriesLists) | yes |
| multiplySeriesLists(sourceSeriesList, factorSeriesList) | yes |
| pearson(seriesList, seriesList, windowSize) | yes |
| pearsonClosest(seriesList, seriesList, n, direction) | yes |
| polyfit(seriesList, degree=1, offset="0d") | yes |
| powSeriesLists(sourceSeriesList, factorSeriesList) | yes |
| removeZeroSeries(seriesList, xFilesFactor=None) | yes |
| stdev(seriesList, points, windowTolerance=0.1) | yes |
| tukeyAbove(seriesList, basis, n, interval=0) | yes |
| tukeyBelow(seriesList, basis, n, interval=0) | yes |
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
