carbonapi: replacement graphite API server
------------------------------------------

[![Build Status](https://drone.io/github.com/dgryski/carbonapi/status.png)](https://drone.io/github.com/dgryski/carbonapi/latest)
[![GoDoc](https://godoc.org/github.com/dgryski/carbonapi?status.svg)](https://godoc.org/github.com/dgryski/carbonapi)


CarbonAPI supports a significant subset of graphite functions [see [COMPATIBILITY](COMPATIBILITY.md)].
In our testing it has shown to be 5x-10x faster than requesting data from graphite-web.

To use this, you must have a [carbonzipper](https://github.com/dgryski/carbonzipper)
install, which in turn requires that your
carbon stores are running [carbonserver](https://github.com/grobian/carbonserver)

The only required parameter is the address of the zipper to connect to.

$ ./carbonapi -z=http://zipper:8080

Request metrics will be dumped to graphite if the -graphite flag is provided,
or if the GRAPHITEHOST/GRAPHITEPORT environment variables are found.

Request data will be stored in memory (default) or in memcache.

OSX Build Notes
---------------
Some additional steps may be needed to build carbonapi with cairo rendering on MacOSX.

Install cairo:

$ brew install Caskroom/cask/xquartz

$ brew install cairo --with-x11

Set `PKG_CONFIG_PATH` and `-tags cairo` when building:

$ PKG_CONFIG_PATH=/opt/X11/lib/pkgconfig go build -v -tags cairo

Acknowledgement
---------------
This program was originally developed for Booking.com.  With approval
from Booking.com, the code was generalised and published as Open Source
on github, for which the author would like to express his gratitude.

License
-------

This code is licensed under the MIT license.
