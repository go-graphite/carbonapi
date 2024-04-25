package stringutils

// TrimLeft is the equivalent of strings.TrimLeft
func TrimLeft(s string, cutset byte) string {
	lenStr, start := len(s), 0
	for start < lenStr && s[start] == cutset {
		start++
	}
	return s[start:]
}

// Trim is the equivalent of strings.Trim
func Trim(s string, cutset byte) string {
	i, j := 0, len(s)-1
	for ; i <= j; i++ {
		if s[i] != cutset {
			break
		}
	}
	for ; i < j; j-- {
		if s[j] != cutset {
			break
		}
	}

	return s[i : j+1]
}

// TrimRight is the equivalent of strings.TrimRight
func TrimRight(s string, cutset byte) string {
	lenStr := len(s)
	for lenStr > 0 && s[lenStr-1] == cutset {
		lenStr--
	}
	return s[:lenStr]
}
