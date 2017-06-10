carbonapi: replacement graphite API server
------------------------------------------

[![Build Status](https://travis-ci.org/go-graphite/carbonapi.svg?branch=master)](https://travis-ci.org/go-graphite/carbonapi)
[![GoDoc](https://godoc.org/github.com/go-graphite/carbonapi?status.svg)](https://godoc.org/github.com/go-graphite/carbonapi)
[![PR Stats](http://issuestats.com/github/go-graphite/carbonapi/badge/pr)](http://issuestats.com/github/go-graphite/carbonapi)
[![Issues Stats](http://issuestats.com/github/go-graphite/carbonapi/badge/issue)](http://issuestats.com/github/go-graphite/carbonapi)

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

The only required parameter is the path to a config file. For an example see `carbonapi.example.yaml`

`$ ./carbonapi -config /etc/carbonapi.yaml`

Request metrics will be dumped to graphite if coresponding config options are set,
or if the GRAPHITEHOST/GRAPHITEPORT environment variables are found.

Request data will be stored in memory (default) or in memcache.


Requirements
------------

CarbonAPI uses protobuf-based protocol to talk with underlying storages. For current version the compatibility list is:

1. [carbonzipper](https://github.com/go-graphite/carbonzipper) >= 0.50
2. [go-carbon](https://github.com/lomik/go-carbon) >= 0.9.0 (Note: you need to enable carbonserver in go-carbon). Limitations: /info handler won't work properly
3. [carbonserver](https://github.com/grobian/carbonserver)@master (Note: you should probably switch to go-carbon in that case). Limitations: /info handler won't work properly
4. [graphite-clickhouse](https://github.com/lomik/graphite-clickhouse) any. That's alternative storage that doesn't use Whisper. Limitations: /info handler won't work properly.
5. [carbonapi](https://github.com/go-graphite/carbonapi) >= 0.5. Note: we are not sure if there is any point in running carbonapi over carbonapi at this moment.

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
