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

// licence.go implements propertypbv1.LicenceServiceServer on the same Server:
// one CRUD surface for the regulatory licences of a property and its units.
// Every licence is parented by the property; a per-unit licence names its unit
// in the `unit` field, and `target` derives from that.

// ListLicences returns a page of licences under a property — property-wide and
// per-unit ones alike. Filter by target, unit, type, state, or expiry_date
// (`expiry_date <= 2026-08-01`) to find licences due for renewal.
func (s *Server) ListLicences(ctx context.Context, req *propertypbv1.ListLicencesRequest) (*propertypbv1.ListLicencesResponse, error) {
	var out *propertypbv1.ListLicencesResponse
	err := rpc.Traced(ctx, "PropertyService", "ListLicences", func(ctx context.Context) error {
		items, next, err := s.repo.ListLicences(ctx, req.GetParent(), repox.ListInput{
			PageSize:  req.GetPageSize(),
			PageToken: req.GetPageToken(),
			OrderBy:   req.GetOrderBy(),
			Filter:    req.GetFilter(),
		})
		if err != nil {
			return rpc.ToStatusErr(err)
		}
		out = &propertypbv1.ListLicencesResponse{Licences: items, NextPageToken: next}
		return nil
	})
	return out, err
}

// GetLicence returns a single licence by resource name.
func (s *Server) GetLicence(ctx context.Context, req *propertypbv1.GetLicenceRequest) (*propertypbv1.Licence, error) {
	var out *propertypbv1.Licence
	err := rpc.Traced(ctx, "PropertyService", "GetLicence", func(ctx context.Context) error {
		l, err := s.repo.GetLicence(ctx, req.GetName())
		if err != nil {
			return rpc.ToStatusErr(err)
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
	pid, err := types.PropertyID(req.GetParent())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid parent")
	}
	l := proto.Clone(req.GetLicence()).(*propertypbv1.Licence)
	if id := req.GetLicenceId(); id != "" {
		name, err := types.LicenceName(pid, id)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid licence_id")
		}
		l.Name = name
	}
	var out *propertypbv1.Licence
	err = rpc.Traced(ctx, "PropertyService", "CreateLicence", func(ctx context.Context) error {
		created, err := s.repo.CreateLicence(ctx, req.GetParent(), l)
		if err != nil {
			return rpc.ToStatusErr(err)
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
	var out *propertypbv1.Licence
	err := rpc.Traced(ctx, "PropertyService", "UpdateLicence", func(ctx context.Context) error {
		updated, err := s.repo.UpdateLicence(ctx, l, req.GetUpdateMask().GetPaths())
		if err != nil {
			return rpc.ToStatusErr(err)
		}
		out = updated
		return nil
	})
	return out, err
}

// DeleteLicence removes a licence by resource name.
func (s *Server) DeleteLicence(ctx context.Context, req *propertypbv1.DeleteLicenceRequest) (*emptypb.Empty, error) {
	err := rpc.Traced(ctx, "PropertyService", "DeleteLicence", func(ctx context.Context) error {
		return rpc.ToStatusErr(s.repo.DeleteLicence(ctx, req.GetName()))
	})
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
