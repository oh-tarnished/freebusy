package identity

import (
	"context"
	"github.com/oh-tarnished/freebusy/internal/database"

	"github.com/oh-tarnished/freebusy/internal/database/repository/repox"
	"github.com/oh-tarnished/freebusy/internal/runtime/rpc"
	"github.com/oh-tarnished/freebusy/internal/service/identity/db"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/identity/v1/identitypbv1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// meAlias is the reserved name that resolves to the signed-in caller.
const meAlias = "users/me"

// callerHeader is the metadata key an upstream OIDC gateway sets to the
// authenticated user's bare id, after it has validated the token.
const callerHeader = "x-user-id"

// Server implements identitypbv1.IdentityServiceServer on top of a
// provider-agnostic db.UserRepository.
type Server struct {
	identitypbv1.UnimplementedIdentityServiceServer
	repo db.UserRepository
}

// New builds the identity service on conn: the provider-selected repository
// wrapped in the gRPC server implementation.
func New(conn *database.Connection) *Server {
	return NewServer(db.New(conn))
}

// NewServer returns a Server backed by repo.
func NewServer(repo db.UserRepository) *Server {
	return &Server{repo: repo}
}

// resolveName resolves the "users/me" alias to the caller's resource name (from
// the authenticated context); any other name passes through unchanged.
func resolveName(ctx context.Context, name string) (string, error) {
	if name != meAlias {
		return name, nil
	}
	md, _ := metadata.FromIncomingContext(ctx)
	if vals := md.Get(callerHeader); len(vals) > 0 && vals[0] != "" {
		return types.UserName(vals[0])
	}
	return "", status.Error(codes.Unauthenticated, `"users/me" requires an authenticated caller`)
}

// GetUser returns a user by resource name; "users/me" resolves to the caller.
func (s *Server) GetUser(ctx context.Context, req *identitypbv1.GetUserRequest) (*identitypbv1.User, error) {
	var out *identitypbv1.User
	err := rpc.Traced(ctx, "IdentityService", "GetUser", func(ctx context.Context) error {
		name, err := resolveName(ctx, req.GetName())
		if err != nil {
			return err
		}
		u, err := s.repo.GetUser(ctx, name)
		if err != nil {
			return rpc.ToStatusErr(err)
		}
		out = u
		return nil
	})
	return out, err
}

// ListUsers returns a page of users.
func (s *Server) ListUsers(ctx context.Context, req *identitypbv1.ListUsersRequest) (*identitypbv1.ListUsersResponse, error) {
	var out *identitypbv1.ListUsersResponse
	err := rpc.Traced(ctx, "IdentityService", "ListUsers", func(ctx context.Context) error {
		items, next, err := s.repo.ListUsers(ctx, repox.ListInput{
			PageSize:  req.GetPageSize(),
			PageToken: req.GetPageToken(),
			OrderBy:   req.GetOrderBy(),
			Filter:    req.GetFilter(),
		})
		if err != nil {
			return rpc.ToStatusErr(err)
		}
		out = &identitypbv1.ListUsersResponse{Users: items, NextPageToken: next}
		return nil
	})
	return out, err
}

// UpdateUser updates the signed-in user's profile; "users/me" resolves to the caller.
func (s *Server) UpdateUser(ctx context.Context, req *identitypbv1.UpdateUserRequest) (*identitypbv1.User, error) {
	u := req.GetUser()
	var out *identitypbv1.User
	err := rpc.Traced(ctx, "IdentityService", "UpdateUser", func(ctx context.Context) error {
		name, err := resolveName(ctx, u.GetName())
		if err != nil {
			return err
		}
		u.Name = name
		updated, err := s.repo.UpdateUser(ctx, u, req.GetUpdateMask().GetPaths())
		if err != nil {
			return rpc.ToStatusErr(err)
		}
		out = updated
		return nil
	})
	return out, err
}
