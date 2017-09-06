package strftime

import (
	"testing"
	"time"
)

var testTime = time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)

func TestBasic(t *testing.T) {
	s, err := Format("%a", testTime)
	if err != nil {
		t.Fatalf("go error - %s", err)
	}
	if s != "Tue" {
		t.Fatalf("Bad day for %s, got %s - expected Tue", testTime, s)
	}

}

func TestUnknown(t *testing.T) {
	_, err := Format("%g", testTime)
	if err == nil {
		t.Fatalf("managed to expand %g")
	}
}

func TestDayOfYear(t *testing.T) {
	s, err := Format("%j", testTime)
	if err != nil {
		t.Fatalf("error expanding %j", err)
	}

	if s != "314" {
		t.Fatalf("day of year != 314 (got %s)", s)
	}
}

func TestWeekday(t *testing.T) {
	s, err := Format("%w", testTime)
	if err != nil {
		t.Fatalf("error expanding %w", err)
	}

	if s != "2" {
		t.Fatalf("day of week != 2 (got %s)", s)
	}
}

func checkWeek(format string, t *testing.T) {
	s, err := Format(format, testTime)
	if err != nil {
		t.Fatalf("error expanding %s - %s", format, err)
	}

	if s != "45" {
		t.Fatalf("[%s] week num != 45 (got %s)", format, s)
	}
}

func TestWeekNum(t *testing.T) {
	checkWeek("%W", t)
	checkWeek("%U", t)
}
