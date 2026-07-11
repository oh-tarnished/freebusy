package property

import (
	"context"

	"github.com/oh-tarnished/freebusy/internal/database/repository/repox"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/property/v1/propertypbv1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

// licence.go implements propertypbv1.LicenceServiceServer on the same Server:
// one CRUD surface for the regulatory licences of a property and its units.
// Every licence is parented by the property; a per-unit licence names its unit
// in the `unit` field, and `target` derives from that.

// ListLicences returns a page of licences under a property — property-wide and
// per-unit ones alike. Filter by target, unit, type, state, or expiry_date
// (`expiry_date <= 2026-08-01`) to find licences due for renewal.
func (s *Server) ListLicences(ctx context.Context, req *propertypbv1.ListLicencesRequest) (*propertypbv1.ListLicencesResponse, error) {
	if req.GetParent() == "" {
		return nil, status.Error(codes.InvalidArgument, "parent is required")
	}
	var out *propertypbv1.ListLicencesResponse
	err := traced(ctx, "ListLicences", func(ctx context.Context) error {
		items, next, err := s.repo.ListLicences(ctx, req.GetParent(), repox.ListInput{
			PageSize:  req.GetPageSize(),
			PageToken: req.GetPageToken(),
			OrderBy:   req.GetOrderBy(),
			Filter:    req.GetFilter(),
		})
		if err != nil {
			return toStatusErr(err)
		}
		out = &propertypbv1.ListLicencesResponse{Licences: items, NextPageToken: next}
		return nil
	})
	return out, err
}

// GetLicence returns a single licence by resource name.
func (s *Server) GetLicence(ctx context.Context, req *propertypbv1.GetLicenceRequest) (*propertypbv1.Licence, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	var out *propertypbv1.Licence
	err := traced(ctx, "GetLicence", func(ctx context.Context) error {
		l, err := s.repo.GetLicence(ctx, req.GetName())
		if err != nil {
			return toStatusErr(err)
		}
		out = l
		return nil
	})
	return out, err
}

// CreateLicence creates a licence under a property. Setting licence.unit makes
// it a per-unit licence; the unit must belong to the parent property. A
// caller-supplied licence_id fixes the resource name.
func (s *Server) CreateLicence(ctx context.Context, req *propertypbv1.CreateLicenceRequest) (*propertypbv1.Licence, error) {
	l := req.GetLicence()
	switch {
	case req.GetParent() == "":
		return nil, status.Error(codes.InvalidArgument, "parent is required")
	case l == nil:
		return nil, status.Error(codes.InvalidArgument, "licence is required")
	case l.GetType() == propertypbv1.LicenceType_LICENCE_TYPE_UNSPECIFIED:
		return nil, status.Error(codes.InvalidArgument, "licence.type is required")
	}
	pid, err := types.PropertyID(req.GetParent())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid parent")
	}
	if u := l.GetUnit(); u != "" {
		unitPropertyID, _, err := types.ParseUnitParent(u)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid licence.unit")
		}
		if unitPropertyID != pid {
			return nil, status.Error(codes.InvalidArgument, "licence.unit must belong to the parent property")
		}
	}
	l = proto.Clone(l).(*propertypbv1.Licence)
	if id := req.GetLicenceId(); id != "" {
		name, err := types.LicenceName(pid, id)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid licence_id")
		}
		l.Name = name
	}
	var out *propertypbv1.Licence
	err = traced(ctx, "CreateLicence", func(ctx context.Context) error {
		created, err := s.repo.CreateLicence(ctx, req.GetParent(), l)
		if err != nil {
			return toStatusErr(err)
		}
		out = created
		return nil
	})
	return out, err
}

// UpdateLicence applies the update mask to an existing licence. The target and
// unit are immutable.
func (s *Server) UpdateLicence(ctx context.Context, req *propertypbv1.UpdateLicenceRequest) (*propertypbv1.Licence, error) {
	l := req.GetLicence()
	if l == nil || l.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "licence.name is required")
	}
	var out *propertypbv1.Licence
	err := traced(ctx, "UpdateLicence", func(ctx context.Context) error {
		updated, err := s.repo.UpdateLicence(ctx, l, req.GetUpdateMask().GetPaths())
		if err != nil {
			return toStatusErr(err)
		}
		out = updated
		return nil
	})
	return out, err
}

// DeleteLicence removes a licence by resource name.
func (s *Server) DeleteLicence(ctx context.Context, req *propertypbv1.DeleteLicenceRequest) (*emptypb.Empty, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	err := traced(ctx, "DeleteLicence", func(ctx context.Context) error {
		return toStatusErr(s.repo.DeleteLicence(ctx, req.GetName()))
	})
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
