package hasura

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	commonschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/commonql/schemaql"
	unitsql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/unitsql"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/the-protobuf-project/runtime-go/network/graphql"
	"google.golang.org/genproto/googleapis/type/money"
)

const rfc3339 = time.RFC3339

func deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

func derefInt32(p *int32) int32 {
	if p == nil {
		return 0
	}
	return *p
}

// fromBigdec parses a numeric(precision,scale) column's decimal-string value.
func fromBigdec(b graphql.Bigdecimal) float64 {
	f, _ := strconv.ParseFloat(string(b), 64)
	return f
}

func durationFromStr(s string) time.Duration {
	if s == "" {
		return 0
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0
	}
	return d
}

// parseTS parses a stored RFC 3339 timestamp string into a UTC time.
func parseTS(s string) time.Time {
	t, err := time.Parse(rfc3339, s)
	if err != nil {
		return time.Time{}
	}
	return t.UTC()
}

// parseDate parses a stored "2006-01-02" date at midnight in loc.
func parseDate(s string, loc *time.Location) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
}

func moneyFromSchema(m *commonschema.CommonMoneys) *money.Money {
	if m == nil {
		return nil
	}
	return &money.Money{
		CurrencyCode: deref(m.CurrencyCode),
		Units:        int64(deref(m.Units)),
		Nanos:        deref(m.Nanos),
	}
}

// protoWeekday maps the stored proto weekday name to a Go weekday.
var protoWeekday = map[string]time.Weekday{
	"WEEKDAY_SUNDAY":    time.Sunday,
	"WEEKDAY_MONDAY":    time.Monday,
	"WEEKDAY_TUESDAY":   time.Tuesday,
	"WEEKDAY_WEDNESDAY": time.Wednesday,
	"WEEKDAY_THURSDAY":  time.Thursday,
	"WEEKDAY_FRIDAY":    time.Friday,
	"WEEKDAY_SATURDAY":  time.Saturday,
}

func weekdaysFromStr(s *string) []time.Weekday {
	if s == nil || *s == "" {
		return nil
	}
	var out []time.Weekday
	for _, name := range strings.Split(*s, ",") {
		if wd, ok := protoWeekday[strings.TrimSpace(name)]; ok {
			out = append(out, wd)
		}
	}
	return out
}

// unitFilterPredicate translates the SearchAvailability filter into a unit where
// predicate. Supported fields: type (=/!=), display_name (=/!=/:), and a bareword
// term matching display_name. Returns nil when there are no conditions.
func unitFilterPredicate(filter string) (*graphql.Predicate, error) {
	conds, err := types.ParseFilter(filter)
	if err != nil {
		return nil, err
	}
	preds := make([]graphql.Predicate, 0, len(conds))
	for _, c := range conds {
		switch c.Field {
		case "":
			preds = append(preds, unitsql.DisplayName.ILike("%"+c.Value+"%"))
		case "display_name":
			switch c.Op {
			case types.FilterEq:
				preds = append(preds, unitsql.DisplayName.Eq(c.Value))
			case types.FilterNeq:
				preds = append(preds, unitsql.DisplayName.Neq(c.Value))
			case types.FilterHas:
				preds = append(preds, unitsql.DisplayName.ILike("%"+c.Value+"%"))
			default:
				return nil, fmt.Errorf("%w: unsupported operator for display_name", types.ErrInvalidArgument)
			}
		case "type":
			val := strings.TrimPrefix(strings.ToUpper(c.Value), "UNIT_TYPE_")
			switch c.Op {
			case types.FilterEq:
				preds = append(preds, unitsql.Type.Eq(val))
			case types.FilterNeq:
				preds = append(preds, unitsql.Type.Neq(val))
			default:
				return nil, fmt.Errorf("%w: unsupported operator for type", types.ErrInvalidArgument)
			}
		default:
			return nil, fmt.Errorf("%w: cannot filter by %q", types.ErrInvalidArgument, c.Field)
		}
	}
	switch len(preds) {
	case 0:
		return nil, nil
	case 1:
		return &preds[0], nil
	default:
		p := graphql.And(preds...)
		return &p, nil
	}
}

// mapErr translates GraphQL/runtime errors into the repository sentinels.
func mapErr(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, graphql.ErrConflict):
		return types.ErrConflict
	}
	return err
}
