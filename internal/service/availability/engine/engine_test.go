package engine

import (
	"testing"
	"time"

	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/genproto/googleapis/type/money"
)

func inr(u int64) *money.Money { return &money.Money{CurrencyCode: "INR", Units: u} }

func d(y, m, day int) *date.Date { return &date.Date{Year: int32(y), Month: int32(m), Day: int32(day)} }

// A 3-night stay (Dec 24–27) on a 2-unit pool: one existing booking occupies the
// middle night, and a closure blacks out the first night.
func TestComputeNights(t *testing.T) {
	u := &UnitInfo{Mode: ModeNightly, Capacity: 2, TimeZone: "Asia/Kolkata", Price: inr(5000)}
	// Booking on the night of Dec 25 (IST) = [Dec 24 18:30 UTC, Dec 25 18:30 UTC):
	// Dec 25 04:00–10:00 UTC falls squarely inside it.
	res := []Reservation{{
		Start: time.Date(2026, 12, 25, 4, 0, 0, 0, time.UTC),
		End:   time.Date(2026, 12, 25, 10, 0, 0, 0, time.UTC),
		Units: 1,
	}}
	// Closure over Dec 24 night: [24 00:00 IST, 25 00:00 IST) = 23 18:30 → 24 18:30 UTC.
	clo := []Closure{{
		Start: time.Date(2026, 12, 23, 18, 30, 0, 0, time.UTC),
		End:   time.Date(2026, 12, 24, 18, 30, 0, 0, time.UTC),
	}}

	nights := ComputeNights(u, d(2026, 12, 24), d(2026, 12, 27), 1, res, clo)
	if len(nights) != 3 {
		t.Fatalf("got %d nights, want 3", len(nights))
	}
	if !nights[0].GetClosed() {
		t.Fatalf("night 0 (Dec 24) should be closed")
	}
	if nights[1].GetFreeUnits() != 1 {
		t.Fatalf("night 1 (Dec 25) free = %d, want 1 (2 − 1 booked)", nights[1].GetFreeUnits())
	}
	if nights[2].GetFreeUnits() != 2 || nights[2].GetClosed() {
		t.Fatalf("night 2 (Dec 26) should be fully free and open, got free=%d closed=%v", nights[2].GetFreeUnits(), nights[2].GetClosed())
	}
}

// A time-slot unit: hourly slots over a 3-hour window, one slot already fully
// booked (capacity 1).
func TestComputeSlots(t *testing.T) {
	u := &UnitInfo{Mode: ModeTimeSlot, Capacity: 1, TimeZone: "UTC", Duration: time.Hour, Price: inr(800)}
	start := time.Date(2026, 12, 24, 9, 0, 0, 0, time.UTC)
	end := start.Add(3 * time.Hour)
	res := []Reservation{{Start: start.Add(time.Hour), End: start.Add(2 * time.Hour), Units: 1}}

	slots := ComputeSlots(u, start, end, 0, 1, res, nil, start.Add(-24*time.Hour))
	if len(slots) != 3 {
		t.Fatalf("got %d slots, want 3", len(slots))
	}
	if !slots[0].GetBookable() || slots[1].GetBookable() || !slots[2].GetBookable() {
		t.Fatalf("bookable pattern = %v/%v/%v, want true/false/true", slots[0].GetBookable(), slots[1].GetBookable(), slots[2].GetBookable())
	}
}

// CheckSpan reports MIN_NIGHTS and NO_CAPACITY when the stay is too short and full.
func TestCheckSpanReasons(t *testing.T) {
	u := &UnitInfo{Mode: ModeNightly, Capacity: 1, TimeZone: "UTC", MinNights: 2, Price: inr(5000)}
	start := time.Date(2026, 12, 24, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 1) // 1 night, below min 2
	res := []Reservation{{Start: start, End: end, Units: 1}}

	bookable, free, reasons := CheckSpan(u, start, end, 1, 1, res, nil, start.Add(-48*time.Hour))
	if bookable {
		t.Fatal("span should not be bookable")
	}
	if free != 0 {
		t.Fatalf("free = %d, want 0", free)
	}
	var haveMin, haveCap bool
	for _, r := range reasons {
		switch r.GetCode().String() {
		case "CODE_MIN_NIGHTS":
			haveMin = true
		case "CODE_NO_CAPACITY":
			haveCap = true
		}
	}
	if !haveMin || !haveCap {
		t.Fatalf("reasons missing: minNights=%v noCapacity=%v", haveMin, haveCap)
	}
}
