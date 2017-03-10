carbonzipper: carbonserver proxy for graphite-web
=================================================

CarbonZipper is the central part of a replacement graphite storage stack.  It
proxies requests from graphite-web to a cluster of carbon storage backends.
Previous versions (available in the git history) were able to talk to python
carbon stores, but the current version requires the use of
[carbonserver](https://github.com/grobian/carbonserver).

Configuration is done via a JSON file loaded at startup.  The only required
field is the list of carbonserver backends to connect to.

Other pieces of the stack are:
   - [carbonapi](https://github.com/dgryski/carbonapi)
   - [carbonmem](https://github.com/dgryski/carbonmem)
   - [carbonsearch](https://github.com/kanatohodets/carbonsearch)


Changes
-------
**0.62**
   - Fix carbonsearch queries with recent carbonapi version
   - Fix pathCache to handle render requests with globs.
   - Add cache for carbonsearch results

**0.61**
   - Fix rewrite for internal queries, because of an error some queries were sent as protobuf not as protobuf3
   - gofmt the code!

**0.60**
   - **BREAKING CHANGE** Carbonzipper backend protocol changed to protobuf3. Though output for /render, /info /find can be both (format=protobuf3 for protobuf3, format=protobuf for protobuf2).

**0.50**
   - See commit log.


Upgrading to 0.60 from 0.50 or earlier
--------------------------------------

Starting from 0.60, carbonzipper will be able to talk **only** with storages compatible with **protobuf3**.

At this moment (0.60) it's only go-carbon, starting from commit ee2bc24 (post 0.9.1)

Carbonzipper can still return results in protobuf and compatibility won't be removed at least until Summer 2017.

If you want to upgrade, the best option is to do follwing steps:

1. Migrate to go-carbon post 0.9.1 release. (note: carbonserver isn't compatible with this version of zipper)
2. Migrate to carbonsearch 0.16.0 (if you are using any)
3. Upgrade carbonzipper to 0.60 or newer.
4. Upgrade carbonapi to 0.6.0 (commit 119e346 or newer) (optional, but advised)


Acknowledgement
---------------
This program was originally developed for Booking.com.  With approval
from Booking.com, the code was generalised and published as Open Source
on github, for which the author would like to express his gratitude.

License
-------

This code is licensed under the MIT license.
