package main

import (
	"fmt"
	"testing"
	"time"
)

func TestDateParamToEpoch(t *testing.T) {

	timeNow = func() time.Time {
		//16 Aug 1994 15:30
		return time.Date(1994, time.August, 16, 15, 30, 0, 100, defaultTimeZone)
	}

	const shortForm = "15:04 2006-Jan-02"

	var tests = []struct {
		input  string
		output string
	}{
		{"midnight", "00:00 1994-Aug-16"},
		{"noon", "12:00 1994-Aug-16"},
		{"teatime", "16:00 1994-Aug-16"},

		{"noon 08/16/94", "12:00 1994-Aug-16"},
		{"midnight 20060816", "00:00 2006-Aug-16"},

		{"15:04 19940816", "15:04 1994-Aug-16"},
		{"-1day", "15:30 1994-Aug-15"},
	}

	for _, tt := range tests {
		got := dateParamToEpoch(tt.input, 0)
		ts, err := time.ParseInLocation(shortForm, tt.output, defaultTimeZone)
		if err != nil {
			panic(fmt.Sprintf("error parsing time: %q: %v", tt.output, err))
		}

		want := int32(ts.Unix())
		if got != want {
			t.Errorf("dateParamToEpoch(%q, 0)=%v, want %v", tt.input, got, want)
		}
	}
}
