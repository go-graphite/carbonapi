package stringutils

import (
	"strings"
	"unicode/utf8"
)

// Replace returns a copy of the string s with the first n
// non-overlapping instances of old replaced by new.
// Also return change flag.
// If old is empty, it matches at the beginning of the string
// and after each UTF-8 sequence, yielding up to k+1 replacements
// for a k-rune string.
// If n < 0, there is no limit on the number of replacements.
func Replace(s, old, new string, n int) (string, bool) {
	if old == new || n == 0 {
		return s, false // avoid allocation
	}

	// Compute number of replacements.
	if m := strings.Count(s, old); m == 0 {
		return s, false // avoid allocation
	} else if n < 0 || m < n {
		n = m
	}

	// Apply replacements to buffer.
	var b strings.Builder
	b.Grow(len(s) + n*(len(new)-len(old)))
	start := 0
	if len(old) == 0 {
		for i := 0; i < n; i++ {
			j := start
			if i > 0 {
				_, wid := utf8.DecodeRuneInString(s[start:])
				j += wid
			}
			b.WriteString(s[start:j])
			b.WriteString(new)
			start = j + len(old)
		}
	} else {
		for i := 0; i < n; i++ {
			j := strings.Index(s[start:], old)
			if j == -1 {
				break
			}
			j += start
			b.WriteString(s[start:j])
			b.WriteString(new)
			start = j + len(old)
		}
	}
	b.WriteString(s[start:])
	return b.String(), true
}

// ReplaceAll returns a copy of the string s with all
// non-overlapping instances of old replaced by new.
// Also return change flag.
// If old is empty, it matches at the beginning of the string
// and after each UTF-8 sequence, yielding up to k+1 replacements
// for a k-rune string.
func ReplaceAll(s, old, new string) (string, bool) {
	return Replace(s, old, new, -1)
}
