carbonapi: replacement graphite API server
------------------------------------------

[![Build Status](https://travis-ci.org/dgryski/carbonapi.svg?branch=master)](https://travis-ci.org/dgryski/carbonapi)
[![GoDoc](https://godoc.org/github.com/dgryski/carbonapi?status.svg)](https://godoc.org/github.com/dgryski/carbonapi)


CarbonAPI supports a significant subset of graphite functions [see [COMPATIBILITY](COMPATIBILITY.md)].
In our testing it has shown to be 5x-10x faster than requesting data from graphite-web.

To use this, you must have a [carbonzipper](https://github.com/dgryski/carbonzipper)
install, which in turn requires that your
carbon stores are running [go-carbon](https://github.com/lomik/go-carbon) (Note: you need to enable carbonserver in go-carbon)

It's possible to talk directly with `go-carbon`'s carbonserver module, but custom `/info` handler won't work.

The only required parameter is the path to a config file. For an example see `carbonapi.example.yaml`

`$ ./carbonapi -config /etc/carbonapi.yaml`

Request metrics will be dumped to graphite if coresponding config options are set,
or if the GRAPHITEHOST/GRAPHITEPORT environment variables are found.

Request data will be stored in memory (default) or in memcache.

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

This code is licensed under the MIT license.
