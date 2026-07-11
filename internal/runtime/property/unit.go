// Unit RPCs.
package property

import (
	"context"

	"github.com/oh-tarnished/freebusy/internal/database/repository/repox"
	"github.com/oh-tarnished/freebusy/internal/runtime/rpc"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/property/v1/propertypbv1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ListUnits returns a page of units under a property.
func (s *Server) ListUnits(ctx context.Context, req *propertypbv1.ListUnitsRequest) (*propertypbv1.ListUnitsResponse, error) {
	var out *propertypbv1.ListUnitsResponse
	err := rpc.Traced(ctx, "PropertyService", "ListUnits", func(ctx context.Context) error {
		items, next, err := s.repo.ListUnits(ctx, req.GetParent(), repox.ListInput{
			PageSize:  req.GetPageSize(),
			PageToken: req.GetPageToken(),
			OrderBy:   req.GetOrderBy(),
			Filter:    req.GetFilter(),
		})
		if err != nil {
			return rpc.ToStatusErr(err)
		}
		out = &propertypbv1.ListUnitsResponse{Units: items, NextPageToken: next}
		return nil
	})
	return out, err
}

// GetUnit returns a single unit by resource name.
func (s *Server) GetUnit(ctx context.Context, req *propertypbv1.GetUnitRequest) (*propertypbv1.Unit, error) {
	var out *propertypbv1.Unit
	err := rpc.Traced(ctx, "PropertyService", "GetUnit", func(ctx context.Context) error {
		u, err := s.repo.GetUnit(ctx, req.GetName())
		if err != nil {
			return rpc.ToStatusErr(err)
		}
		out = u
		return nil
	})
	return out, err
}

// CreateUnit creates a unit under a property. A caller-supplied unit_id fixes the
// resource name.
func (s *Server) CreateUnit(ctx context.Context, req *propertypbv1.CreateUnitRequest) (*propertypbv1.Unit, error) {
	u := proto.Clone(req.GetUnit()).(*propertypbv1.Unit)
	if id := req.GetUnitId(); id != "" {
		pid, err := types.PropertyID(req.GetParent())
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid parent")
		}
		name, err := types.UnitName(pid, id)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid unit_id")
		}
		u.Name = name
	}
	var out *propertypbv1.Unit
	err := rpc.Traced(ctx, "PropertyService", "CreateUnit", func(ctx context.Context) error {
		created, err := s.repo.CreateUnit(ctx, req.GetParent(), u)
		if err != nil {
			return rpc.ToStatusErr(err)
		}
		out = created
		return nil
	})
	return out, err
}

// UpdateUnit applies the update mask to an existing unit.
func (s *Server) UpdateUnit(ctx context.Context, req *propertypbv1.UpdateUnitRequest) (*propertypbv1.Unit, error) {
	u := req.GetUnit()
	var out *propertypbv1.Unit
	err := rpc.Traced(ctx, "PropertyService", "UpdateUnit", func(ctx context.Context) error {
		updated, err := s.repo.UpdateUnit(ctx, u, req.GetUpdateMask().GetPaths())
		if err != nil {
			return rpc.ToStatusErr(err)
		}
		out = updated
		return nil
	})
	return out, err
}

// DeleteUnit removes a unit by resource name. Child licences block the delete
// unless force is set.
func (s *Server) DeleteUnit(ctx context.Context, req *propertypbv1.DeleteUnitRequest) (*emptypb.Empty, error) {
	err := rpc.Traced(ctx, "PropertyService", "DeleteUnit", func(ctx context.Context) error {
		return rpc.ToStatusErr(s.repo.DeleteUnit(ctx, req.GetName(), req.GetForce()))
	})
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
