package stringutils

import (
	"unicode"
	"unicode/utf8"
)

// based on strings.Map
// Map returns a copy of the string s with all its characters modified
// according to the mapping function. If mapping returns a negative value, the character is
// dropped from the string with no replacement.
func (sb *Builder) Map(mapping func(rune) rune, s string) {
	// In the worst case, the string can grow when mapped, making
	// things unpleasant. But it's so rare we barge in assuming it's
	// fine. It could also shrink but that falls out naturally.

	// The output buffer b is initialized on demand, the first
	// time a character differs.

	updated := false
	sb.Grow(sb.Len() + len(s) + utf8.UTFMax)

	for i, c := range s {
		r := mapping(c)
		if r == c && c != utf8.RuneError {
			continue
		}

		var width int
		if c == utf8.RuneError {
			c, width = utf8.DecodeRuneInString(s[i:])
			if width != 1 && r == c {
				continue
			}
		} else {
			width = utf8.RuneLen(c)
		}

		updated = true
		sb.WriteString(s[:i])
		if r >= 0 {
			sb.WriteRune(r)
		}

		s = s[i+width:]
		break
	}

	// Fast path for unchanged input
	if updated { // didn't call b.Grow above
		for _, c := range s {
			r := mapping(c)

			if r >= 0 {
				// common case
				// Due to inlining, it is more performant to determine if WriteByte should be
				// invoked rather than always call WriteRune
				if r < utf8.RuneSelf {
					sb.WriteByte(byte(r))
				} else {
					// r is not a ASCII rune.
					sb.WriteRune(r)
				}
			}
		}
	} else {
		sb.WriteString(s)
	}
}

// based on strings.ToUpper
// ToUpper returns s with all Unicode letters mapped to their upper case.
func (sb *Builder) WriteStringUpper(s string) {
	isASCII, hasLower := true, false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= utf8.RuneSelf {
			isASCII = false
			break
		}
		hasLower = hasLower || ('a' <= c && c <= 'z')
	}

	if isASCII { // optimize for ASCII-only strings.
		if hasLower {
			sb.Grow(sb.Len() + len(s))
			for i := 0; i < len(s); i++ {
				c := s[i]
				if 'a' <= c && c <= 'z' {
					c -= 'a' - 'A'
				}
				sb.WriteByte(c)
			}
		} else {
			sb.WriteString(s)
		}
	} else {
		sb.Map(unicode.ToUpper, s)
	}
}

// based on strings.ToLower
// ToLower returns s with all Unicode letters mapped to their lower case.
func (sb *Builder) WriteStringLower(s string) {
	isASCII, hasUpper := true, false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= utf8.RuneSelf {
			isASCII = false
			break
		}
		hasUpper = hasUpper || ('A' <= c && c <= 'Z')
	}

	if isASCII { // optimize for ASCII-only strings.
		if !hasUpper {
			sb.WriteString(s)
			return
		}
		sb.Grow(sb.Len() + len(s))
		for i := 0; i < len(s); i++ {
			c := s[i]
			if 'A' <= c && c <= 'Z' {
				c += 'a' - 'A'
			}
			sb.WriteByte(c)
		}
	} else {
		sb.Map(unicode.ToLower, s)
	}
}
