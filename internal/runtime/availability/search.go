// The storefront catalog search RPC and its ordering.
package availability

import (
	"context"
	"sort"
	"time"

	"github.com/oh-tarnished/freebusy/internal/runtime/rpc"
	"github.com/oh-tarnished/freebusy/internal/service/availability/engine"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/availability/v1/availabilitypbv1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SearchAvailability sweeps the catalog for units bookable over a period.
func (s *Server) SearchAvailability(ctx context.Context, req *availabilitypbv1.SearchAvailabilityRequest) (*availabilitypbv1.SearchAvailabilityResponse, error) {
	var out *availabilitypbv1.SearchAvailabilityResponse
	err := rpc.Traced(ctx, "AvailabilityService", "SearchAvailability", func(ctx context.Context) error {
		desc, err := priceDescending(req.GetOrderBy())
		if err != nil {
			return err
		}
		units, err := s.reader.SearchUnits(ctx, req.GetProperty(), req.GetOrganisation(), req.GetFilter())
		if err != nil {
			return rpc.ToStatusErr(err)
		}

		// Batch the per-unit reads: one bookings query and one closures query over
		// the whole candidate set, bounded by the widest period across the units.
		unitIDs := make([]string, 0, len(units))
		tzByUnit := make(map[string]string, len(units))
		var lo, hi time.Time
		for _, u := range units {
			p, perr := resolvePeriod(u, req.GetWindow(), req.GetDateRange())
			if perr != nil {
				return perr
			}
			unitIDs = append(unitIDs, u.ID)
			tzByUnit[u.ID] = u.TimeZone
			if lo.IsZero() || p.start.Before(lo) {
				lo = p.start
			}
			if p.end.After(hi) {
				hi = p.end
			}
		}
		bookingsByUnit, err := s.reader.ActiveBookingsForUnits(ctx, unitIDs, lo, hi)
		if err != nil {
			return rpc.ToStatusErr(err)
		}
		closuresByUnit, err := s.reader.ClosuresForUnits(ctx, unitIDs, tzByUnit)
		if err != nil {
			return rpc.ToStatusErr(err)
		}

		now := time.Now().UTC()
		matches := make([]*availabilitypbv1.AvailabilityMatch, 0, len(units))
		for _, u := range units {
			p, _ := resolvePeriod(u, req.GetWindow(), req.GetDateRange())
			bookable, _, _ := engine.CheckSpan(u, p.start, p.end, req.GetUnits(), p.nights, bookingsByUnit[u.ID], closuresByUnit[u.ID], now)
			if !bookable && !req.GetIncludeUnavailable() {
				continue
			}
			matches = append(matches, &availabilitypbv1.AvailabilityMatch{
				Unit:        u.Name,
				DisplayName: u.DisplayName,
				Mode:        modeProto(u.Mode),
				Bookable:    bookable,
				Price:       engine.LeadPrice(u, p.nights),
			})
		}
		sortMatches(matches, desc)

		limit, offset := types.PageBounds(req.GetPageSize(), req.GetPageToken())
		resp := &availabilitypbv1.SearchAvailabilityResponse{}
		if offset < len(matches) {
			end := offset + limit
			if end > len(matches) {
				end = len(matches)
			}
			resp.Matches = matches[offset:end]
			if end < len(matches) {
				resp.NextPageToken = types.EncodeOffset(end)
			}
		}
		out = resp
		return nil
	})
	return out, err
}

// priceDescending parses order_by, accepting only "price" (asc/desc). Empty means
// price ascending.
func priceDescending(orderBy string) (bool, error) {
	if orderBy == "" {
		return false, nil
	}
	terms, err := types.ParseOrderBy(orderBy)
	if err != nil {
		return false, err
	}
	for _, t := range terms {
		if t.Field != "price" {
			return false, status.Errorf(codes.InvalidArgument, "cannot sort by %q", t.Field)
		}
		return t.Desc, nil
	}
	return false, nil
}

func sortMatches(m []*availabilitypbv1.AvailabilityMatch, desc bool) {
	sort.SliceStable(m, func(i, j int) bool {
		a, b := priceNanos(m[i]), priceNanos(m[j])
		if desc {
			return a > b
		}
		return a < b
	})
}

func priceNanos(m *availabilitypbv1.AvailabilityMatch) int64 {
	p := m.GetPrice()
	return p.GetUnits()*1_000_000_000 + int64(p.GetNanos())
}
