package hasura

import (
	"strings"

	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/promocode/v1/promocodepbv1"
)

// inMask aliases the shared field-mask predicate so update semantics stay
// identical across the gorm and hasura adapters.
var inMask = types.InMask

// groupTouched reports whether an update mask selects a nested message group,
// matching the group itself ("discount") or any field beneath it
// ("discount.percent_off"). An empty mask selects every group (replace-all).
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

// applyMask merges the masked fields of pc onto merged (a proto built from the
// stored record). Scalars are copied field-by-field; the nested discount, window,
// limits, and scope are replaced wholesale when their group is selected, so the
// repository can re-materialize each child row from the merged proto. Identity,
// timestamps, derived state, and etag are managed by the repository.
func applyMask(merged, pc *promocodepbv1.PromoCode, paths []string) {
	if inMask(paths, "code") {
		merged.Code = pc.GetCode()
	}
	if inMask(paths, "display_name") {
		merged.DisplayName = pc.GetDisplayName()
	}
	if inMask(paths, "description") {
		merged.Description = pc.GetDescription()
	}
	if inMask(paths, "disabled") {
		merged.Disabled = pc.GetDisabled()
	}
	if groupTouched(paths, "discount") {
		merged.Discount = pc.GetDiscount()
	}
	if groupTouched(paths, "window") {
		merged.Window = pc.GetWindow()
	}
	if groupTouched(paths, "limits") {
		merged.Limits = pc.GetLimits()
	}
	if groupTouched(paths, "scope") {
		merged.Scope = pc.GetScope()
	}
}
