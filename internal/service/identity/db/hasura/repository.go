package hasura

import (
	"context"
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql"
	usersql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/identityql/usersql"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/identity/v1/identitypbv1"
	"github.com/oh-tarnished/generateql/runtime/go/graphql"
	"github.com/oh-tarnished/runtime-go/ulid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// UserRepository is the Hasura-backed user repository.
type UserRepository struct {
	svc *freebusyql.Service
}

// NewUserRepository returns a Hasura-backed UserRepository bound to svc.
func NewUserRepository(svc *freebusyql.Service) *UserRepository {
	return &UserRepository{svc: svc}
}

// GetUser returns the user addressed by its resource name.
func (r *UserRepository) GetUser(ctx context.Context, name string) (*identitypbv1.User, error) {
	id, err := types.UserID(name)
	if err != nil {
		return nil, err
	}
	res, err := r.svc.Query.Identity.Users.Get(ctx, id)
	if err != nil {
		return nil, mapHasuraErr(err)
	}
	if res == nil {
		return nil, types.ErrNotFound
	}
	return userFromSchema(res), nil
}

// ListUsers returns a page of users ordered by params.OrderBy.
func (r *UserRepository) ListUsers(ctx context.Context, params types.ListParams) ([]*identitypbv1.User, string, error) {
	order, err := userOrderTerms(params.OrderBy)
	if err != nil {
		return nil, "", err
	}
	where, hasWhere, err := userFilterPredicate(params.Filter)
	if err != nil {
		return nil, "", err
	}
	limit, offset := types.PageBounds(params)
	req := usersql.List().Limit(limit + 1).Offset(offset)
	if len(order) > 0 {
		req = req.OrderBy(order...)
	}
	if hasWhere {
		req = req.Where(where)
	}
	rows, err := r.svc.Query.Identity.Users.List(ctx, req)
	if err != nil {
		return nil, "", mapHasuraErr(err)
	}
	next := ""
	if len(rows) > limit {
		rows = rows[:limit]
		next = types.EncodeOffset(offset + limit)
	}
	items := make([]*identitypbv1.User, 0, len(rows))
	for i := range rows {
		items = append(items, userFromSchema(&rows[i]))
	}
	return items, next, nil
}

// UpdateUser applies the masked profile fields of u and returns the result. Email
// and identity are IdP-owned and never written.
func (r *UserRepository) UpdateUser(ctx context.Context, u *identitypbv1.User, paths []string) (*identitypbv1.User, error) {
	id, err := types.UserID(u.GetName())
	if err != nil {
		return nil, err
	}
	res, err := r.svc.Query.Identity.Users.Get(ctx, id)
	if err != nil {
		return nil, mapHasuraErr(err)
	}
	if res == nil {
		return nil, types.ErrNotFound
	}
	if u.GetEtag() != "" && res.Etag != nil && u.GetEtag() != *res.Etag {
		return nil, types.ErrConflict
	}

	patch := usersql.UpdateInput{
		Etag:       graphql.Value(ulid.GenerateString()),
		UpdateTime: graphql.Value(tsToStr(timestamppb.New(time.Now().UTC()))),
	}
	if fieldSelected(paths, "display_name") {
		patch.DisplayName = nullableStr(u.GetDisplayName())
	}
	if fieldSelected(paths, "avatar_url") {
		patch.AvatarUrl = nullableStr(u.GetAvatarUrl())
	}
	if fieldSelected(paths, "locale") {
		patch.Locale = nullableStr(u.GetLocale())
	}
	if fieldSelected(paths, "time_zone") {
		patch.TimeZone = nullableStr(u.GetTimeZone())
	}
	if _, err := r.svc.Mutation.Identity.Users.Update(ctx, id, patch); err != nil {
		return nil, mapHasuraErr(err)
	}
	return r.GetUser(ctx, u.GetName())
}

// fieldSelected reports whether an update mask selects field. An empty mask
// selects every field.
func fieldSelected(paths []string, field string) bool {
	if len(paths) == 0 {
		return true
	}
	for _, p := range paths {
		if p == field {
			return true
		}
	}
	return false
}
