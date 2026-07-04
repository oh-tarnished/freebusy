package hasura

import (
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/organisation/v1/orgpbv1"
)

// applyOrgMask merges the masked mutable fields of o onto merged (a proto built
// from the stored record). Identity, state, member_count, timestamps, and etag
// are repository-managed.
func applyOrgMask(merged, o *orgpbv1.Organisation, paths []string) {
	if types.InMask(paths, "display_name") {
		merged.DisplayName = o.GetDisplayName()
	}
	if types.InMask(paths, "slug") {
		merged.Slug = o.GetSlug()
	}
	if types.InMask(paths, "billing_email") {
		merged.BillingEmail = o.GetBillingEmail()
	}
	if types.InMask(paths, "settings") {
		merged.Settings = o.GetSettings()
	}
}
