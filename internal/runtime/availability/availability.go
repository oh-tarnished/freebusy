package availability

import (
	"context"
	"sort"
	"time"

	availdb "github.com/oh-tarnished/freebusy/internal/service/availability/db"
	"github.com/oh-tarnished/freebusy/internal/service/availability/engine"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/availability/v1/availabilitypbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/shared/v1/sharedpbv1"
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server implements availabilitypbv1.AvailabilityServiceServer over the read port.
type Server struct {
	availabilitypbv1.UnimplementedAvailabilityServiceServer
	reader availdb.AvailabilityReader
}

// NewServer returns a Server backed by reader.
func NewServer(reader availdb.AvailabilityReader) *Server {
	return &Server{reader: reader}
}

// period is a request's resolved period, in both instant and calendar forms.
type period struct {
	start, end time.Time // UTC instants
	dateStart  *date.Date
	dateEnd    *date.Date
	nights     int32
}

// resolvePeriod turns a request's window/date_range oneof into a period evaluated
// in the unit's timezone. Exactly one of window/dateRange must be set.
func resolvePeriod(u *engine.UnitInfo, window *sharedpbv1.TimeWindow, dr *sharedpbv1.DateRange) (period, error) {
	loc, err := time.LoadLocation(u.TimeZone)
	if err != nil {
		loc = time.UTC
	}
	switch {
	case dr != nil && dr.GetStartDate() != nil && dr.GetEndDate() != nil:
		ds, de := dr.GetStartDate(), dr.GetEndDate()
		return period{
			start:     startOfDate(ds, loc),
			end:       startOfDate(de, loc),
			dateStart: ds,
			dateEnd:   de,
			nights:    engine.NightsBetween(ds, de, loc),
		}, nil
	case window != nil && window.GetStartTime() != nil && window.GetEndTime() != nil:
		s := window.GetStartTime().AsTime()
		e := window.GetEndTime().AsTime()
		ds := dateOf(s.In(loc))
		de := dateOf(e.In(loc))
		return period{
			start:     s.UTC(),
			end:       e.UTC(),
			dateStart: ds,
			dateEnd:   de,
			nights:    engine.NightsBetween(ds, de, loc),
		}, nil
	default:
		return period{}, status.Error(codes.InvalidArgument, "a window or date_range period is required")
	}
}

// computeOne runs a single ComputeAvailabilityRequest against the engine.
func (s *Server) computeOne(ctx context.Context, req *availabilitypbv1.ComputeAvailabilityRequest) (*availabilitypbv1.UnitAvailability, error) {
	if req.GetUnit() == "" {
		return nil, status.Error(codes.InvalidArgument, "unit is required")
	}
	info, err := s.reader.GetUnit(ctx, req.GetUnit())
	if err != nil {
		return nil, toStatusErr(err)
	}
	p, err := resolvePeriod(info, req.GetWindow(), req.GetDateRange())
	if err != nil {
		return nil, err
	}
	res, err := s.reader.ActiveBookings(ctx, info.ID, p.start, p.end)
	if err != nil {
		return nil, toStatusErr(err)
	}
	clo, err := s.reader.Closures(ctx, info.ID, info.TimeZone)
	if err != nil {
		return nil, toStatusErr(err)
	}
	out := &availabilitypbv1.UnitAvailability{Unit: info.Name, Mode: modeProto(info.Mode)}
	if info.Mode == engine.ModeNightly {
		out.Nights = engine.ComputeNights(info, p.dateStart, p.dateEnd, req.GetUnits(), res, clo)
	} else {
		out.Slots = engine.ComputeSlots(info, p.start, p.end, req.GetDuration().AsDuration(), req.GetUnits(), res, clo, time.Now().UTC())
	}
	return out, nil
}

// ComputeAvailability computes availability for a unit over a period.
func (s *Server) ComputeAvailability(ctx context.Context, req *availabilitypbv1.ComputeAvailabilityRequest) (*availabilitypbv1.ComputeAvailabilityResponse, error) {
	var out *availabilitypbv1.ComputeAvailabilityResponse
	err := traced(ctx, "ComputeAvailability", func(ctx context.Context) error {
		ua, err := s.computeOne(ctx, req)
		if err != nil {
			return err
		}
		out = &availabilitypbv1.ComputeAvailabilityResponse{Mode: ua.GetMode(), Slots: ua.GetSlots(), Nights: ua.GetNights()}
		return nil
	})
	return out, err
}

// CheckAvailability tests whether one exact span is bookable.
func (s *Server) CheckAvailability(ctx context.Context, req *availabilitypbv1.CheckAvailabilityRequest) (*availabilitypbv1.CheckAvailabilityResponse, error) {
	if req.GetUnit() == "" {
		return nil, status.Error(codes.InvalidArgument, "unit is required")
	}
	var out *availabilitypbv1.CheckAvailabilityResponse
	err := traced(ctx, "CheckAvailability", func(ctx context.Context) error {
		info, err := s.reader.GetUnit(ctx, req.GetUnit())
		if err != nil {
			return toStatusErr(err)
		}
		p, err := resolvePeriod(info, req.GetWindow(), req.GetDateRange())
		if err != nil {
			return err
		}
		res, err := s.reader.ActiveBookings(ctx, info.ID, p.start, p.end)
		if err != nil {
			return toStatusErr(err)
		}
		clo, err := s.reader.Closures(ctx, info.ID, info.TimeZone)
		if err != nil {
			return toStatusErr(err)
		}
		bookable, free, reasons := engine.CheckSpan(info, p.start, p.end, req.GetUnits(), p.nights, res, clo, time.Now().UTC())
		out = &availabilitypbv1.CheckAvailabilityResponse{Bookable: bookable, FreeCount: free, Reasons: reasons}
		return nil
	})
	return out, err
}

// ComputeBookableRanges computes contiguous bookable ranges within a window.
func (s *Server) ComputeBookableRanges(ctx context.Context, req *availabilitypbv1.ComputeBookableRangesRequest) (*availabilitypbv1.ComputeBookableRangesResponse, error) {
	if req.GetUnit() == "" {
		return nil, status.Error(codes.InvalidArgument, "unit is required")
	}
	var out *availabilitypbv1.ComputeBookableRangesResponse
	err := traced(ctx, "ComputeBookableRanges", func(ctx context.Context) error {
		info, err := s.reader.GetUnit(ctx, req.GetUnit())
		if err != nil {
			return toStatusErr(err)
		}
		p, err := resolvePeriod(info, req.GetWindow(), req.GetDateRange())
		if err != nil {
			return err
		}
		res, err := s.reader.ActiveBookings(ctx, info.ID, p.start, p.end)
		if err != nil {
			return toStatusErr(err)
		}
		clo, err := s.reader.Closures(ctx, info.ID, info.TimeZone)
		if err != nil {
			return toStatusErr(err)
		}
		out = &availabilitypbv1.ComputeBookableRangesResponse{}
		if info.Mode == engine.ModeNightly {
			loc, _ := time.LoadLocation(info.TimeZone)
			if loc == nil {
				loc = time.UTC
			}
			nights := engine.ComputeNights(info, p.dateStart, p.dateEnd, req.GetUnits(), res, clo)
			out.Ranges = engine.NightRanges(nights, req.GetUnits(), loc)
		} else {
			slots := engine.ComputeSlots(info, p.start, p.end, req.GetDuration().AsDuration(), req.GetUnits(), res, clo, time.Now().UTC())
			out.Ranges = engine.SlotRanges(slots)
		}
		return nil
	})
	return out, err
}

// BatchComputeAvailability computes availability for several units at once.
func (s *Server) BatchComputeAvailability(ctx context.Context, req *availabilitypbv1.BatchComputeAvailabilityRequest) (*availabilitypbv1.BatchComputeAvailabilityResponse, error) {
	var out *availabilitypbv1.BatchComputeAvailabilityResponse
	err := traced(ctx, "BatchComputeAvailability", func(ctx context.Context) error {
		resp := &availabilitypbv1.BatchComputeAvailabilityResponse{Units: make([]*availabilitypbv1.UnitAvailability, 0, len(req.GetRequests()))}
		for _, r := range req.GetRequests() {
			ua, err := s.computeOne(ctx, r)
			if err != nil {
				return err
			}
			resp.Units = append(resp.Units, ua)
		}
		out = resp
		return nil
	})
	return out, err
}

// SearchAvailability sweeps the catalog for units bookable over a period.
func (s *Server) SearchAvailability(ctx context.Context, req *availabilitypbv1.SearchAvailabilityRequest) (*availabilitypbv1.SearchAvailabilityResponse, error) {
	var out *availabilitypbv1.SearchAvailabilityResponse
	err := traced(ctx, "SearchAvailability", func(ctx context.Context) error {
		desc, err := priceDescending(req.GetOrderBy())
		if err != nil {
			return err
		}
		units, err := s.reader.SearchUnits(ctx, req.GetProperty(), req.GetOrganisation(), req.GetFilter())
		if err != nil {
			return toStatusErr(err)
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
			return toStatusErr(err)
		}
		closuresByUnit, err := s.reader.ClosuresForUnits(ctx, unitIDs, tzByUnit)
		if err != nil {
			return toStatusErr(err)
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

		limit, offset := types.PageBounds(types.ListParams{PageSize: req.GetPageSize(), PageToken: req.GetPageToken()})
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

func modeProto(m string) sharedpbv1.BookingMode {
	switch m {
	case engine.ModeNightly:
		return sharedpbv1.BookingMode_BOOKING_MODE_NIGHTLY
	case engine.ModeTimeSlot:
		return sharedpbv1.BookingMode_BOOKING_MODE_TIME_SLOT
	default:
		return sharedpbv1.BookingMode_BOOKING_MODE_UNSPECIFIED
	}
}

func startOfDate(d *date.Date, loc *time.Location) time.Time {
	return time.Date(int(d.GetYear()), time.Month(d.GetMonth()), int(d.GetDay()), 0, 0, 0, 0, loc).UTC()
}

func dateOf(t time.Time) *date.Date {
	return &date.Date{Year: int32(t.Year()), Month: int32(t.Month()), Day: int32(t.Day())}
}
