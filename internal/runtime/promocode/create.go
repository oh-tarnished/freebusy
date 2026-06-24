package promocode

import (
	"strings"

	"github.com/oh-tarnished/freebusy/internal/runtime/promocode/codegen"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/promocode/v1/promocodepbv1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// resolveCode derives the human-entered code for a create from the request's
// code_generation mode: AUTO mints a fresh server-generated code, while MANUAL
// (the default when unset) uses promo_code.code verbatim and requires it to be
// non-empty.
func resolveCode(req *promocodepbv1.CreatePromoCodeRequest) (string, error) {
	if req.GetCodeGeneration() == promocodepbv1.CodeGeneration_CODE_GENERATION_AUTO {
		return codegen.Generate(), nil
	}
	code := strings.TrimSpace(req.GetPromoCode().GetCode())
	if code == "" {
		return "", status.Error(codes.InvalidArgument, "promo_code.code is required when code_generation is MANUAL")
	}
	return code, nil
}
