package hasura

import (
	"strings"

	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/property/v1/propertypbv1"
)

var inMask = types.InMask

// groupTouched reports whether an update mask selects a nested message group
// (the group itself or any field beneath it). An empty mask selects everything.
func groupTouched(paths []string, group string) bool {
	if len(paths) == 0 {
		return true
	}
	prefix := group + "."
	for _, p := range paths {
		if p == group || strings.HasPrefix(p, prefix) {
			return true
		}
	}
	return false
}

// applyPropertyMask merges the masked fields of p onto merged (a proto built from
// the stored record); the repository re-materializes the child graph from the
// result. Identity, timestamps, state, units, and etag are repository-managed.
func applyPropertyMask(merged, p *propertypbv1.Property, paths []string) {
	if inMask(paths, "organisation") {
		merged.Organisation = p.GetOrganisation()
	}
	if inMask(paths, "display_name") {
		merged.DisplayName = p.GetDisplayName()
	}
	if inMask(paths, "description") {
		merged.Description = p.GetDescription()
	}
	if inMask(paths, "time_zone") {
		merged.TimeZone = p.GetTimeZone()
	}
	if inMask(paths, "tags") {
		merged.Tags = p.GetTags()
	}
	if inMask(paths, "attributes") {
		merged.Attributes = p.GetAttributes()
	}
	if groupTouched(paths, "address") {
		merged.Address = p.GetAddress()
	}
	if groupTouched(paths, "policy") {
		merged.Policy = p.GetPolicy()
	}
	if groupTouched(paths, "media") {
		merged.Media = p.GetMedia()
	}
}

// applyUnitMask merges the masked fields of u onto merged. booking_mode is
// immutable and never merged; the pricing groups, media, and
// applicable_promo_codes are replaced wholesale when selected.
func applyUnitMask(merged, u *propertypbv1.Unit, paths []string) {
	if inMask(paths, "display_name") {
		merged.DisplayName = u.GetDisplayName()
	}
	if inMask(paths, "description") {
		merged.Description = u.GetDescription()
	}
	if inMask(paths, "type") {
		merged.Type = u.GetType()
	}
	if inMask(paths, "capacity") {
		merged.Capacity = u.GetCapacity()
	}
	if inMask(paths, "max_occupancy") {
		merged.MaxOccupancy = u.GetMaxOccupancy()
	}
	if inMask(paths, "time_zone") {
		merged.TimeZone = u.GetTimeZone()
	}
	if inMask(paths, "pricing_unit") {
		merged.PricingUnit = u.GetPricingUnit()
	}
	if inMask(paths, "duration") {
		merged.Duration = u.GetDuration()
	}
	if inMask(paths, "tags") {
		merged.Tags = u.GetTags()
	}
	if inMask(paths, "attributes") {
		merged.Attributes = u.GetAttributes()
	}
	if groupTouched(paths, "price") {
		merged.Price = u.GetPrice()
	}
	if groupTouched(paths, "rate_overrides") {
		merged.RateOverrides = u.GetRateOverrides()
	}
	if groupTouched(paths, "los_discounts") {
		merged.LosDiscounts = u.GetLosDiscounts()
	}
	if groupTouched(paths, "fees") {
		merged.Fees = u.GetFees()
	}
	if groupTouched(paths, "taxes") {
		merged.Taxes = u.GetTaxes()
	}
	if groupTouched(paths, "media") {
		merged.Media = u.GetMedia()
	}
	if groupTouched(paths, "applicable_promo_codes") {
		merged.ApplicablePromoCodes = u.GetApplicablePromoCodes()
	}
}

// applyLicenceMask merges the masked fields of l onto merged. The attachment
// is replaced wholesale when selected. Identity, target, unit, timestamps,
// state, and etag are repository-managed (target and unit are immutable).
func applyLicenceMask(merged, l *propertypbv1.Licence, paths []string) {
	if inMask(paths, "type") {
		merged.Type = l.GetType()
	}
	if inMask(paths, "licence_number") {
		merged.LicenceNumber = l.GetLicenceNumber()
	}
	if inMask(paths, "issuing_authority") {
		merged.IssuingAuthority = l.GetIssuingAuthority()
	}
	if inMask(paths, "issue_date") {
		merged.IssueDate = l.GetIssueDate()
	}
	if inMask(paths, "expiry_date") {
		merged.ExpiryDate = l.GetExpiryDate()
	}
	if inMask(paths, "notes") {
		merged.Notes = l.GetNotes()
	}
	if groupTouched(paths, "attachment") {
		merged.Attachment = l.GetAttachment()
	}
}
