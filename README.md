carbonapi: replacement graphite API server
------------------------------------------

[![Build Status](https://travis-ci.org/go-graphite/carbonapi.svg?branch=master)](https://travis-ci.org/go-graphite/carbonapi)
[![GoDoc](https://godoc.org/github.com/go-graphite/carbonapi?status.svg)](https://godoc.org/github.com/go-graphite/carbonapi)

We are using <a href="https://packagecloud.io/"><img alt="Private Maven, RPM, DEB, PyPi and RubyGem Repository | packagecloud" height="46" src="https://packagecloud.io/images/packagecloud-badge.png" width="158" /></a> to host our packages!

CarbonAPI supports a significant subset of graphite functions [see [COMPATIBILITY](COMPATIBILITY.md)].
In our testing it has shown to be 5x-10x faster than requesting data from graphite-web.

For requirements see **Requirements** section below.

Installation
------------

At this moment we are building packages for CentOS 7, Rockylinux 8 (should be compatible with RHEL 8), Debian 10, Debian 11, Debian 12 (testing), Ubuntu 18.04, Ubuntu 20.04, Ubuntu 22.04. Installation guides are available on packagecloud (see the links below):

* Stable versions: [Stable repo](https://packagecloud.io/go-graphite/stable/install)
* *Autobuilds* (master, might be unstable): [Autobuild repo](https://packagecloud.io/go-graphite/autobuilds/install)
* *Docker* images: [ghcr.io](https://ghcr.io/go-graphite/carbonapi)

Configuration
------------

*CarbonAPI* can be configured by *config file* or by *environment variables*.

### Configuration file

```bash
$ ./carbonapi -config /etc/carbonapi.yaml
```

* [Configuration guides](doc/configuration.md),
* [Example config](cmd/carbonapi/carbonapi.example.yaml).

There are multiple example configurations available for different backends:

* [Prometheus](cmd/carbonapi/carbonapi.example.prometheus.yaml),
* [graphite-clickhouse](cmd/carbonapi/carbonapi.example.clickhouse.yaml),
* [go-carbon](cmd/carbonapi/carbonapi.example.yaml),
* [VictoriaMetrics](cmd/carbonapi/carbonapi.example.victoriametrics.yaml),
* [IRONdb](cmd/carbonapi/carbonapi.example.irondb.yaml).

### Configuration by environment variables

Every parameter in config file are mapped to environment variable. I.E.

```yaml
concurency: 20
cache:
   # Type of caching. Valid: "mem", "memcache", "null"
   type: "mem"
upstreams:
    backends:
        - "http://10.0.0.1:8080"
        - "http://10.0.0.2:8080"
```
That config can be replaced by

```bash
CARBONAPI_CONCURENCY=20
CARBONAPI_CACHE_TYPE=mem
CARBONAPI_UPSTREAMS_BACKENDS="http://10.0.0.1:8080 http://10.0.0.2:8080"
```

You should be only aware of logging: because carbonapi support a list of logger, env variables will replace
only first logger.

If you apply variable `LOGGER_FILE=stdout` to config:

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
```

it will be equal to config:

```yaml
logger:
    - logger: ""
      file: "stdout" # Changed only here
      level: "debug"
      encoding: "console"
      encodingTime: "iso8601"
      encodingDuration: "seconds"
    - logger: ""
      file: "carbonapi.log" # Not changed
      level: "info"
      encoding: "json"
```

Supported protocols
-------------------

 * `auto` - carbonapi will do it's best to determine backend's protocol. Currently it can identify only `carbonapi_v2_pb` or `carbonapi_v3_pb`
 * `carbonapi_v2_pb`, `pb`, `pb3`, `protobuf` - carbonapi <0.11 style protocol. Supported by [go-carbon](https://github.com/go-graphite/go-carbon) and [graphite-clickhouse](https://github.com/lomik/graphite-clickhouse) older or equal version v0.11.7
 * `carbonapi_v3_pb` - new carbonapi protocol, that supports passing metadata through. Supported by carbonzipper >=1.0.0.alpha.3, [graphite-clickhouse](https://github.com/lomik/graphite-clickhouse) newer then v0.12.0 and go-carbon newer then v0.13.0
 * `carbonapi_v3_grpc` - grpc version of new carbonapi protocol. Currently no known implementation exists.
 * `msgpack` - messagepack based protocol, used in graphite-web 1.1 and metrictank. It's still experimental and might contain bugs.
 * `prometheus` - prometheus HTTP API
 * `victoriametrics` - special version of prometheus backend to use with [VictoriaMetrics](https://github.com/VictoriaMetrics/VictoriaMetrics).
 * `irondb` - supports reading Graphite-compatible metrics from [IRONdb](https://docs.circonus.com/irondb/) from [Circonus](https://www.circonus.com/).


Requirements
------------

Golang compatibility matrix:

| Golang Version | Last supported carbonapi version |
|----------------|----------------------------------|
| 1.10           | 0.10.0.1                         |
| 1.12           | 0.12.6                           |
| 1.16 / 1.17    | 0.15.6                           |
| 1.18           | 0.16.0-patch2                    |
| 1.20           | 0.16.1                           |
| 1.21           | 0.17.0                           |

Overall rule of thumb is that carbonapi supports last 2 major go versions. E.x. at this moment Go 1.22 and 1.21 are supported.

You can verify current versions that are being tested in [CI Configuration](https://github.com/go-graphite/carbonapi/blob/main/.github/workflows/tests.yml#L14).

CarbonAPI uses protobuf-based protocol to talk with underlying storages. For current version the compatibility list is:

1. [go-carbon](https://github.com/go-graphite/go-carbon) >= 0.9.0 (Note: you need to enable carbonserver in go-carbon). Recommended to run latest version, that currently supports `carbonapi_v3_pb`
2. [graphite-clickhouse](https://github.com/go-graphite/graphite-clickhouse) any. That's alternative storage that doesn't use Whisper.
3. [metrictank](https://github.com/grafana/metrictank) - supported via `msgpack` protocol. Support is not very well tested and might contain bugs. Use with cautions. Tags are not supported.
4. [carbonapi](https://github.com/go-graphite/carbonapi) >= 0.5. Note: starting from carbonapi 1274333ebd1fe50946cb4d51561e3e0f1060bc79 separate binary of carbonzipper is deprecated.
5. [carbonserver](https://github.com/grobian/carbonserver)@master (Note: you should probably switch to go-carbon in that case).
6. [carbonzipper](https://github.com/go-graphite/carbonzipper) >= 0.50. **Please note**, carbonzipper functionality was merged to carbonapi and it's no longer needed to run separate zipper.

Supported architectures and OSs
-------------------------------

Currently building is tested regularly on amd64 (automated) and arm64 (manual) only. However from time to time, riscv64 is also tested manually.

For OS support: **Linux** is the only OS that is well tested for production usage. Theoretically nothing prevents from running carbonapi on **\*BSD**, however its not tested by developers (but bugs will be accepted and eventually fixed). Running on **macos** is supported for testing purposes but it is not tested for any production use case. Other platforms are not tested and not supported.

For any other OS or Architectures bugs **won't be actively worked on**, but PRs that fixes the OS and doesn't break any other supported platforms are more than welcome.

Some remarks on different backends
----------------------------------

For backends that uses proper database (e.x. `graphite-clickhouse`) you should set `maxBatchSize: 0` in your config file for this backend group.

For other backends (e.x. go-carbon) you should set it to some reasonable value. It increases response speed, but the cost is increased memory consumption.

Tag support was only tested with `graphite-clickhouse`, however it should work with any other database.

Internal Metrics
----------------------------------

Internal metrics will be dumped to *Graphite* if [corresponding config options](doc/configuration.md#graphite) are set,
or if the `GRAPHITEHOST`/`GRAPHITEPORT` environment variables are found.

The metrics are:

| Metric Name | Description |
| ----------- | ----------- |
| `cache_items` | if caching is enabled, this metric will contain many metrics are stored in cache |
| `cache_size` | configured query cache size in bytes |
| `request_cache_hits` | how many requests were served from cache. (this is for requests to /render endpoint) |
| `request_cache_misses` | how many requests were not in cache. (this is for requests to /render endpoint) |
| `request_cache_overhead_ns` | how much time in ns it took to talk to cache (that is useful to assess if cache actually helps you in terms of latency) (this is for `requests` to `/render` endpoint) |
| `find_requests` | requests server by endpoint /metrics/find |
| `requests` | requests served by endpoint `/render` |
| `requests_in_XX_to_XX` | request response times in percentiles |
| `timeouts` | number of timeouts while fetching from backend |
| `backend_cache_hits` | how many requests were not read from backend |
| `backend_cache_misses` | how many requests were not found in the backend |

OSX Build Notes
---------------
Some additional steps may be needed to build carbonapi with cairo rendering on MacOSX.

Install cairo:

```
$ brew install Caskroom/cask/xquartz

$ brew install cairo --with-x11
```

Acknowledgement
---------------
This program was originally developed for Booking.com.  With approval
from Booking.com, the code was generalised and published as Open Source
on github, for which the author would like to express his gratitude.

Booking.com's Fork
------------------

In summer 2018, Booking.com forked version 0.11 of carbonapi and continued development in their own repo: [github.com/bookingcom/carbonapi](https://github.com/bookingcom/carbonapi).

License
-------

This code is licensed under the BSD-2 license.
