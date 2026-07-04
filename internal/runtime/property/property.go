// Package property is the gRPC/protobuf layer for the PropertyService: it
// implements propertypbv1.PropertyServiceServer, owning request validation,
// observability, and the mapping of repository errors to gRPC status codes. All
// protobuf concerns live here; persistence stays behind the provider-agnostic
// db.PropertyRepository, so the database layer is agnostic to protobuf and gRPC.
package property

import (
	"context"

	propertydb "github.com/oh-tarnished/freebusy/internal/service/property/db"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/property/v1/propertypbv1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Server implements propertypbv1.PropertyServiceServer on top of a
// provider-agnostic db.PropertyRepository.
type Server struct {
	propertypbv1.UnimplementedPropertyServiceServer
	repo propertydb.PropertyRepository
}

// NewServer returns a Server backed by repo.
func NewServer(repo propertydb.PropertyRepository) *Server {
	return &Server{repo: repo}
}

// --- Property ----------------------------------------------------------------

// ListProperties returns a page of properties for the given pagination request.
func (s *Server) ListProperties(ctx context.Context, req *propertypbv1.ListPropertiesRequest) (*propertypbv1.ListPropertiesResponse, error) {
	filter, err := types.ParseFilter(req.GetFilter())
	if err != nil {
		return nil, toStatusErr(err)
	}
	var out *propertypbv1.ListPropertiesResponse
	err = traced(ctx, "ListProperties", func(ctx context.Context) error {
		items, next, err := s.repo.ListProperties(ctx, types.ListParams{
			PageSize:  req.GetPageSize(),
			PageToken: req.GetPageToken(),
			OrderBy:   req.GetOrderBy(),
			Filter:    filter,
		})
		if err != nil {
			return toStatusErr(err)
		}
		out = &propertypbv1.ListPropertiesResponse{Properties: items, NextPageToken: next}
		return nil
	})
	return out, err
}

// GetProperty returns a single property by resource name.
func (s *Server) GetProperty(ctx context.Context, req *propertypbv1.GetPropertyRequest) (*propertypbv1.Property, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	var out *propertypbv1.Property
	err := traced(ctx, "GetProperty", func(ctx context.Context) error {
		p, err := s.repo.GetProperty(ctx, req.GetName())
		if err != nil {
			return toStatusErr(err)
		}
		out = p
		return nil
	})
	return out, err
}

// CreateProperty creates a property. A caller-supplied property_id fixes the
// resource name.
func (s *Server) CreateProperty(ctx context.Context, req *propertypbv1.CreatePropertyRequest) (*propertypbv1.Property, error) {
	p := req.GetProperty()
	switch {
	case p == nil:
		return nil, status.Error(codes.InvalidArgument, "property is required")
	case p.GetOrganisation() == "":
		return nil, status.Error(codes.InvalidArgument, "property.organisation is required")
	case p.GetDisplayName() == "":
		return nil, status.Error(codes.InvalidArgument, "property.display_name is required")
	case p.GetTimeZone() == "":
		return nil, status.Error(codes.InvalidArgument, "property.time_zone is required")
	}
	p = proto.Clone(p).(*propertypbv1.Property)
	if id := req.GetPropertyId(); id != "" {
		name, err := types.PropertyName(id)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid property_id")
		}
		p.Name = name
	}
	var out *propertypbv1.Property
	err := traced(ctx, "CreateProperty", func(ctx context.Context) error {
		created, err := s.repo.CreateProperty(ctx, p)
		if err != nil {
			return toStatusErr(err)
		}
		out = created
		return nil
	})
	return out, err
}

// UpdateProperty applies the update mask to an existing property. An empty mask
// replaces every mutable field.
func (s *Server) UpdateProperty(ctx context.Context, req *propertypbv1.UpdatePropertyRequest) (*propertypbv1.Property, error) {
	p := req.GetProperty()
	if p == nil || p.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "property.name is required")
	}
	var out *propertypbv1.Property
	err := traced(ctx, "UpdateProperty", func(ctx context.Context) error {
		updated, err := s.repo.UpdateProperty(ctx, p, req.GetUpdateMask().GetPaths())
		if err != nil {
			return toStatusErr(err)
		}
		out = updated
		return nil
	})
	return out, err
}

// ArchiveProperty hides a property from the storefront and new bookings.
func (s *Server) ArchiveProperty(ctx context.Context, req *propertypbv1.ArchivePropertyRequest) (*propertypbv1.Property, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	var out *propertypbv1.Property
	err := traced(ctx, "ArchiveProperty", func(ctx context.Context) error {
		p, err := s.repo.ArchiveProperty(ctx, req.GetName())
		if err != nil {
			return toStatusErr(err)
		}
		out = p
		return nil
	})
	return out, err
}

// UnarchiveProperty restores an archived property to the active state.
func (s *Server) UnarchiveProperty(ctx context.Context, req *propertypbv1.UnarchivePropertyRequest) (*propertypbv1.Property, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	var out *propertypbv1.Property
	err := traced(ctx, "UnarchiveProperty", func(ctx context.Context) error {
		p, err := s.repo.UnarchiveProperty(ctx, req.GetName())
		if err != nil {
			return toStatusErr(err)
		}
		out = p
		return nil
	})
	return out, err
}

// --- Unit --------------------------------------------------------------------

// ListUnits returns a page of units under a property.
func (s *Server) ListUnits(ctx context.Context, req *propertypbv1.ListUnitsRequest) (*propertypbv1.ListUnitsResponse, error) {
	if req.GetParent() == "" {
		return nil, status.Error(codes.InvalidArgument, "parent is required")
	}
	filter, err := types.ParseFilter(req.GetFilter())
	if err != nil {
		return nil, toStatusErr(err)
	}
	var out *propertypbv1.ListUnitsResponse
	err = traced(ctx, "ListUnits", func(ctx context.Context) error {
		items, next, err := s.repo.ListUnits(ctx, req.GetParent(), types.ListParams{
			PageSize:  req.GetPageSize(),
			PageToken: req.GetPageToken(),
			OrderBy:   req.GetOrderBy(),
			Filter:    filter,
		})
		if err != nil {
			return toStatusErr(err)
		}
		out = &propertypbv1.ListUnitsResponse{Units: items, NextPageToken: next}
		return nil
	})
	return out, err
}

// GetUnit returns a single unit by resource name.
func (s *Server) GetUnit(ctx context.Context, req *propertypbv1.GetUnitRequest) (*propertypbv1.Unit, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	var out *propertypbv1.Unit
	err := traced(ctx, "GetUnit", func(ctx context.Context) error {
		u, err := s.repo.GetUnit(ctx, req.GetName())
		if err != nil {
			return toStatusErr(err)
		}
		out = u
		return nil
	})
	return out, err
}

// CreateUnit creates a unit under a property. A caller-supplied unit_id fixes the
// resource name.
func (s *Server) CreateUnit(ctx context.Context, req *propertypbv1.CreateUnitRequest) (*propertypbv1.Unit, error) {
	u := req.GetUnit()
	switch {
	case req.GetParent() == "":
		return nil, status.Error(codes.InvalidArgument, "parent is required")
	case u == nil:
		return nil, status.Error(codes.InvalidArgument, "unit is required")
	case u.GetDisplayName() == "":
		return nil, status.Error(codes.InvalidArgument, "unit.display_name is required")
	case u.GetType() == propertypbv1.UnitType_UNIT_TYPE_UNSPECIFIED:
		return nil, status.Error(codes.InvalidArgument, "unit.type is required")
	case u.GetBookingMode() == 0:
		return nil, status.Error(codes.InvalidArgument, "unit.booking_mode is required")
	case u.GetTimeZone() == "":
		return nil, status.Error(codes.InvalidArgument, "unit.time_zone is required")
	}
	u = proto.Clone(u).(*propertypbv1.Unit)
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
	err := traced(ctx, "CreateUnit", func(ctx context.Context) error {
		created, err := s.repo.CreateUnit(ctx, req.GetParent(), u)
		if err != nil {
			return toStatusErr(err)
		}
		out = created
		return nil
	})
	return out, err
}

// UpdateUnit applies the update mask to an existing unit.
func (s *Server) UpdateUnit(ctx context.Context, req *propertypbv1.UpdateUnitRequest) (*propertypbv1.Unit, error) {
	u := req.GetUnit()
	if u == nil || u.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "unit.name is required")
	}
	var out *propertypbv1.Unit
	err := traced(ctx, "UpdateUnit", func(ctx context.Context) error {
		updated, err := s.repo.UpdateUnit(ctx, u, req.GetUpdateMask().GetPaths())
		if err != nil {
			return toStatusErr(err)
		}
		out = updated
		return nil
	})
	return out, err
}

// DeleteUnit removes a unit by resource name.
func (s *Server) DeleteUnit(ctx context.Context, req *propertypbv1.DeleteUnitRequest) (*emptypb.Empty, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	err := traced(ctx, "DeleteUnit", func(ctx context.Context) error {
		return toStatusErr(s.repo.DeleteUnit(ctx, req.GetName()))
	})
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
