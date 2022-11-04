package types

import (
	"strings"
	"unicode/utf8"
)

var allowedCharactersInMetricName = map[byte]struct{}{
	'=': struct{}{},
	'@': struct{}{},
}

func byteAllowedInName(b byte) bool {
	_, ok := allowedCharactersInMetricName[b]
	return ok
}

// ExtractNameLoc extracts name start:end location out of function list with . Only for use in MetrciData name parse, it's has more allowed chard in names, like =
func ExtractNameLoc(s string) (int, int) {
	// search for a metric name in 's'
	// metric name is defined to be a Series of name characters terminated by a ',' or ')'

	var w, i, start, braces int
	var c rune
FOR:
	for i, c = range s {
		w = i
		switch c {
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
		case '(':
			if i >= 11 {
				n := i - 11
				if s[n:i] == "seriesByTag" {
					end := strings.IndexRune(s[n:], ')')
					if end == -1 {
						return n, len(s) // broken end of args, can't exist in correctly parsed functions
					} else {
						return n, n + end + 1
					}
				}
			}
			start = i + 1
		case ')':
			break FOR
		}
		w += utf8.RuneLen(c)
	}

	return start, w
}

// ExtractName extracts name out of function list. Only for use in MetrciData name parse, it's has more allowed chard in names, like =
func ExtractName(s string) string {
	start, end := ExtractNameLoc(s)
	return s[start:end]
}

// ExtractNameTag extracts name tag out of function list with . Only for use in MetrciData name parse, it's has more allowed chard in names, like =
func ExtractNameTag(s string) string {
	// search for a metric name in 's'
	// metric name is defined to be a Series of name characters terminated by a ',' or ')'

	var w, i, start, braces int
	var c rune
FOR:
	for i, c = range s {
		w = i
		switch c {
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
		case '(':
			if i >= 11 {
				n := i - 11
				if s[n:i] == "seriesByTag" {
					end := strings.IndexRune(s[n:], ')')
					if end == -1 {
						return s[n:len(s)] // broken end of args, can't exist in correctly parsed functions
					} else {
						return s[n : n+end+1]
					}
				}
			}
			start = i + 1
		case ')', ';':
			break FOR
		}
		w += utf8.RuneLen(c)
	}

	return s[start:w]
}
