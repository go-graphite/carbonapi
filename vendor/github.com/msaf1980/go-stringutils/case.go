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

const (
	toLowerTable = "\x00\x01\x02\x03\x04\x05\x06\a\b\t\n\v\f\r\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f !\"#$%&'()*+,-./0123456789:;<=>?@abcdefghijklmnopqrstuvwxyz[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~\u007f\x80\x81\x82\x83\x84\x85\x86\x87\x88\x89\x8a\x8b\x8c\x8d\x8e\x8f\x90\x91\x92\x93\x94\x95\x96\x97\x98\x99\x9a\x9b\x9c\x9d\x9e\x9f\xa0\xa1\xa2\xa3\xa4\xa5\xa6\xa7\xa8\xa9\xaa\xab\xac\xad\xae\xaf\xb0\xb1\xb2\xb3\xb4\xb5\xb6\xb7\xb8\xb9\xba\xbb\xbc\xbd\xbe\xbf\xc0\xc1\xc2\xc3\xc4\xc5\xc6\xc7\xc8\xc9\xca\xcb\xcc\xcd\xce\xcf\xd0\xd1\xd2\xd3\xd4\xd5\xd6\xd7\xd8\xd9\xda\xdb\xdc\xdd\xde\xdf\xe0\xe1\xe2\xe3\xe4\xe5\xe6\xe7\xe8\xe9\xea\xeb\xec\xed\xee\xef\xf0\xf1\xf2\xf3\xf4\xf5\xf6\xf7\xf8\xf9\xfa\xfb\xfc\xfd\xfe\xff"
	toUpperTable = "\x00\x01\x02\x03\x04\x05\x06\a\b\t\n\v\f\r\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`ABCDEFGHIJKLMNOPQRSTUVWXYZ{|}~\u007f\x80\x81\x82\x83\x84\x85\x86\x87\x88\x89\x8a\x8b\x8c\x8d\x8e\x8f\x90\x91\x92\x93\x94\x95\x96\x97\x98\x99\x9a\x9b\x9c\x9d\x9e\x9f\xa0\xa1\xa2\xa3\xa4\xa5\xa6\xa7\xa8\xa9\xaa\xab\xac\xad\xae\xaf\xb0\xb1\xb2\xb3\xb4\xb5\xb6\xb7\xb8\xb9\xba\xbb\xbc\xbd\xbe\xbf\xc0\xc1\xc2\xc3\xc4\xc5\xc6\xc7\xc8\xc9\xca\xcb\xcc\xcd\xce\xcf\xd0\xd1\xd2\xd3\xd4\xd5\xd6\xd7\xd8\xd9\xda\xdb\xdc\xdd\xde\xdf\xe0\xe1\xe2\xe3\xe4\xe5\xe6\xe7\xe8\xe9\xea\xeb\xec\xed\xee\xef\xf0\xf1\xf2\xf3\xf4\xf5\xf6\xf7\xf8\xf9\xfa\xfb\xfc\xfd\xfe\xff"
)

// ToLower converts ascii string to lower-case
func ToLower(b string) string {
	res := make([]byte, len(b))
	copy(res, b)
	for i := 0; i < len(res); i++ {
		res[i] = toLowerTable[res[i]]
	}

	return UnsafeString(res)
}

// ToUpper converts ascii string to upper-case
func ToUpper(b string) string {
	res := make([]byte, len(b))
	copy(res, b)
	for i := 0; i < len(res); i++ {
		res[i] = toUpperTable[res[i]]
	}

	return UnsafeString(res)
}

// EqualFold tests ascii strings for equality case-insensitively
func EqualFold(b, s string) bool {
	if len(b) != len(s) {
		return false
	}
	for i := len(b) - 1; i >= 0; i-- {
		if toUpperTable[b[i]] != toUpperTable[s[i]] {
			return false
		}
	}
	return true
}
