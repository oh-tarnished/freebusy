package engine

import (
	"time"

	"github.com/oh-tarnished/freebusy/protobuf/generated/go/availability/v1/availabilitypbv1"
	sharedpbv1 "github.com/oh-tarnished/freebusy/protobuf/generated/go/shared/v1/sharedpbv1"
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/genproto/googleapis/type/money"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// loc returns the unit's timezone location, falling back to UTC.
func (u *UnitInfo) loc() *time.Location {
	l, err := time.LoadLocation(u.TimeZone)
	if err != nil {
		return time.UTC
	}
	return l
}

// overlapUnits sums the units of reservations overlapping [s,e) (half-open).
func overlapUnits(res []Reservation, s, e time.Time) int32 {
	var sum int32
	for i := range res {
		if res[i].Start.Before(e) && res[i].End.After(s) {
			u := res[i].Units
			if u < 1 {
				u = 1
			}
			sum += u
		}
	}
	return sum
}

func closedOver(closures []Closure, s, e time.Time) bool {
	for i := range closures {
		if closures[i].Start.Before(e) && closures[i].End.After(s) {
			return true
		}
	}
	return false
}

// ComputeNights returns per-night availability for a NIGHTLY unit over the date
// range [startDate, endDate) evaluated in the unit's timezone (endDate is the
// checkout date, exclusive). unitsReq is the party/pool count required free.
func ComputeNights(u *UnitInfo, startDate, endDate *date.Date, unitsReq int32, res []Reservation, closures []Closure) []*availabilitypbv1.NightAvailability {
	if unitsReq < 1 {
		unitsReq = 1
	}
	loc := u.loc()
	cur := dateToTime(startDate, loc)
	end := dateToTime(endDate, loc)
	var out []*availabilitypbv1.NightAvailability
	for cur.Before(end) {
		next := cur.AddDate(0, 0, 1)
		s, e := cur.UTC(), next.UTC()
		free := u.Capacity - overlapUnits(res, s, e)
		if free < 0 {
			free = 0
		}
		out = append(out, &availabilitypbv1.NightAvailability{
			Night:     timeToDate(cur),
			FreeUnits: free,
			Closed:    closedOver(closures, s, e),
			Price:     cloneMoney(u.Price),
		})
		cur = next
	}
	return out
}

// ComputeSlots returns discrete bookable slots for a TIME_SLOT unit across
// [start,end). slotDur overrides the unit duration when > 0. now is used for the
// minimum-notice check.
func ComputeSlots(u *UnitInfo, start, end time.Time, slotDur time.Duration, unitsReq int32, res []Reservation, closures []Closure, now time.Time) []*availabilitypbv1.Slot {
	if unitsReq < 1 {
		unitsReq = 1
	}
	dur := slotDur
	if dur <= 0 {
		dur = u.Duration
	}
	if dur <= 0 {
		return nil
	}
	var out []*availabilitypbv1.Slot
	for s := start.UTC(); !s.Add(dur).After(end.UTC()); s = s.Add(dur) {
		e := s.Add(dur)
		free := u.Capacity - overlapUnits(res, s, e)
		if free < 0 {
			free = 0
		}
		bookable := !u.Archived &&
			free >= unitsReq &&
			!closedOver(closures, s, e) &&
			passesNotice(u, s, now)
		out = append(out, &availabilitypbv1.Slot{
			StartTime: timestamppb.New(s),
			EndTime:   timestamppb.New(e),
			FreeCount: free,
			Bookable:  bookable,
			Price:     cloneMoney(u.Price),
		})
	}
	return out
}

func passesNotice(u *UnitInfo, start, now time.Time) bool {
	if u.MinNotice > 0 && start.Sub(now) < u.MinNotice {
		return false
	}
	if u.MaxAdvance > 0 && start.Sub(now) > u.MaxAdvance {
		return false
	}
	return true
}

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
	if u.Mode == ModeNightly && nights > 0 {
		if u.MinNights > 0 && nights < u.MinNights {
			add(availabilitypbv1.Code_CODE_MIN_NIGHTS, "shorter than the minimum stay")
		}
		if u.MaxNights > 0 && nights > u.MaxNights {
			add(availabilitypbv1.Code_CODE_MAX_NIGHTS, "longer than the maximum stay")
		}
	}

	free := minFreeOverSpan(u, start.UTC(), end.UTC(), res)
	if free < unitsReq {
		add(availabilitypbv1.Code_CODE_NO_CAPACITY, "not enough free units for the requested count")
	}
	return len(reasons) == 0, free, reasons
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

// NightRanges coalesces consecutive open, sufficiently-free nights into bookable
// ranges (for NIGHTLY ComputeBookableRanges).
func NightRanges(nights []*availabilitypbv1.NightAvailability, unitsReq int32, loc *time.Location) []*availabilitypbv1.BookableRange {
	if unitsReq < 1 {
		unitsReq = 1
	}
	var out []*availabilitypbv1.BookableRange
	var runStart *date.Date
	var runEnd time.Time
	flush := func(end time.Time) {
		if runStart != nil {
			out = append(out, &availabilitypbv1.BookableRange{
				Window: &sharedpbv1.TimeWindow{
					StartTime: timestamppb.New(dateToTime(runStart, loc).UTC()),
					EndTime:   timestamppb.New(end.UTC()),
				},
				Bookable: true,
			})
			runStart = nil
		}
	}
	for _, n := range nights {
		open := !n.GetClosed() && n.GetFreeUnits() >= unitsReq
		nightStart := dateToTime(n.GetNight(), loc)
		if open {
			if runStart == nil {
				runStart = n.GetNight()
			}
			runEnd = nightStart.AddDate(0, 0, 1)
		} else {
			flush(runEnd)
		}
	}
	flush(runEnd)
	return out
}

// SlotRanges coalesces consecutive bookable slots into contiguous ranges.
func SlotRanges(slots []*availabilitypbv1.Slot) []*availabilitypbv1.BookableRange {
	var out []*availabilitypbv1.BookableRange
	var start, end *timestamppb.Timestamp
	flush := func() {
		if start != nil {
			out = append(out, &availabilitypbv1.BookableRange{
				Window:   &sharedpbv1.TimeWindow{StartTime: start, EndTime: end},
				Bookable: true,
			})
			start = nil
		}
	}
	for _, s := range slots {
		if s.GetBookable() {
			if start == nil {
				start = s.GetStartTime()
			}
			end = s.GetEndTime()
		} else {
			flush()
		}
	}
	flush()
	return out
}

// LeadPrice is the price used for search sorting/display: the stay total for
// NIGHTLY (base × nights), or the single slot/unit price for TIME_SLOT.
func LeadPrice(u *UnitInfo, nights int32) *money.Money {
	if u.Price == nil {
		return nil
	}
	if u.Mode == ModeNightly && nights > 1 {
		total := (u.Price.GetUnits()*1_000_000_000 + int64(u.Price.GetNanos())) * int64(nights)
		return &money.Money{CurrencyCode: u.Price.GetCurrencyCode(), Units: total / 1_000_000_000, Nanos: int32(total % 1_000_000_000)}
	}
	return cloneMoney(u.Price)
}

// --- date/time helpers -------------------------------------------------------

func dateToTime(d *date.Date, loc *time.Location) time.Time {
	if d == nil {
		return time.Time{}
	}
	return time.Date(int(d.GetYear()), time.Month(d.GetMonth()), int(d.GetDay()), 0, 0, 0, 0, loc)
}

func timeToDate(t time.Time) *date.Date {
	return &date.Date{Year: int32(t.Year()), Month: int32(t.Month()), Day: int32(t.Day())}
}

func cloneMoney(m *money.Money) *money.Money {
	if m == nil {
		return nil
	}
	return &money.Money{CurrencyCode: m.GetCurrencyCode(), Units: m.GetUnits(), Nanos: m.GetNanos()}
}

// NightsBetween counts calendar nights of a date range in loc.
func NightsBetween(start, end *date.Date, loc *time.Location) int32 {
	s := dateToTime(start, loc)
	e := dateToTime(end, loc)
	n := int32(e.Sub(s).Hours() / 24)
	if n < 1 {
		return 1
	}
	return n
}
