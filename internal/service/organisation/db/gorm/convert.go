package gorm

import (
	"strings"
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/organisation"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/organisation/v1/orgpbv1"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// This file holds the pure conversions between the protobuf Organisation/Member
// domain types and their GORM storage models. Both are flat rows (no nested
// value-objects); only the settings jsonb and the enums need special handling.

func ptr[T any](v T) *T { return &v }

func deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

func strOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func timeToTS(t *time.Time) *timestamppb.Timestamp {
	if t == nil || t.IsZero() {
		return nil
	}
	return timestamppb.New(*t)
}

func structToJSON(s *structpb.Struct) []byte {
	if s == nil {
		return nil
	}
	b, err := s.MarshalJSON()
	if err != nil {
		return nil
	}
	return b
}

func jsonToStruct(b []byte) *structpb.Struct {
	if len(b) == 0 {
		return nil
	}
	s := &structpb.Struct{}
	if err := s.UnmarshalJSON(b); err != nil {
		return nil
	}
	return s
}

// userName rebuilds a "users/{id}" resource name from the bare id the member's
// user / inviter FK columns store (both reference identity users.id).
func userName(id *string) string {
	if id == nil || *id == "" {
		return ""
	}
	return "users/" + *id
}

// --- Organisation ------------------------------------------------------------

// orgToModel builds a storage row from a proto Organisation for a create; state
// defaults to ACTIVE and member_count is server-derived (left unset).
func orgToModel(o *orgpbv1.Organisation) *organisation.Organisation {
	state := organisation.OrganisationStateActive
	return &organisation.Organisation{
		DisplayName:  o.GetDisplayName(),
		Slug:         strOrNil(o.GetSlug()),
		BillingEmail: strOrNil(o.GetBillingEmail()),
		State:        &state,
		Settings:     structToJSON(o.GetSettings()),
	}
}

func orgFromModel(m *organisation.Organisation) *orgpbv1.Organisation {
	return &orgpbv1.Organisation{
		Name:         m.Name,
		DisplayName:  m.DisplayName,
		Slug:         deref(m.Slug),
		BillingEmail: deref(m.BillingEmail),
		State:        orgStateFromModel(m.State),
		Settings:     jsonToStruct(m.Settings),
		MemberCount:  deref(m.MemberCount),
		CreateTime:   timeToTS(&m.CreateTime),
		UpdateTime:   timeToTS(&m.UpdateTime),
		Etag:         deref(m.Etag),
	}
}

func orgStateFromModel(s *organisation.OrganisationState) orgpbv1.OrganisationState {
	if s == nil {
		return orgpbv1.OrganisationState_ORGANISATION_STATE_UNSPECIFIED
	}
	return orgpbv1.OrganisationState(orgpbv1.OrganisationState_value["ORGANISATION_STATE_"+string(*s)])
}

// --- Member ------------------------------------------------------------------

// memberToModel builds a storage row from a proto Member for an invite; state
// defaults to INVITED and the user/inviter are server-set (left unset).
func memberToModel(mem *orgpbv1.Member) *organisation.Member {
	state := organisation.MemberStateInvited
	return &organisation.Member{
		Email: mem.GetEmail(),
		Role:  roleToModel(mem.GetRole()),
		State: &state,
	}
}

func memberFromModel(m *organisation.Member) *orgpbv1.Member {
	return &orgpbv1.Member{
		Name:        m.Name,
		User:        userName(m.UserID),
		Email:       m.Email,
		DisplayName: deref(m.DisplayName),
		Role:        roleFromModel(m.Role),
		State:       memberStateFromModel(m.State),
		Inviter:     userName(m.InviterID),
		CreateTime:  timeToTS(&m.CreateTime),
		UpdateTime:  timeToTS(&m.UpdateTime),
		Etag:        deref(m.Etag),
	}
}

func roleToModel(r orgpbv1.OrganisationRole) organisation.OrganisationRole {
	return organisation.OrganisationRole(strings.TrimPrefix(r.String(), "ORGANISATION_ROLE_"))
}

func roleFromModel(r organisation.OrganisationRole) orgpbv1.OrganisationRole {
	return orgpbv1.OrganisationRole(orgpbv1.OrganisationRole_value["ORGANISATION_ROLE_"+string(r)])
}

func memberStateFromModel(s *organisation.MemberState) orgpbv1.MemberState {
	if s == nil {
		return orgpbv1.MemberState_MEMBER_STATE_UNSPECIFIED
	}
	return orgpbv1.MemberState(orgpbv1.MemberState_value["MEMBER_STATE_"+string(*s)])
}
