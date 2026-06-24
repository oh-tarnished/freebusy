package gorm

import (
	"fmt"
	"strings"
	"time"

	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/gorm/freebusy/promocode"
	"gorm.io/gorm"
)

// Schema-qualified identifiers for the derived `state` predicate. The promo code
// resource and its window / limits children live in the promocode schema; the
// state filter mirrors discount.EffectiveState in SQL by joining to them.
const (
	promoTable  = `"promocode"."resource"`
	windowTable = `"promocode"."redemption_windows"`
	limitsTable = `"promocode"."usage_limits"`
)

// applyPromoFilter narrows q with the parsed filter conditions (AND-combined).
// Supported fields: code and display_name (`=`, `!=`, `:` substring), disabled
// (`=`/`!=` bool), and the derived state (`=`/`!=` ACTIVE|DISABLED|EXPIRED). A
// bareword term with no field is a case-insensitive search across code and
// display_name. An unknown field or operator yields types.ErrInvalidArgument.
func applyPromoFilter(q *gorm.DB, conds []types.FilterCondition, now time.Time) (*gorm.DB, error) {
	for _, c := range conds {
		clause, args, err := promoCondition(c, now)
		if err != nil {
			return nil, err
		}
		q = q.Where(clause, args...)
	}
	return q, nil
}

func promoCondition(c types.FilterCondition, now time.Time) (string, []any, error) {
	switch c.Field {
	case "":
		// Free-text term: match across the searchable text columns.
		pat := likeContains(c.Value)
		return `(` + promoTable + `."code" ILIKE ? ESCAPE '\' OR ` + promoTable + `."display_name" ILIKE ? ESCAPE '\')`, []any{pat, pat}, nil
	case "code", "display_name":
		return textCondition(c)
	case "disabled":
		return disabledCondition(c)
	case "state":
		return stateCondition(c, now)
	default:
		return "", nil, fmt.Errorf("%w: cannot filter by %q", types.ErrInvalidArgument, c.Field)
	}
}

// textCondition handles code / display_name, which support exact, negated, and
// substring (`:`) matching.
func textCondition(c types.FilterCondition) (string, []any, error) {
	col := promoTable + `."` + c.Field + `"`
	switch c.Op {
	case types.FilterEq:
		return col + " = ?", []any{c.Value}, nil
	case types.FilterNeq:
		return col + " <> ?", []any{c.Value}, nil
	case types.FilterHas:
		return col + " ILIKE ? ESCAPE '\\'", []any{likeContains(c.Value)}, nil
	default:
		return "", nil, fmt.Errorf("%w: unsupported operator for %q", types.ErrInvalidArgument, c.Field)
	}
}

func disabledCondition(c types.FilterCondition) (string, []any, error) {
	if c.Op == types.FilterHas {
		return "", nil, fmt.Errorf("%w: operator `:` is not valid for disabled", types.ErrInvalidArgument)
	}
	var want bool
	switch strings.ToLower(c.Value) {
	case "true":
		want = true
	case "false":
		want = false
	default:
		return "", nil, fmt.Errorf("%w: disabled must be true or false, got %q", types.ErrInvalidArgument, c.Value)
	}
	if c.Op == types.FilterNeq {
		want = !want
	}
	return `COALESCE(` + promoTable + `."disabled", false) = ?`, []any{want}, nil
}

// stateCondition translates a state filter into the same derived predicate as
// discount.EffectiveState: DISABLED when disabled is set; otherwise EXPIRED once
// the redemption window has closed or the redemption cap is reached; otherwise
// ACTIVE.
func stateCondition(c types.FilterCondition, now time.Time) (string, []any, error) {
	if c.Op != types.FilterEq && c.Op != types.FilterNeq {
		return "", nil, fmt.Errorf("%w: operator `:` is not valid for state", types.ErrInvalidArgument)
	}

	disabled := `COALESCE(` + promoTable + `."disabled", false)`
	expired := `(EXISTS (SELECT 1 FROM ` + windowTable + ` w WHERE w."id" = ` + promoTable + `."window_id"` +
		` AND w."end_time" IS NOT NULL AND w."end_time" < ?)` +
		` OR EXISTS (SELECT 1 FROM ` + limitsTable + ` l WHERE l."id" = ` + promoTable + `."limits_id"` +
		` AND l."max_redemptions" IS NOT NULL AND l."max_redemptions" <= COALESCE(` + promoTable + `."redemption_count", 0)))`

	var clause string
	var args []any
	switch strings.ToUpper(strings.TrimPrefix(strings.ToUpper(c.Value), "PROMO_CODE_STATE_")) {
	case string(promocode.PromoCodeStateActive):
		clause = "(NOT " + disabled + " AND NOT " + expired + ")"
		args = []any{now}
	case string(promocode.PromoCodeStateDisabled):
		clause = disabled
	case string(promocode.PromoCodeStateExpired):
		clause = "(NOT " + disabled + " AND " + expired + ")"
		args = []any{now}
	default:
		return "", nil, fmt.Errorf("%w: unknown state %q (want ACTIVE, DISABLED, or EXPIRED)", types.ErrInvalidArgument, c.Value)
	}
	if c.Op == types.FilterNeq {
		clause = "NOT " + clause
	}
	return clause, args, nil
}

// likeContains builds a case-insensitive "contains" pattern, escaping the LIKE
// wildcards in the user value so they match literally.
func likeContains(v string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return "%" + r.Replace(v) + "%"
}
