VERSION := $(shell git describe --always --tags)

GO ?= go

all: test

FORCE:

test: FORCE
	$(GO) test -race -coverprofile coverage.txt  ./...

lint:
	golangci-lint run
