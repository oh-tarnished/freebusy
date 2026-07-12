package engine

import (
	"sync"
	"time"

	"github.com/oh-tarnished/freebusy/protobuf/generated/go/availability/v1/availabilitypbv1"
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// locations caches loaded timezone locations by IANA name: time.LoadLocation
// reads tzdata on every call, and search resolves one location per unit.
var locations sync.Map

// loc returns the unit's timezone location, falling back to UTC.
func (u *UnitInfo) loc() *time.Location {
	if l, ok := locations.Load(u.TimeZone); ok {
		return l.(*time.Location)
	}
	l, err := time.LoadLocation(u.TimeZone)
	if err != nil {
		l = time.UTC
	}
	locations.Store(u.TimeZone, l)
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
	loc := u.loc()
	var out []*availabilitypbv1.Slot
	for s := start.UTC(); !s.Add(dur).After(end.UTC()); s = s.Add(dur) {
		e := s.Add(dur)
		free := u.Capacity - overlapUnits(res, s, e)
		if free < 0 {
			free = 0
		}
		bookable := !u.Archived &&
			bufferedFree(u, s, e, res) >= unitsReq &&
			!closedOver(closures, s, e) &&
			passesNotice(u, s, now) &&
			u.Recurring.Covers(s.In(loc), e.In(loc))
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

// bufferedFree is the free count when the candidate span carries its setup/
// turnover deltas and existing reservations are padded by the required gap — so a
// booking too close to an adjacent one counts against capacity.
func bufferedFree(u *UnitInfo, s, e time.Time, res []Reservation) int32 {
	if u.Gap <= 0 && u.StartDelta <= 0 && u.EndDelta <= 0 {
		free := u.Capacity - overlapUnits(res, s, e)
		if free < 0 {
			free = 0
		}
		return free
	}
	occStart := s.Add(-u.StartDelta)
	occEnd := e.Add(u.EndDelta)
	free := u.Capacity - overlapUnitsPadded(res, occStart, occEnd, u.Gap)
	if free < 0 {
		free = 0
	}
	return free
}

// overlapUnitsPadded sums units of reservations overlapping [s,e) after each
// reservation is padded by pad on both sides (the gap between bookings).
func overlapUnitsPadded(res []Reservation, s, e time.Time, pad time.Duration) int32 {
	var sum int32
	for i := range res {
		rs := res[i].Start.Add(-pad)
		re := res[i].End.Add(pad)
		if rs.Before(e) && re.After(s) {
			u := res[i].Units
			if u < 1 {
				u = 1
			}
			sum += u
		}
	}
	return sum
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

// --- date/time helpers -------------------------------------------------------
