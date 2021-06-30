package irondb

import (
	"fmt"
	"math"
	"regexp"
	"strings"
)

// graphiteExprListToIronDBTagQuery - converts list of Graphite Tag expressions to IronDB Tag query
// e.g. ["name=~cpu\..*", "tag1!=value1"] => and(__name:/cpu\..*/,not(tag1:value1))
func graphiteExprListToIronDBTagQuery(exprList []string) string {
	//r := make([]string, len(exprList))
	var r []string

	for _, expr := range exprList {
		// remove escaped quotes
		expr = strings.ReplaceAll(expr, "\"", "")
		eq_idx := strings.Index(expr, "=")
		if eq_idx > 0 {
			tag_name := expr[:eq_idx]
			ne_idx := strings.Index(expr, "!")
			if ne_idx != -1 {
				tag_name = expr[:ne_idx]
			}
			if tag_name == "name" {
				tag_name = "__name"
			}
			if ne_idx == -1 {
				// Handle = and =~
				if expr[eq_idx+1] == '~' {
					// op = "=~"
					// tag_value = expr[eq_idx+2:]
					r = append(r, fmt.Sprintf("and(%s:/%s/)", tag_name, expr[eq_idx+2:]))
				} else {
					// op = "="
					// tag_value = expr[eq_idx+1:]
					r = append(r, fmt.Sprintf("and(%s:%s)", tag_name, expr[eq_idx+1:]))
				}
			} else {
				// Handle != and !=~
				if expr[eq_idx+1] == '~' {
					// op = "!~"
					// tag_value = expr[eq_idx+2:]
					r = append(r, fmt.Sprintf("not(%s:/%s/)", tag_name, expr[eq_idx+2:]))
				} else {
					// op = "!="
					// tag_value = expr[idx+1:]
					r = append(r, fmt.Sprintf("not(%s:%s)", tag_name, expr[eq_idx+1:]))
				}
			}
		}
	}
	if len(r) > 0 {
		if len(r) > 1 {
			return fmt.Sprintf("and(%s)", strings.Join(r, ","))
		}
		return r[0]
	}
	return ""
}

// convertNameToGraphite - convert IronDB tagged name to Graphite-web
func convertNameToGraphite(name string) string {
	// if name contains tags - convert to same format as python graphite-irondb uses
	// remove all MT tags
	if strings.Contains(name, "|MT{") {
		mt_re := regexp.MustCompile(`\|MT{([^}]*)}`)
		name = mt_re.ReplaceAllString(name, "")
	}
	if strings.Contains(name, "|ST[") {
		name = strings.ReplaceAll(name, "|ST[", ";")
		name = strings.ReplaceAll(name, "]", "")
		name = strings.ReplaceAll(name, ",", ";")
		name = strings.ReplaceAll(name, ":", "=")
	}
	return name
}

// AdjustStep adjusts step keeping in mind default/configurable limit of maximum points per query
// Steps sequence is aligned with Grafana. Step progresses in the following order:
// minimal configured step if not default => 20 => 30 => 60 => 120 => 300 => 600 => 900 => 1200 => 1800 => 3600 => 7200 => 10800 => 21600 => 43200 => 86400
func adjustStep(start, stop, maxPointsPerQuery, minStep int64) int64 {
	safeStep := int64(math.Ceil(float64(stop-start) / float64(maxPointsPerQuery)))

	step := minStep
	if safeStep > minStep {
		step = safeStep
	}

	switch {
	case step <= minStep:
		return minStep // minimal configured step
	case step <= 20:
		return 20 // 20s
	case step <= 30:
		return 30 // 30s
	case step <= 60:
		return 60 // 1m
	case step <= 120:
		return 120 // 2m
	case step <= 300:
		return 300 // 5m
	case step <= 600:
		return 600 // 10m
	case step <= 900:
		return 900 // 15m
	case step <= 1200:
		return 1200 // 20m
	case step <= 1800:
		return 1800 // 30m
	case step <= 3600:
		return 3600 // 1h
	case step <= 7200:
		return 7200 // 2h
	case step <= 10800:
		return 10800 // 3h
	case step <= 21600:
		return 21600 // 6h
	case step <= 43200:
		return 43200 // 12h
	default:
		return 86400 // 24h
	}
}
