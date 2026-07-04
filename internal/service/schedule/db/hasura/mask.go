package hasura

import (
	"strings"

	"github.com/oh-tarnished/freebusy/protobuf/generated/go/schedule/v1/schedulepbv1"
)

// groupTouched reports whether an update mask selects a nested section (the
// section itself or any field beneath it). An empty mask selects everything.
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

// applyScheduleMask merges the masked sections of s onto merged (a proto built
// from the stored record, or empty for a first update). Each section is replaced
// wholesale when selected, so the repository re-materializes its child rows. The
// name, exceptions, and etag are repository-managed.
func applyScheduleMask(merged, s *schedulepbv1.Schedule, paths []string) {
	if groupTouched(paths, "recurring_rules") {
		merged.RecurringRules = s.GetRecurringRules()
	}
	if groupTouched(paths, "buffers") {
		merged.Buffers = s.GetBuffers()
	}
	if groupTouched(paths, "stay_constraints") {
		merged.StayConstraints = s.GetStayConstraints()
	}
	if groupTouched(paths, "cancellation_policy") {
		merged.CancellationPolicy = s.GetCancellationPolicy()
	}
}
