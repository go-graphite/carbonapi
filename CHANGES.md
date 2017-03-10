CHANGELOG
---------

**0.7.0**
 - Render requests with globs now sent as-is.
 - Added ProxyHeaders handler to get real client's ip address from X-Real-IP or X-Forwarded-For headers.
 - Highest\* and Lowest\* now assumes '1' if no second argument was passed
 - Implement legendValue expression (see http://graphite.readthedocs.io/en/latest/functions.html#graphite.render.functions.legendValue for more information)

Notes on migration:
It's highly recommended, if you use carbonzipper to also upgrade it to version 0.62

**0.6.0**
 - Migrate to protobuf3

**<=0.5.0**
There is no dedicated changelog for older versions of carbonapi. Please see commit log for more information about what changed for each commit.
