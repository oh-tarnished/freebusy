# Exercise the PromoCodeService (server must be running; needs grpcurl).
# Invoked as `just promo <recipe>`.
set working-directory := '../..'

grpc_port := env_var_or_default("FREEBUSY_GRPC_PORT", "50051")

# List promo codes.
[group('api')]
list:
    grpcurl -plaintext -d '{}' localhost:{{ grpc_port }} freebusy.promocode.v1.PromoCodeService/ListPromoCodes

# Create a sample 25%-off promo code.
[group('api')]
create:
    grpcurl -plaintext -d '{"promo_code":{"code":"SUMMER25","display_name":"Summer Sale","discount_type":"DISCOUNT_TYPE_PERCENTAGE","percent_off":25}}' \
        localhost:{{ grpc_port }} freebusy.promocode.v1.PromoCodeService/CreatePromoCode

# Validate a code against a $100 subtotal.
[group('api')]
validate code="SUMMER25":
    grpcurl -plaintext -d '{"code":"{{ code }}","subtotal":{"currency_code":"USD","units":100}}' \
        localhost:{{ grpc_port }} freebusy.promocode.v1.PromoCodeService/ValidatePromoCode
