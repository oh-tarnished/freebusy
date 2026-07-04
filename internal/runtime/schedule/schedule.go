// Package schedule is the gRPC/protobuf layer for the ScheduleService: it
// implements schedulepbv1.ScheduleServiceServer, owning request validation,
// observability, and the mapping of repository errors to gRPC status codes.
// Persistence stays behind the provider-agnostic db.ScheduleRepository.
package schedule

import (
	"context"

	scheduledb "github.com/oh-tarnished/freebusy/internal/service/schedule/db"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/schedule/v1/schedulepbv1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Server implements schedulepbv1.ScheduleServiceServer on top of a
// provider-agnostic db.ScheduleRepository.
type Server struct {
	schedulepbv1.UnimplementedScheduleServiceServer
	repo scheduledb.ScheduleRepository
}

// NewServer returns a Server backed by repo.
func NewServer(repo scheduledb.ScheduleRepository) *Server {
	return &Server{repo: repo}
}

// --- Schedule ----------------------------------------------------------------

// GetSchedule returns a unit's full availability configuration.
func (s *Server) GetSchedule(ctx context.Context, req *schedulepbv1.GetScheduleRequest) (*schedulepbv1.Schedule, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	var out *schedulepbv1.Schedule
	err := traced(ctx, "GetSchedule", func(ctx context.Context) error {
		sc, err := s.repo.GetSchedule(ctx, req.GetName())
		if err != nil {
			return toStatusErr(err)
		}
		out = sc
		return nil
	})
	return out, err
}

// UpdateSchedule upserts the masked sections of a unit's schedule.
func (s *Server) UpdateSchedule(ctx context.Context, req *schedulepbv1.UpdateScheduleRequest) (*schedulepbv1.Schedule, error) {
	sc := req.GetSchedule()
	if sc == nil || sc.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "schedule.name is required")
	}
	var out *schedulepbv1.Schedule
	err := traced(ctx, "UpdateSchedule", func(ctx context.Context) error {
		updated, err := s.repo.UpdateSchedule(ctx, sc, req.GetUpdateMask().GetPaths())
		if err != nil {
			return toStatusErr(err)
		}
		out = updated
		return nil
	})
	return out, err
}

// --- AvailabilityException ---------------------------------------------------

// ListAvailabilityExceptions returns a page of a unit's exceptions.
func (s *Server) ListAvailabilityExceptions(ctx context.Context, req *schedulepbv1.ListAvailabilityExceptionsRequest) (*schedulepbv1.ListAvailabilityExceptionsResponse, error) {
	if req.GetParent() == "" {
		return nil, status.Error(codes.InvalidArgument, "parent is required")
	}
	filter, err := types.ParseFilter(req.GetFilter())
	if err != nil {
		return nil, toStatusErr(err)
	}
	var out *schedulepbv1.ListAvailabilityExceptionsResponse
	err = traced(ctx, "ListAvailabilityExceptions", func(ctx context.Context) error {
		items, next, err := s.repo.ListAvailabilityExceptions(ctx, req.GetParent(), types.ListParams{
			PageSize:  req.GetPageSize(),
			PageToken: req.GetPageToken(),
			OrderBy:   req.GetOrderBy(),
			Filter:    filter,
		})
		if err != nil {
			return toStatusErr(err)
		}
		out = &schedulepbv1.ListAvailabilityExceptionsResponse{AvailabilityExceptions: items, NextPageToken: next}
		return nil
	})
	return out, err
}

// GetAvailabilityException returns an exception by resource name.
func (s *Server) GetAvailabilityException(ctx context.Context, req *schedulepbv1.GetAvailabilityExceptionRequest) (*schedulepbv1.AvailabilityException, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	var out *schedulepbv1.AvailabilityException
	err := traced(ctx, "GetAvailabilityException", func(ctx context.Context) error {
		e, err := s.repo.GetAvailabilityException(ctx, req.GetName())
		if err != nil {
			return toStatusErr(err)
		}
		out = e
		return nil
	})
	return out, err
}

// CreateAvailabilityException adds a closure or extra-hours exception to a unit.
func (s *Server) CreateAvailabilityException(ctx context.Context, req *schedulepbv1.CreateAvailabilityExceptionRequest) (*schedulepbv1.AvailabilityException, error) {
	e := req.GetAvailabilityException()
	switch {
	case req.GetParent() == "":
		return nil, status.Error(codes.InvalidArgument, "parent is required")
	case e == nil:
		return nil, status.Error(codes.InvalidArgument, "availability_exception is required")
	case e.GetKind() == schedulepbv1.ExceptionKind_EXCEPTION_KIND_UNSPECIFIED:
		return nil, status.Error(codes.InvalidArgument, "availability_exception.kind is required")
	case e.GetWindow() == nil && e.GetDateRange() == nil:
		return nil, status.Error(codes.InvalidArgument, "availability_exception must set a span: window or date_range")
	}
	e = proto.Clone(e).(*schedulepbv1.AvailabilityException)
	if id := req.GetAvailabilityExceptionId(); id != "" {
		propertyID, unitID, uerr := types.ParseUnitParent(req.GetParent())
		if uerr != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid parent")
		}
		name, nerr := types.AvailabilityExceptionName(propertyID, unitID, id)
		if nerr != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid availability_exception_id")
		}
		e.Name = name
	}
	var out *schedulepbv1.AvailabilityException
	err := traced(ctx, "CreateAvailabilityException", func(ctx context.Context) error {
		created, err := s.repo.CreateAvailabilityException(ctx, req.GetParent(), e)
		if err != nil {
			return toStatusErr(err)
		}
		out = created
		return nil
	})
	return out, err
}

// DeleteAvailabilityException removes an exception by resource name.
func (s *Server) DeleteAvailabilityException(ctx context.Context, req *schedulepbv1.DeleteAvailabilityExceptionRequest) (*emptypb.Empty, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	err := traced(ctx, "DeleteAvailabilityException", func(ctx context.Context) error {
		return toStatusErr(s.repo.DeleteAvailabilityException(ctx, req.GetName()))
	})
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
