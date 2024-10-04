Changes
=================================================

[Fix] - bugfix

**[Breaking]** - breaking change

[Feature] - new feature

[Improvement] - non-breaking improvement

[Code] - code quality related change that shouldn't make any significant difference for end-user

[Config] - related to config options or default parameters

[Build] - only related to ways how someone can build application

CHANGELOG
---------

**0.17.0**

 - [Feature] return error on partial targets fetch
 - [Feature] add a config option to pass consolidateBy to the storage backend
 - [Feature] add -exact-config command line argument
 - [Improvement] :bangbang: refactor for avoid global evaluator usage
 - [Improvement] Add movingWindow to list of functions that might adjust timerange
 - [Improvement] tags/autoComplete: return detailed error code instead of 500
 - [Improvement] MaxDataPoints consolidation: support nudging for consistent bucketing
 - [Fix] incorrect positional parameters
 - [Fix] transformNull name tag
 - [Fix] Deadlock on uninitialized (nil) limiter
 - [Fix] Check if query doesn't exceed allowed length limit
 - [Fix] return code for render_handler
 - [Fix] gracefully shutdown http servers
 - [Fix] sortBy: substitute NaN values for negative infinity
 - [Fix]  validate consolidateBy arguments
 - [Fix] fix `"name"` tag overrides in various functions
 - [Fix] runtime error highest current
 - [Fix] PromethizeTagValue panic on empty value


**0.16.1**
 - [Build] Update build version of golang to 1.21.0
 - [Improvement] Better error messages
 - [Improvement] Implement /metrics/expand API
 - [Improvement] Implement aliasQuery
 - [Fix] aliasByNode/Tag should not disacrd tags anymore
 - [Fix] VicotriaMetrics cluster now should support autocomplete for tags
 - [Fix] ConstantSeries should now work as in graphite-web
 - [Fix] Fix logic of hitcount function (should behave like in graphite-web)
 - [Fix] alignToInterval should work well in hitcount function
 - [Fix] smartSummarize should handle stop time properly
 - [Fix] smartSummarize should now handle alignTo properly
 - [Fix] smartSummarize should handle consolidation functiosn better
 - [Fix] summarize should work as in graphite-web for functions that have only NaN values
 - [Fix] toUpper and toLower now should work correctly with function names
 - [Fix] Skip all whitespace during expression parsing
 - [Fix] time units now case insensitive
 - [Fix] use specified bootstrapInterval to adjust start time in holtWinters* class of functions
 - [Fix] groupByNode should return exepcted amount of series
 - [Fix] multiple fixes for exponentialMovingAverage
 - [Fix] holtWinters supports seasonality argument
 - [Fix] fix holtWintersAbberation behavior
 - [Fix] fix behavior of Below function
 - [Fix] fix sin and exp functions (description was not accurate or correct)

**0.16.0.1**
 - [Build] Fix automation that builds docker images
 - [Build] Add rockylinux-9 packages (RHEL 9)
 - [Build] Update build version of golang to 1.19.4

**0.16.0**
 - [Improvement] Deprecate and remove carbonzipper binary (thx to @msaf1980)
 - [Improvement] Remove deprecated carbonsearch support
 - [Improvement] Refactor HTTP client (used to talk to databases) to properly do keepAlive and respect specified amount of connections
 - [Improvement] HTTP client should now support mTLS checking
 - [Improvement] Listeners now support TLS (including mTLS)
 - [Improvement] Update all vendored libraries to their latest stable version
 - [Code] fix various panics in tags and use copy tags to avoid mutating input (thx to @npazosmendez)
 - [Feature] Port `join` function from Avito carbonapi fork (https://github.com/el-yurchito/carbonapi/commit/bccdb90a90492314c18696eb4f064b818c5fff70 and https://github.com/el-yurchito/carbonapi/commit/47177771d60b7af0f5ac865634ecf3d1c3aa2802)
 - [Feature] Implement exp() function (thx to @carrieedwards)
 - [Feature] Implement logit() function (thx to @carrieedwards)
 - [Feature] Implement unique() function (thx to @carrieedwards)
 - [Feature] Implement sinFunction() function (thx to @carrieedwards)
 - [Feature] Implement legendValue() function (thx to @carrieedwards)
 - [Feature] Implement aggregateWithWildcards() function (thx to @carrieedwards)
 - [Feature] Implement powSeries() function (thx to @leizor)
 - [Feature] Implement averageOutsidePercentile() function (thx to @carrieedwards)
 - [Feature] Implement sigmoid() function (thx to @carrieedwards)
 - [Feature] Support for xFilesFactor and setXFilesFactor (thx to @carrieedwards)
 - [Feature] Implement removeBetweenPercentile() function (thx to @carrieedwards)
 - [Feature] Implement identity() function (thx to @carrieedwards)
 - [Feature] Implement minMax() function (thx to @carrieedwards)
 - [Feature] Implement aggregateSeriesLists() function (thx to @leizor)
 - [Feature] Implement movingWindow() function (thx to @carrieedwards)
 - [Feature] Implement exponentialMovingAverage() function (thx to @carrieedwards)
 - [Feature] Implement verticalLine() function (thx to @leizor)
 - [Feature] Implement toLowerCase() function (thx to @carrieedwards)
 - [Feature] Implement toUpperCase() function (thx to @carrieedwards)
 - [Feature] Implement sumSeriesLists() function (thx to @carrieedwards)
 - [Feature] Implement compressPeriodicGaps() function (thx to @carrieedwards)
 - [Feature] Implement holtWintersConfidenceArea() function (thx to @carrieedwards)
 - [Fix] Numerous compatibility fixes for functions, mostly for graphite-web compatibility (thx to @carrieedwards and @npazosmendez, @msaf1980)
 - [Fix] Refactor scaling and re-aligning series (thx to @msaf1980)
 - [Fix] Fix logic that extracts names when metric name contains non-ascii characters (thx to @msaf1980)
 - [Fix] Parser now accepts escaped characters
 - [Fix] IronDB adjuststep logic now works properly
 - [Fix] Fix tag extractions in seriesByTag (thx to @msaf1980)
 - [Fix] Fix tag handling in nested functions (thx to @carrieedwards)

**0.15.6**
 - [Improvement] Significant improvement in sorting metrics from backends (thx to @Felixoid)
 - [Fix] Get rid of bashism in postinst script (thx to @Glandos)

**0.15.5**
 - **[Security]** Update vendored dependency for GoGo Proto that should fix improper index validation (CVE-2021-3121)
 - [Feature] carbonapi supports `-check-config` flag just to run config validation (thx to @msaf1980)
 - [Feature] 2-level cache lifetime support (see docs for more information) (thx to @msaf1980)
 - [Feature] Support IronDB backend (thx to @deniszh)
 - [Improvement] Debug logs should now log a bit more information about metrics
 - [Improvement] Better performance for metrics/find on relatively new VictoriaMetrics (thx to @iklfst)
 - [Improvement] Add support for Cluster Tenant ID for VictoriaMetrics backend (thx to @iklfst)
 - [Fix] asPercent handles empty values correctly (thx to @msaf1980)
 - [Fix] Cairo functions supports context passing now (thx to @faceair)
 - [Fix] Handling multi-fetch request parameters for VictoriaMetrics backend respects start & stop Time as well as step for every metric (thx to @alexey-mylnkov)
 - [Fix] Fix a bug in removeEmptySeries
 - [Fix] Properly pass step to VictoriaMetrics (thx to @tantra35)
 - [Fix] Fix substr function to behave in a way like graphite-web
 - [Fix] timeShift properly hanldes resetEnd parameter (thx to @msaf1980)
 - [Fix] Fix `moving*` to work properly when step is greater than window size (thx to @jonasbleyl)
 - [Fix] Pass request UUID as header (thx to @msaf1980)
 - [Fix] `groupByNode(s)` supports negative node indecies (thx to @Felixoid)

**0.15.4**
 - [Fix] zipper requests net error (encapsulated) (thx to @msaf1980)
 - [Fix] preserve original points for some functions (thx to @msaf1980)
 - [Fix] aliasByNode: strip function in name (thx to @msaf1980)
 - [Improvement] build: allow to replace fpm with cli-compability program (thx to @msaf1980)
 - [Improvement] build: equal version for carbonapi/carbonzipper package (thx to @msaf1980)

**0.15.3**
 - [Fix] Time parsing is now closer to how graphite do it
 - [Fix] Aggregation functions now scale input to common step (thx to @Felixoid)
 - [Fix] Fix the way how partialy retrived requests that had globs were handled
 - [Fix] Fix resolver to accept `0.0.0.0` and `[::]` as valid addresses for listeners
 - [Fix] Handling of unicode characters when common table is added.
 - [Fix] AlignVluaes in Prometheus and VictoriaMetrics backends shouldn't drop last value (thx to @limpasha)
 - [Fix] maxDataPoints should be passed to Prometheus and VictoriaMetrics\
 - [Fix] Various packaging problems (e.x. logrotate) (thx to @deniszh)
 - [Improvement] xFilesFactor support in SUmmarizeValues
 - [Improvement] Port functions that were implemented in avito's fork of carbonapi (aliasByHash, lowestMin, highestMin, integralWithReset and much more): https://github.com/go-graphite/carbonapi/commit/322dae560e1f4c79aa4b410cf3701c9c677c386b for the diff for COMPATIBILITY.md file
 - [Improvement] Refactor error handling (thx to @msaf1980)
 - [Improvement] Default systemd unit should drop privileges (thx to @deniszh)
 - [Improvement] Enable cairo support for docker images (thx to @deniszh)

**0.15.2**
 - [Fix] Honor isLeaf attribute in replies (makes possible to have metric called "metric.foo" and metric called "metric.foo.bar" and see both in find queries (thx to @tantra35)
 - [Fix] Fix bootstrapInterval paramter handling in holtWinters functions
 - [Fix] Fix potential panics when metirc do not have required amount of data for `holtWinters` functions
 - [Fix] Fix panics with several functions if they are applied before `alias` function (mostly `holtWinter` functions)

**0.15.1**
 - [Fix] Fix back compatibility for default listen on `localhost:8081` (thx to @Felixoid)

**0.15.0**
 - **[Breaking]** - cache (including memcache) now uses sha256 for a hash function for keys. This might break existing setup
 - **[Breaking]** - new config variable `slowLogThreshold` under `upstreams` section controls what request time would be required for request to be logged in `slow` logger.
 - **[Breaking]** - remove semi-broken support for graceful restarts (See #552)
 - [Fix] Fix parsing tags from backend replies (adds support of `=` in tag values, make them more compatible with graphite-web)
 - [Fix] Fix how request splitting worked for maxBatchSize != 0 (thx to @tantra35)
 - [Improvement] Bad requests will return error 400 instead of 500 (thx to @Felixoid)
 - [Improvement] VictoriaMetrics >= 1.53.1 backend should work a bit faster for find queries
 - [Improvement] Do not apply tag-deduplication for VictoriaMetrics >= 1.50.0 (less memory consumption, faster queries)
 - [Improvement] Use VictoriaMetrics's graphite-compatibe API to query for tags (less memory consumption, faster queres)
 - [Feature] Support `add` function (thx to @Felixoid)
 - [Feature] removeEmptySeries supports xFilesFactor parameter (thx to @faceair)
 - [Feature] Allow to specify multiple listen addresses
 - [Feature] Optional caching DNS resolver

**0.14.2.1**
 - [Fix] Fix test for timeShift function. This doesn't affect the way how carbonapi works, just makes CI happy

**0.14.2**
 - [Feature] Separate backend protocol for VictoriaMetrics. Based on prometheus protocol (and share some of it's code), but uses some of VM-specific features to improve performance for /metrics/find queries (Related to #521)
 - [Feature] timeShift function supports `resetEnd` parameter (thx to @faceair). Current default is set to `false` to match carbonapi behavior, however in 0.15.0 it will be changed to `true`.
 - [Feature] Resepct pixelRatio parameter from referer if not specified in request. (thx to @lomik)
 - [Fix] Handling of maxBatchSize (maxGlobs) in config file. Respect overrides on backend level.
 - [Fix] Warnings about duplicate functions when carbonapi starts.
 - [Fix] Resulting tags in groupByTags are now correct (thx to @Felixoid)
 - [Fix] Fix handling of requests that fetches data with different start/end times (thx to @Felixoid). Related to #526
 - [Fix] Fix the way how `pow` function works with NaN in values (thx to @zhelyabuzhsky)
 - [Fix] CSV format now produce dates in UTC (like in graphite-web) (thx to @jonasbleyl)
 - [Fix] Fix from/util timestamp aligning in all `moving*` functions (thx to @Felixoid)

**0.14.1**
 - [Feature] Implement `doMultipleRequestsIfSplit` config option which could be useful for go-carbon and huge requests (See #509)
 - [Improvement] Return stacktrace on panic (thx to @Felixoid)
 - [Fix] Fix case where some metrics passed to functions like group or sumSeries were missing. Fixes #438
 - [Fix] Accept `1` and `0` as bool arguments for `True` and `False`. (thx to @Felixoid)
 - [Fix] Align precision of multiple metrics (thx to @Felixoid, see #500 and #501)
 - [Fix] fallbackSeries now works properly (thx to @Egor Redozubov)
 - [Fix] Panic when trying to render png/svg of an empty response (Fixes #503)
 - [Fix] Fix Error 500 when sendGlobsAsIs is false (Fixes #506)
 - [Fix] prometheus backend: carbonapi should send 'end' instead of 'stop' in queries (thx to Alexandre Vincent)
 - [Fix] change metrics resulting tags to match graphite-web in some cases (thx to @Felixoid)
 - [Fix] Prometheus backend: trust timestamps from the backend (Fixes #504, Fixes #514)
 - [Fix] percentileOfSeries: return first of filtered datapoints in case only one series have valid data (Fixes #516)
 - [Code] merge aliasByNode and aliasByTags (thx to @Felixoid)

**0.14.0**
 - **[Breaking]**[Code] expr library and all functions now requires caller to pass context. See #485
 - **[Breaking]**[Config] for protocol `auto` there is now no default implied concurrency limit of `100` as it was before.
 - **[Breaking]**[Config] Changed default value for `notFoundStatusCode` to 200 to match graphite-web behavior
 - [Feature] Add a `backendCache` option that implements dedicated cache for backend responses. See #480 (thx to @jaroslawr)
 - [Feature] For Prometheus backend it is now possible to specify max\_points\_per\_query
 - [Feature] weightedAverage function (thx to @Felixoid)
 - [Improvement] carbonapi now pass maxDataPoints to backends that support carbonapi\_v3\_pb format. Previously 0 was passed.
 - [Fix] metric find requests to backend now pass start and end time (thx to @faceair)
 - [Fix] Fix 404 status code if backend have errors (thx to @lexx-bright)
 - [Fix] Fix sorting in \*seriesLists functions (thx to Egor Redozubov)
 - [Fix] Potential panic during groupByNode evaluation if callback is invalid expression
 - [Fix] Partially overlapping backend groups caused some queries to return empty result
 - [Fix] Sorting metrics should work now in the same way as in graphite-web (thx to @Felixoid)
 - [Fix] Time of the first timestamp was wrong if multiplySeries was applied (it matched request `from`)


**0.13.0**
 - [Fix] smartSummarize now supports wildcards in the metric names (thx to @Peter-Sh)
 - [Fix] json format correctly distinguishes between +-inf and nan (thx to @faceair)
 - [Fix] prometheus backend: fix a bunch of problems related to globs and regex escaping (thx to @rodio)

**0.13.0.rc.3**
 - [Fix] Proper fix for prometheus backend and non-taged render requests with groupByNodes function
 - [Fix] Align timestamps in prometheus backend (thx to @rodio, #467)

**0.13.0-rc.2**
 - [Fix] Prometheus backend wasn't working correctly for non-taged render requests (#465, thx to @menai34 for proposed fix)
 - [Fix] Fix panic when using prometheus tagged response for non-tagged queries in groupByTags function (and maybe more)
 - [Fix] Add proper aliases for `aggregate` functions - that would make groupByTags properly usable with functions like `diff` and `total`.

**0.13.0-rc.1**
 - [Improvement] Redesign error handling and logging. Logging should be now less noisy and all error messages should contain better reasoning about error cause
 - [Improvement] Move some of the logging messages to Debug level - that should make logs less noisy and still preserve ability to see detailed errors on Debug level
 - [Improvement] Add a config parameter to disable tldCache (useful for clickhouse-based backends)
 - [Improvement] Implement noNullPoints query parameter. Works only with JSON as in graphite-web
 - [Improvement] For all SeriesLists functions, allow to specify default argument (thx to kolobaev@)
 - [Improvement] Add support for `round` function (thx to kolobaev@)
 - [Improvement] Add integralByInterval function (thx to faceair@)
 - [Improvement] Add sortBy function (thx to misiek08@)
 - [Improvement] Add smartSummarize function (thx to misiek08@)
 - [Fix] /render and /metrics/find URLs now works correctly for format=carbonapi\_v2\_pb (protov3) and for new format (carbonapi as a carbonapi's backend)
 - [Fix] Allow '%' in metric names
 - [Fix] zipper metrics now exported again
 - [Fix] applyByNode - fix various incompatibilities with graphite-web (node starts with 1, some rewrite related issues, etc) (thx to faceair@)
 - [Fix] Honor SendGlobAsIs and AlwaysSendGlobAsIs (important for pre-0.12 configs)
 - [Fix] Fix panic in some cases when one of the metrics is missing
 - [Fix] Compatbility fixes to useAboveSeries (now it's behavior matches graphite-web's)
 - [Fix] Fix seriesList functions in case of unsorted responses (\*seriesList sorts denomniators and numerators first of all)
 - [Fix] Fix pprof endpoint routing (thx to faceair@)
 - [Fix] Various fixes around error handling (thx to faceair@)
 - [Fix] Avoid multiple requests for time moving based functions (thx to faceair@)
 - [Code] Make linters much more happier about the code (thx to faceair@ for contribution)
 - **[Breaking]** [Code] Comment out support for gRPC backend type. It was never properly tested and likely need complete rework before it will be usable
 - **[Breaking]** [Build] Minimum Supported golang version is 1.13.0

**0.12.6**

 - [Fix] Fix aliasByTags to work correclty with other functions
 - [Fix] panic when using protov2 and backend that doesn't support tags (thx to @gekmihesg)

**0.12.5**
 - [Feature] Implement 'highest' function
 - [Feature] Implement 'lowest' function
 - [Feature] Implement 'aggregateLine' function
 - [Feature] Implement 'filterSeries' function

**0.12.4**
 - [Feature] Function Defines - allows to do custom metric aliases (thx to @lomik). See [doc/configuration.md](https://github.com/go-graphite/carbonapi/blob/master/doc/configuration.md#define) for config format
 - [Improvement] New config options that allows to prefix all URLs and to enable /debug/vars on a separate address:port. See [docs/configuration.md](https://github.com/go-graphite/carbonapi/blob/master/doc/configuration.md#expvar) for more information
 - [Improvement] `/render` queries now returns tags in json (as graphite-web do)
 - [Improvement] groupByTags should now support all available aggregation functions (#410)
 - [Fix] Fix panic when using carbonapi\_proto\_v3 and doing tag-related queries (#407)
 - [Fix] Add missing alias for averageSeries (thx to @msaf1980)

**0.12.3**
 - [Fix] Fix graphiteWeb proxy function (thx. to @sylvain-beugin)
 - [Fix] Prometheus Backend: correctly handle groups, fixes #405
 - [Fix] Prometheus Backend: convert target that doesn't contain seriesByTags in a same way that's used for /metrics/find
 - [Fix] change behavior of aliasSub to match graphite-web (fixes #290)
 - [Improvement] Prometheus backend: Allow to specify "start" parameter (via backendOptions)

**0.12.2**
 - [Fix] Fix stacked cairo-based graphs (doesn't affect json or grafana rendering)
 - [Fix] Correctly deduplicate requests for cases where same metric fetched in the same expression and cache is disabled. Fixes #401
 - [Fix] Make all zipper's errors non-fatal (carbonapi will try to fetch construct full response no matter what error it was)
 - [Fix] Make zipper less noisy

**0.12.1**
 - [Improvement] Port 'minValue' parameter handling for nonNegativeDerivative and perSecond: https://github.com/bookingcom/carbonapi/commit/5bfda0d24 and https://github.com/bookingcom/carbonapi/commit/790c05d8
 - [Improvement] Config option "headersToPass" to control what request headers will be passed to backend (default: none). Fixes #398
 - [Improvement] Config option "headersToLog" to control what request headers will be logged by carbonapi (default: none).
 - [Fix] Cherry-pick https://github.com/bookingcom/carbonapi/commit/946ca8b (fixes small png render issues)
 - [Fix] #260 - parsing of bool values as Names.
 - [Fix] Fix limiter waiting for a wrong server in some cases
 - [Code] Cherry-pick https://github.com/bookingcom/carbonapi/pull/172/commits/9f0b3f611 to simplify tests (author: @gksinghjsr)

**0.12.0**
 - Add support for Prometheus as Backend. This allows to use Prometheus-compatible software as carbonapi's backends (e.x. Prometheus and VictoriaMetrics)
 - Add support for querying msgpack-compatible backends. This should make carbonapi compatible with graphite-web 1.1 and [grafana/metrictank](https://github.com/grafana/metrictank)
 - **[Breaking][Code]** Migrate all internal structures to `github.com/go-graphite/protocol/carbonapi_v3_pb`. This removes redundant IsAbsent slice and changes all timestamps to int64 (they are still expected to have uint32 timestamps there)
 - **[Breaking][Improvement]** Integrate carbonzipper 1.0.0. This introduces better loadbalancing support, but significantly changes config file format. It might behave differently with the same settings. It also removes carbonzipper dependency.
 - [Improvement] seriesByTag Support (thx. to Vladimir Kolobaev)
 - [Improvement] aliasByTag Support (thx. to Vladimir Kolobaev)
 - [Improvement] Added support for more aggregation functions (thx. to Oleg Matrokhin)
 - [Improvement] Add 'aggregate' function.
 - [Improvement] pixelRatio for png render (thx. to Roman Lomonosov)
 - [Fix] Supported functions were updated to be more compatible with graphtie-web 1.1.0+
 - [Fix] fix movingXyz error on intervals greater than 30 days (thx. to Safronov Michail)
 - [Fix] Fix timeShift function (thx. to Gunnar Þór Magnússon)

**0.12.0-rc.1**
 - Fix panic when using Prometheus backend and query /tags/autoComplete/tags without parameters
 - Fix long going issue with stuck requests towards backend.

**0.12.0-rc.0**
 - Add experimental support for Prometheus as Backend.
 - Add experimental support for querying msgpack-compatible backends. This should make carbonapi compatible with graphite-web 1.1 and [grafana/metrictank](https://github.com/grafana/metrictank)
 - **[Breaking][Code]** Migrate all internal structures to `github.com/go-graphite/protocol/carbonapi_v3_pb`. This removes redundant IsAbsent slice and changes all timestamps to int64 (they are still expected to have uint32 timestamps there)
 - **[Breaking][Improvement]** Migrate to carbonzipper 1.0.0. This introduces better loadbalancing support, but significantly changes config file format. It might behave differently with the same settings.
 - seriesByTag Support (thx. to Vladimir Kolobaev)
 - aliasByTag Support (thx. to Vladimir Kolobaev)
 - Supported functions were updated to be more compatible with graphite-web 1.1.0+
 - Added support for more aggregation functions (thx. to Oleg Matrokhin)

**0.11.0**
 - **[Breaking][Fix] Allow to specify prefix for environment variables through `-envprefix` command line parameter. Default now is "CARBONAPI_" which might break some environments**
 - [Improvement] graphiteWeb function that implements force-fallback to graphiteWeb
 - [Improvement] graphiteWeb can query graphite-web 1.1.0+ for a list of supported functions and automatically do fallback to graphite in case user calls unimplemented function
 - [Improvement] New config switch to disable find request before render. This is helpful for graphite-clickhouse backends
 - [Fix] Fix typo in polyfit function description
 - [Fix] Skip metrics without pairs in reduceSeries (thx. to @errx)
 - [Fix] Fix config file parsing (thx. to @errx)
 - [Code] Move MakeResponse to expr/types and rename it to MakeMetricData. This is sometimes useful outside of tests. Thx. to @borovskyav
 - [Code] Add helper package that will ease doing per-function tests. See `expr/functions/absolute/function_test.go` for an example

**0.10.0.1**
 - [Fix] Autobuild scripts. Version bump to avoid retagging.

**0.10.0**
 - [Fix] `lineWidth` param is not working
 - [Fix] Support `&` in metric name
 - [Fix] default colors for png and svg are now defined as in graphite-web in 1.1.0, but add example of how to set them to new colors
 - [Improvement] You can override default colors through config
 - [Improvement] It's now possible to specify 'template=' option to png or svg renders. It will use templates that's specified by 'graphTemplates' option in config. Template config format is not compatible with graphite-web though and uses either toml or yaml. Options naming is also different and follows URL override style.
 - [Code] Minor API changes in expr/png: MarshalPNGRequest and MarshalSVGRequest now requires to specify template.

**0.9.2**
 - [Improvement] "pipe" function chaining syntax. Same way as in https://github.com/graphite-project/graphite-web/pull/2042 (thx. to @lomik)

**0.9.1**
 - [Code] Refactor expr. Split it into pkg/parser. Simplify tests.
 - [Code] Refactor expr. Split functions into directories.
 - [Code] Simple docs about how to add a new function.
 - [Fix] All aggregating functions (e.x. summarize) now adds math.NaN() to begin and end of value list to make them same size.
 - [Fix] Experimental option `ExtrapolateExperiment` that allows aggregating functions to extrapolate data. Function is not tested properly yet and might be buggy, disabled by default
 - [Improvement] API to get list of functions (graphite-web 1.1 compatibility)
 - [Improvement] Add 'first' and 'last' as consolidateBy options

**0.9.0**
 - [Improvement] Merge carbonzipper into carbonapi. Please migrate to new 'upstreams' in the config file.
 - [Improvement] Make default cache timeout configurable.
 - [Improvement] Add divideSeriesLists (thx. to Oleg Matrokhin)
 - [Improvement] Add groupByNodes (thx to @errx)
 - [Improvement] Add cumulative(thx to Oleg Matrokhin)
 - [Improvement] Add diffSeriesLists and multiplySeriesLists functions (works the same way as divideSeriesLists vs divideSeries)
 - [Improvement] Add linearRegression function (thx. to Oleg Matrokhin)
 - [Improvement] Allow applyByNode to generate new targets
 - [Improvement] Allow diffSeries to substract series with different Steps
 - [Improvement] Make cache timeouts configurable
 - [Improvement] Make batch size configurable for sendGlobAsIs
 - [Improvement] Export GoVersion through expvars
 - [Improvement] Allow overriding config variables through ENV
 - [Improvement] Allow to use .toml in addition to .yaml as config file (determined by extension)
 - [Improvement] Add fallbackSeries (thx to @gksinghjsr)
 - [Improvement] Implement areaMode in cairo rendering (thx. to @ibuclaw)
 - [Improvement] Implement lineWidth, rightWidth, leftWidth (for cairo rendering) (thx. to @ibuclaw)
 - [Improvement] Add Unicode support for series names (thx. to @korservick)
 - [Improvement] Add multiplySeriesWithWildcards (thx. to @kamaev)
 - [Improvement] Add stddevSeries (thx. to @kamaev)
 - [Improvement] Add mapSeries (thx. to @kipwoker)
 - [Improvement] Add reduceSeries (thx. to @kipwoker)
 - [Improvement] Add delay() function (thx. to @kipwoker)
 - [Improvement] Update asPercent to graphite-web 1.1 (thx. to @kipwoker)
 - [Fix] Fix rounding errors when generating yLables for Cairo-based renders (thx. to @ibuclaw)
 - [Fix] Fix integer overflow when building for 32bit systems
 - [Fix] Don't set path in treejson if '.' not found in name (thx. to @ibuclaw)
 - [Fix] Fix derivative() if first arg is missing (thx. to Oleg Matrokhin)
 - [Fix] Fix calculation of yTop when using log base (thx. to @ibuclaw)
 - [Fix] Fix asPercent() to handle multiple series (thx. to @ibuclaw)
 - [Fix] Swap red and darkred colors (thx to @ibuclaw)
 - [Fix] Make asPercent graphite-compatible #213
 - [Fix] Fix summarize for cases when bucketSize < stepTime
 - [Fix] Fixed #90 - carbonapi shouldn't fail if one of the series doesn't exist
 - [Fix] Make PNG rendering looks like graphite-web (thx. to @lomik)
 - [Fix] Don't fail if argument is empty series (graphite-web compatibility, thx. to @gksinghjsr)
 - [Fix] Fix group() function #225 (thx. to @gksinghjsr)
 - [Code] Small refactor of PictureParams structs. Makes GetPictureParams and DefaultColorList public.

**0.8.0**
 - **[Breaking]** Most of the configuration options moved to a config file
 - **[Breaking]** Logging migrated to zap (structured logging). Log format changed significantly. Old command line options removed. Please consult carbonapi.example.yaml for a new config options and explanations
 - [Improvement] CarbonAPI now passes extra headers to keep track of requests
 - [Improvement] Allow '^' and '$' as name characters. Required for carbonsearch text-match pattern anchoring
 - [Improvement] Allow to bind to arbitrary IP:Port. (thx to @zdykstra, #194), Fixes #176
 - [Improvement] Extend legendValues to append multiple values
 - [Improvement] Add a switch to turn off propagating globbed requests to carbonzipper (was introduced in 0.7.0)
 - [Improvement] Add `{find,render}_cache_overhead_ns` metrics that shows cache query overhead in nanoseconds
 - [Improvement] Use dep as a vendoring tool
 - [Improvement] Add a Makefile that will hide some magic from user.
 - [Improvement] Send glob heruistic -- if request with globs is relatively small (<100 metrics) carbonapi will send it as is.
 - [Improvement] divideSeries support seriesList in dividents (#203, thx @cldellow)
 - [Improvement] added movingMin/movingMax/movingSum (#206, thx @cldellow)
 - [Feature] Implement following functions from graphite-web: linewidth, areaBetween, applyByNode
 - [Feature] Add fft, ifft, LowPass functions. See COMPATIBILITY.md for more information
 - [Feature] Publish current configuration as an expvar (GET /debug/vars)
 - [Feature] Make prefix configurable (thx to @StephenPCG, #189)
 - [Fix] Fix issue with missing labels on graphs > 14d
 - [Fix] Allow expressions to use 'total' instead of 'sum'. Required for graphite compatibility
 - **[Fix]** Fix incompatibility between carbonapi and older versions of carbonzipper/carbonserver/go-carbon (protobuf2-only)

Notes on upgrading:

Even though there are several changes that's marked as breaking, it only breaks local config parsing and changes logging format.

**This release fixes the incompatibility introduced in 0.6.0 and can work with both new and old versions of carbonserver, carbonzipper, go-carbon**

**0.7.0**
 - [Improvement] Render requests with globs now sent as-is.
 - [Improvement] Added ProxyHeaders handler to get real client's ip address from X-Real-IP or X-Forwarded-For headers.
 - [Improvement] Highest\* and Lowest\* now assumes '1' if no second argument was passed
 - [Feature] Implement legendValue expression (see http://graphite.readthedocs.io/en/latest/functions.html#graphite.render.functions.legendValue for more information)

Notes on upgrading to 0.7.0:

It's highly recommended, if you use carbonzipper to also upgrade it to version 0.62

**0.6.0**
 - **[Breaking]** Migrate to protobuf3

Notes on upgrading to 0.6.0:

You need to upgrade carbonzipper or go-carbon to the version that supports protobuf3

**<=0.5.0**
There is no dedicated changelog for older versions of carbonapi. Please see commit log for more information about what changed for each commit.
