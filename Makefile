VERSION=0.26
distdir=carbonzipper-$(VERSION)
REV=`cat revision.txt`

carbonzipper: fetchdeps
	GOPATH=`pwd`/_deps go build -ldflags "-X main.BuildVersion `cat revision.txt`" -o carbonzipper

fetchdeps:
	GOPATH=`pwd`/_deps go get -d

updatedeps:
	GOPATH=`pwd`/_deps go get -du

dist: fetchdeps
	mkdir $(distdir)
	mv _deps $(distdir)
	cp Makefile *.go $(distdir)
	echo "REV is" $(REV)
	git rev-parse HEAD >$(distdir)/revision.txt
	tar zvcf $(distdir).tar.gz $(distdir)
	rm -rf $(distdir)
