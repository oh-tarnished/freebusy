package types

import (
	"errors"
	"fmt"

	filterx "github.com/oh-tarnished/freebusy/internal/database/gorm/filterx"
)

// This file is the one boundary between the gRPC-layer vocabulary (ListParams,
// FilterCondition — parsed early so bad input fails before any repository work)
// and the generated filterx engines the repositories delegate filtering,
// ordering, and pagination to. The shapes mirror each other one-to-one.

// FilterxInput maps ListParams onto the generated engines' ListInput.
func FilterxInput(params ListParams) filterx.ListInput {
	return filterx.ListInput{
		PageSize:  params.PageSize,
		PageToken: params.PageToken,
		OrderBy:   params.OrderBy,
		Filter:    Filterx(params.Filter),
	}
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
