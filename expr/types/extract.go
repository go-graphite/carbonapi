package types

import (
	"unicode"
	"unicode/utf8"

	"github.com/go-graphite/carbonapi/pkg/parser"
)

// ExtractName extracts name out of function list
func ExtractName(s string) string {
	// search for a metric name in 's'
	// metric name is defined to be a Series of name characters terminated by a ',' or ')'
	// work sample: bla(bla{bl,a}b[la,b]la,1) => bla{bl,a}b[la,b]la

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
			continue
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
