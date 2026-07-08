package gorm

import (
	"fmt"
	"strings"
	"time"

	filterx "github.com/oh-tarnished/freebusy/internal/database/gorm/filterx"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/promocode"
)

// Everything except `state` is handled by the generated PromoCodeFilterSpec +
// filterx.Gorm engine. State is not a stored column — it is derived from
// disabled / the redemption window / the usage cap — so it stays a hand-written
// override registered with the engine (and the Hasura provider, which gets no
// override, keeps rejecting it).

// Schema-qualified identifiers for the derived `state` predicate. The promo code
// resource and its window / limits children live in the promocode schema; the
// state filter mirrors discount.EffectiveState in SQL by joining to them.
const (
	promoTable  = `"promocode"."resource"`
	windowTable = `"promocode"."redemption_windows"`
	limitsTable = `"promocode"."usage_limits"`
)

// stateHandler builds the engine override translating a state filter into the
// same derived predicate as discount.EffectiveState: DISABLED when disabled is
// set; otherwise EXPIRED once the redemption window has closed or the
// redemption cap is reached; otherwise ACTIVE.
func stateHandler(now time.Time) filterx.SQLHandler {
	return func(c filterx.Condition) (string, []any, error) {
		if c.Op != filterx.OpEq && c.Op != filterx.OpNeq {
			return "", nil, fmt.Errorf("%w: operator `:` is not valid for state", filterx.ErrInvalid)
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
			return "", nil, fmt.Errorf("%w: unknown state %q (want ACTIVE, DISABLED, or EXPIRED)", filterx.ErrInvalid, c.Value)
		}
		if c.Op == filterx.OpNeq {
			clause = "NOT " + clause
		}
		return clause, args, nil
	}
}
