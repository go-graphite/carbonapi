package stringutils

import (
	"strings"
	"unicode/utf8"
)

var empthy = ""

// Split2 return the split string results (without memory allocations)
//
//	If sep string not found: 's' '' 1
//	If s or sep string is empthy: 's' '' 1
//	In other cases: 's0' 's2' 2
func Split2(s string, sep string) (string, string, int) {
	if len(sep) == 0 {
		return s, empthy, 1
	}

	if pos := strings.Index(s, sep); pos == -1 {
		return s, empthy, 1
	} else if pos == len(s)-len(sep) {
		return s[0:pos], empthy, 2
	} else {
		return s[0:pos], s[pos+len(sep):], 2
	}
}

// Split return splitted slice (use pre-allocated buffer) (realloc if needed)
func Split(s string, sep string, buf []string) []string {
	buf = buf[:0]

	for {
		if pos := strings.Index(s, sep); pos == -1 {
			buf = append(buf, s)
			break
		} else {
			buf = append(buf, s[0:pos])
			s = s[pos+len(sep):]
		}
	}
	return buf
}

// SplitByte return splitted slice (use pre-allocated buffer) (realloc if needed)
func SplitByte(s string, sep byte, buf []string) []string {
	buf = buf[:0]

	for {
		if pos := strings.IndexByte(s, sep); pos == -1 {
			buf = append(buf, s)
			break
		} else {
			buf = append(buf, s[0:pos])
			s = s[pos+1:]
		}
	}
	return buf
}

// SplitRune return splitted slice (use pre-allocated buffer) (realloc if needed)
func SplitRune(s string, sep rune, buf []string) []string {
	buf = buf[:0]

	w := utf8.RuneLen(sep)

	for {
		if pos := strings.IndexRune(s, sep); pos == -1 {
			buf = append(buf, s)
			break
		} else {
			buf = append(buf, s[0:pos])
			s = s[pos+w:]
		}
	}
	return buf
}
