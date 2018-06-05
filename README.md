carbonapi: replacement graphite API server
------------------------------------------

[![Build Status](https://travis-ci.org/go-graphite/carbonapi.svg?branch=master)](https://travis-ci.org/go-graphite/carbonapi)
[![GoDoc](https://godoc.org/github.com/go-graphite/carbonapi?status.svg)](https://godoc.org/github.com/go-graphite/carbonapi)
[![PR Stats](http://issuestats.com/github/go-graphite/carbonapi/badge/pr?style=flat)](http://issuestats.com/github/go-graphite/carbonapi)
[![Issues Stats](http://issuestats.com/github/go-graphite/carbonapi/badge/issue?style=flat)](http://issuestats.com/github/go-graphite/carbonapi)

We are using <a href="https://packagecloud.io/"><img alt="Private Maven, RPM, DEB, PyPi and RubyGem Repository | packagecloud" height="46" src="https://packagecloud.io/images/packagecloud-badge.png" width="158" /></a> to host our packages!

CarbonAPI supports a significant subset of graphite functions [see [COMPATIBILITY](COMPATIBILITY.md)].
In our testing it has shown to be 5x-10x faster than requesting data from graphite-web.

For requirements see **Requirements** section below.

Installation
------------

At this moment we are building packages for CentOS 6, CentOS 7, Ubuntu 14.04 and Ubuntu 16.04. Installation guides are available on packagecloud (see the links below).

Stable versions: [Stable repo](https://packagecloud.io/go-graphite/stable/install)

Autobuilds (master, might be unstable): [Autobuild repo](https://packagecloud.io/go-graphite/autobuilds/install)


General information
-------------------

Carbonzipper can be configured by environment variables or by config file. For an example see `carbonapi.example.yaml`

`$ ./carbonapi -config /etc/carbonapi.yaml`

Request metrics will be dumped to graphite if coresponding config options are set,
or if the GRAPHITEHOST/GRAPHITEPORT environment variables are found.

Request data will be stored in memory (default) or in memcache.

## Configuration by environment variables

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
 * `carbonapi_v2_pb`, `pb`, `pb3`, `protobuf` - carbonapi pre-0.12 style protocol. Supported by go-carbon, graphite-clickhouse, etc.
 * `carbonapi_v3_pb` - new carbonapi protocol, that supports passing metadata through. Supported by carbonzipper 1.0.0.alpha.3 or later. Implementing support for that is in-progress for go-carbon and graphite-clickhouse
 * `carbonapi_v3_grpc` - grpc version of new carbonapi protocol. Currently no known implementation exists.
 * `msgpack` - messagepack based protocol, used in graphite-web 1.1 and metrictank. It's still experimental and might contain bugs.


Requirements
------------

You need to have Go >= 1.8 to build carbonapi from sources. Building with Go 1.7 is not supported for versions >0.8.0

CarbonAPI uses protobuf-based protocol to talk with underlying storages. For current version the compatibility list is:

1. [go-carbon](https://github.com/lomik/go-carbon) >= 0.9.0 (Note: you need to enable carbonserver in go-carbon). Recommended to run latest version, that currently supports `carbonapi_v3_pb`
2. [graphite-clickhouse](https://github.com/lomik/graphite-clickhouse) any. That's alternative storage that doesn't use Whisper.
3. [metrictank](https://github.com/grafana/metrictank) - supported via `msgpack` protocol. Support is not very well tested and might contain bugs. Use with cautions. Tags are not supported.
4. [carbonapi](https://github.com/go-graphite/carbonapi) >= 0.5. Note: starting from carbonapi 3596e9647611e1f833a911d663747271623ec003 (post 0.8) carbonapi can be used as a zipper's replacement
5. [carbonserver](https://github.com/grobian/carbonserver)@master (Note: you should probably switch to go-carbon in that case).
6. [carbonzipper](https://github.com/go-graphite/carbonzipper) >= 0.50. **Please note**, carbonzipper functionality was merged to carbonapi and it's no longer needed to run separate zipper. 


Some remarks on different backends
----------------------------------

For backends that uses normal database (e.x. graphite-clickhouse) you should set `maxGlobs: 0` in your config file for this backend group.

For other backends (e.x. go-carbon) you should set it to some reasonable value. It increases response speed, but the cost is increased memory consumption.

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

License
-------

This code is licensed under the BSD-2 license.
