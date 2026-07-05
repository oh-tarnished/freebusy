package engine

import (
	"testing"
	"time"

	"github.com/oh-tarnished/freebusy/protobuf/generated/go/availability/v1/availabilitypbv1"
	"github.com/oh-tarnished/freebusy/shared/rrule"
)

// A time-slot unit open only Mon–Fri 09:00–17:00: slots outside those hours are
// not bookable even when free.
func TestSlotsGatedToOpenHours(t *testing.T) {
	u := &UnitInfo{
		Mode:      ModeTimeSlot,
		Capacity:  1,
		TimeZone:  "UTC",
		Duration:  time.Hour,
		Price:     inr(800),
		Recurring: []rrule.Rule{{RRule: "FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR", Opens: "09:00", Closes: "17:00"}},
	}
	// 2026-12-24 is a Thursday. Window 08:00–11:00 → slots 08–09 (closed), 09–10, 10–11 (open).
	start := time.Date(2026, 12, 24, 8, 0, 0, 0, time.UTC)
	slots := ComputeSlots(u, start, start.Add(3*time.Hour), 0, 1, nil, nil, start.Add(-24*time.Hour))
	if len(slots) != 3 {
		t.Fatalf("got %d slots, want 3", len(slots))
	}
	if slots[0].GetBookable() {
		t.Fatal("08:00–09:00 is before open hours, should not be bookable")
	}
	if !slots[1].GetBookable() || !slots[2].GetBookable() {
		t.Fatal("09:00–11:00 is within open hours, should be bookable")
	}
}

// A gap buffer blocks a slot adjacent to an existing booking even though the slot
// itself does not overlap it: reported as CODE_BUFFER_CONFLICT, distinct from
// NO_CAPACITY.
func TestBufferConflict(t *testing.T) {
	u := &UnitInfo{Mode: ModeTimeSlot, Capacity: 1, TimeZone: "UTC", Duration: time.Hour, Gap: 30 * time.Minute, Price: inr(800)}
	start := time.Date(2026, 12, 24, 10, 0, 0, 0, time.UTC)
	// Existing booking 09:00–10:00; a 10:00–11:00 slot doesn't overlap but is within
	// the 30-min gap.
	res := []Reservation{{Start: start.Add(-time.Hour), End: start, Units: 1}}

	bookable, free, reasons := CheckSpan(u, start, start.Add(time.Hour), 1, 0, res, nil, start.Add(-48*time.Hour))
	if bookable {
		t.Fatal("should be blocked by the buffer")
	}
	if free != 1 {
		t.Fatalf("raw free = %d, want 1 (no direct overlap)", free)
	}
	if !hasCode(reasons, "CODE_BUFFER_CONFLICT") || hasCode(reasons, "CODE_NO_CAPACITY") {
		t.Fatalf("want BUFFER_CONFLICT and not NO_CAPACITY, got %v", codes(reasons))
	}
}

// Check-in on a disallowed weekday yields CODE_CHECKIN_DAY.
func TestCheckinWeekday(t *testing.T) {
	// Allow check-in only on Fridays.
	u := &UnitInfo{Mode: ModeNightly, Capacity: 1, TimeZone: "UTC", CheckinWeekdays: []time.Weekday{time.Friday}, Price: inr(5000)}
	// 2026-12-24 is a Thursday.
	start := time.Date(2026, 12, 24, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 2)

	bookable, _, reasons := CheckSpan(u, start, end, 1, 2, nil, nil, start.Add(-72*time.Hour))
	if bookable || !hasCode(reasons, "CODE_CHECKIN_DAY") {
		t.Fatalf("Thursday check-in should be rejected with CHECKIN_DAY, got bookable=%v reasons=%v", bookable, codes(reasons))
	}
}

func hasCode(reasons []*availabilitypbv1.UnbookableReason, code string) bool {
	for _, r := range reasons {
		if r.GetCode().String() == code {
			return true
		}
	}
	return false
}

func codes(reasons []*availabilitypbv1.UnbookableReason) []string {
	out := make([]string, 0, len(reasons))
	for _, r := range reasons {
		out = append(out, r.GetCode().String())
	}
	return out
}
