all: carbonapi
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
        EXTRA_PKG_CONFIG_PATH=/opt/X11/lib/pkgconfig
endif
VERSION ?= $(shell git describe --abbrev=4 --dirty --always --tags)

GO ?= go

carbonapi:
	PKG_CONFIG_PATH="$(EXTRA_PKG_CONFIG_PATH)" $(GO) build -v -tags cairo -ldflags '-X main.BuildVersion=$(VERSION)'

debug:
	PKG_CONFIG_PATH="$(EXTRA_PKG_CONFIG_PATH)" $(GO) build -v -tags cairo -ldflags '-X main.BuildVersion=$(VERSION)' -gcflags=all='-l -N'

nocairo:
	$(GO) build -ldflags '-X main.BuildVersion=$(VERSION)'

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
	cp ./carbonapi.example.yaml $(DESTDIR)/usr/share/carbonapi/


clean:
	rm -f carbonapi
	rm -f *.deb
	rm -f *.rpm
