package service

import (
	"context"
	"errors"
	"time"

	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/internal/discount"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/promocode/v1/promocodepbv1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ValidatePromoCode resolves a human-entered code and evaluates it against the
// prospective booking (subtotal, optional resource and offering). A code that is
// absent or not redeemable is reported as Valid=false with a Reason rather than a
// gRPC error, since validation failures are an expected, non-exceptional outcome.
func (s *PromoCodeServer) ValidatePromoCode(ctx context.Context, req *promocodepbv1.ValidatePromoCodeRequest) (*promocodepbv1.ValidatePromoCodeResponse, error) {
	if req.GetCode() == "" {
		return nil, status.Error(codes.InvalidArgument, "code is required")
	}
	if req.GetSubtotal() == nil {
		return nil, status.Error(codes.InvalidArgument, "subtotal is required")
	}

	var out *promocodepbv1.ValidatePromoCodeResponse
	err := traced(ctx, "ValidatePromoCode", func(ctx context.Context) error {
		pc, err := s.repo.FindByCode(ctx, req.GetCode())
		if err != nil {
			if errors.Is(err, types.ErrNotFound) {
				out = &promocodepbv1.ValidatePromoCodeResponse{Valid: false, Reason: "promo code not found"}
				return nil
			}
			return toStatusErr(err)
		}

		result := discount.Evaluate(pc, req.GetSubtotal(), req.GetResource(), req.GetOffering(), time.Now().UTC())
		out = &promocodepbv1.ValidatePromoCodeResponse{
			Valid:          result.Valid,
			Reason:         result.Reason,
			PromoCode:      pc.GetName(),
			DiscountAmount: result.Discount,
			FinalTotal:     result.FinalTotal,
		}
		return nil
	})
	return out, err
}
