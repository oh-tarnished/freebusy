// Scalar codecs shared by the schedule graph: timestamps, dates, durations, weekdays, kinds.
package hasura

import (
	"strings"
	"time"

	"github.com/oh-tarnished/freebusy/protobuf/generated/go/schedule/v1/schedulepbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/shared/v1/sharedpbv1"
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	rfc3339    = time.RFC3339
	dateLayout = "2006-01-02"
)

func strOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func strToTS(s string) *timestamppb.Timestamp {
	if s == "" {
		return nil
	}
	t, err := time.Parse(rfc3339, s)
	if err != nil {
		return nil
	}
	return timestamppb.New(t)
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
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return nil
	}
	return &date.Date{Year: int32(t.Year()), Month: int32(t.Month()), Day: int32(t.Day())}
}

func durationToStr(d *durationpb.Duration) string {
	if d == nil {
		return ""
	}
	return d.AsDuration().String()
}

func durationFromStr(s string) *durationpb.Duration {
	if s == "" {
		return nil
	}
	d, err := time.ParseDuration(s)
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

func kindToStr(k schedulepbv1.ExceptionKind) string {
	if k == schedulepbv1.ExceptionKind_EXCEPTION_KIND_UNSPECIFIED {
		return ""
	}
	return strings.TrimPrefix(k.String(), "EXCEPTION_KIND_")
}

func kindFromStr(s string) schedulepbv1.ExceptionKind {
	if s == "" {
		return schedulepbv1.ExceptionKind_EXCEPTION_KIND_UNSPECIFIED
	}
	return schedulepbv1.ExceptionKind(schedulepbv1.ExceptionKind_value["EXCEPTION_KIND_"+s])
}
