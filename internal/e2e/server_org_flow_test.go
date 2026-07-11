// Interceptor-rejection and organisation/member flows of the server e2e suite.
package e2e

import (
	"context"
	"fmt"
	"testing"

	"github.com/oh-tarnished/freebusy/protobuf/generated/go/availability/v1/availabilitypbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/identity/v1/identitypbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/organisation/v1/orgpbv1"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// rejectionFlow proves buf.validate rules reach the wire as InvalidArgument
// before any handler runs.
func rejectionFlow(t *testing.T, c *e2eClients) {
	t.Helper()
	ctx := context.Background()

	_, err := c.orgs.CreateOrganisation(ctx, &orgpbv1.CreateOrganisationRequest{})
	wantCode(t, err, codes.InvalidArgument, "create org without body")
	_, err = c.orgs.GetOrganisation(ctx, &orgpbv1.GetOrganisationRequest{Name: "orgs/x"})
	wantCode(t, err, codes.InvalidArgument, "get org with bad name")
	_, err = c.avail.ComputeAvailability(ctx, &availabilitypbv1.ComputeAvailabilityRequest{
		Unit: "properties/p/units/u",
	})
	wantCode(t, err, codes.InvalidArgument, "compute availability without period")
}

// orgFlow walks the organisation + member lifecycle — create with database
// defaults, filtered list, masked update, stale-etag rejection, invite, the
// force-delete guard — and returns the organisation name for the flows built
// on top of it. Cleanup force-deletes the organisation last (LIFO).
func orgFlow(t *testing.T, c *e2eClients) string {
	t.Helper()
	ctx := context.Background()

	org, err := c.orgs.CreateOrganisation(ctx, &orgpbv1.CreateOrganisationRequest{
		Organisation: &orgpbv1.Organisation{DisplayName: "e2e-org-" + c.suffix},
	})
	if err != nil {
		t.Fatalf("CreateOrganisation: %v", err)
	}
	if org.GetState() != orgpbv1.OrganisationState_ORGANISATION_STATE_ACTIVE {
		t.Fatalf("org state = %v, want ACTIVE (database default)", org.GetState())
	}
	t.Cleanup(func() {
		if _, err := c.orgs.DeleteOrganisation(ctx, &orgpbv1.DeleteOrganisationRequest{Name: org.GetName(), Force: true}); err != nil {
			t.Logf("DeleteOrganisation: %v", err)
		}
	})

	page, err := c.orgs.ListOrganisations(ctx, &orgpbv1.ListOrganisationsRequest{
		Filter: fmt.Sprintf("display_name = %q", org.GetDisplayName()),
	})
	if err != nil || len(page.GetOrganisations()) != 1 {
		t.Fatalf("ListOrganisations filter: err=%v n=%d, want 1", err, len(page.GetOrganisations()))
	}

	renamed, err := c.orgs.UpdateOrganisation(ctx, &orgpbv1.UpdateOrganisationRequest{
		Organisation: &orgpbv1.Organisation{Name: org.GetName(), DisplayName: "e2e-org-renamed-" + c.suffix},
		UpdateMask:   &fieldmaskpb.FieldMask{Paths: []string{"display_name"}},
	})
	if err != nil || renamed.GetDisplayName() != "e2e-org-renamed-"+c.suffix {
		t.Fatalf("UpdateOrganisation: %v (display_name %q)", err, renamed.GetDisplayName())
	}
	// The pre-rename etag is now stale: optimistic concurrency must reject it.
	_, err = c.orgs.UpdateOrganisation(ctx, &orgpbv1.UpdateOrganisationRequest{
		Organisation: &orgpbv1.Organisation{Name: org.GetName(), DisplayName: "x", Etag: org.GetEtag()},
		UpdateMask:   &fieldmaskpb.FieldMask{Paths: []string{"display_name"}},
	})
	wantCode(t, err, codes.Aborted, "update org with stale etag")

	invited, err := c.orgs.InviteMember(ctx, &orgpbv1.InviteMemberRequest{
		Parent: org.GetName(),
		Email:  "e2e-" + c.suffix + "@example.com",
		Role:   orgpbv1.OrganisationRole_ORGANISATION_ROLE_ADMIN,
	})
	if err != nil {
		t.Fatalf("InviteMember: %v", err)
	}
	if invited.GetMember().GetState() != orgpbv1.MemberState_MEMBER_STATE_INVITED {
		t.Fatalf("member state = %v, want INVITED (database default)", invited.GetMember().GetState())
	}
	// The force-delete guard: an organisation with members refuses a plain delete.
	_, err = c.orgs.DeleteOrganisation(ctx, &orgpbv1.DeleteOrganisationRequest{Name: org.GetName()})
	wantCode(t, err, codes.Aborted, "delete org with members without force")
	if _, err := c.orgs.DeleteMember(ctx, &orgpbv1.DeleteMemberRequest{Name: invited.GetMember().GetName()}); err != nil {
		t.Fatalf("DeleteMember: %v", err)
	}
	return org.GetName()
}

// identityFlow: users are IdP-provisioned, so the wire surface is list plus
// the authenticated-caller alias.
func identityFlow(t *testing.T, c *e2eClients) {
	t.Helper()
	ctx := context.Background()
	if _, err := c.identity.ListUsers(ctx, &identitypbv1.ListUsersRequest{}); err != nil {
		t.Fatalf("ListUsers: %v", err)
	}
	// "users/me" needs an authenticated caller; the bufconn client sends none.
	_, err := c.identity.GetUser(ctx, &identitypbv1.GetUserRequest{Name: "users/me"})
	wantCode(t, err, codes.Unauthenticated, "get users/me without caller identity")
}
