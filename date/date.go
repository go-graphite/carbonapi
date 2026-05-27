package date

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/go-graphite/carbonapi/pkg/parser"
)

var errBadTime = errors.New("bad time")
var timeNow = time.Now

// MockTimeNow replaces the package clock and returns the previous one so tests can restore it.
func MockTimeNow(f func() time.Time) func() time.Time {
	prev := timeNow
	timeNow = f
	return prev
}

// parseTime parses a time and returns hours and minutes
func parseTime(s string) (hour, minute int, err error) {

	switch s {
	case "midnight":
		return 0, 0, nil
	case "noon":
		return 12, 0, nil
	case "teatime":
		return 16, 0, nil
	}

	parts := strings.Split(s, ":")

	if len(parts) != 2 {
		return 0, 0, errBadTime
	}

	hour, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, errBadTime
	}

	minute, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, errBadTime
	}

	return hour, minute, nil
}

var TimeFormats = []string{"20060102", "01/02/06"}

// ParseAtTimeOr is ParseAtTime with a fallback: on parse error it returns d
// instead of an error.
func ParseAtTimeOr(s, qtz string, defaultTimeZone *time.Location, d int64) int64 {
	epoch, err := ParseAtTime(s, qtz, defaultTimeZone)
	if err != nil {
		return d
	}
	return epoch
}

// ParseAtTime parses graphite-web's at-time grammar into a unix epoch.
// E.g. "today-2d", "noon+3h", "-1d", "now", "00:00 20140101".
// See render/attime.py upstream.
func ParseAtTime(s, qtz string, defaultTimeZone *time.Location) (int64, error) {
	if s == "" {
		return 0, errBadTime
	}

	var tz = defaultTimeZone
	if qtz != "" {
		if z, err := time.LoadLocation(qtz); err == nil {
			tz = z
		}
	}

	if s[0] == '-' || s[0] == '+' {
		offset, err := parser.IntervalString(s, -1)
		if err != nil {
			return 0, err
		}
		return timeNow().In(tz).Add(time.Duration(offset) * time.Second).Unix(), nil
	}

	// handle <ref>±<offset> form (e.g. "today-2d", "noon+3h").
	for i := 1; i < len(s); i++ {
		if s[i] == '+' || s[i] == '-' {
			refEpoch, refErr := parseTimeReference(s[:i], tz)
			if refErr != nil {
				break
			}
			offset, err := parser.IntervalString(s[i:], 1)
			if err != nil {
				return 0, err
			}
			return refEpoch + int64(offset), nil
		}
	}

	return parseTimeReference(s, tz)
}

// parseTimeReference parses a time reference (no offset) into an epoch.
func parseTimeReference(s string, tz *time.Location) (int64, error) {
	switch s {
	case "now":
		return timeNow().In(tz).Unix(), nil
	case "midnight", "noon", "teatime":
		yy, mm, dd := timeNow().In(tz).Date()
		hh, min, _ := parseTime(s) // error ignored, we know it's valid
		dt := time.Date(yy, mm, dd, hh, min, 0, 0, tz)
		return dt.Unix(), nil
	}

	sint, err := strconv.Atoi(s)
	// need to check that len(s) != 8 to avoid turning 20060102 into seconds
	if err == nil && len(s) != 8 {
		return int64(sint), nil // We got a timestamp so returning it
	}

	s = strings.Replace(s, "_", " ", 1) // Go can't parse _ in date strings

	var ts, ds string
	split := strings.Fields(s)

	switch {
	case len(split) == 1:
		ds = s
	case len(split) == 2:
		ts, ds = split[0], split[1]
	case len(split) > 2:
		return 0, errBadTime
	}

	var t time.Time
dateStringSwitch:
	switch ds {
	case "today":
		t = timeNow().In(tz)
		// nothing
	case "yesterday":
		t = timeNow().In(tz).AddDate(0, 0, -1)
	case "tomorrow":
		t = timeNow().In(tz).AddDate(0, 0, 1)
	default:
		for _, format := range TimeFormats {
			t, err = time.ParseInLocation(format, ds, tz)
			if err == nil {
				break dateStringSwitch
			}
		}

		return 0, errBadTime
	}

	var hour, minute int
	if ts != "" {
		hour, minute, _ = parseTime(ts)
		// defaults to hour=0, minute=0 on error, which is midnight, which is fine for now
	}

	yy, mm, dd := t.Date()
	t = time.Date(yy, mm, dd, hour, minute, 0, 0, tz)

	return t.Unix(), nil
}
