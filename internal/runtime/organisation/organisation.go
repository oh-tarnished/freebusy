// Package organisation is the gRPC/protobuf layer for the OrganisationService:
// it implements orgpbv1.OrganisationServiceServer, owning request validation,
// observability, and the mapping of repository errors to gRPC status codes.
// Persistence stays behind the provider-agnostic db.OrganisationRepository.
package organisation

import (
	"context"

	organisationdb "github.com/oh-tarnished/freebusy/internal/service/organisation/db"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/organisation/v1/orgpbv1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Server implements orgpbv1.OrganisationServiceServer on top of a
// provider-agnostic db.OrganisationRepository.
type Server struct {
	orgpbv1.UnimplementedOrganisationServiceServer
	repo organisationdb.OrganisationRepository
}

// NewServer returns a Server backed by repo.
func NewServer(repo organisationdb.OrganisationRepository) *Server {
	return &Server{repo: repo}
}

// --- Organisation ------------------------------------------------------------

func (s *Server) ListOrganisations(ctx context.Context, req *orgpbv1.ListOrganisationsRequest) (*orgpbv1.ListOrganisationsResponse, error) {
	filter, err := types.ParseFilter(req.GetFilter())
	if err != nil {
		return nil, toStatusErr(err)
	}
	var out *orgpbv1.ListOrganisationsResponse
	err = traced(ctx, "ListOrganisations", func(ctx context.Context) error {
		items, next, err := s.repo.ListOrganisations(ctx, types.ListParams{
			PageSize:  req.GetPageSize(),
			PageToken: req.GetPageToken(),
			OrderBy:   req.GetOrderBy(),
			Filter:    filter,
		})
		if err != nil {
			return toStatusErr(err)
		}
		out = &orgpbv1.ListOrganisationsResponse{Organisations: items, NextPageToken: next}
		return nil
	})
	return out, err
}

func (s *Server) GetOrganisation(ctx context.Context, req *orgpbv1.GetOrganisationRequest) (*orgpbv1.Organisation, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	var out *orgpbv1.Organisation
	err := traced(ctx, "GetOrganisation", func(ctx context.Context) error {
		o, err := s.repo.GetOrganisation(ctx, req.GetName())
		if err != nil {
			return toStatusErr(err)
		}
		out = o
		return nil
	})
	return out, err
}

func (s *Server) CreateOrganisation(ctx context.Context, req *orgpbv1.CreateOrganisationRequest) (*orgpbv1.Organisation, error) {
	o := req.GetOrganisation()
	switch {
	case o == nil:
		return nil, status.Error(codes.InvalidArgument, "organisation is required")
	case o.GetDisplayName() == "":
		return nil, status.Error(codes.InvalidArgument, "organisation.display_name is required")
	}
	o = proto.Clone(o).(*orgpbv1.Organisation)
	if id := req.GetOrganisationId(); id != "" {
		name, err := types.OrganisationName(id)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid organisation_id")
		}
		o.Name = name
	}
	var out *orgpbv1.Organisation
	err := traced(ctx, "CreateOrganisation", func(ctx context.Context) error {
		created, err := s.repo.CreateOrganisation(ctx, o)
		if err != nil {
			return toStatusErr(err)
		}
		out = created
		return nil
	})
	return out, err
}

func (s *Server) UpdateOrganisation(ctx context.Context, req *orgpbv1.UpdateOrganisationRequest) (*orgpbv1.Organisation, error) {
	o := req.GetOrganisation()
	if o == nil || o.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "organisation.name is required")
	}
	var out *orgpbv1.Organisation
	err := traced(ctx, "UpdateOrganisation", func(ctx context.Context) error {
		updated, err := s.repo.UpdateOrganisation(ctx, o, req.GetUpdateMask().GetPaths())
		if err != nil {
			return toStatusErr(err)
		}
		out = updated
		return nil
	})
	return out, err
}

func (s *Server) DeleteOrganisation(ctx context.Context, req *orgpbv1.DeleteOrganisationRequest) (*emptypb.Empty, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	err := traced(ctx, "DeleteOrganisation", func(ctx context.Context) error {
		return toStatusErr(s.repo.DeleteOrganisation(ctx, req.GetName(), req.GetForce()))
	})
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// --- Member ------------------------------------------------------------------

func (s *Server) InviteMember(ctx context.Context, req *orgpbv1.InviteMemberRequest) (*orgpbv1.InviteMemberResponse, error) {
	switch {
	case req.GetParent() == "":
		return nil, status.Error(codes.InvalidArgument, "parent is required")
	case req.GetEmail() == "":
		return nil, status.Error(codes.InvalidArgument, "email is required")
	case req.GetRole() == orgpbv1.OrganisationRole_ORGANISATION_ROLE_UNSPECIFIED:
		return nil, status.Error(codes.InvalidArgument, "role is required")
	}
	member := &orgpbv1.Member{Email: req.GetEmail(), Role: req.GetRole()}
	var out *orgpbv1.InviteMemberResponse
	err := traced(ctx, "InviteMember", func(ctx context.Context) error {
		created, err := s.repo.CreateMember(ctx, req.GetParent(), member)
		if err != nil {
			return toStatusErr(err)
		}
		out = &orgpbv1.InviteMemberResponse{Member: created}
		return nil
	})
	return out, err
}

func (s *Server) ListMembers(ctx context.Context, req *orgpbv1.ListMembersRequest) (*orgpbv1.ListMembersResponse, error) {
	if req.GetParent() == "" {
		return nil, status.Error(codes.InvalidArgument, "parent is required")
	}
	filter, err := types.ParseFilter(req.GetFilter())
	if err != nil {
		return nil, toStatusErr(err)
	}
	var out *orgpbv1.ListMembersResponse
	err = traced(ctx, "ListMembers", func(ctx context.Context) error {
		items, next, err := s.repo.ListMembers(ctx, req.GetParent(), types.ListParams{
			PageSize:  req.GetPageSize(),
			PageToken: req.GetPageToken(),
			OrderBy:   req.GetOrderBy(),
			Filter:    filter,
		})
		if err != nil {
			return toStatusErr(err)
		}
		out = &orgpbv1.ListMembersResponse{Members: items, NextPageToken: next}
		return nil
	})
	return out, err
}

func (s *Server) GetMember(ctx context.Context, req *orgpbv1.GetMemberRequest) (*orgpbv1.Member, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	var out *orgpbv1.Member
	err := traced(ctx, "GetMember", func(ctx context.Context) error {
		m, err := s.repo.GetMember(ctx, req.GetName())
		if err != nil {
			return toStatusErr(err)
		}
		out = m
		return nil
	})
	return out, err
}

func (s *Server) UpdateMember(ctx context.Context, req *orgpbv1.UpdateMemberRequest) (*orgpbv1.Member, error) {
	m := req.GetMember()
	if m == nil || m.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "member.name is required")
	}
	var out *orgpbv1.Member
	err := traced(ctx, "UpdateMember", func(ctx context.Context) error {
		updated, err := s.repo.UpdateMember(ctx, m, req.GetUpdateMask().GetPaths())
		if err != nil {
			return toStatusErr(err)
		}
		out = updated
		return nil
	})
	return out, err
}

func (s *Server) DeleteMember(ctx context.Context, req *orgpbv1.DeleteMemberRequest) (*emptypb.Empty, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	err := traced(ctx, "DeleteMember", func(ctx context.Context) error {
		return toStatusErr(s.repo.DeleteMember(ctx, req.GetName()))
	})
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
