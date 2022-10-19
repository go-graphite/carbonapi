package metrics

import (
	"sort"
	"strings"
)

// MergeTags merge two tag maps into one tag map
func MergeTags(a, b map[string]string) map[string]string {
	var dst map[string]string
	if len(a) == 0 {
		if len(b) == 0 {
			return nil
		}
		dst = b
	} else if len(b) == 0 {
		dst = a
	} else {
		dst = make(map[string]string)
		for k, v := range a {
			dst[k] = v
		}
		for k, v := range b {
			if _, exist := dst[k]; !exist {
				dst[k] = v
			}
		}
	}
	return dst
}

// JoinTags convert tags map sorted tags string representation (separated by comma), like tags or Graphite
func JoinTags(tagsMap map[string]string) string {
	tags := make([]string, 0, len(tagsMap))
	for k, v := range tagsMap {
		tags = append(tags, k+"="+v)
	}
	sort.Strings(tags)
	return ";" + strings.Join(tags, ";")
}
