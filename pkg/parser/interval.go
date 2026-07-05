package parser

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// IntervalString converts a sign and string into a number of seconds
func IntervalString(s string, defaultSign int) (int32, error) {
	if len(s) == 0 {
		return 0, ErrUnknownTimeUnits
	}

	if s == "-" || s == "+" {
		return 0, ErrUnknownTimeUnits
	}

	original := s
	sign := defaultSign

	switch s[0] {
	case '-':
		sign = -1
		s = s[1:]
	case '+':
		sign = 1
		s = s[1:]
	}

	var totalInterval int64
	for len(s) > 0 {
		var j int
		for j < len(s) && '0' <= s[j] && s[j] <= '9' {
			j++
		}
		var offsetStr string
		offsetStr, s = s[:j], s[j:]

		j = 0
		for j < len(s) && (s[j] < '0' || '9' < s[j]) {
			j++
		}
		var unitStr string
		unitStr, s = s[:j], s[j:]

		var units int64
		switch strings.ToLower(unitStr) {
		case "s", "sec", "secs", "second", "seconds":
			units = 1
		case "m", "min", "mins", "minute", "minutes":
			units = 60
		case "h", "hr", "hrs", "hour", "hours":
			units = 60 * 60
		case "d", "day", "days":
			units = 24 * 60 * 60
		case "w", "wk", "wks", "week", "weeks":
			units = 7 * 24 * 60 * 60
		case "mon", "month", "months":
			units = 30 * 24 * 60 * 60
		case "y", "yr", "yrs", "year", "years":
			units = 365 * 24 * 60 * 60
		default:
			return 0, ErrUnknownTimeUnits
		}

		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			return 0, err
		}
		totalInterval += int64(sign) * int64(offset) * units
	}

	if totalInterval > math.MaxInt32 || totalInterval < math.MinInt32 {
		return 0, fmt.Errorf("interval %q out of range", original)
	}
	return int32(totalInterval), nil
}

func TruthyBool(s string) bool {
	switch s {
	case "", "0", "false", "False", "no", "No":
		return false
	case "1", "true", "True", "yes", "Yes":
		return true
	}
	return false
}
