---
name: Bug report
about: Create a report to help us improve
title: "[BUG] "
labels: bug
assignees: ''

---

**Describe the bug**
A clear and concise description of what the bug is.

**CarbonAPI Version**
What's the CarbonAPI version you are running? Does this issue reproducible on current master?

**Logs**
If applicable, add logs (please use log level debug) that shows the whole duration of the request. Request can be identified by UUID if needed.

**CarbonAPI Configuration:**
Please include carbonapi config file that you are using (please replace all information you don't want to share, like domain names or IP addresses).

**Simplified query (if applicable)**
Please provide a query that triggered the issue, ideally narrowed down to smallest possible set of functions.

If you have a complex query like `someFunction(otherFunction(yetAnotherFunction(metric.name), some_paramters), some.other.metric.name)` it is very helpful to try to remove some of the functions around the `metric.name` and check which one triggers the problem.

**Backend metric retention and aggregation schemas**
Please provide backend's schema (most important thing - if query cross retention period or not), aggregation function, xFilesFactor (if applicable).

**Backend response (if possible)**
If that's possible - please share some sample set of backend responses. Most important thing here:
1. Are all backend responses the same? If not, what's the difference?
2. Do they contain special values, like Inf/NaN values?
3. Do all of them have same step?

To get the backend response, you can send request for the same metrics towards the backend:
1. You should remove all the functions from the request as most backends do not support any of them
2. You might need to convert relative time offsets to unix time, e.x. if you have "from=-1h" you might need to pass `from={current-timestamp}-3600&until={current-timestamp}` where `{current timestamp}` should be replaced by actual value. On some systems you can use command line to get current timestamp `date +%s` or even you can ask for a timestamp that was 1 hour ago: `date +%s --date="1 hour ago"`

**Additional context**
Add any other context about the problem here. Like version of a backend.
