all: carbonapi
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
        EXTRA_PKG_CONFIG_PATH=/opt/X11/lib/pkgconfig
endif
VERSION ?= $(shell git describe --abbrev=4 --dirty --always --tags)

GO ?= go

PKG_CARBONAPI=github.com/go-graphite/carbonapi/cmd/carbonapi
PKG_CARBONZIPPER=github.com/go-graphite/carbonzipper/cmd/carbonzipper

carbonapi:
	PKG_CONFIG_PATH="$(EXTRA_PKG_CONFIG_PATH)" $(GO) build -v -tags cairo -ldflags '-X main.BuildVersion=$(VERSION)' $(PKG_CARBONAPI)

debug:
	PKG_CONFIG_PATH="$(EXTRA_PKG_CONFIG_PATH)" $(GO) build -v -tags cairo -ldflags '-X main.BuildVersion=$(VERSION)' -gcflags=all='-l -N' $(PKG_CARBONAPI)

nocairo:
	$(GO) build -ldflags '-X main.BuildVersion=$(VERSION)' $(PKG_CARBONAPI)

carbonzipper: dep
	$(GO) build --ldflags '-X main.BuildVersion=$(VERSION)' $(PKG_CARBONZIPPER)

test:
	PKG_CONFIG_PATH="$(EXTRA_PKG_CONFIG_PATH)" $(GO) test -tags cairo ./... -race

vet:
	$(GO) vet

dep:
	@which dep 2>/dev/null || $(GO) get github.com/golang/dep/cmd/dep
	dep ensure

depupd:
	@which dep 2>/dev/null || $(GO) get github.com/golang/dep/cmd/dep
	dep ensure -update

install:
	mkdir -p $(DESTDIR)/usr/bin/
	mkdir -p $(DESTDIR)/usr/share/carbonapi/
	cp ./carbonapi $(DESTDIR)/usr/bin/
	cp ./cmd/carbonapi/carbonapi.example.yaml $(DESTDIR)/usr/share/carbonapi/

install_carbonzipper:
	mkdir -p $(DESTDIR)/usr/bin/
	mkdir -p $(DESTDIR)/usr/share/carbonzipper/
	cp ./carbonzipper $(DESTDIR)/usr/bin/
	cp ./cmd/carbonzipper/example.conf $(DESTDIR)/usr/share/carbonzipper/

clean:
	rm -f carbonapi carbonzipper
	rm -f *.deb
	rm -f *.rpm
