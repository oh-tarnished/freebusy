// Package rrule is a small, reusable evaluator for recurring open-hours rules: an
// RFC 5545 RRULE that names the days a window recurs, paired with a local
// "HH:MM" open/close time-of-day. It answers one question the availability engine
// (and any other caller) needs — is a given [start,end) span fully inside the
// open hours? — without pulling in a full iCalendar dependency.
//
// Supported RRULE subset (what open-hours actually use): FREQ=DAILY, and
// FREQ=WEEKLY with BYDAY=MO,TU,… A bare BYDAY (no FREQ) is treated as weekly.
// INTERVAL/UNTIL/COUNT and other parts are ignored — an open-hours rule recurs
// indefinitely, so they do not affect "is this span open". Unrecognized or empty
// rules match every day, so a malformed rule never wrongly blocks a booking.
package rrule

import (
	"strings"
	"time"
)

// Rule is one open-hours window: the recurrence (RRule) and the local time-of-day
// span it is open, as 24-hour "HH:MM" strings. Empty Opens and Closes mean the
// whole day is open on matching days; Closes <= Opens means the window crosses
// midnight (e.g. Opens "22:00", Closes "02:00").
type Rule struct {
	RRule  string
	Opens  string
	Closes string
}

// Covers reports whether the span [start,end) falls entirely within the open
// hours defined by rules. An empty rule set means "always open" (returns true),
// preserving the no-schedule default. start/end should be in the location the
// rule's local times are expressed in (the unit's timezone).
func Covers(rules []Rule, start, end time.Time) bool {
	if len(rules) == 0 {
		return true
	}
	for i := range rules {
		if rules[i].covers(start, end) {
			return true
		}
	}
	return false
}

// covers reports whether [start,end) fits within this single rule's window.
func (r Rule) covers(start, end time.Time) bool {
	days, everyDay := parseDays(r.RRule)
	open, close := parseHM(r.Opens), parseHM(r.Closes)
	if open < 0 && close < 0 {
		open, close = 0, 24*60 // whole day
	} else {
		if open < 0 {
			open = 0
		}
		if close < 0 {
			close = 24 * 60
		}
	}
	if close <= open {
		close += 24 * 60 // crosses midnight
	}

	startMin := start.Hour()*60 + start.Minute()
	endMin := startMin + int(end.Sub(start)/time.Minute)

	// Window anchored to start's day.
	if (everyDay || days[start.Weekday()]) && startMin >= open && endMin <= close {
		return true
	}
	// A window that opened the previous day (midnight-crossing) may still cover an
	// early-morning span; shift the window back a day into start's minute frame.
	if close > 24*60 {
		prev := start.AddDate(0, 0, -1).Weekday()
		if (everyDay || days[prev]) && startMin >= open-24*60 && endMin <= close-24*60 {
			return true
		}
	}
	return false
}

// weekdayByday maps RFC 5545 BYDAY codes to Go weekdays.
var weekdayByday = map[string]time.Weekday{
	"SU": time.Sunday,
	"MO": time.Monday,
	"TU": time.Tuesday,
	"WE": time.Wednesday,
	"TH": time.Thursday,
	"FR": time.Friday,
	"SA": time.Saturday,
}

// parseDays extracts the matching weekdays from an RRULE. It returns (days, true)
// when the rule matches every day (FREQ=DAILY, or an empty/unrecognized rule),
// otherwise (days, false) with the BYDAY set.
func parseDays(rrule string) (map[time.Weekday]bool, bool) {
	if strings.TrimSpace(rrule) == "" {
		return nil, true
	}
	var freq, byday string
	for _, part := range strings.Split(rrule, ";") {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch strings.ToUpper(strings.TrimSpace(kv[0])) {
		case "FREQ":
			freq = strings.ToUpper(strings.TrimSpace(kv[1]))
		case "BYDAY":
			byday = strings.ToUpper(strings.TrimSpace(kv[1]))
		}
	}
	if byday != "" {
		days := map[time.Weekday]bool{}
		for _, code := range strings.Split(byday, ",") {
			// Tolerate an ordinal prefix like "2MO" by taking the trailing 2 letters.
			code = strings.TrimSpace(code)
			if len(code) >= 2 {
				code = code[len(code)-2:]
			}
			if wd, ok := weekdayByday[code]; ok {
				days[wd] = true
			}
		}
		if len(days) > 0 {
			return days, false
		}
	}
	if freq == "DAILY" {
		return nil, true
	}
	// WEEKLY without BYDAY, or anything we do not model: match every day rather
	// than wrongly block.
	return nil, true
}

// parseHM parses a 24-hour "HH:MM" time into minutes since midnight, or -1 when
// empty or malformed.
func parseHM(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return -1
	}
	t, err := time.Parse("15:04", s)
	if err != nil {
		return -1
	}
	return t.Hour()*60 + t.Minute()
}
