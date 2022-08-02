package types

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

	var (
		start, braces, i, w int
	)

FOR:
	for braces, i, w = 0, 0, 0; i < len(s); i += w {

		w = 1

		switch s[i] {
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
			start = i + 1
		case ')':
			break FOR
		}
	}

	return start, i
}

// ExtractName extracts name out of function list. Only for use in MetrciData name parse, it's has more allowed chard in names, like =
func ExtractName(s string) string {
	start, end := ExtractNameLoc(s)
	return s[start:end]
}
