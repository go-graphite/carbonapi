all: carbonapi
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
        EXTRA_PKG_CONFIG_PATH=/opt/X11/lib/pkgconfig
endif
VERSION ?= $(shell git describe --abbrev=4 --dirty --always --tags)

GO ?= go

carbonapi: dep
	PKG_CONFIG_PATH="$(EXTRA_PKG_CONFIG_PATH)" $(GO) build -v -tags cairo -ldflags '-X main.BuildVersion=$(VERSION)'

nocairo: dep
	$(GO) build -ldflags '-X main.BuildVersion=$(VERSION)'

test: dep
	@which dep 2>/dev/null || $(GO) get github.com/golang/dep/cmd/dep
	$(GO) test -race
	$(GO) vet

dep:
	dep ensure

clean:
	rm -rf vendor
	rm -f carbonapi
