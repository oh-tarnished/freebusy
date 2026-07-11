package hasura

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/oh-tarnished/freebusy/protobuf/generated/go/shared/v1/sharedpbv1"
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/genproto/googleapis/type/timeofday"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// This file holds the pure conversions between the protobuf Property/Unit domain
// types and the normalized Hasura/GraphQL schema. Timestamps cross the boundary
// as RFC 3339 strings, dates as "2006-01-02", times as "15:04:05"; enums as their
// bare value name (no proto prefix); FK ids as strings (empty means unset).

const dateLayout = "2006-01-02"

// orgName rebuilds the "organisations/{id}" resource name from the bare id the
// property row's organisation FK column stores (the FK references
// organisations.id).
func orgName(id string) string {
	if id == "" {
		return ""
	}
	return "organisations/" + id
}

func strToTS(s string) *timestamppb.Timestamp {
	if s == "" {
		return nil
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02T15:04:05.999999Z07:00", "2006-01-02T15:04:05"} {
		if t, err := time.Parse(layout, s); err == nil {
			return timestamppb.New(t)
		}
	}
	return nil
}

func dateToStr(d *date.Date) string {
	if d == nil {
		return ""
	}
	return time.Date(int(d.GetYear()), time.Month(d.GetMonth()), int(d.GetDay()), 0, 0, 0, 0, time.UTC).Format(dateLayout)
}

func strToDate(s string) *date.Date {
	if s == "" {
		return nil
	}
	t, err := time.Parse(dateLayout, s[:min(len(s), 10)])
	if err != nil {
		return nil
	}
	return &date.Date{Year: int32(t.Year()), Month: int32(t.Month()), Day: int32(t.Day())}
}

func todToStr(t *timeofday.TimeOfDay) string {
	if t == nil {
		return ""
	}
	return time.Date(0, 1, 1, int(t.GetHours()), int(t.GetMinutes()), int(t.GetSeconds()), 0, time.UTC).Format("15:04:05")
}

func strToTOD(s *string) *timeofday.TimeOfDay {
	if s == nil || *s == "" {
		return nil
	}
	t, err := time.Parse("15:04:05", (*s)[:min(len(*s), 8)])
	if err != nil {
		return nil
	}
	return &timeofday.TimeOfDay{Hours: int32(t.Hour()), Minutes: int32(t.Minute()), Seconds: int32(t.Second())}
}

func durationToStr(d *durationpb.Duration) string {
	if d == nil {
		return ""
	}
	return d.AsDuration().String()
}

func strToDuration(s *string) *durationpb.Duration {
	if s == nil || *s == "" {
		return nil
	}
	d, err := time.ParseDuration(*s)
	if err != nil {
		return nil
	}
	return durationpb.New(d)
}

func weekdaysToStr(days []sharedpbv1.Weekday) string {
	if len(days) == 0 {
		return ""
	}
	parts := make([]string, 0, len(days))
	for _, d := range days {
		parts = append(parts, d.String())
	}
	return strings.Join(parts, ",")
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

// strSliceToPtrs adapts a proto repeated string to the []*string GraphQL arrays
// use; strPtrsToSlice reverses it (dropping nils).
func strSliceToPtrs(in []string) []*string {
	if len(in) == 0 {
		return nil
	}
	out := make([]*string, len(in))
	for i := range in {
		v := in[i]
		out[i] = &v
	}
	return out
}

func strPtrsToSlice(in []*string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, 0, len(in))
	for _, p := range in {
		if p != nil {
			out = append(out, *p)
		}
	}
	return out
}

func structToJSON(s *structpb.Struct) []byte {
	if s == nil {
		return nil
	}
	b, err := s.MarshalJSON()
	if err != nil {
		return nil
	}
	return b
}

// jsonBytes unwraps the *json.RawMessage the GraphQL schema uses for jsonb
// columns into the []byte the struct conversions expect.
func jsonBytes(r *json.RawMessage) []byte {
	if r == nil {
		return nil
	}
	return []byte(*r)
}

func jsonToStruct(b []byte) *structpb.Struct {
	if len(b) == 0 {
		return nil
	}
	s := &structpb.Struct{}
	if err := s.UnmarshalJSON(b); err != nil {
		return nil
	}
	return s
}

// --- value-object inputs/reads ----------------------------------------------

// --- enum <-> bare value-name conversions -----------------------------------
