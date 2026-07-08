DELETE FROM "identity"."guests"
WHERE "booking_id" = {{booking_id}}
RETURNING "id"