package stringutils

import "strings"

// Reverse return reversed string (rune-wise left to right).
func Reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; {
		r[i], r[j] = r[j], r[i]
		i++
		j--
	}
	return string(r)
}

// ReverseSegments return reversed string by segments around delimiter.
func ReverseSegments(target, delim string) string {
	if len(delim) == 0 || len(target) == 0 {
		return target
	}
	a := strings.Split(target, delim)
	l := len(a)
	for i := 0; i < l/2; i++ {
		a[i], a[l-i-1] = a[l-i-1], a[i]
	}

	return strings.Join(a, delim)
}
