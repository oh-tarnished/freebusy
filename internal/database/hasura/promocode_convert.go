package hasura

import (
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/promocodeql/resourceql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/promocodeql/schemaql"
	"github.com/oh-tarnished/freebusy/internal/discount"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/promocode/v1/promocodepbv1"
	"github.com/oh-tarnished/generateql/runtime/go/graphql"
	"google.golang.org/genproto/googleapis/type/money"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// This file holds the pure conversions between the protobuf PromoCode and the
// Hasura/GraphQL types. Enums and timestamps cross as strings; integers cross as
// graphql.Int64; Money and the applicable lists are read back from the nested
// relations on schemaql.PromocodeResource (a single Get returns the whole graph).

// Enum string maps. Hasura persists the same string values as the relational
// schema, so PERCENTAGE / ACTIVE etc. travel verbatim.
var (
	discountTypeToStr = map[promocodepbv1.DiscountType]string{
		promocodepbv1.DiscountType_DISCOUNT_TYPE_PERCENTAGE:   "PERCENTAGE",
		promocodepbv1.DiscountType_DISCOUNT_TYPE_FIXED_AMOUNT: "FIXED_AMOUNT",
	}
	discountTypeFromStr = map[string]promocodepbv1.DiscountType{
		"PERCENTAGE":   promocodepbv1.DiscountType_DISCOUNT_TYPE_PERCENTAGE,
		"FIXED_AMOUNT": promocodepbv1.DiscountType_DISCOUNT_TYPE_FIXED_AMOUNT,
	}
	stateToStr = map[promocodepbv1.PromoCodeState]string{
		promocodepbv1.PromoCodeState_PROMO_CODE_STATE_ACTIVE:   "ACTIVE",
		promocodepbv1.PromoCodeState_PROMO_CODE_STATE_DISABLED: "DISABLED",
		promocodepbv1.PromoCodeState_PROMO_CODE_STATE_EXPIRED:  "EXPIRED",
	}
)

func deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

func int64Of(v *graphql.Int64) int64 {
	if v == nil {
		return 0
	}
	return int64(*v)
}

// tsToRFC formats a protobuf timestamp as an RFC3339 string, or "" when unset.
func tsToRFC(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}
	return ts.AsTime().UTC().Format(time.RFC3339Nano)
}

// rfcToTS parses a (possibly nil) RFC3339 string into a protobuf timestamp,
// tolerating the common Postgres/Hasura layouts. Unparseable input yields nil.
func rfcToTS(s *string) *timestamppb.Timestamp {
	if s == nil || *s == "" {
		return nil
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02T15:04:05.999999Z07:00"} {
		if t, err := time.Parse(layout, *s); err == nil {
			return timestamppb.New(t)
		}
	}
	return nil
}

func rfcStrToTS(s string) *timestamppb.Timestamp {
	if s == "" {
		return nil
	}
	return rfcToTS(&s)
}

// stateForWrite returns the lifecycle state string to persist on a write: a
// manually disabled code is DISABLED, otherwise ACTIVE (window-based EXPIRED is
// derived at read/validate time).
func stateForWrite(pc *promocodepbv1.PromoCode) string {
	if pc.GetDisabled() {
		return "DISABLED"
	}
	return "ACTIVE"
}

// resourceFromModel converts the Hasura read model (with its nested Money and
// applicable-list relations) into a protobuf PromoCode.
func resourceFromModel(m *schemaql.PromocodeResource) *promocodepbv1.PromoCode {
	pc := &promocodepbv1.PromoCode{
		Name:             m.Name,
		Code:             m.Code,
		DisplayName:      deref(m.DisplayName),
		Description:      deref(m.Description),
		DiscountType:     discountTypeFromStr[m.DiscountType],
		PercentOff:       deref(m.PercentOff),
		RedeemStartTime:  rfcToTS(m.RedeemStartTime),
		RedeemEndTime:    rfcToTS(m.RedeemEndTime),
		MaxRedemptions:   int64Of(m.MaxRedemptions),
		PerCustomerLimit: deref(m.PerCustomerLimit),
		RedemptionCount:  int64Of(m.RedemptionCount),
		Disabled:         deref(m.Disabled),
		Etag:             deref(m.Etag),
		CreateTime:       rfcStrToTS(m.CreateTime),
		UpdateTime:       rfcStrToTS(m.UpdateTime),
	}
	// Derive the lifecycle state from the window/flags rather than the stored value.
	pc.State = discount.EffectiveState(pc, time.Now().UTC())
	// Only surface Money when the relation actually resolved (a non-empty nested
	// id); a dangling foreign key must read back as nil, not a zero-value amount.
	if m.AmountOffId != nil && *m.AmountOffId != "" && m.BookingMoney.Id != "" {
		pc.AmountOff = &money.Money{
			CurrencyCode: deref(m.BookingMoney.CurrencyCode),
			Units:        int64Of(m.BookingMoney.Units),
			Nanos:        deref(m.BookingMoney.Nanos),
		}
	}
	if m.MinSubtotalId != nil && *m.MinSubtotalId != "" && m.BookingMoneyByMinSubtotalId.Id != "" {
		pc.MinSubtotal = &money.Money{
			CurrencyCode: deref(m.BookingMoneyByMinSubtotalId.CurrencyCode),
			Units:        int64Of(m.BookingMoneyByMinSubtotalId.Units),
			Nanos:        deref(m.BookingMoneyByMinSubtotalId.Nanos),
		}
	}
	for _, row := range m.PromocodeApplicableResources {
		pc.ApplicableResources = append(pc.ApplicableResources, row.ResourceId)
	}
	for _, row := range m.PromocodeApplicableOfferings {
		pc.ApplicableOfferings = append(pc.ApplicableOfferings, row.OfferingId)
	}
	return pc
}

// toCreateInput builds the insert payload for a promo code resource row. Money
// foreign keys (already inserted) and the server-assigned id/name/etag/timestamps
// are passed in; omitzero json tags drop unset optional fields from the mutation.
func toCreateInput(pc *promocodepbv1.PromoCode, id, name, etag, amountID, minID string) resourceql.CreateInput {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	discountType := discountTypeToStr[pc.GetDiscountType()]
	if discountType == "" {
		discountType = "PERCENTAGE"
	}
	return resourceql.CreateInput{
		Id:               id,
		Name:             name,
		Code:             pc.GetCode(),
		DisplayName:      pc.GetDisplayName(),
		Description:      pc.GetDescription(),
		DiscountType:     discountType,
		PercentOff:       pc.GetPercentOff(),
		AmountOffId:      amountID,
		MinSubtotalId:    minID,
		RedeemStartTime:  tsToRFC(pc.GetRedeemStartTime()),
		RedeemEndTime:    tsToRFC(pc.GetRedeemEndTime()),
		MaxRedemptions:   graphql.Int64(pc.GetMaxRedemptions()),
		PerCustomerLimit: pc.GetPerCustomerLimit(),
		RedemptionCount:  graphql.Int64(pc.GetRedemptionCount()),
		State:            stateForWrite(pc),
		Disabled:         pc.GetDisabled(),
		Etag:             etag,
		CreateTime:       now,
		UpdateTime:       now,
	}
}
