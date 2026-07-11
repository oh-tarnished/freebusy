// Row-scalar helpers for the reader.
package gorm

import (
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/common"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/schedule"
	"github.com/oh-tarnished/freebusy/internal/service/availability/engine"
	"google.golang.org/genproto/googleapis/type/money"
)

// closureOf converts one exception row into an engine.Closure span in loc.
func closureOf(e *schedule.AvailabilityException, loc *time.Location) (engine.Closure, bool) {
	switch {
	case e.Window != nil:
		return engine.Closure{Start: e.Window.StartTime.UTC(), End: e.Window.EndTime.UTC()}, true
	case e.DateRange != nil:
		s, en := e.DateRange.StartDate, e.DateRange.EndDate
		start := time.Date(s.Year(), s.Month(), s.Day(), 0, 0, 0, 0, loc)
		end := time.Date(en.Year(), en.Month(), en.Day(), 0, 0, 0, 0, loc)
		return engine.Closure{Start: start.UTC(), End: end.UTC()}, true
	default:
		return engine.Closure{}, false
	}
}

func derefInt32(p *int32) int32 {
	if p == nil {
		return 0
	}
	return *p
}

func durationFromStr(s *string) time.Duration {
	if s == nil || *s == "" {
		return 0
	}
	d, err := time.ParseDuration(*s)
	if err != nil {
		return 0
	}
	return d
}

func moneyFromModel(m *common.Money) *money.Money {
	if m == nil {
		return nil
	}
	out := &money.Money{}
	if m.CurrencyCode != nil {
		out.CurrencyCode = *m.CurrencyCode
	}
	if m.Units != nil {
		out.Units = *m.Units
	}
	if m.Nanos != nil {
		out.Nanos = *m.Nanos
	}
	return out
}
