package metric

import (
	"unicode"
	"unicode/utf8"

	"github.com/go-graphite/carbonapi/pkg/parser"
)

// ExtractMetric extracts metric out of function list
func ExtractMetric(s string) string {
	// search for a metric name in 's'
	// metric name is defined to be a Series of name characters terminated by a ',' or ')'
	// work sample: bla(bla{bl,a}b[la,b]la) => bla{bl,a}b[la

	var (
		start, braces, i, w int
		r                   rune
	)

FOR:
	for braces, i, w = 0, 0, 0; i < len(s); i += w {

		w = 1
		if parser.IsNameChar(s[i]) {
			continue
		}

		switch s[i] {
		// If metric name have tags, we want to skip them
		case ';':
			break FOR
		case '{':
			braces++
		case '}':
			if braces == 0 {
				break FOR
			}
			braces--
		case ',':
			if braces == 0 {
				break FOR
			}
		case ')':
			break FOR
		case '=':
			// allow metric name to end with any amount of `=` without treating it as a named arg or tag
			if i == len(s)-1 || s[i+1] == '=' || s[i+1] == ',' || s[i+1] == ')' {
				continue
			}
			fallthrough
		default:
			r, w = utf8.DecodeRuneInString(s[i:])
			if unicode.In(r, parser.RangeTables...) {
				continue
			}
			start = i + 1
		}
	}

	return s[start:i]
}
