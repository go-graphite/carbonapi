VERSION=0.26
distdir=carbonzipper-$(VERSION)

carbonzipper: fetchdeps
	GOPATH=`pwd`/Godeps/_workspace go build -o carbonzipper

fetchdeps:
	GOPATH=`pwd`/Godeps/_workspace go get -d

updatedeps:
	GOPATH=`pwd`/Godeps/_workspace go get -du

dist: fetchdeps
	godep save
	mkdir $(distdir)
	mv Godeps $(distdir)
	cp Makefile *.go $(distdir)
	tar zvcf $(distdir).tar.gz $(distdir)
	rm -rf $(distdir)
