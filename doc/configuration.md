Table of Contents
=================

* [General configuration for carbonapi](#general-configuration-for-carbonapi)
  * [listen](#listen)
    * [Example:](#example)
  * [useCachingDNSResolver](#useCachingDNSResolver)
  * [prefix](#prefix)
    * [Example:](#example-1)
  * [headersToPass](#headerstopass)
    * [Example:](#example-2)
  * [headersToLog](#headerstolog)
    * [Example:](#example-3)
  * [headersToLog](#define)
    * [Example:](#example-4)
  * [notFoundStatusCode](#notfoundstatuscode)
    * [Example:](#example-5)
  * [httpResponseStackTrace](#httpresponsestacktrace)
  * [unicodeRangeTables](#unicoderangetables)
    * [Example](#example-6)
  * [cache](#cache)
    * [Example](#example-7)
  * [cpus](#cpus)
    * [Example](#example-8)
  * [tz](#tz)
    * [Example](#example-9)
  * [functionsConfig](#functionsconfig)
    * [Example](#example-10)
    * [Example for timeShift](#example-for-timeshift)
  * [graphite](#graphite)
    * [Example](#example-11)
  * [pidFile](#pidfile)
    * [Example](#example-12)
  * [graphTemplates](#graphtemplates)
    * [Example](#example-13)
  * [defaultColors](#defaultcolors)
    * [Example](#example-14)
  * [expvar](#expvar)
    * [Example](#example-15)
  * [logger](#logger)
    * [Example](#example-16)
* [Carbonzipper configuration](#carbonzipper-configuration)
  * [concurency](#concurency)
    * [Example](#example-17)
  * [maxBatchSize](#maxbatchsize)
    * [Example](#example-18)
  * [idleConnections](#idleconnections)
  * [upstreams](#upstreams)
    * [Example](#example-19)
      * [For go\-carbon and prometheus](#for-go-carbon-and-prometheus)
      * [For VictoriaMetrics](#for-victoriametrics)
      * [For graphite\-clickhouse](#for-graphite-clickhouse)
      * [For metrictank](#for-metrictank)
      * [For IRONdb](#for-irondb)
  * [expireDelaySec](#expiredelaysec)
    * [Example](#example-20)

# General configuration for carbonapi

## listen

Describe the port and address that carbonapi will bind to. This is the one that you can use to connect to it.

### Example:
This will make it available on http://localhost:8081:
```yaml
listen: "localhost:8081"
```

This will make it available on all addresses, port 8080:
```yaml
listen: ":8080"
```

This will make it available on all IPv4 addresses, port 8080:
```yaml
listen: "0.0.0.0:8080"
```

***
## useCachingDNSResolver

**You shouldn't use it unless you know what you are doing.**

Use custom DNS resolver that have internal cache instead of default Golang one. This is global setting and cannot be overridden on backend level.

This option might help with environments where using DNS names is highly encouraged but DNS server provided have troubles keeping up with request rate. For example - some older versions of k8s or very specific settings on kube-dns side might rate-limit DNS requests.

Please note that this DNS resolver is deliberately non-RFC compliant and will ignore TTL for the domain. `cachingDNSRefreshTime` will be used as a TTL instead.

Default: false

***
## cachingDNSRefreshTime

If `useCachingDNSResolver` is set to true, this is the TTL for DNS records to be valid.

Default: 1m

***
## prefix

Specify prefix for all URLs. Might be useful when you cannot afford to listen on different port.

Default: None

### Example:
This will make carbonapi handlers accessible on `/graphite`, e.x. `http://localhost:8080/render` will become `http://localhost:8080/graphite/render` 

```yaml
prefix: "graphite"
```

***
## headersToPass

This option controls what headers (if passed by upstream client) will be passed to backends..

Default: none

### Example:
This is example to pass all dashboard/panel ids from Grafana
```yaml
headersToPass:
    - "X-Dashboard-Id"
    - "X-Grafana-Org-Id"
    - "X-Panel-Id"
```

***
## headersToLog

This option controls what headers will be logged by carbonapi (accessLog).

If header is not present, it won't be logged.

Headers will be appended to access log and to any other carbonapi logs for this handler if there is any used.
They won't be logged at zipper's level (currently).

Default: none

### Example:
This is example to log all dashboard/panel ids from Grafana
```yaml
headersToLog:
    - "X-Dashboard-Id"
    - "X-Grafana-Org-Id"
    - "X-Panel-Id"
```

***
## notFoundStatusCode

This option controls what status code will be returned if `/render` or `/metrics/find` won't return any metrics 

In some cases someone would like to override this to "200". Example use case - when you create a dashboard before
service starts to send any data out and don't want to have errors from Grafana.

Default: 200

### Example:
This is example to return HTTP code 200
```yaml
notFoundStatusCode: 200
```

***
## httpResponseStackTrace

This option controls if stack trace should be sent as http answer in case of a panic during `/render` proceeding.

Default: true

***
## define

List of custom function aliases (defines)

Defines are done by templating this custom aliases to known set of functions.

Templating is done by utilizing golang text template language.

Supported variables:
 - argString - argument as a string. E.x. in query `myDefine(foo, bar, baz)`, argString will be `foo, bar, baz`
 - args - indexed array of arguments. E.x. in case of `myDefine(foo, bar)`, `index .args 0` will be first argument, `index .args 1` will be second
 - kwargs - key-value arguments (map). This is required to support cases like `myDefine(foo, bar=baz)`, in this case `index .args 0` will contain `foo`, and `index .kwargs "bar"` will contain `baz`

### Example:
Create a perMinute function, that do "perSecond" and then scale it by 60

Config:
```yaml
define:
  -
    name: "perMinute"
    template: "perSecond({{.argString}})|scale(60)" 
```

Example Query:

`/render/?target=perMinute(foo.bar)`

***
## unicodeRangeTables

Allow extra charsets in metric names. By default only "Latin" is allowed

Please note that each unicodeRangeTables will slow down metric parsing a bit

For list of supported tables, see: https://golang.org/src/unicode/tables.go?#L3437

Special name "all" reserved to append all tables that's currently supported by Go

### Example
This will allow support of Latin, Cyrillic and Japanese characters in metric names:
```yaml
unicodeRangeTables:
   - "Latin"
   - "Cyrillic"
   - "Hiragana"
   - "Katakana"
   - "Han"
```

Please note that you need to specify "Latin" if you are redefining this list.

This will allow support of all unicode characters that's supported by Go
```yaml
unicodeRangeTables:
   - "all"
 ```

***
## cache
Specify what storage to use for response cache. This cache stores the final
carbonapi response right before sending it to the client. A cache hit to this
cache avoids almost all computations, including rendering images etc. On the
other hand, a request will cause a cache hit only if a previous request with
exactly the same response format and with same maxDataPoints param populated the
cache. Grafana sets maxDataPoints depending on client screen width, reducing the
hit ratio for this cache.

Supported cache types:
 - `mem` - will use integrated in-memory cache. Not distributed. Fast.
 - `memcache` - will use specified memcache servers. Could be shared. Slow.
 - `null` - disable cache
 
Extra options:
 - `size_mb` - specify max size of cache, in MiB
 - `defaultTimeoutSec` - specify default cache duration. Identical to `DEFAULT_CACHE_DURATION` in graphite-web
### Example
```yaml
cache:
   type: "memcache"
   size_mb: 0
   defaultTimeoutSec: 60
   memcachedServers:
       - "127.0.0.1:1234"
       - "127.0.0.2:1235"
```

## backendCache
Specify what storage to use for backend cache. This cache stores the responses
from the backends. It should have more cache hits than the response cache since
the response format and the maxDataPoints paramter are not part of the cache
key, but results from cache still need to be postprocessed (e.g. serialized to
desired response format).

Supports same options as the response cache.
### Example
```yaml
backendCache:
   type: "memcache"
   size_mb: 0
   defaultTimeoutSec: 60
   memcachedServers:
       - "127.0.0.1:1234"
       - "127.0.0.2:1235"
```

### Example for 2-level cache lifetime (short timeout on 1-hour window and long timeout for other) and round timestamps for tune cache efficience
```yaml
cache:
   type: "mem"
   size_mb: 0
   defaultTimeoutSec: 10800 # Default cache timeout
   shortTimeoutSec: 60 # if until - from <= 1hour && now-until < 1min, by default is equal to defaultTimeoutSec
backendCache:
   type: "mem"
   size_mb: 0
   defaultTimeoutSec: 10800 # Default cache timeout
   shortTimeoutSec: 60 # if until - from <= 1hour && now-until < 1min, by default is equal to defaultTimeoutSec

truncateTime: # truncate from/until for identifical results between carbonapi instances. Also reduce load on long-range queries
  "8760h": "1h"     # Timestamp will be truncated to 1 hour round if (until - from) > 365 days
  "2160h": "10m"     # Timestamp will be truncated to 10 minute round if (until - from) > 90 days
  "1h": "1m"         # Timestamp will be truncated to 1 minute round if (until - from) > 1 hour
  "0": "10s"         # Timestamp will be truncated to 10 seconds round by default
```

***
## cpus

Specify amount of CPU Cores that golang can use. 0 - unlimited

### Example
```yaml
cpus: 0
```

***
## tz
Specify timezone to use.

Format: `name,offset`

You need to specify the timezone to use and it's offset from UTC
 
Default: "local"

### Example
Use timezone that will be called "Europe/Zurich" with offset "7200" seconds (UTC+2)
```yaml
tz: "Europe/Zurich,7200"
```

***
## functionsConfig

Extra config files for specific functions

Only the following functions currently support having their own config:
  - `graphiteWeb`
  - `aliasByPostgres`
  - `movingMedian`
  - `moving` (applies to `movingAverage`, `movingMin`, `movingMax`, `movingSum`)

### Example
```yaml
functionsConfig:
    graphiteWeb: ./graphiteWeb.example.yaml
```

### Example for timeShift
```yaml
functionsConfig:
    timeShift: ./timeShift.example.yaml
```

`timeShift.example.yaml`:
```yaml
resetEndDefaultValue: false
```

***
## graphite
Specify configuration on how to send internal metrics to graphite.

Available parameters:
  - `host` - specify host where to send metrics. Leave empty to disable
  - `interval` - specify how often to send statistics (e.x. `60s` or `1m` for every minute)
  - `prefix` - specify metrics prefix
  - `pattern` - allow to control how metrics are named.
  Special word `{prefix}` will be replaced with content of `prefix` variable.
  Special word `{fqdn}` will be replaced with host's full hostname (fqdn)
  
  
Specifying tags currently not supported.
### Example
```yaml
graphite:
    host: ""
    interval: "60s"
    prefix: "carbon.api"
    pattern: "{prefix}.{fqdn}"
```

***
## pidFile

Specify pidfile. Useful for systemd units

### Example
```yaml
pidFile: ""
```

***
## graphTemplates
Specify file with graphTemplates.

### Example
```yaml
graphTemplates: graphTemplates.example.yaml
```

***
## defaultColors

Specify default color maps to html-style colors, used in png/svg rendering only

### Example
This will make the behavior same as in graphite-web as proposed in https://github.com/graphite-project/graphite-web/pull/2239

Beware this will make dark background graphs less readable
```yaml
defaultColors:
      "red": "ff0000"
      "green": "00ff00"
      "blue": "#0000ff"
      "darkred": "#c80032"
      "darkgreen": "00c800"
      "darkblue": "002173"
```

***
## expvar

Controls whether expvar (contains internal metrics, config, etc) is enabled and if it's accessible on a separate address:port.
Also allows to enable pprof handlers (useful for profiling and debugging).

Please note, that exposing pprof handlers to untrusted network is *dangerous* and might lead to data leak.

Exposing expvars to untrusted network is not recommended as it might give 3rd party unnecessary amount of data about your infrastructure.

### Example
This describes current defaults: expvar enabled, pprof handlers disabled, listen on the same address-port as main application.
```yaml
expvar:
      enabled: true
      pprofEnabled: false
      listen: ""
```

This is useful to enable debugging and to move all related handlers and add exposed only on localhost, port 7070.
```yaml
expvar:
      enabled: true
      pprofEnabled: true
      listen: "localhost:7070"
```

***
## logger

Allows to fine-tune logger

Supported loggers:
 - `zipper` for all zipper-related messages
 - `access` - for access logs
 - `slow` - for slow queries
 - `functionInit` - for function-specific messages (during initialization, e.x. configs)
 - `main` - logger that's used during initial startup
 - `registerFunction` - logger that's used when new functions are registered (should be quite)

Supported options (per-logger):
  - `logger` - specify logger name (see above)
  - `file` - where log will be written to. Can be file name or `stderr` or `stdout`
  - `level` - loglevel. Please note that `debug` is rather verbose, but `info` should mostly stay quiet
  - `encoding` - `console`, `json` or `mixed`, first one should be a bit more readable for human eyes
  - `encodingTime` - specify how time-dates should be encoded. `iso8601` will follow ISO-8601, `millis` will be epoch with milliseconds, everything else will be epoch only.
  - `encodingDuration` - specify how duration should be encoded

### Example
```yaml
logger:
    - logger: ""
      file: "stderr"
      level: "debug"
      encoding: "console"
      encodingTime: "iso8601"
      encodingDuration: "seconds"
    - logger: ""
      file: "carbonapi.log"
      level: "info"
      encoding: "json"
    # disable slow log completely
    - logger: "slow"
      level: "error"
```


# Carbonzipper configuration
There are two types of configurations supported:
 1. Old-style - this is the one that was used in standalone zipper or in bookingcom's zipper
 2. New-style - supported since carbonapi 0.12.0 and allows you to specify different type of load-balancing algorithms, etc. 
## concurency
Specify max metric requests that can be fetched in parallel.

Default: 1000

It's overall recommended to set that value to at least `requests_per_second*average_time_per_request`

If you want to have not more than 20 requests per second (any type of them) and on average request takes about 3 seconds, you should set this value to at least 60.

For high-performance setup it's **not** recommended to set this value to lower than default.

### Example
```yaml
concurency: 1000
```

***
## maxBatchSize
(old-style option)

Specify maximum number of metrics per request (used with old upstream style configuration)

### Example
```yaml
maxBatchSize: 100
```

***
## idleConnections
(old-style option)

Maximium idle connections to carbonzipper

##Example
```yaml
idleConnections: 10
```

***
## upstreams

(Required)

Main configuration for backends.

Supported options:
  - `graphite09compat` - enables compatibility with graphite-web 0.9.x in terms of cluster response, default: false
  - `buckets` - Number of 100ms buckets to track request distribution in. Used to build `carbon.zipper.hostname.requests_in_0ms_to_100ms` metric and friends.
  
     The last bucket is **not** called 'requests_in_Xms_to_inf' on purpose, so we can change our minds about how many buckets we want to have and have their names remain consistent.
  - `slowLogThreshold` -  threshold for slow requests to be logged.

    If you don't want it to be logged at all, please see [logger](#example-16) section for more details

    Default: "1s"
  - `timeouts` - structure that allow to set timeout for `find`, `render` and `connect` phases
  - `backendOptions` - extra options to pass for the backend.

    currently, only prometheus backend supports options.

    valid options:
      - `step` - (`prometheus` or `victoriametrics` only) define default step for the request
      - `start` - (`prometheus` or `victoriametrics` only) define "start" parameter for `/api/v1/series` requests

        supports either unix timestamp or delta from now(). For delta you should specify it in duration format.

        For example `-5m` will mean "5 minutes ago", time will be resolved every time you do find query.
      - `max_points_per_query` - (`prometheus` or `victoriametrics` only) define maximum datapoints per query. It will be used to adjust step for queries over big range. Default limit for Prometheus is 11000.
      - `force_min_step_interval` - (`prometheus` or `victoriametrics` only) define to force using `step` in all requests ignoring MaxDataPoints param for given interval. Default value for Prometheus and VictoriaMetrics is `0s` so feature is disabled.
      - `probe_version_interval` - (`victoriametrics` only) define how often VictoriaMetrics version will be checked (as VM supports certain API endpoints starting from a specific version). Special value to disable: `never`. Default: `600s`.
      - `fallback_version` - (`victoriametrics` only) define version string that will be used as a fallback if version_short will be empty (useful when you run master builds, as they will have it empty). Format: "vX.Y.Z", Default: `v0.0.0` (all special VM optimizations will be disabled)
      - `vmClusterTenantID` - `victoriametrics` in **cluster mode** only. Use this option to configure `accountID` and `projectID` in the VM-cluster API urls. Tenants are identified by "accountID" or "accountID:projectID". Type: `string`. Default: none (single node VictoriaMetrics).
      - `irondb_account_id` - (`irondb` only) Client AccountID, default - `1`
      - `irondb_graphite_rollup`- (`irondb` only) Graphite rollup for IRONdb, in seconds. Default - `60`
      - `irondb_graphite_prefix`- (`irondb` only) Optional Graphite prefix for IRONdb. Default - `` (empty)
      - `irondb_timeout` - (`irondb` only) Timeout gets the timeout duration for HTTP requests to IRONdb. The default value is `10s`, but please make it lower than top level `find` and `render` timeouts.
      - `irondb_dial_timeout` - (`irondb` only) DialTimeout gets the initial connection timeout duration for attempts to connect to IRONdb. The default value is `500ms`.
      - `irondb_watch_interval` - (`irondb` only) WatchInterval gets the frequency at which a SnowthClient will check for updates to the active status of its nodes if WatchAndUpdate() is called. Default value - `30s`
      `irondb_connect_retries` - (`irondb` only) ConnectRetries gets the number of times requests will be retried on other nodes when network errors occur. Default - `-1`, that means unlimited.
      `irondb_retries`- (`irondb` only) Retries gets the number of times requests will be retried. Default is taken from `retries` value.
  - `concurrencyLimitPerServer` - limit of max connections per server. Likely should be >= maxIdleConnsPerHost. Default: 0 - unlimited
  - `maxIdleConnsPerHost` - as we use KeepAlive to keep connections opened, this limits amount of connections that will be left opened. Tune with care as some backends might have issues handling larger number of connections.
  - `keepAliveInterval` - KeepAlive interval
  - `scaleToCommonStep` - controls if metrics in one target should be aggregated to common step. `true` by default
  - `backends` - old-style backend configuration.
  
    Contains list of servers. Requests will be sent to **ALL** of them. There is a small optimization here - every once in a while, carbonapi will ask all backends about top-level parts of metric names and will try to send requests only to servers which have that in their name.
    
    This doesn't yet work if there are tags involved.
    
    Note: `backend` section will override `backendv2` if both specified.
  - `carbonsearch` - old-style carbonsearch configuration.
  
    It supports 2 options:
        * `backend` - specify the url where carbonsearch is
        * `prefix` - specify metric prefix that will be sent to carbonsearch.
  
    carbonsearch is an old attempt to implement tags for go-graphite stack: https://github.com/kanatohodets/carbonsearch
    
    It's not known if it was widely used outside of Booking.com and it's no longer known if Booking.com still use that functionality.
    
    Example carbonsearch query:
    
    `virt.v1.*.lb-pool:www.server-state:installed`
    
    It will fetch all metrics that have tag `lb-pool` set to `www` and  `server-state` set to `installed`
    
    It's mostly equivalent of:
    
    `seriesByTags('lb-pool=www','server-state=installed')`
    
    However metrics will be resolved by a separate server in this case.
    
  - `carbonsearchv2` - (new-style) configuration for carbonsearch
  
     Supports following extra options:
       * `backends` - list of backend groups. Request will be sent to all backend groups. However inside each of them it might be treated as broadcast or round-robin.
         
         Should contain:
           * `groupName` - name of the carbonsearch backend
           * `protocol` - only `carbonapi_v2_pb` make any sense as of now, as the only known implementation implements that protocol.
           * `lbMethod` - load-balancing method.
           
             Supported methods:             
               * `broadcast`, `all` - will send query to all of the servers and combine the response
               * `roundrobin`, `rr`, `any` - will send requests in round-robin manner. This means that all servers will be treated as equals and they all should contain full set of data
    
  - `backendv2` - (new-style) configuration for backends
  
     Supports following extra options:
       * `backends` - list of backend groups. Request will be sent to all backend groups. However inside each of them it might be treated as broadcast or round-robin.
         
         Should contain:
           * `groupName` - name of the carbonapi's backend
           * `protocol` - specify protocol for the backend.
           
             Supported protocols:
               * `carbonapi_v3_pb` - new native protocol, over http. Should be fastest. Currently supported by [lomik/go-carbon](https://github.com/lomik/go-carbon), [lomik/graphite-clickhouse](https://github.com/lomik/graphite-clickhouse) and [go-graphite/carbonapi](https://github.com/go-graphite/carbonapi)
               * `carbonapi_v3_grpc` - new experimental protocol that instead of HTTP requests, uses gRPC. No known backend support that.
               * `carbonapi_v2_pb`, `protobuf`, `pb`, `pb3` - older protobuf-based protocol. Supported by [lomik/go-carbon](https://github.com/lomik/go-carbon) and [lomik/graphite-clickhouse](https://github.com/lomik/graphite-clickhouse)
               * `msgpack` - message pack encoding, supported by [graphite-project/graphite-web](https://github.com/graphite-project/graphite-web) and [grafana/metrictank](https://github.com/grafana/metrictank)
               * `prometheus` - prometheus HTTP Request API. Can be used with [prometheus](https://prometheus.io) and should be usable with other backends that supports PromQL (backend can do basic fetching at this moment and doesn't offload any functions to the backend).
               * `victoriametrics`, `vm` - special version of prometheus backend, that take advantage of some APIs that's not supported by prometheus. Can be used with [VictoriaMetrics](https://github.com/VictoriaMetrics/VictoriaMetrics).
               * `snowthd`, `irondb` - supports reading Graphite-compatible metrics from [IRONdb](https://docs.circonus.com/irondb/) from [Circonus](https://www.circonus.com/).
               * `auto` - attempts to detect if carbonapi can use `carbonapi_v3_pb` or `carbonapi_v2_pb`
           * `lbMethod` - load-balancing method.
           
             Supported methods:             
               * `broadcast`, `all` - will send query to all of the servers and combine the response
               
                 It's best suited for independent backends, like go-carbon
               * `roundrobin`, `rr`, `any` - will send requests in round-robin manner. This means that all servers will be treated as equals and they all should contain full set of data
               
                 It's best suited for backends in cluster mode, like Clickhouse.
           * `maxTries` - specify amount of retries if query fails
           * `maxBatchSize` - max metrics per request.
           
             0 - unlimited.
             
             If not 0, carbonapi will do `find` request to determine how many metrics matches criteria and only then will fetch them, not more than `maxBatchSize` per request.
             
           * `keepAliveInterval` - override global `keepAliveInterval` for this backend group
           * `concurrencyLimit` - override global `concurrencyLimit` for this backend group
           * `maxIdleConnsPerHost` - override global `maxIdleConnsPerHost` for this backend group
           * `timeouts` - override global `timeouts` struct for this backend group
           * `servers` - list of sever URLs in this backend groups

### Example

Old-style configuration:
```yaml
upstreams:
    graphite09compat: false
    buckets: 10

    timeouts:
        find: "2s"
        render: "10s"
        connect: "200ms"

    concurrencyLimitPerServer: 0
    maxIdleConnsPerHost: 100
    keepAliveInterval: "30s"

    carbonsearch:
        backend: "http://127.0.0.1:8070"
        prefix: "virt.v1.*"
    backends:
        - "http://127.0.0.2:8080"
        - "http://127.0.0.3:8080"
        - "http://127.0.0.4:8080"
        - "http://127.0.0.5:8080"
```

#### For go-carbon and prometheus
```yaml
upstreams:
    graphite09compat: false
    buckets: 10

    concurrencyLimitPerServer: 0
    keepAliveInterval: "30s"
    maxIdleConnsPerHost: 100
    timeouts:
        find: "2s"
        render: "10s"
        connect: "200ms"

    carbonsearchv2:
        prefix: "virt.v1.*"
        backends:
            -
              groupName: "shard-1"
              protocol: "carbonapi_v2_pb"
              lbMethod: "rr"
              servers:
                  - "http://192.168.1.1:8080"
                  - "http://192.168.1.2:8080"
            -
              groupName: "shard-2"
              protocol: "carbonapi_v2_pb"
              lbMethod: "rr"
              servers:
                  - "http://192.168.1.3:8080"
                  - "http://192.168.1.4:8080"
    #backends section will override this one!
    backendsv2:
        backends:
          -
            groupName: "go-carbon-group1"
            protocol: "carbonapi_v3_pb"
            lbMethod: "broadcast"
            maxTries: 3
            maxBatchSize: 100
            keepAliveInterval: "10s"
            concurrencyLimit: 0
            maxIdleConnsPerHost: 1000
            timeouts:
                find: "2s"
                render: "50s"
                connect: "200ms"
            servers:
                - "http://192.168.0.1:8080"
                - "http://192.168.0.2:8080"
          -
            groupName: "go-carbon-legacy"
            maxBatchSize: 10
            concurrencyLimit: 0
            maxIdleConnsPerHost: 100
            protocol: "carbonapi_v2_pb"
            lbMethod: "broadcast"
            servers:
                - "http://192.168.0.3:8080"
                - "http://192.168.0.4:8080"
          -
            groupName: "prometheus"
            maxBatchSize: 0
            concurrencyLimit: 0
            maxIdleConnsPerHost: 1000
            protocol: "prometheus"
            lbMethod: "broadcast"
            servers:
                - "http://192.168.0.5:9090"
                - "http://192.168.0.6:9090"
```

#### For VictoriaMetrics
```yaml
upstreams:
    graphite09compat: false
    buckets: 10
    concurrencyLimitPerServer: 0
    keepAliveInterval: "30s"
    maxIdleConnsPerHost: 100
    timeouts:
        find: "2s"
        render: "10s"
        connect: "200ms"
    backendsv2:
        backends:
          -
            groupName: "victoriametrics"
            protocol: "victoriametrics"
            lbMethod: "broadcast"
            maxBatchSize: 0
            concurrencyLimit: 0
            maxIdleConnsPerHost: 1000
            servers:
                - "http://192.168.0.5:8428"
                - "http://192.168.0.6:8428"
```

#### For graphite-clickhouse
```yaml
upstreams:
    graphite09compat: false
    buckets: 10

    concurrencyLimitPerServer: 0
    keepAliveInterval: "30s"
    maxIdleConnsPerHost: 100
    timeouts:
        find: "2s"
        render: "10s"
        connect: "200ms"

    #backends section will override this one!
    backendsv2:
        backends:
          -
            groupName: "clickhouse-cluster1"
            protocol: "carbonapi_v2_pb" # "carbonapi_v3_pb" for the latest master
            lbMethod: "rr"
            maxTries: 3
            maxBatchSize: 0
            keepAliveInterval: "10s"
            concurrencyLimit: 0
            maxIdleConnsPerHost: 1000
            timeouts:
                find: "2s"
                render: "50s"
                connect: "200ms"
            servers:
                - "http://192.168.0.1:8080"
                - "http://192.168.0.2:8080"
          -
            groupName: "clickhouse-cluster2"
            protocol: "carbonapi_v2_pb" # "carbonapi_v3_pb" for the latest master
            lbMethod: "rr"
            maxTries: 3
            maxBatchSize: 0
            backendOptions:
                step: "60"
                start: "-5m"
            keepAliveInterval: "10s"
            concurrencyLimit: 0
            maxIdleConnsPerHost: 1000
            servers:
                - "http://192.168.0.3:8080"
                - "http://192.168.0.4:8080"
```

#### For metrictank
```yaml
upstreams:
    graphite09compat: false
    buckets: 10

    concurrencyLimitPerServer: 0
    keepAliveInterval: "30s"
    maxIdleConnsPerHost: 100
    timeouts:
        find: "2s"
        render: "10s"
        connect: "200ms"

    #backends section will override this one!
    backendsv2:
        backends:
          -
            groupName: "metrictank"
            protocol: "msgpack"
            lbMethod: "rr"
            maxTries: 3
            maxBatchSize: 0
            keepAliveInterval: "10s"
            concurrencyLimit: 0
            maxIdleConnsPerHost: 1000
            timeouts:
                find: "2s"
                render: "50s"
                connect: "200ms"
            servers:
                - "http://192.168.0.1:6060"
                - "http://192.168.0.2:6060"
          -
            groupName: "graphite-web"
            protocol: "msgpack"
            lbMethod: "broadcast"
            maxTries: 3
            maxBatchSize: 0
            keepAliveInterval: "10s"
            concurrencyLimit: 0
            maxIdleConnsPerHost: 1000
            servers:
                - "http://192.168.0.3:8080?format=msgpack"
                - "http://192.168.0.4:8080?format=msgpack"
```
#### For IronDB
```yaml
upstreams:
    graphite09compat: false
    buckets: 10

    concurrencyLimitPerServer: 0
    keepAliveInterval: "30s"
    maxIdleConnsPerHost: 100
    timeouts:
        find: "2s"
        render: "10s"
        connect: "200ms"

    #backends section will override this one!
    backendsv2:
        backends:
          -
            groupName: "snowthd"
            protocol: "irondb"
            lbMethod: "rr" # please use "roundrobin" - broadcast has not much sense here
            maxTries: 3
            maxBatchSize: 0 # recommended value
            keepAliveInterval: "10s"
            concurrencyLimit: 0
            maxIdleConnsPerHost: 1000
            doMultipleRequestsIfSplit: false # recommended value
            backendOptions:
              irondb_account_id: 1
              irondb_timeout: "5s" # ideally shold be less then find or render timeout
              irondb_graphite_rollup: 60
            servers:
                - "http://192.168.0.1:8112"
                - "http://192.168.0.2:8112"
                - "http://192.168.0.3:8112"

```


***
## expireDelaySec
If not zero, enabled cache for find requests this parameter controls when it will expire (in seconds)

Default: 600 (10 minutes)

### Example
```yaml
expireDelaySec: 10
```
