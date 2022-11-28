package parser

import (
	"fmt"
	"github.com/leekchan/timeutil"
	"strconv"
	"strings"
	"time"
)

var months = []string{"jan", "feb", "mar", "apr", "may", "jun", "jul", "aug", "sep", "oct", "nov", "dec"}
var weekdays = []string{"sun", "mon", "tue", "wed", "thu", "fri", "sat"}

type dateTime struct {
	year        int
	month       int
	day         int
	seconds     int
	nanoseconds int
	location    time.Location
}

func ParseDateTime(dateTime string, defaultSign int, now time.Time) (int64, error) {
	var parsed []string
	var offset string
	var ref string
	r := strings.NewReplacer(
		" ", "",
		",", "",
		"_", "",
	)
	parsedTime := strings.TrimSpace(dateTime)
	parsedTime = r.Replace(parsedTime)

	val, err := strconv.Atoi(parsedTime)
	if err == nil {
		year, _ := strconv.Atoi(parsedTime[:4])
		month, _ := strconv.Atoi(parsedTime[4:6])
		day, _ := strconv.Atoi(parsedTime[6:])
		if len(parsedTime) != 8 || year < 1900 || month > 13 || day > 32 {
			return int64(val), nil
		}
	}
	if strings.Contains(parsedTime, "-") {
		parsed = strings.SplitN(parsedTime, "-", 2)
		offset = "-" + parsed[1]
		ref = parsed[0]
	} else if strings.Contains(parsedTime, "+") {
		parsed = strings.SplitN(parsedTime, "+", 2)
		offset = "+" + parsed[1]
		ref = parsed[0]
	} else {
		offset = ""
		ref = parsedTime
	}

	refTime, _ := parseTimeReference(ref, now)
	interval, _ := parseInterval(offset, defaultSign)

	total := refTime + interval
	return total, nil

}

func parseTimeReference(ref string, now time.Time) (int64, error) {
	if ref == "" || ref == "now" {
		return now.Unix(), nil
	}

	var hour int
	var minute int
	var rawRef = ref
	//var err error
	var refDate = now

	hour, minute, ref = getReferenceDate(ref)
	if ref == "" {
		refDate = time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, time.UTC)
		return refDate.Unix(), nil
	}

	// Day reference
	if strings.Contains(ref, "today") || strings.Contains(ref, "yesterday") || strings.Contains(ref, "tomorrow") {
		if strings.Contains(ref, "yesterday") {
			refDate = time.Date(now.Year(), now.Month(), -1, hour, minute, 0, 0, time.UTC)
		} else if strings.Contains(ref, "tomorrow") {
			refDate = time.Date(now.Year(), now.Month(), 1, hour, minute, 0, 0, time.UTC)
		}
	} else if strings.Count(ref, "/") == 2 { // MM/DD/YY[YY] format
		parsed := strings.SplitN(ref, "/", 3)
		year, _ := strconv.Atoi(parsed[2])
		month, _ := strconv.Atoi(parsed[0])
		day, _ := strconv.Atoi(parsed[1])
		if year < 1900 {
			year += 1900
		}
		if year < 1970 {
			year += 100
		}
		refDate = time.Date(year, time.Month(month), day, hour, minute, 0, 0, time.UTC)
	} else if _, err := strconv.Atoi(ref); err == nil && len(ref) == 8 { // YYYYMMDD format
		year, _ := strconv.Atoi(ref[:4])
		month, _ := strconv.Atoi(ref[4:6])
		day, _ := strconv.Atoi(ref[6:])
		refDate = time.Date(year, time.Month(month), day, hour, minute, 0, 0, time.UTC)
	} else if len(ref) >= 3 && stringMatchesList(ref[:3], months) { // MonthName DayOfMonth
		var day int
		if val, err := strconv.Atoi(ref[(len(ref) - 2):]); err == nil {
			day = val
		} else if val, err := strconv.Atoi(ref[(len(ref) - 1):]); err == nil {
			day = val
		} else {
			return 0, fmt.Errorf("Day of month required after month name: %s", rawRef)
		}
		refDate = refDate.AddDate(0, 0, day)
	} else if len(ref) >= 3 && stringMatchesList(ref[:3], weekdays) { // DayOfWeek (Monday, etc)
		dayName := timeutil.Strftime(&refDate, "%a")
		dayName = strings.ToLower(dayName)
		today := stringMatchesListIndex(dayName, weekdays)
		twoWeeks := append(weekdays, weekdays...)
		dayOffset := today - stringMatchesListIndex(ref[:3], twoWeeks)
		if dayOffset < 0 {
			dayOffset += 7
		}
		refDate = refDate.AddDate(0, 0, -(dayOffset))
	} else {
		return 0, fmt.Errorf("Unknown day reference: %s", rawRef)
	}

	return refDate.Unix(), nil
}

// IntervalString converts a sign and string into a number of seconds
func parseInterval(s string, defaultSign int) (int64, error) {
	if len(s) == 0 {
		return 0, nil
	}
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

		var units int
		switch unitStr {
		case "s", "sec", "secs", "second", "seconds":
			units = 1
		case "m", "min", "mins", "minute", "minutes":
			units = 60
		case "h", "hour", "hours":
			units = 60 * 60
		case "d", "day", "days":
			units = 24 * 60 * 60
		case "w", "week", "weeks":
			units = 7 * 24 * 60 * 60
		case "mon", "month", "months":
			units = 30 * 24 * 60 * 60
		case "y", "year", "years":
			units = 365 * 24 * 60 * 60
		default:
			return 0, ErrUnknownTimeUnits
		}

		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			return 0, err
		}
		totalInterval += int64(sign * offset * units)
	}

	return totalInterval, nil
}

func stringMatchesList(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func getReferenceDate(ref string) (hr int, min int, remaining string) {
	// Time-of-day reference
	var hour = 0
	var minute = 0
	i := strings.Index(ref, ":")
	if i > 0 && i < 3 {
		hour, _ = strconv.Atoi(ref[:i])
		minute, _ = strconv.Atoi(ref[i+1 : i+3])
		ref = ref[i+3:]
		if ref == "" {
			return hour, minute, ref
		} else if ref[:2] == "am" {
			ref = ref[2:]
		} else if ref[:2] == "pm" {
			hour = (hour + 12) % 24
			ref = ref[2:]
		}
	}

	// X am or XXam
	i = strings.Index(ref, "am")
	if i > 0 && i < 3 {
		hour, _ = strconv.Atoi(ref[:i])
		ref = ref[i+2:]
	}

	// X pm or XX pm
	i = strings.Index(ref, "pm")
	if i > 0 && i < 3 {
		hr, _ := strconv.Atoi(ref[:i])
		hour = (hr + 12) % 24
		ref = ref[i+2:]
	}

	if strings.HasPrefix(ref, "noon") {
		hour = 12
		minute = 0
		ref = ref[4:]
	} else if strings.HasPrefix(ref, "midnight") {
		hour = 0
		minute = 0
		ref = ref[8:]
	} else if strings.HasPrefix(ref, "teatime") {
		hour = 16
		minute = 16
		ref = ref[7:]
	}

	//now := time.Now().UTC()
	//timeZone := time.UTC
	//refDate := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, timeZone)

	return hour, minute, ref
}

func setHourMinute(hour, minute int) time.Time {
	now := time.Now().UTC()
	timeZone := time.UTC
	return time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, timeZone)
}

func stringMatchesListIndex(a string, list []string) int {
	for i, b := range list {
		if b == a {
			return i
		}
	}
	return -1
}
