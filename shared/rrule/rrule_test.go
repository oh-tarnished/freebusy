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

// Compile is the hot-path form: parse once, evaluate many spans.
func TestCompiledCovers(t *testing.T) {
	c := Compile([]Rule{{RRule: "FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR", Opens: "09:00", Closes: "17:00"}})
	// 2026-12-24 is a Thursday.
	if !c.Covers(at(2026, 12, 24, 10, 0), at(2026, 12, 24, 11, 0)) {
		t.Fatal("Thu 10:00-11:00 should be open")
	}
	if c.Covers(at(2026, 12, 26, 10, 0), at(2026, 12, 26, 11, 0)) {
		t.Fatal("Sat should be closed (not in BYDAY)")
	}
}

// The zero value and an empty set both mean "always open".
func TestCompiledZeroValueAlwaysOpen(t *testing.T) {
	var zero Compiled
	if !zero.Covers(at(2026, 1, 1, 3, 0), at(2026, 1, 1, 4, 0)) {
		t.Fatal("zero-value Compiled should mean always open")
	}
	if !Compile(nil).Covers(at(2026, 1, 1, 3, 0), at(2026, 1, 1, 4, 0)) {
		t.Fatal("Compile(nil) should mean always open")
	}
}

// A window crossing midnight covers late-evening and early-morning spans on
// the anchoring day, and rejects a span outside it.
func TestCompiledMidnightCrossing(t *testing.T) {
	c := Compile([]Rule{{RRule: "FREQ=WEEKLY;BYDAY=TH", Opens: "22:00", Closes: "02:00"}})
	// 2026-12-24 is a Thursday: open 22:00 Thu through 02:00 Fri.
	if !c.Covers(at(2026, 12, 24, 22, 30), at(2026, 12, 24, 23, 30)) {
		t.Fatal("Thu 22:30-23:30 should be open")
	}
	if !c.Covers(at(2026, 12, 25, 0, 30), at(2026, 12, 25, 1, 30)) {
		t.Fatal("Fri 00:30-01:30 should be open (window opened Thu)")
	}
	if c.Covers(at(2026, 12, 25, 22, 30), at(2026, 12, 25, 23, 30)) {
		t.Fatal("Fri 22:30-23:30 should be closed (window only anchors on Thu)")
	}
}
