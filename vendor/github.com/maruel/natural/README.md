# natural

Yet another natural sort, with 100% test coverage and a benchmark. It allocates
no memory, doesn't depend on package `sort` and hence doesn't depend on `reflect`.

[![GoDoc](https://godoc.org/github.com/maruel/natural?status.svg)](https://godoc.org/github.com/maruel/natural)


## Benchmarks

On Go 1.10.1.

On a Xeon:

```
$ go test -bench=. -benchmem -cpu 1
goos: linux
goarch: amd64
pkg: github.com/maruel/natural
BenchmarkNative                200000000       8.63 ns/op     0 B/op     0 allocs/op
BenchmarkLessStringOnly        100000000       18.7 ns/op     0 B/op     0 allocs/op
BenchmarkLessDigits             50000000       30.5 ns/op     0 B/op     0 allocs/op
BenchmarkLessStringDigits       50000000       31.3 ns/op     0 B/op     0 allocs/op
BenchmarkLessDigitsTwoGroups    20000000       64.4 ns/op     0 B/op     0 allocs/op
```

On a Raspberry Pi 3:

```
$ go test -bench=. -benchmem -cpu 1
goos: linux
goarch: arm
pkg: github.com/maruel/natural
BenchmarkNative                 10000000        148 ns/op     0 B/op      0 allocs/op
BenchmarkLessStringOnly          5000000        312 ns/op     0 B/op      0 allocs/op
BenchmarkLessDigits              2000000        656 ns/op     0 B/op      0 allocs/op
BenchmarkLessStringDigits        2000000        679 ns/op     0 B/op      0 allocs/op
BenchmarkLessDigitsTwoGroups     1000000       1480 ns/op     0 B/op      0 allocs/op
```

Coverage:

```
$ go test -cover
PASS
coverage: 100.0% of statements
ok     github.com/maruel/natural       0.012s
```

