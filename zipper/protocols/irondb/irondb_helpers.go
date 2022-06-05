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
	var r []string
	var irondbOp string
	for _, expr := range exprList {
		// remove escaped quotes
		expr = strings.ReplaceAll(expr, "\"", "")
		eqIdx := strings.Index(expr, "=")
		if eqIdx > 0 {
			tagName := expr[:eqIdx]
			neqIdx := strings.Index(expr, "!")
			if neqIdx == -1 {
				irondbOp = "and"
			} else {
				irondbOp = "not"
				tagName = expr[:neqIdx]
			}
			if tagName == "name" {
				tagName = "__name"
			}
			if expr[eqIdx+1] == '~' {
				// op = "=~" or op = "!~"
				// tagValue = expr[eq_idx+2:]
				r = append(r, fmt.Sprintf("%s(%s:/%s/)", irondbOp, tagName, expr[eqIdx+2:]))
			} else {
				// op = "=" or op = "!="
				// tagValue = expr[eq_idx+1:]
				r = append(r, fmt.Sprintf("%s(%s:%s)", irondbOp, tagName, expr[eqIdx+1:]))
			}
		}
	}
	switch len(r) {
	case 0:
		return ""
	case 1:
		return r[0]
	default:
		return fmt.Sprintf("and(%s)", strings.Join(r, ","))
	}
}

// convertNameToGraphite - convert IronDB tagged name to Graphite-web
func convertNameToGraphite(name string) string {
	// if name contains tags - convert to same format as python graphite-irondb uses
	// remove all MT tags
	if strings.Contains(name, "|MT{") {
		mtRe := regexp.MustCompile(`\|MT{([^}]*)}`)
		name = mtRe.ReplaceAllString(name, "")
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
	safeStep := minStep
	if maxPointsPerQuery != 0 {
		safeStep = int64(math.Ceil(float64(stop-start) / float64(maxPointsPerQuery)))
	}

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
