package date

import (
	"fmt"
	"testing"
	"time"
)

func TestDateParamToEpoch(t *testing.T) {

	defaultTimeZone := time.UTC
	// 16 Aug 1994 15:30 UTC
	defer MockTimeNow(func() time.Time {
		return time.Date(1994, time.August, 16, 15, 30, 0, 100, defaultTimeZone)
	})()

	const shortForm = "15:04 2006-Jan-02"

	var tests = []struct {
		input  string
		output string
	}{
		// qtz="America/Los_Angeles": output expressed in UTC, so midnight LA = 07:00 UTC.
		{"midnight", "07:00 1994-Aug-16"},
		{"noon", "19:00 1994-Aug-16"},
		{"teatime", "23:00 1994-Aug-16"},
		{"tomorrow", "07:00 1994-Aug-17"},

		{"noon 08/12/94", "19:00 1994-Aug-12"},
		{"midnight 20060812", "07:00 2006-Aug-12"},
		{"noon tomorrow", "19:00 1994-Aug-17"},

		{"17:04 19940812", "00:04 1994-Aug-13"},
		{"-1day", "15:30 1994-Aug-15"},
		{"19940812", "07:00 1994-Aug-12"},

		// <ref>±<offset> form
		{"today-2d", "07:00 1994-Aug-14"},
		{"today-1h", "06:00 1994-Aug-16"},
		{"yesterday+12h", "19:00 1994-Aug-15"},
		{"now-1h", "14:30 1994-Aug-16"},
		{"now+30min", "16:00 1994-Aug-16"},
		{"noon+3h", "22:00 1994-Aug-16"},
		{"midnight-30min", "06:30 1994-Aug-16"},

		// case-insensitive
		{"NOW", "15:30 1994-Aug-16"},
		{"Today-1h", "06:00 1994-Aug-16"},
		{"MIDNIGHT", "07:00 1994-Aug-16"},

		// 4-digit year in MM/DD/YYYY
		{"01/02/2014", "08:00 2014-Jan-02"},
		{"noon 08/12/2006", "19:00 2006-Aug-12"},
	}

	for _, tt := range tests {
		got := DateParamToEpoch(tt.input, "America/Los_Angeles", 0, defaultTimeZone)
		ts, err := time.ParseInLocation(shortForm, tt.output, defaultTimeZone)
		if err != nil {
			panic(fmt.Sprintf("error parsing time: %q: %v", tt.output, err))
		}

		want := int64(ts.Unix())
		if got != want {
			t.Errorf("DateParamToEpoch(%q)=%v, want %v", tt.input, got, want)
		}
	}
}
