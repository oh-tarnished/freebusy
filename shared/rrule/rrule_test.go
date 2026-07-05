package rrule

import (
	"testing"
	"time"
)

func at(y, m, d, hh, mm int) time.Time {
	return time.Date(y, time.Month(m), d, hh, mm, 0, 0, time.UTC)
}

func TestCoversWeeklyBusinessHours(t *testing.T) {
	rules := []Rule{{RRule: "FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR", Opens: "09:00", Closes: "17:00"}}

	// 2026-12-24 is a Thursday.
	if !Covers(rules, at(2026, 12, 24, 10, 0), at(2026, 12, 24, 11, 0)) {
		t.Fatal("Thu 10:00–11:00 should be open")
	}
	if Covers(rules, at(2026, 12, 24, 16, 30), at(2026, 12, 24, 17, 30)) {
		t.Fatal("Thu 16:30–17:30 spills past close, should be closed")
	}
	// 2026-12-26 is a Saturday — not in BYDAY.
	if Covers(rules, at(2026, 12, 26, 10, 0), at(2026, 12, 26, 11, 0)) {
		t.Fatal("Sat should be closed (not in BYDAY)")
	}
}

func TestCoversEmptyAlwaysOpen(t *testing.T) {
	if !Covers(nil, at(2026, 1, 1, 3, 0), at(2026, 1, 1, 4, 0)) {
		t.Fatal("no rules should mean always open")
	}
}

func TestCoversDailyWholeDay(t *testing.T) {
	rules := []Rule{{RRule: "FREQ=DAILY"}}
	if !Covers(rules, at(2026, 12, 26, 2, 0), at(2026, 12, 26, 23, 0)) {
		t.Fatal("FREQ=DAILY with no hours should be open all day, every day")
	}
}
