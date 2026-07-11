package availability

import (
	"context"
	"github.com/oh-tarnished/freebusy/internal/database"
	"time"

	"github.com/oh-tarnished/freebusy/internal/runtime/rpc"
	"github.com/oh-tarnished/freebusy/internal/service/availability/db"
	"github.com/oh-tarnished/freebusy/internal/service/availability/engine"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/availability/v1/availabilitypbv1"
)

// Server implements availabilitypbv1.AvailabilityServiceServer over the read port.
type Server struct {
	availabilitypbv1.UnimplementedAvailabilityServiceServer
	reader db.AvailabilityReader
}

// New builds the availability service on conn: the provider-selected repository
// wrapped in the gRPC server implementation.
func New(conn *database.Connection) *Server {
	return NewServer(db.New(conn))
}

// NewServer returns a Server backed by reader.
func NewServer(reader db.AvailabilityReader) *Server {
	return &Server{reader: reader}
}

// computeOne runs a single ComputeAvailabilityRequest against the engine.
func (s *Server) computeOne(ctx context.Context, req *availabilitypbv1.ComputeAvailabilityRequest) (*availabilitypbv1.UnitAvailability, error) {
	info, err := s.reader.GetUnit(ctx, req.GetUnit())
	if err != nil {
		return nil, rpc.ToStatusErr(err)
	}
	p, err := resolvePeriod(info, req.GetWindow(), req.GetDateRange())
	if err != nil {
		return nil, err
	}
	res, err := s.reader.ActiveBookings(ctx, info.ID, p.start, p.end)
	if err != nil {
		return nil, rpc.ToStatusErr(err)
	}
	clo, err := s.reader.Closures(ctx, info.ID, info.TimeZone)
	if err != nil {
		return nil, rpc.ToStatusErr(err)
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
	err := rpc.Traced(ctx, "AvailabilityService", "ComputeAvailability", func(ctx context.Context) error {
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
	var out *availabilitypbv1.CheckAvailabilityResponse
	err := rpc.Traced(ctx, "AvailabilityService", "CheckAvailability", func(ctx context.Context) error {
		info, err := s.reader.GetUnit(ctx, req.GetUnit())
		if err != nil {
			return rpc.ToStatusErr(err)
		}
		p, err := resolvePeriod(info, req.GetWindow(), req.GetDateRange())
		if err != nil {
			return err
		}
		res, err := s.reader.ActiveBookings(ctx, info.ID, p.start, p.end)
		if err != nil {
			return rpc.ToStatusErr(err)
		}
		clo, err := s.reader.Closures(ctx, info.ID, info.TimeZone)
		if err != nil {
			return rpc.ToStatusErr(err)
		}
		bookable, free, reasons := engine.CheckSpan(info, p.start, p.end, req.GetUnits(), p.nights, res, clo, time.Now().UTC())
		out = &availabilitypbv1.CheckAvailabilityResponse{Bookable: bookable, FreeCount: free, Reasons: reasons}
		return nil
	})
	return out, err
}

// ComputeBookableRanges computes contiguous bookable ranges within a window.
func (s *Server) ComputeBookableRanges(ctx context.Context, req *availabilitypbv1.ComputeBookableRangesRequest) (*availabilitypbv1.ComputeBookableRangesResponse, error) {
	var out *availabilitypbv1.ComputeBookableRangesResponse
	err := rpc.Traced(ctx, "AvailabilityService", "ComputeBookableRanges", func(ctx context.Context) error {
		info, err := s.reader.GetUnit(ctx, req.GetUnit())
		if err != nil {
			return rpc.ToStatusErr(err)
		}
		p, err := resolvePeriod(info, req.GetWindow(), req.GetDateRange())
		if err != nil {
			return err
		}
		res, err := s.reader.ActiveBookings(ctx, info.ID, p.start, p.end)
		if err != nil {
			return rpc.ToStatusErr(err)
		}
		clo, err := s.reader.Closures(ctx, info.ID, info.TimeZone)
		if err != nil {
			return rpc.ToStatusErr(err)
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
	err := rpc.Traced(ctx, "AvailabilityService", "BatchComputeAvailability", func(ctx context.Context) error {
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
