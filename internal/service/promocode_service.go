package service

import (
	"context"
	"errors"

	"github.com/oh-tarnished/freebusy/internal/database/repository"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/promocode/v1/promocodepbv1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

// PromoCodeServer implements promocodepbv1.PromoCodeServiceServer on top of a
// provider-agnostic repository.PromoCodeRepository, which is selected (GORM or
// Hasura) by the database factory.
type PromoCodeServer struct {
	promocodepbv1.UnimplementedPromoCodeServiceServer
	repo repository.PromoCodeRepository
}

// NewPromoCodeServer returns a PromoCodeServer backed by repo.
func NewPromoCodeServer(repo repository.PromoCodeRepository) *PromoCodeServer {
	return &PromoCodeServer{repo: repo}
}

// ListPromoCodes returns a page of promo codes for the given pagination request.
func (s *PromoCodeServer) ListPromoCodes(ctx context.Context, req *promocodepbv1.ListPromoCodesRequest) (*promocodepbv1.ListPromoCodesResponse, error) {
	if req.GetFilter() != "" {
		return nil, status.Error(codes.Unimplemented, "filter is not yet supported")
	}
	var out *promocodepbv1.ListPromoCodesResponse
	err := traced(ctx, "ListPromoCodes", func(ctx context.Context) error {
		items, next, err := s.repo.List(ctx, types.ListParams{
			PageSize:  req.GetPageSize(),
			PageToken: req.GetPageToken(),
			OrderBy:   req.GetOrderBy(),
		})
		if err != nil {
			return toStatusErr(err)
		}
		out = &promocodepbv1.ListPromoCodesResponse{PromoCodes: items, NextPageToken: next}
		return nil
	})
	return out, err
}

// GetPromoCode returns a single promo code by resource name.
func (s *PromoCodeServer) GetPromoCode(ctx context.Context, req *promocodepbv1.GetPromoCodeRequest) (*promocodepbv1.PromoCode, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	var out *promocodepbv1.PromoCode
	err := traced(ctx, "GetPromoCode", func(ctx context.Context) error {
		pc, err := s.repo.Get(ctx, req.GetName())
		if err != nil {
			return toStatusErr(err)
		}
		out = pc
		return nil
	})
	return out, err
}

// CreatePromoCode creates a promo code. A caller-supplied promo_code_id fixes the
// resource name; validate_only runs request validation without persisting.
func (s *PromoCodeServer) CreatePromoCode(ctx context.Context, req *promocodepbv1.CreatePromoCodeRequest) (*promocodepbv1.PromoCode, error) {
	pc := req.GetPromoCode()
	if pc == nil {
		return nil, status.Error(codes.InvalidArgument, "promo_code is required")
	}
	if pc.GetCode() == "" {
		return nil, status.Error(codes.InvalidArgument, "promo_code.code is required")
	}
	if pc.GetDiscountType() == promocodepbv1.DiscountType_DISCOUNT_TYPE_PERCENTAGE {
		if p := pc.GetPercentOff(); p < 1 || p > 100 {
			return nil, status.Error(codes.InvalidArgument, "promo_code.percent_off must be between 1 and 100 for a percentage discount")
		}
	}
	if id := req.GetPromoCodeId(); id != "" {
		name, err := types.PromoCodeName(id)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid promo_code_id")
		}
		// Clone so the caller-chosen name doesn't mutate the inbound request proto.
		pc = proto.Clone(pc).(*promocodepbv1.PromoCode)
		pc.Name = name
	}
	if req.GetValidateOnly() {
		// Dry run: surface what a real create would reject (duplicate code) without
		// persisting, rather than blindly echoing the request.
		var out *promocodepbv1.PromoCode
		err := traced(ctx, "CreatePromoCode.validateOnly", func(ctx context.Context) error {
			switch _, err := s.repo.FindByCode(ctx, pc.GetCode()); {
			case err == nil:
				return status.Error(codes.AlreadyExists, "a promo code with this code already exists")
			case !errors.Is(err, types.ErrNotFound):
				return toStatusErr(err)
			}
			out = pc
			return nil
		})
		return out, err
	}
	var out *promocodepbv1.PromoCode
	err := traced(ctx, "CreatePromoCode", func(ctx context.Context) error {
		created, err := s.repo.Create(ctx, pc)
		if err != nil {
			return toStatusErr(err)
		}
		out = created
		return nil
	})
	return out, err
}

// UpdatePromoCode applies the update mask to an existing promo code. An empty mask
// replaces every mutable field; validate_only skips persistence.
func (s *PromoCodeServer) UpdatePromoCode(ctx context.Context, req *promocodepbv1.UpdatePromoCodeRequest) (*promocodepbv1.PromoCode, error) {
	pc := req.GetPromoCode()
	if pc == nil || pc.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "promo_code.name is required")
	}
	if req.GetValidateOnly() {
		// Dry run: confirm the target exists and the etag matches (the checks a real
		// update would enforce) without persisting, rather than echoing the request.
		var out *promocodepbv1.PromoCode
		err := traced(ctx, "UpdatePromoCode.validateOnly", func(ctx context.Context) error {
			existing, err := s.repo.Get(ctx, pc.GetName())
			if err != nil {
				return toStatusErr(err)
			}
			if pc.GetEtag() != "" && existing.GetEtag() != "" && pc.GetEtag() != existing.GetEtag() {
				return status.Error(codes.Aborted, "etag mismatch")
			}
			out = pc
			return nil
		})
		return out, err
	}
	var out *promocodepbv1.PromoCode
	err := traced(ctx, "UpdatePromoCode", func(ctx context.Context) error {
		updated, err := s.repo.Update(ctx, pc, req.GetUpdateMask().GetPaths())
		if err != nil {
			return toStatusErr(err)
		}
		out = updated
		return nil
	})
	return out, err
}

// DeletePromoCode removes a promo code by resource name.
func (s *PromoCodeServer) DeletePromoCode(ctx context.Context, req *promocodepbv1.DeletePromoCodeRequest) (*emptypb.Empty, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	err := traced(ctx, "DeletePromoCode", func(ctx context.Context) error {
		return toStatusErr(s.repo.Delete(ctx, req.GetName()))
	})
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
