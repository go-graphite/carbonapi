Changes
=================================================

[Fix] - bugfix

**[Breaking]** - breaking change

[Feature] - new feature

[Improvement] - non-breaking improvement

[Code] - code quality related change that shouldn't make any significant difference for end-user

CHANGELOG
---------

**WIP**
 - **[Breaking]** Most of the configuration options moved to a config file.
 - **[Breaking]** Logging migrated to zap (structured logging). Log format changed significantly. Old command line options removed. Please consult carbonapi.example.yaml for a new config options and explanations
 - [Improvement] CarbonAPI now passes extra headers to keep track of requests.
 - [Improvement] Allow '^' and '$' as name characters. Required for carbonsearch text-match pattern anchoring.
 - [Improvement] Allow to bind to arbitrary IP:Port. (thx to @zdykstra, #194), Fixes #176.
 - [Improvement] Extend legendValues to append multiple values.
 - [Improvement] Add a switch to turn off propagating globbed requests to carbonzipper (was introduced in 0.7.0)
 - [Feature] Make prefix configurable (thx to @StephenPCG, #189)
 - [Fix] Fi issue with missing labels on graphs > 14d
 - [Fix] Allow expressions to use 'total' instead of 'sum'. Required for graphite compatibility.

**0.7.0**
 - [Improvement] Render requests with globs now sent as-is.
 - [Improvement] Added ProxyHeaders handler to get real client's ip address from X-Real-IP or X-Forwarded-For headers.
 - [Improvement] Highest\* and Lowest\* now assumes '1' if no second argument was passed
 - [Feature] Implement legendValue expression (see http://graphite.readthedocs.io/en/latest/functions.html#graphite.render.functions.legendValue for more information)

Notes on migration:
It's highly recommended, if you use carbonzipper to also upgrade it to version 0.62

**0.6.0**
 - **[Breaking]** Migrate to protobuf3

**<=0.5.0**
There is no dedicated changelog for older versions of carbonapi. Please see commit log for more information about what changed for each commit.
