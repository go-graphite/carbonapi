VERSION=0.21
distdir=carbonzipper-$(VERSION)

carbonzipper:
	GOPATH=`pwd`/Godeps/_workspace go build -o carbonzipper

dist:
	godep save
	mkdir $(distdir)
	mv Godeps $(distdir)
	cp Makefile *.go $(distdir)
	tar zvcf $(distdir).tar.gz $(distdir)
	rm -rf $(distdir)
