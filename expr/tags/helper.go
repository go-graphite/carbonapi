package tags

import (
	"strings"
)

// ExtractTags extracts all graphite-style tags out of metric name
// E.x. cpu.usage_idle;cpu=cpu-total;host=test => {"name": "cpu.usage_idle", "cpu": "cpu-total", "host": "test"}
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
			kv := strings.Split(newS, "=")
			result[kv[0]] = kv[1]
			break
		}

		kv := strings.Split(newS[:idx], "=")
		result[kv[0]] = kv[1]
		newS = newS[idx+1:]
	}

	return result
}
