package types

import (
	"errors"
	"fmt"

	filterx "github.com/oh-tarnished/freebusy/internal/database/gorm/filterx"
	"github.com/oh-tarnished/freebusy/internal/database/repository/repox"
)

// This file is the one boundary between the gRPC-layer vocabulary
// (FilterCondition — parsed early so bad input fails before any repository
// work) and the generated filterx engines the repositories delegate filtering,
// ordering, and pagination to. The shapes mirror each other one-to-one.

// FilterxFromRaw parses a raw AIP-160 expression off a repox.ListInput and
// assembles the generated engines' ListInput — the bridge providers use when
// the gRPC layer passes filters through unparsed. The error carries the
// repox.ErrInvalidArgument sentinel.
func FilterxFromRaw(in repox.ListInput) (filterx.ListInput, error) {
	conds, err := filterx.Parse(in.Filter)
	if err != nil {
		return filterx.ListInput{}, repox.MapFilterxErr(err)
	}
	return filterx.ListInput{
		PageSize:  in.PageSize,
		PageToken: in.PageToken,
		OrderBy:   in.OrderBy,
		Filter:    conds,
	}, nil
}

// Filterx maps parsed filter conditions onto the generated Condition type.
// The operator enums declare in the same order, so the cast is exact.
func Filterx(conds []FilterCondition) []filterx.Condition {
	out := make([]filterx.Condition, len(conds))
	for i, c := range conds {
		out[i] = filterx.Condition{Field: c.Field, Op: filterx.Op(c.Op), Value: c.Value}
	}
	return out
}

// MapFilterxErr rewraps the engines' invalid-input sentinel as
// ErrInvalidArgument so the service layer's status mapping stays unchanged.
func MapFilterxErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, filterx.ErrInvalid) {
		return fmt.Errorf("%w: %s", ErrInvalidArgument, err.Error())
	}
	return err
}
