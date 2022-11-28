package parser

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseTime(t *testing.T) {
	now := time.Date(2015, time.Month(3), 8, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		s              string
		expectedResult int64
	}{
		{
			s:              "12:0020150308",
			expectedResult: 1425816000,
		},
		{
			s:              "9:0020150308",
			expectedResult: 1425805200,
		},
		{
			s:              "20150110",
			expectedResult: 1420848000,
		},
		{
			s:              "midnight",
			expectedResult: 1425772800,
		},
		{
			s:              "midnight+1h",
			expectedResult: 1425776400,
		},
		{
			s:              "midnight_tomorrow",
			expectedResult: 1425168000,
		},
		{
			s:              "midnight_tomorrow+3h",
			expectedResult: 1425178800,
		},
		{
			s:              "now",
			expectedResult: 1425816000,
		},
		{
			s:              "-1h",
			expectedResult: 1425812400,
		},
		{
			s:              "8:50",
			expectedResult: 1425804600,
		},
		{
			s:              "8:50am",
			expectedResult: 1425804600,
		},
		{
			s:              "8:50pm",
			expectedResult: 1425847800,
		},
		{
			s:              "8am",
			expectedResult: 1425801600,
		},
		{
			s:              "10pm",
			expectedResult: 1425852000,
		},
		{
			s:              "noon",
			expectedResult: 1425816000,
		},
		{
			s:              "midnight",
			expectedResult: 1425772800,
		},
		{
			s:              "teatime",
			expectedResult: 1425831360,
		},
		{
			s:              "yesterday",
			expectedResult: 1424995200,
		},
		{
			s:              "today",
			expectedResult: 1425816000,
		},
		{
			s:              "tomorrow",
			expectedResult: 1425168000,
		},
		{
			s:              "02/25/15",
			expectedResult: 1424822400,
		},
		{
			s:              "20140606",
			expectedResult: 1402012800,
		},
		{
			s:              "january8",
			expectedResult: 1426507200,
		},
		{
			s:              "january10",
			expectedResult: 1426680000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			assert := assert.New(t)

			time, err := ParseDateTime(tt.s, -1, now)
			assert.NoError(err)
			fmt.Println("time: ", time)
			assert.Equal(tt.expectedResult, time, tt.s)
		})
	}
}

func testInvalidTimes(t *testing.T) {
	now := time.Date(2015, time.Month(3), 8, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		s              string
		expectedResult int64
	}{
		{
			s:              "12:0020150308",
			expectedResult: 1664985600,
		},
	}

	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			assert := assert.New(t)

			time, err := ParseDateTime(tt.s, -1, now)
			assert.NoError(err)
			fmt.Println("time: ", time)
			assert.Equal(tt.expectedResult, time, tt.s)
		})
	}
}
