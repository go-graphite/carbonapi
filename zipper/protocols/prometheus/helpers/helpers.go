package helpers

import (
	"math"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/go-graphite/carbonapi/zipper/protocols/prometheus/types"
)

// ConvertGraphiteTargetToPromQL - converts graphite target string to PromQL friendly format
func ConvertGraphiteTargetToPromQL(query string) string {
	var sb strings.Builder

	for {
		n := strings.IndexAny(query, "*[{")
		if n < 0 {
			sb.WriteString(regexp.QuoteMeta(query))
			return sb.String()
		}

		sb.WriteString(regexp.QuoteMeta(query[:n]))
		ch := query[n]
		query = query[n+1:]

		switch ch {
		case '*':
			if query == "" {
				// needed to support find requests when asterisk is the last character and dots should be included
				sb.WriteString(".*")
				break
			}

			sb.WriteString("[^.]*?")

		case '[':
			n = strings.Index(query, "]")
			if n < 0 {
				sb.WriteString(regexp.QuoteMeta("[" + query))
				return sb.String()
			}
			sb.WriteString("[" + query[:n+1])
			query = query[n+1:]

		case '{':
			n = strings.Index(query, "}")
			if n < 0 {
				sb.WriteString(regexp.QuoteMeta("{" + query))
				return sb.String()
			}
			alts := strings.Split(query[:n], ",")
			query = query[n+1:]
			for i := range alts {
				alts[i] = regexp.QuoteMeta(alts[i])
			}
			sb.WriteString("(" + strings.Join(alts, "|") + ")")
		}
	}
}

// AlignValues inserts math.NaN() in place of gaps in data from Prometheus
func AlignValues(startTime, stopTime, step int64, promValues []types.Value) []float64 {
	var (
		promValuesCtr = 0
		resValues     = make([]float64, (stopTime-startTime)/step+1)
	)

	for i := range resValues {
		nextTimestamp := float64(startTime + int64(i)*step)

		if promValuesCtr < len(promValues) && promValues[promValuesCtr].Timestamp == nextTimestamp {
			resValues[i] = promValues[promValuesCtr].Value
			promValuesCtr++
			continue
		}

		resValues[i] = math.NaN()
	}

	return resValues
}

// AdjustStep adjusts step keeping in mind default/configurable limit of maximum points per query
// Steps sequence is aligned with Grafana. Step progresses in the following order:
// minimal configured step if not default => 20 => 30 => 60 => 120 => 300 => 600 => 900 => 1200 => 1800 => 3600 => 7200 => 10800 => 21600 => 43200 => 86400
func AdjustStep(start, stop, maxPointsPerQuery, minStep int64, forceMinStepInterval time.Duration) int64 {

	interval := float64(stop - start)

	if forceMinStepInterval.Seconds() > interval {
		return minStep
	}

	safeStep := int64(math.Ceil(interval / float64(maxPointsPerQuery)))

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

// PromethizeTagValue - accept 'Tag=value' or 'Tag=~value' string and return sanitized version of it
func PromethizeTagValue(tagValue string) (string, types.Tag) {
	// Handle = and =~
	var (
		t       types.Tag
		tagName string
		idx     = strings.Index(tagValue, "=")
	)

	if idx < 0 {
		return tagName, t
	}

	if idx > 0 && tagValue[idx-1] == '!' {
		t.OP = "!"
		tagName = tagValue[:idx-1]
	} else {
		tagName = tagValue[:idx]
	}

	switch {
	case idx+1 == len(tagValue): // != or = with empty value
		t.OP += "="
	case tagValue[idx+1] == '~':
		if len(t.OP) > 0 { // !=~
			t.OP += "~"
		} else { // =~
			t.OP = "=~"
		}

		if idx+2 < len(tagValue) { // check is not empty value
			t.TagValue = tagValue[idx+2:]
		}
	default: // != or = with value
		t.OP += "="
		t.TagValue = tagValue[idx+1:]
	}

	return tagName, t
}

// SplitTagValues - For given tag-value list converts it to more usable map[string]Tag, where string is TagName
func SplitTagValues(query string) map[string]types.Tag {
	tags := strings.Split(query, ",")
	result := make(map[string]types.Tag)
	for _, tvString := range tags {
		tvString = strings.TrimSpace(tvString)
		name, tag := PromethizeTagValue(tvString[1 : len(tvString)-1])
		result[name] = tag
	}
	return result
}

// PromMetricToGraphite converts prometheus metric name to a format expected by graphite
func PromMetricToGraphite(metric map[string]string) string {
	var res strings.Builder

	res.WriteString(metric["__name__"])
	delete(metric, "__name__")

	keys := make([]string, 0, len(metric))
	for k := range metric {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		res.WriteString(";" + k + "=" + metric[k])
	}

	return res.String()
}

// SeriesByTagToPromQL converts graphite SeriesByTag to PromQL
// will return step if __step__ is passed
func SeriesByTagToPromQL(step, target string) (string, string) {
	firstTag := true
	var queryBuilder strings.Builder
	tagsString := target[len("seriesByTag(") : len(target)-1]
	tvs := SplitTagValues(tagsString)
	// It's ok to have empty "__name__"
	if v, ok := tvs["__name__"]; ok {
		if v.OP == "=" {
			queryBuilder.WriteString(v.TagValue)
		} else {
			firstTag = false
			queryBuilder.WriteByte('{')
			queryBuilder.WriteString("__name__" + v.OP + "\"" + v.TagValue + "\"")
		}

		delete(tvs, "__name__")
	}
	for tagName, t := range tvs {
		if tagName == "__step__" {
			step = t.TagValue
			continue
		}
		if firstTag {
			firstTag = false
			queryBuilder.WriteByte('{')
			queryBuilder.WriteString(tagName + t.OP + "\"" + t.TagValue + "\"")
		} else {
			queryBuilder.WriteString(", " + tagName + t.OP + "\"" + t.TagValue + "\"")
		}
	}
	if !firstTag {
		queryBuilder.WriteByte('}')
	}
	return step, queryBuilder.String()
}
