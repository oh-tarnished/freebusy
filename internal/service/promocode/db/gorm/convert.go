package gorm

import (
	"strings"
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/common"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/promocode"
	"github.com/oh-tarnished/freebusy/internal/service/promocode/discount"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/promocode/v1/promocodepbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
	"google.golang.org/genproto/googleapis/type/money"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// This file holds the pure (side-effect-free) conversions between the protobuf
// PromoCode and the normalized GORM storage models. The protobuf API nests the
// discount, redemption window, usage limits, and scope as sub-messages; the
// schema stores each as its own belongs-to child table (promocode.discounts,
// promocode.redemption_windows, promocode.usage_limits, promocode.scopes) with a
// foreign key on promocode.resource. Money value-objects normalize into the
// shared common.moneys table, and a scope's applicable resources / offerings
// become join rows. The join columns store the full API resource name verbatim
// so the list values round-trip exactly.

func ptr[T any](v T) *T { return &v }

func deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

// strOrNil maps an empty proto string (which cannot represent NULL) to a nil
// column pointer, so unset optional strings stay NULL in the database.
func strOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func tsToTime(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime()
	return &t
}

func timeToTS(t *time.Time) *timestamppb.Timestamp {
	if t == nil || t.IsZero() {
		return nil
	}
	return timestamppb.New(*t)
}

// lastSegment returns the final path component of an AIP resource name
// ("resources/r1/offerings/o1" -> "o1"), used to populate the join row's id
// column while the full name round-trips via a separate column.
func lastSegment(name string) string {
	if i := strings.LastIndex(name, "/"); i >= 0 {
		return name[i+1:]
	}
	return name
}

// moneyToModel builds a new Money row (with a fresh ULID id) from a proto Money,
// or returns nil when m is nil.
func moneyToModel(m *money.Money) *common.Money {
	if m == nil {
		return nil
	}
	return &common.Money{
		ID:           ulid.GenerateString(),
		CurrencyCode: strOrNil(m.GetCurrencyCode()),
		Units:        ptr(m.GetUnits()),
		Nanos:        ptr(m.GetNanos()),
	}
}

func moneyFromModel(m *common.Money) *money.Money {
	if m == nil {
		return nil
	}
	return &money.Money{
		CurrencyCode: deref(m.CurrencyCode),
		Units:        deref(m.Units),
		Nanos:        deref(m.Nanos),
	}
}

func int64Wrapper(w *wrapperspb.Int64Value) *int64 {
	if w == nil {
		return nil
	}
	return ptr(w.GetValue())
}

func int32Wrapper(w *wrapperspb.Int32Value) *int32 {
	if w == nil {
		return nil
	}
	return ptr(w.GetValue())
}

// promoGraph is the full set of rows a single PromoCode materializes into: the
// resource row plus its belongs-to children, the Money rows they reference, and
// the scope's applicable-resources / -offerings join rows. The repository
// persists it in one transaction (moneys and children before the resource that
// references them; join rows after the scope they belong to).
type promoGraph struct {
	promo     *promocode.PromoCode
	discount  *promocode.Discount
	window    *promocode.RedemptionWindow
	limits    *promocode.UsageLimits
	scope     *promocode.Scope
	moneys     []*common.Money
	properties []*promocode.ScopeApplicableProperties
	units      []*promocode.ScopeApplicableUnits
}

// buildGraph turns a proto PromoCode into the row graph that backs it, minting a
// fresh ULID for every row and wiring the foreign keys. Identity (ID/Name/Etag)
// of the resource row is left to the caller, which owns id assignment and the
// transaction.
func buildGraph(pc *promocodepbv1.PromoCode) *promoGraph {
	g := &promoGraph{}

	state := promocode.PromoCodeStateActive
	if pc.GetDisabled() {
		state = promocode.PromoCodeStateDisabled
	}
	g.promo = &promocode.PromoCode{
		Code:        pc.GetCode(),
		DisplayName: strOrNil(pc.GetDisplayName()),
		Description: strOrNil(pc.GetDescription()),
		State:       &state,
		Disabled:    ptr(pc.GetDisabled()),
	}

	// Discount is required: always materialize a row. A non-nil amount_off is the
	// fixed-amount arm; otherwise it's the percentage arm.
	g.discount = &promocode.Discount{ID: ulid.GenerateString()}
	if amt := moneyToModel(pc.GetDiscount().GetAmountOff()); amt != nil {
		g.moneys = append(g.moneys, amt)
		g.discount.AmountOffID = &amt.ID
		g.discount.AmountCase = ptr(promocode.DiscountAmountCaseAmountOff)
	} else {
		g.discount.PercentOff = ptr(pc.GetDiscount().GetPercentOff())
		g.discount.AmountCase = ptr(promocode.DiscountAmountCasePercentOff)
	}
	g.promo.DiscountID = g.discount.ID

	if w := pc.GetWindow(); w != nil {
		g.window = &promocode.RedemptionWindow{
			ID:        ulid.GenerateString(),
			StartTime: tsToTime(w.GetStartTime()),
			EndTime:   tsToTime(w.GetEndTime()),
		}
		g.promo.WindowID = &g.window.ID
	}

	if l := pc.GetLimits(); l != nil {
		g.limits = &promocode.UsageLimits{
			ID:               ulid.GenerateString(),
			MaxRedemptions:   int64Wrapper(l.GetMaxRedemptions()),
			PerCustomerLimit: int32Wrapper(l.GetPerCustomerLimit()),
		}
		g.promo.LimitsID = &g.limits.ID
	}

	if sc := pc.GetScope(); sc != nil {
		g.scope = &promocode.Scope{ID: ulid.GenerateString()}
		if min := moneyToModel(sc.GetMinSubtotal()); min != nil {
			g.moneys = append(g.moneys, min)
			g.scope.MinSubtotalID = &min.ID
		}
		for _, name := range sc.GetApplicableProperties() {
			g.properties = append(g.properties, &promocode.ScopeApplicableProperties{
				ID:         ulid.GenerateString(),
				ScopeID:    g.scope.ID,
				PropertyID: name,
			})
		}
		for _, name := range sc.GetApplicableUnits() {
			g.units = append(g.units, &promocode.ScopeApplicableUnits{
				ID:       ulid.GenerateString(),
				ScopeID:  g.scope.ID,
				UnitID:   lastSegment(name),
				UnitName: name,
			})
		}
		g.promo.ScopeID = &g.scope.ID
	}

	return g
}

// fromModel assembles the protobuf PromoCode from a stored resource row and its
// preloaded associations (discount + amount money, window, limits, scope + min
// money + applicable join rows).
func fromModel(m *promocode.PromoCode) *promocodepbv1.PromoCode {
	pc := &promocodepbv1.PromoCode{
		Name:            m.Name,
		Code:            m.Code,
		DisplayName:     deref(m.DisplayName),
		Description:     deref(m.Description),
		Discount:        discountFromModel(m.Discount),
		Window:          windowFromModel(m.Window),
		Limits:          limitsFromModel(m.Limits),
		Scope:           scopeFromModel(m.Scope),
		RedemptionCount: deref(m.RedemptionCount),
		Disabled:        deref(m.Disabled),
		CreateTime:      timeToTS(&m.CreateTime),
		UpdateTime:      timeToTS(&m.UpdateTime),
		Etag:            deref(m.Etag),
	}
	// Derive the lifecycle state from the window/flags rather than trusting the
	// possibly-stale stored value (a code becomes EXPIRED purely with time).
	pc.State = discount.EffectiveState(pc, time.Now().UTC())
	return pc
}

func discountFromModel(d *promocode.Discount) *promocodepbv1.Discount {
	if d == nil {
		return nil
	}
	out := &promocodepbv1.Discount{}
	if d.AmountCase != nil && *d.AmountCase == promocode.DiscountAmountCaseAmountOff {
		out.Amount = &promocodepbv1.Discount_AmountOff{AmountOff: moneyFromModel(d.AmountOff)}
	} else {
		out.Amount = &promocodepbv1.Discount_PercentOff{PercentOff: deref(d.PercentOff)}
	}
	return out
}

func windowFromModel(w *promocode.RedemptionWindow) *promocodepbv1.RedemptionWindow {
	if w == nil {
		return nil
	}
	return &promocodepbv1.RedemptionWindow{
		StartTime: timeToTS(w.StartTime),
		EndTime:   timeToTS(w.EndTime),
	}
}

func limitsFromModel(l *promocode.UsageLimits) *promocodepbv1.UsageLimits {
	if l == nil {
		return nil
	}
	out := &promocodepbv1.UsageLimits{}
	if l.MaxRedemptions != nil {
		out.MaxRedemptions = wrapperspb.Int64(*l.MaxRedemptions)
	}
	if l.PerCustomerLimit != nil {
		out.PerCustomerLimit = wrapperspb.Int32(*l.PerCustomerLimit)
	}
	return out
}

func scopeFromModel(s *promocode.Scope) *promocodepbv1.Scope {
	if s == nil {
		return nil
	}
	out := &promocodepbv1.Scope{MinSubtotal: moneyFromModel(s.MinSubtotal)}
	for i := range s.ScopeApplicableProperties {
		out.ApplicableProperties = append(out.ApplicableProperties, s.ScopeApplicableProperties[i].PropertyID)
	}
	for i := range s.ScopeApplicableUnits {
		out.ApplicableUnits = append(out.ApplicableUnits, s.ScopeApplicableUnits[i].UnitName)
	}
	return out
}
