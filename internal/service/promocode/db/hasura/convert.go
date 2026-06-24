package hasura

import (
	"strings"
	"time"

	commonschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/commonql/schemaql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/commonql/moneysql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/promocodeql/discountsql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/promocodeql/redemptionwindowsql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/promocodeql/resourceql"
	pcschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/promocodeql/schemaql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/promocodeql/scopeapplicableofferingsql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/promocodeql/scopeapplicableresourcesql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/promocodeql/scopesql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/promocodeql/usagelimitsql"
	"github.com/oh-tarnished/freebusy/internal/service/promocode/discount"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/promocode/v1/promocodepbv1"
	"github.com/oh-tarnished/generateql/runtime/go/graphql"
	"github.com/oh-tarnished/runtime-go/ulid"
	"google.golang.org/genproto/googleapis/type/money"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// This file holds the pure conversions between the protobuf PromoCode and the
// normalized Hasura/GraphQL schema. The protobuf API nests the discount,
// redemption window, usage limits, and scope as sub-messages; the schema stores
// each as its own table (promocode.discounts, .redemption_windows, .usage_limits,
// .scopes) referenced by a foreign key on promocode.resource. Money value-objects
// normalize into common.moneys, and a scope's applicable resources / offerings
// become join rows. Timestamps cross the GraphQL boundary as RFC 3339 strings.

// State strings persisted in the resource.state column (match the DB CHECK).
const (
	stateActive   = "ACTIVE"
	stateDisabled = "DISABLED"
	stateExpired  = "EXPIRED"
)

// Discriminator values for the discount.amount_case column.
const (
	amountCasePercentOff = "PERCENT_OFF"
	amountCaseAmountOff  = "AMOUNT_OFF"
)

func deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

// lastSegment returns the final path component of an AIP resource name
// ("resources/r1/offerings/o1" -> "o1").
func lastSegment(name string) string {
	if i := strings.LastIndex(name, "/"); i >= 0 {
		return name[i+1:]
	}
	return name
}

// tsToStr renders a protobuf timestamp as the RFC 3339 string Hasura expects for
// a timestamptz column; the empty string (omitted on input) means NULL.
func tsToStr(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}
	return ts.AsTime().UTC().Format(time.RFC3339Nano)
}

// strToTS parses an RFC 3339 timestamptz string from Hasura back into a protobuf
// timestamp, tolerating the few layouts the engine emits. A blank/unparseable
// value yields nil.
func strToTS(s *string) *timestamppb.Timestamp {
	if s == nil || *s == "" {
		return nil
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02T15:04:05.999999Z07:00", "2006-01-02T15:04:05"} {
		if t, err := time.Parse(layout, *s); err == nil {
			return timestamppb.New(t)
		}
	}
	return nil
}

// moneyInput builds a common.moneys insert with the given id.
func moneyInput(id string, m *money.Money) moneysql.CreateInput {
	return moneysql.CreateInput{
		Id:           id,
		CurrencyCode: m.GetCurrencyCode(),
		Units:        graphql.Int64(m.GetUnits()),
		Nanos:        m.GetNanos(),
	}
}

func moneyFromModel(m *commonschema.CommonMoneys) *money.Money {
	if m == nil {
		return nil
	}
	return &money.Money{
		CurrencyCode: deref(m.CurrencyCode),
		Units:        int64(deref(m.Units)),
		Nanos:        deref(m.Nanos),
	}
}

// graph is the full set of GraphQL inserts a single PromoCode materializes into:
// the resource row plus its child rows, the Money rows they reference, and the
// scope's applicable join rows. The repository commits it as one atomic Tx batch
// (Money rows and children before the resource that references them; join rows
// after the scope they belong to).
type graph struct {
	resource  resourceql.CreateInput
	discount  discountsql.CreateInput
	window    *redemptionwindowsql.CreateInput
	limits    *usagelimitsql.CreateInput
	scope     *scopesql.CreateInput
	moneys    []moneysql.CreateInput
	resources []scopeapplicableresourcesql.CreateInput
	offerings []scopeapplicableofferingsql.CreateInput
}

// buildGraph turns a proto PromoCode into the insert graph that backs it, minting
// a fresh ULID for every row and wiring the foreign keys. Identity (Id/Name/Etag)
// of the resource row is left to the caller. now stamps create/update time.
func buildGraph(pc *promocodepbv1.PromoCode, now time.Time) *graph {
	g := &graph{}

	state := stateActive
	if pc.GetDisabled() {
		state = stateDisabled
	}
	nowStr := tsToStr(timestamppb.New(now))
	g.resource = resourceql.CreateInput{
		Code:        pc.GetCode(),
		DisplayName: pc.GetDisplayName(),
		Description: pc.GetDescription(),
		Disabled:    pc.GetDisabled(),
		State:       state,
		CreateTime:  nowStr,
		UpdateTime:  nowStr,
	}

	// Discount is required: always materialize a row. A non-nil amount_off is the
	// fixed-amount arm; otherwise it's the percentage arm.
	dID := ulid.GenerateString()
	g.discount = discountsql.CreateInput{Id: dID}
	if amt := pc.GetDiscount().GetAmountOff(); amt != nil {
		mID := ulid.GenerateString()
		g.moneys = append(g.moneys, moneyInput(mID, amt))
		g.discount.AmountOffId = mID
		g.discount.AmountCase = amountCaseAmountOff
	} else {
		g.discount.PercentOff = pc.GetDiscount().GetPercentOff()
		g.discount.AmountCase = amountCasePercentOff
	}
	g.resource.DiscountId = dID

	if w := pc.GetWindow(); w != nil {
		wID := ulid.GenerateString()
		g.window = &redemptionwindowsql.CreateInput{
			Id:        wID,
			StartTime: tsToStr(w.GetStartTime()),
			EndTime:   tsToStr(w.GetEndTime()),
		}
		g.resource.WindowId = wID
	}

	if l := pc.GetLimits(); l != nil {
		lID := ulid.GenerateString()
		ul := usagelimitsql.CreateInput{Id: lID}
		if l.GetMaxRedemptions() != nil {
			ul.MaxRedemptions = graphql.Int64(l.GetMaxRedemptions().GetValue())
		}
		if l.GetPerCustomerLimit() != nil {
			ul.PerCustomerLimit = l.GetPerCustomerLimit().GetValue()
		}
		g.limits = &ul
		g.resource.LimitsId = lID
	}

	if sc := pc.GetScope(); sc != nil {
		sID := ulid.GenerateString()
		scp := scopesql.CreateInput{Id: sID}
		if min := sc.GetMinSubtotal(); min != nil {
			mID := ulid.GenerateString()
			g.moneys = append(g.moneys, moneyInput(mID, min))
			scp.MinSubtotalId = mID
		}
		g.scope = &scp
		g.resource.ScopeId = sID
		for _, name := range sc.GetApplicableResources() {
			g.resources = append(g.resources, scopeapplicableresourcesql.CreateInput{
				Id:         ulid.GenerateString(),
				ScopeId:    sID,
				ResourceId: name,
			})
		}
		for _, name := range sc.GetApplicableOfferings() {
			g.offerings = append(g.offerings, scopeapplicableofferingsql.CreateInput{
				Id:           ulid.GenerateString(),
				ScopeId:      sID,
				OfferingId:   lastSegment(name),
				OfferingName: name,
			})
		}
	}

	return g
}

// parts holds a stored resource row and the child rows fetched to hydrate it.
type parts struct {
	res       *pcschema.PromocodeResource
	discount  *pcschema.PromocodeDiscounts
	amountOff *commonschema.CommonMoneys
	window    *pcschema.PromocodeRedemptionWindows
	limits    *pcschema.PromocodeUsageLimits
	scope     *pcschema.PromocodeScopes
	minSub    *commonschema.CommonMoneys
	resources []pcschema.PromocodeScopeApplicableResources
	offerings []pcschema.PromocodeScopeApplicableOfferings
}

// fromParts assembles the protobuf PromoCode from a stored resource row and its
// fetched child rows.
func fromParts(p parts) *promocodepbv1.PromoCode {
	res := p.res
	pc := &promocodepbv1.PromoCode{
		Name:            res.Name,
		Code:            res.Code,
		DisplayName:     deref(res.DisplayName),
		Description:     deref(res.Description),
		Discount:        discountFromModel(p.discount, p.amountOff),
		Window:          windowFromModel(p.window),
		Limits:          limitsFromModel(p.limits),
		Scope:           scopeFromParts(p.scope, p.minSub, p.resources, p.offerings),
		RedemptionCount: int64(deref(res.RedemptionCount)),
		Disabled:        deref(res.Disabled),
		CreateTime:      strToTS(&res.CreateTime),
		UpdateTime:      strToTS(&res.UpdateTime),
		Etag:            deref(res.Etag),
	}
	// Derive the lifecycle state from the window/flags rather than trusting the
	// possibly-stale stored value (a code becomes EXPIRED purely with time).
	pc.State = discount.EffectiveState(pc, time.Now().UTC())
	return pc
}

func discountFromModel(d *pcschema.PromocodeDiscounts, amountOff *commonschema.CommonMoneys) *promocodepbv1.Discount {
	if d == nil {
		return nil
	}
	out := &promocodepbv1.Discount{}
	if d.AmountCase != nil && *d.AmountCase == amountCaseAmountOff {
		out.Amount = &promocodepbv1.Discount_AmountOff{AmountOff: moneyFromModel(amountOff)}
	} else {
		out.Amount = &promocodepbv1.Discount_PercentOff{PercentOff: deref(d.PercentOff)}
	}
	return out
}

func windowFromModel(w *pcschema.PromocodeRedemptionWindows) *promocodepbv1.RedemptionWindow {
	if w == nil {
		return nil
	}
	return &promocodepbv1.RedemptionWindow{
		StartTime: strToTS(w.StartTime),
		EndTime:   strToTS(w.EndTime),
	}
}

func limitsFromModel(l *pcschema.PromocodeUsageLimits) *promocodepbv1.UsageLimits {
	if l == nil {
		return nil
	}
	out := &promocodepbv1.UsageLimits{}
	if l.MaxRedemptions != nil {
		out.MaxRedemptions = wrapperspb.Int64(int64(*l.MaxRedemptions))
	}
	if l.PerCustomerLimit != nil {
		out.PerCustomerLimit = wrapperspb.Int32(*l.PerCustomerLimit)
	}
	return out
}

func scopeFromParts(s *pcschema.PromocodeScopes, minSub *commonschema.CommonMoneys, res []pcschema.PromocodeScopeApplicableResources, off []pcschema.PromocodeScopeApplicableOfferings) *promocodepbv1.Scope {
	if s == nil {
		return nil
	}
	out := &promocodepbv1.Scope{MinSubtotal: moneyFromModel(minSub)}
	for i := range res {
		out.ApplicableResources = append(out.ApplicableResources, res[i].ResourceId)
	}
	for i := range off {
		out.ApplicableOfferings = append(out.ApplicableOfferings, off[i].OfferingName)
	}
	return out
}
