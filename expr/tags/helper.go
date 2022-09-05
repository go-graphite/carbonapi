package tags

import (
	"strings"
	"unicode/utf8"
)

type ParseStep int8

const (
	WantTag ParseStep = iota
	WantCmp
	WantDelim
)

// . (dot) is a special symbol in regexp, so if user use =~ instead of =, name with dots a transformed (and also can produce unwanted results)
// use \. for consistent results
// `.*` prepended to regexp without ^ at start
// `.*` appended to regexp without $ at end
func sanitizeRegex(s string) string {
	var result strings.Builder
	result.Grow(len(s) + 20)

	if strings.HasPrefix(s, "^") {
		s = s[1:]
	} else {
		result.WriteString("__*")
	}
	var closedRegex bool
	if strings.HasSuffix(s, "$") {
		s = s[0 : len(s)-1]
		closedRegex = true
	}
	// cleanup regexp
	s = strings.TrimRight(s, ".*")

	var prev rune
	for _, c := range s {
		switch c {
		case '.':
			if prev != '\\' {
				result.WriteString("__")
			} else {
				result.WriteRune(c)
			}
		default:
			result.WriteRune(c)
		}
		prev = c
	}

	if !closedRegex {
		result.WriteString("__*")
	}
	return result.String()
}

// parse seriesByTag args
func ExtractSeriesByTags(s, defaultName string) map[string]string {
	if s == "" || len(s) < 13 {
		return nil
	}
	s = s[12:]

	tags := make(map[string]string)

	startTag := 0
	startVal := 0
	step := WantTag
	var (
		i, w  int
		c     rune
		delim rune
	)
LOOP:
	for i < len(s) {
		c, w = utf8.DecodeRuneInString(s[i:])
		switch c {
		case ',':
			if step == WantDelim {
				step = WantTag
			}
			i++
		case ')':
			if step == WantTag || step == WantDelim {
				break LOOP
			}
			i++
		case '\'', '"':
			if step == WantTag {
				// new segment found
				step = WantCmp
				delim = c
				startTag = i + 1
			} else {
				step = WantDelim
			}
			i++
		case '=', '!', '~':
			var add bool
			var isRegex bool
			if step == WantCmp {
				tag := s[startTag:i]
				if tag == "__name__" {
					tag = "name"
				}
				p := s[i:]
				if strings.HasPrefix(p, "!=~") {
					isRegex = true
					i += 3
				} else if strings.HasPrefix(p, "!=") {
					i += 2
				} else if strings.HasPrefix(p, "=~") {
					isRegex = true
					add = true
					i += 2
				} else if strings.HasPrefix(p, "=") {
					add = true
					i++
				} else {
					i += w
					// broken comparator, skip
					continue
				}
				startVal = i
				end := strings.IndexRune(s[startVal:], delim)
				if add && tag != "" && end > 0 {
					var v string
					if isRegex {
						v = sanitizeRegex(s[startVal : startVal+end])
					} else {
						v = s[startVal : startVal+end]
					}
					tags[tag] = v
				}
				step = WantDelim
				i = startVal + end + 1
			} else {
				i += w
			}
		default:
			i += w
		}
	}

	if _, exist := tags["name"]; !exist {
		tags["name"] = defaultName
	}

	return tags
}

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
