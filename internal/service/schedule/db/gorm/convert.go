package gorm

import (
	"strings"
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/schedule"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/shared"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/schedule/v1/schedulepbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/shared/v1/sharedpbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// This file holds the pure conversions between the protobuf Schedule/
// AvailabilityException domain types and the normalized GORM storage models. A
// schedule nests recurring rules, buffers, stay constraints, and a cancellation
// policy (with refund tiers); an exception carries a TimeWindow or DateRange
// span. Durations serialize as Go-duration strings, weekday lists as comma-joined
// enum names, matching the single string columns the ORM generates.

func ptr[T any](v T) *T { return &v }

func deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

func strOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func timeToTS(t *time.Time) *timestamppb.Timestamp {
	if t == nil || t.IsZero() {
		return nil
	}
	return timestamppb.New(*t)
}

func dateToTime(d *date.Date) time.Time {
	if d == nil {
		return time.Time{}
	}
	return time.Date(int(d.GetYear()), time.Month(d.GetMonth()), int(d.GetDay()), 0, 0, 0, 0, time.UTC)
}

func timeToDate(t time.Time) *date.Date {
	if t.IsZero() {
		return nil
	}
	return &date.Date{Year: int32(t.Year()), Month: int32(t.Month()), Day: int32(t.Day())}
}

// durationToStr serializes a proto Duration into the string column the ORM
// generates (Go duration syntax, e.g. "48h0m0s"), round-tripping exactly.
func durationToStr(d *durationpb.Duration) *string {
	if d == nil {
		return nil
	}
	return ptr(d.AsDuration().String())
}

func durationFromStr(s *string) *durationpb.Duration {
	if s == nil || *s == "" {
		return nil
	}
	d, err := time.ParseDuration(*s)
	if err != nil {
		return nil
	}
	return durationpb.New(d)
}

func weekdaysToStr(days []sharedpbv1.Weekday) *string {
	if len(days) == 0 {
		return nil
	}
	parts := make([]string, 0, len(days))
	for _, d := range days {
		parts = append(parts, d.String())
	}
	return ptr(strings.Join(parts, ","))
}

func weekdaysFromStr(s *string) []sharedpbv1.Weekday {
	if s == nil || *s == "" {
		return nil
	}
	names := strings.Split(*s, ",")
	out := make([]sharedpbv1.Weekday, 0, len(names))
	for _, n := range names {
		out = append(out, sharedpbv1.Weekday(sharedpbv1.Weekday_value[strings.TrimSpace(n)]))
	}
	return out
}

func timeWindowToModel(w *sharedpbv1.TimeWindow) *shared.TimeWindow {
	if w == nil {
		return nil
	}
	return &shared.TimeWindow{
		ID:        ulid.GenerateString(),
		StartTime: w.GetStartTime().AsTime(),
		EndTime:   w.GetEndTime().AsTime(),
	}
}

func timeWindowFromModel(w *shared.TimeWindow) *sharedpbv1.TimeWindow {
	if w == nil {
		return nil
	}
	return &sharedpbv1.TimeWindow{
		StartTime: timestamppb.New(w.StartTime),
		EndTime:   timestamppb.New(w.EndTime),
	}
}

func dateRangeToModel(d *sharedpbv1.DateRange) *shared.DateRange {
	if d == nil {
		return nil
	}
	return &shared.DateRange{
		ID:        ulid.GenerateString(),
		StartDate: dateToTime(d.GetStartDate()),
		EndDate:   dateToTime(d.GetEndDate()),
	}
}

func dateRangeFromModel(d *shared.DateRange) *sharedpbv1.DateRange {
	if d == nil {
		return nil
	}
	return &sharedpbv1.DateRange{
		StartDate: timeToDate(d.StartDate),
		EndDate:   timeToDate(d.EndDate),
	}
}

func kindToModel(k schedulepbv1.ExceptionKind) schedule.ExceptionKind {
	return schedule.ExceptionKind(strings.TrimPrefix(k.String(), "EXCEPTION_KIND_"))
}

func kindFromModel(k schedule.ExceptionKind) schedulepbv1.ExceptionKind {
	return schedulepbv1.ExceptionKind(schedulepbv1.ExceptionKind_value["EXCEPTION_KIND_"+string(k)])
}
