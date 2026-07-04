package gorm

import (
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/organisation/v1/orgpbv1"
)

// inMask aliases the shared field-mask predicate.
var inMask = types.InMask

// applyOrgMask merges the masked mutable fields of o onto merged (a proto built
// from the stored record). Identity, state, member_count, timestamps, and etag
// are repository-managed.
func applyOrgMask(merged, o *orgpbv1.Organisation, paths []string) {
	if inMask(paths, "display_name") {
		merged.DisplayName = o.GetDisplayName()
	}
	if inMask(paths, "slug") {
		merged.Slug = o.GetSlug()
	}
	if inMask(paths, "billing_email") {
		merged.BillingEmail = o.GetBillingEmail()
	}
	if inMask(paths, "settings") {
		merged.Settings = o.GetSettings()
	}
}
