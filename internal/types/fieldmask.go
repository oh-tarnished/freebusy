package types

// InMask reports whether field should be written for an Update: true when paths
// is empty (an unset field mask means full replace) or when paths explicitly
// names field. Provider adapters share this so update semantics can't drift
// between them.
func InMask(paths []string, field string) bool {
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
