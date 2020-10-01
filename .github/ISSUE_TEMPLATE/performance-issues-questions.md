---
name: Performance issues/questions
about: Any questions or issues related to performance (CPU Usage, Query time, Memory
  usage)
title: "[Performance]"
labels: performance
assignees: ''

---

**Problem description**
In as much details as possible, please explain what's the problem. E.x. My query X runs Y seconds, instead of Z seconds?

**carbonapi's version**
Version of carbonapi. If you build your own packages, what golang version are you using?

**Does this happened before**
In case same queries ran fine before - please provide a version of carbonapi that didn't had that problems.

**carbonapi's config**
Configuration file of carbonapi.

**backend software and config**
What are you running on backend? What it's version? Config for the backend?

**carbonapi performance metrics**
1. Please enable `expvars` in the config and provide it's output
2. What's system load on a server? What's the memory consumption?
3. If possible, please provide profiler's output (https://blog.golang.org/pprof for instruction how to do that, `debug` handlers can be enabled in carbonapi's config on a separate port). Golang allows different types of profiles. It's always better to give CPU Profile, Memory profile and Lock profile, but that depends on what's the symptoms.
4. Debug logs could be helpful as well

**Query that causes problems**
1. What's the query?
2. How many metrics does it fetch?
3. What's the resolution of those metrics? If it's not consistent (e.x. different metrics have different resolution) please specify that and provide average metric resolution
4. How many datapoints per metric do you have?

**Additional context**
Add any other context or screenshots about the feature request here.
