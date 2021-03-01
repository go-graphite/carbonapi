package tags

import (
	"strings"
)

// ExtractTags extracts all graphite-style tags out of metric name
// E.x. cpu.usage_idle;cpu=cpu-total;host=test => {"name": "cpu.usage_idle", "cpu": "cpu-total", "host": "test"}
// There are some differences between how we handle tags and how graphite-web can do that. In our case it is possible
// to have empty value as it doesn't make sense to skip tag in that case but can be potentially useful
// Also we do not fail on invalid cases, but rather than silently skipping broken tags as some backends might accept
// invalid tag and store it and one of the purposes of carbonapi is to keep working even if backends gives us slightly
// broken replies.
func ExtractTags(s string) map[string]string {
	result := make(map[string]string)
	idx := strings.IndexRune(s, ';')
	if idx < 0 {
		result["name"] = s
		return result
	}

	result["name"] = s[:idx]

	newS := s[idx+1:]
	for {
		idx := strings.IndexRune(newS, ';')
		if idx < 0 {
			firstEqualSignIdx := strings.IndexRune(newS, '=')
			// tag starts with `=` sign or have zero length
			if newS == "" || firstEqualSignIdx == 0 {
				break
			}
			// tag doesn't have = sign at all
			if firstEqualSignIdx == -1 {
				result[newS] = ""
				break
			}

			result[newS[:firstEqualSignIdx]] = newS[firstEqualSignIdx+1:]
			break
		}

		firstEqualSignIdx := strings.IndexRune(newS[:idx], '=')
		// Got an empty tag or tag starts with `=`. That is totally broken, so skipping that
		if idx == 0 || firstEqualSignIdx == 0 {
			newS = newS[idx+1:]
			continue
		}

		// Tag doesn't have value
		if firstEqualSignIdx == -1 {
			result[newS[:idx]] = ""
			newS = newS[idx+1:]
			continue
		}

		result[newS[:firstEqualSignIdx]] = newS[firstEqualSignIdx+1 : idx]
		newS = newS[idx+1:]
	}

	return result
}
