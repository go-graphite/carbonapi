---
name: Bug report
about: Create a report to help us improve
title: "[BUG] "
labels: bug
assignees: ''

---

**Describe the bug**
A clear and concise description of what the bug is.

**Logs**
If applicable, add logs (please use log level debug) that shows the whole duration of the request. Request can be identified by UUID if needed.

**CarbonAPI Configuration:**
Please include carbonapi config file that you are using (please replace all information you don't want to share, like domain names or IP addresses).

**Simplified query (if applicable)**
Please provide a query that triggered the issue, ideally narrowed down to smallest possible set of functions.

**Backend metric retention and aggregation schemas**
Please provide backend's schema (most important thing - if query cross retention period or not), aggregation function, xFilesFactor (if applicable).

**Backend response (if possible)**
If that's possible - please share some sample set of backend responses. Most important thing here:
1. Are all backend responses the same? If not, what's the difference?
2. Do they contain special values, like Inf/NaN values?
3. Do all of them have same step?

**Additional context**
Add any other context about the problem here. Like version of a backend.
