// Span bookability checks and the constraints they apply.
package engine

import (
	"time"

	"github.com/oh-tarnished/freebusy/protobuf/generated/go/availability/v1/availabilitypbv1"
	"github.com/oh-tarnished/freebusy/shared/rrule"
)

// CheckSpan tests whether one exact [start,end) span is bookable, returning the
// minimum free count over the span and the reasons it is not bookable. nights is
// the stay length for NIGHTLY units (0 for TIME_SLOT).
func CheckSpan(u *UnitInfo, start, end time.Time, unitsReq int32, nights int32, res []Reservation, closures []Closure, now time.Time) (bool, int32, []*availabilitypbv1.UnbookableReason) {
	if unitsReq < 1 {
		unitsReq = 1
	}
	var reasons []*availabilitypbv1.UnbookableReason
	add := func(c availabilitypbv1.Code, msg string) {
		reasons = append(reasons, &availabilitypbv1.UnbookableReason{Code: c, Message: msg})
	}

	if u.Archived {
		add(availabilitypbv1.Code_CODE_RESOURCE_ARCHIVED, "the unit is archived")
	}
	if closedOver(closures, start.UTC(), end.UTC()) {
		add(availabilitypbv1.Code_CODE_CLOSED, "a closure covers part of the span")
	}
	if !passesNotice(u, start.UTC(), now) {
		if u.MinNotice > 0 && start.UTC().Sub(now) < u.MinNotice {
			add(availabilitypbv1.Code_CODE_MIN_NOTICE, "the span starts sooner than the minimum notice allows")
		} else {
			add(availabilitypbv1.Code_CODE_MAX_ADVANCE, "the span starts further out than the advance window allows")
		}
	}
	loc := u.loc()
	if u.Mode == ModeTimeSlot && !rrule.Covers(u.Recurring, start.In(loc), end.In(loc)) {
		add(availabilitypbv1.Code_CODE_OUTSIDE_HOURS, "the span falls outside the unit's open hours")
	}
	if u.Mode == ModeNightly {
		if nights > 0 && u.MinNights > 0 && nights < u.MinNights {
			add(availabilitypbv1.Code_CODE_MIN_NIGHTS, "shorter than the minimum stay")
		}
		if nights > 0 && u.MaxNights > 0 && nights > u.MaxNights {
			add(availabilitypbv1.Code_CODE_MAX_NIGHTS, "longer than the maximum stay")
		}
		if len(u.CheckinWeekdays) > 0 && !weekdayAllowed(u.CheckinWeekdays, start.In(loc).Weekday()) {
			add(availabilitypbv1.Code_CODE_CHECKIN_DAY, "check-in falls on a disallowed weekday")
		}
		if len(u.CheckoutWeekdays) > 0 && !weekdayAllowed(u.CheckoutWeekdays, end.In(loc).Weekday()) {
			add(availabilitypbv1.Code_CODE_CHECKOUT_DAY, "check-out falls on a disallowed weekday")
		}
	}

	// Distinguish a true capacity shortfall from a buffer/gap conflict: if the raw
	// free count is short it is NO_CAPACITY; if only the buffered count is short,
	// an adjacent booking's buffer is the blocker.
	free := minFreeOverSpan(u, start.UTC(), end.UTC(), res)
	switch {
	case free < unitsReq:
		add(availabilitypbv1.Code_CODE_NO_CAPACITY, "not enough free units for the requested count")
	case minBufferedFreeOverSpan(u, start.UTC(), end.UTC(), res) < unitsReq:
		add(availabilitypbv1.Code_CODE_BUFFER_CONFLICT, "a buffer or gap around an adjacent booking conflicts")
	}
	return len(reasons) == 0, free, reasons
}

// weekdayAllowed reports whether wd is in the allowed set.
func weekdayAllowed(allowed []time.Weekday, wd time.Weekday) bool {
	for _, a := range allowed {
		if a == wd {
			return true
		}
	}
	return false
}

// minBufferedFreeOverSpan is minFreeOverSpan using the buffered free count (which
// accounts for gap and setup/turnover deltas).
func minBufferedFreeOverSpan(u *UnitInfo, start, end time.Time, res []Reservation) int32 {
	if u.Mode == ModeNightly {
		loc := u.loc()
		min := u.Capacity
		for cur := start.In(loc); cur.Before(end); cur = cur.AddDate(0, 0, 1) {
			if f := bufferedFree(u, cur.UTC(), cur.AddDate(0, 0, 1).UTC(), res); f < min {
				min = f
			}
		}
		return min
	}
	return bufferedFree(u, start, end, res)
}

// minFreeOverSpan returns the minimum free count across the span. For nightly
// spans it evaluates each night; otherwise it treats the span as one interval.
func minFreeOverSpan(u *UnitInfo, start, end time.Time, res []Reservation) int32 {
	free := u.Capacity - overlapUnits(res, start, end)
	if u.Mode == ModeNightly {
		loc := u.loc()
		free = u.Capacity
		for cur := start.In(loc); cur.Before(end); cur = cur.AddDate(0, 0, 1) {
			night := u.Capacity - overlapUnits(res, cur.UTC(), cur.AddDate(0, 0, 1).UTC())
			if night < free {
				free = night
			}
		}
	}
	if free < 0 {
		free = 0
	}
	return free
}
