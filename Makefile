all: carbonapi mockbackend
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
        EXTRA_PKG_CONFIG_PATH ?= /opt/X11/lib/pkgconfig
endif
VERSION ?= $(shell git describe --abbrev=4 --dirty --always --tags)

GO ?= go

PKG_CARBONAPI=github.com/go-graphite/carbonapi/cmd/carbonapi
PKG_MOCKBACKEND=github.com/go-graphite/carbonapi/cmd/mockbackend

carbonapi: $(shell find . -name '*.go' | grep -v 'vendor')
	PKG_CONFIG_PATH="$(EXTRA_PKG_CONFIG_PATH)" GO111MODULE=on $(GO) build -mod=vendor -v -tags cairo -ldflags '-X main.BuildVersion=$(VERSION)' $(PKG_CARBONAPI)

mockbackend: $(shell find . -name '*.go' | grep -v 'vendor')
	GO111MODULE=on $(GO) build -mod=vendor -v -ldflags '-X main.BuildVersion=$(VERSION)' $(PKG_MOCKBACKEND)

debug:
	PKG_CONFIG_PATH="$(EXTRA_PKG_CONFIG_PATH)" GO111MODULE=on $(GO) build -mod=vendor -v -tags cairo -ldflags '-X main.BuildVersion=$(VERSION)' -gcflags=all='-l -N' $(PKG_CARBONAPI)

nocairo:
	GO111MODULE=on $(GO) build -mod=vendor -ldflags '-X main.BuildVersion=$(VERSION)' $(PKG_CARBONAPI)

test:
	PKG_CONFIG_PATH="$(EXTRA_PKG_CONFIG_PATH)" $(GO) test $(EXTRA_LINK_FLAGS) -mod=vendor -tags cairo ./... -race

test_nocairo:
	$(GO) test -mod=vendor ./... -race

vet:
	$(GO) vet

install:
	mkdir -p $(DESTDIR)/usr/bin/
	mkdir -p $(DESTDIR)/usr/share/carbonapi/
	cp ./carbonapi $(DESTDIR)/usr/bin/
	cp ./cmd/carbonapi/carbonapi.example.yaml $(DESTDIR)/usr/share/carbonapi/

clean:
	rm -f carbonapi mockbackend
	rm -f *.deb
	rm -f *.rpm
