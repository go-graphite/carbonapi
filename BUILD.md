General build notes
===================


Before you start
----------------

carbonapi uses dep as a vendoring tool. Makefile will automatically `go get` it for you if it's not installed.

PNG support is optional (but enabled by default if you are using Makefile) and requires cairo library and it's development packages (libcairo-dev on Debian-based, cairo-devel on RHEL-compatible)


OSX Build Notes
---------------
Some additional steps may be needed to build carbonapi with cairo rendering on MacOSX.

Install cairo:

```
$ brew install Caskroom/cask/xquartz

$ brew install cairo --with-x11

```

Then follow build notes for Linux


Build Instructions
------------------

To get version with cairo support (required for PNG and SVG rendering) just run:

```
make
```


To get a version without cairo support, run:

```
make nocairo
```

