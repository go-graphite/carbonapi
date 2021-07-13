# gosnowth

[![Go Reference](https://pkg.go.dev/badge/github.com/circonus-labs/gosnowth.svg)](https://pkg.go.dev/github.com/circonus-labs/gosnowth)

## An IRONdb API client package for Go programs

This codebase contains client code for accessing IRONdb API's. IRONdb is
deployed as a cluster of database nodes which can be queried directly through
an exposed HTTP API. For reference, see the the
[IRONdb API Documentation](https://github.com/circonus/irondb-docs/blob/master/api.md).

Each of the documented interfaces are implemented as methods of the SnowthClient
structure defined in this repository. Documentation for this package is
available through the `go doc` tool using the following command:

``` bash
go doc github.com/circonus-labs/gosnowth
```

## Testing

The following command will run the unit tests for this package:

``` bash
go test -cover github.com/circonus-labs/gosnowth
```

## Using

Examples of using this package are provided in the in the `/examples` directory
which shows how to instantiate a new SnowthClient value, as well as how to use
the SnowthClient to perform operations on IRONdb nodes.

To run the examples use the following command:

``` bash
go run github.com/circonus-labs/gosnowth/examples <host:port> ...
```

Where `<host:port> ...` is a list of one or more space separated IRONdb nodes.

## Other IRONdb Clients

[Here](docs/SnowthClients.md) is a comparison of functionality between gosnowth
and some other IRONdb client libraries.
