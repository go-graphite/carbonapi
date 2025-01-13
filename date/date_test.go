package date

import (
	"fmt"
	"testing"
	"time"
)

func TestDateParamToEpoch(t *testing.T) {

	defaultTimeZone := time.UTC
	timeNow = func() time.Time {
		//16 Aug 1994 15:30
		return time.Date(1994, time.August, 16, 15, 30, 0, 100, defaultTimeZone)
	}

	const shortForm = "15:04 2006-Jan-02"

	var tests = []struct {
		input  string
		output string
	}{
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
	}

	for _, tt := range tests {
		got := DateParamToEpoch(tt.input, "America/Los_Angeles", 0, defaultTimeZone)
		ts, err := time.ParseInLocation(shortForm, tt.output, defaultTimeZone)
		if err != nil {
			panic(fmt.Sprintf("error parsing time: %q: %v", tt.output, err))
		}

		want := int64(ts.Unix())
		if got != want {
			t.Errorf("dateParamToEpoch(%q, 0)=%v, want %v", tt.input, got, want)
		}
	}
}
