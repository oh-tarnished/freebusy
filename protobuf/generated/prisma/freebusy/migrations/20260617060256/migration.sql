-- CreateSchema
CREATE SCHEMA IF NOT EXISTS "booking";

-- CreateSchema
CREATE SCHEMA IF NOT EXISTS "identity";

-- CreateSchema
CREATE SCHEMA IF NOT EXISTS "organisation";

-- CreateSchema
CREATE SCHEMA IF NOT EXISTS "promocode";

-- CreateSchema
CREATE SCHEMA IF NOT EXISTS "resource";

-- CreateSchema
CREATE SCHEMA IF NOT EXISTS "schedule";

-- CreateEnum
CREATE TYPE "booking"."booking_state" AS ENUM ('PENDING_HOLD', 'CONFIRMED', 'CANCELLED', 'EXPIRED', 'COMPLETED', 'NO_SHOW');

-- CreateEnum
CREATE TYPE "booking"."cancel_reason" AS ENUM ('REQUESTED_BY_CUSTOMER', 'REQUESTED_BY_OPERATOR', 'PAYMENT_FAILED', 'NO_SHOW', 'OTHER');

-- CreateEnum
CREATE TYPE "organisation"."organisation_state" AS ENUM ('ACTIVE', 'SUSPENDED');

-- CreateEnum
CREATE TYPE "organisation"."organisation_role" AS ENUM ('OWNER', 'ADMIN', 'MEMBER', 'VIEWER');

-- CreateEnum
CREATE TYPE "promocode"."discount_type" AS ENUM ('PERCENTAGE', 'FIXED_AMOUNT');

-- CreateEnum
CREATE TYPE "promocode"."promocode_state" AS ENUM ('ACTIVE', 'DISABLED', 'EXPIRED');

-- CreateEnum
CREATE TYPE "resource"."resource_type" AS ENUM ('PROVIDER', 'ROOM', 'EQUIPMENT', 'UNIT_TYPE', 'SPACE');

-- CreateEnum
CREATE TYPE "resource"."resource_state" AS ENUM ('ACTIVE', 'ARCHIVED');

-- CreateEnum
CREATE TYPE "resource"."pricing_unit" AS ENUM ('PER_BOOKING', 'PER_NIGHT', 'PER_PERSON');

-- CreateEnum
CREATE TYPE "schedule"."exception_kind" AS ENUM ('CLOSURE', 'EXTRA_HOURS');

-- CreateEnum
CREATE TYPE "schedule"."availability_exception_span_case" AS ENUM ('WINDOW', 'DATE_RANGE');

-- CreateEnum
CREATE TYPE "resource"."booking_mode" AS ENUM ('TIME_SLOT', 'NIGHTLY');

-- CreateEnum
CREATE TYPE "booking"."type" AS ENUM ('BASE', 'FEE', 'TAX', 'DISCOUNT');

-- CreateTable
CREATE TABLE "booking"."bookings" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "resource" TEXT NOT NULL,
    "offering" TEXT,
    "customer" TEXT,
    "units" INTEGER,
    "assigned_unit" TEXT,
    "state" "booking"."booking_state",
    "hold_expire_time" TIMESTAMP(3),
    "price" JSONB,
    "promo_code" TEXT,
    "discount" JSONB,
    "total" JSONB,
    "notes" TEXT,
    "attributes" JSONB,
    "cancel_reason" "booking"."cancel_reason",
    "create_time" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "update_time" TIMESTAMP(3) NOT NULL,
    "confirm_time" TIMESTAMP(3),
    "cancel_time" TIMESTAMP(3),
    "refund_amount" JSONB,
    "refund_percent" INTEGER,
    "hold_ttl" TEXT,
    "etag" TEXT,
    "contact_id" TEXT,
    "window_id" TEXT NOT NULL,

    CONSTRAINT "bookings_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "identity"."users" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "email" TEXT,
    "display_name" TEXT,
    "avatar_url" TEXT,
    "locale" TEXT,
    "time_zone" TEXT,
    "create_time" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "update_time" TIMESTAMP(3) NOT NULL,
    "etag" TEXT,

    CONSTRAINT "users_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "identity"."membership_summaries" (
    "id" TEXT NOT NULL,
    "organisation" TEXT,
    "org_display_name" TEXT,
    "role" TEXT,
    "user_id" TEXT NOT NULL,

    CONSTRAINT "membership_summaries_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "organisation"."organisations" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "display_name" TEXT NOT NULL,
    "slug" TEXT,
    "billing_email" TEXT,
    "state" "organisation"."organisation_state",
    "settings" JSONB,
    "member_count" BIGINT,
    "create_time" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "update_time" TIMESTAMP(3) NOT NULL,
    "etag" TEXT,

    CONSTRAINT "organisations_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "organisation"."members" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "user" TEXT,
    "email" TEXT NOT NULL,
    "display_name" TEXT,
    "role" "organisation"."organisation_role" NOT NULL DEFAULT 'OWNER',
    "state" "organisation"."organisation_state",
    "inviter" TEXT,
    "create_time" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "update_time" TIMESTAMP(3) NOT NULL,
    "etag" TEXT,
    "organisation_id" TEXT NOT NULL,

    CONSTRAINT "members_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "promocode"."promo_codes" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "code" TEXT NOT NULL,
    "display_name" TEXT,
    "description" TEXT,
    "discount_type" "promocode"."discount_type" NOT NULL DEFAULT 'PERCENTAGE',
    "percent_off" INTEGER,
    "amount_off" JSONB,
    "redeem_start_time" TIMESTAMP(3),
    "redeem_end_time" TIMESTAMP(3),
    "max_redemptions" BIGINT,
    "per_customer_limit" INTEGER,
    "min_subtotal" JSONB,
    "redemption_count" BIGINT,
    "state" "promocode"."promocode_state",
    "disabled" BOOLEAN,
    "create_time" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "update_time" TIMESTAMP(3) NOT NULL,
    "etag" TEXT,

    CONSTRAINT "promo_codes_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "promocode"."promo_code_applicable_resources" (
    "id" TEXT NOT NULL,
    "promo_code_id" TEXT NOT NULL,
    "resource_id" TEXT NOT NULL,

    CONSTRAINT "promo_code_applicable_resources_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "promocode"."promo_code_applicable_offerings" (
    "id" TEXT NOT NULL,
    "promo_code_id" TEXT NOT NULL,
    "offering_id" TEXT NOT NULL,

    CONSTRAINT "promo_code_applicable_offerings_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "resource"."resources" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "display_name" TEXT NOT NULL,
    "description" TEXT,
    "type" "resource"."resource_type" NOT NULL DEFAULT 'PROVIDER',
    "booking_mode" "resource"."booking_mode" NOT NULL DEFAULT 'TIME_SLOT',
    "capacity" INTEGER,
    "time_zone" TEXT NOT NULL,
    "tags" TEXT[],
    "attributes" JSONB,
    "state" "resource"."resource_state",
    "create_time" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "update_time" TIMESTAMP(3) NOT NULL,
    "etag" TEXT,

    CONSTRAINT "resources_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "resource"."offerings" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "display_name" TEXT NOT NULL,
    "description" TEXT,
    "duration" TEXT,
    "price" JSONB,
    "pricing_unit" "resource"."pricing_unit",
    "state" "resource"."resource_state",
    "create_time" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "update_time" TIMESTAMP(3) NOT NULL,
    "etag" TEXT,
    "resource_id" TEXT NOT NULL,

    CONSTRAINT "offerings_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "resource"."rate_overrides" (
    "id" TEXT NOT NULL,
    "weekdays" TEXT[],
    "price" JSONB NOT NULL,
    "offering_id" TEXT NOT NULL,
    "date_range_id" TEXT,

    CONSTRAINT "rate_overrides_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "resource"."los_discounts" (
    "id" TEXT NOT NULL,
    "min_nights" INTEGER NOT NULL,
    "percent_off" INTEGER,
    "amount_off" JSONB,
    "offering_id" TEXT NOT NULL,

    CONSTRAINT "los_discounts_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "resource"."fees" (
    "id" TEXT NOT NULL,
    "code" TEXT NOT NULL,
    "display_name" TEXT,
    "amount" JSONB,
    "percent" INTEGER,
    "pricing_unit" "resource"."pricing_unit",
    "taxable" BOOLEAN,
    "offering_id" TEXT NOT NULL,

    CONSTRAINT "fees_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "resource"."taxes" (
    "id" TEXT NOT NULL,
    "code" TEXT NOT NULL,
    "display_name" TEXT,
    "percent" DOUBLE PRECISION NOT NULL,
    "offering_id" TEXT NOT NULL,

    CONSTRAINT "taxes_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "resource"."resource_offerings" (
    "id" TEXT NOT NULL,
    "resource_id" TEXT NOT NULL,
    "offering_id" TEXT NOT NULL,

    CONSTRAINT "resource_offerings_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "schedule"."availability_exceptions" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "kind" "schedule"."exception_kind" NOT NULL DEFAULT 'CLOSURE',
    "reason" TEXT,
    "create_time" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "span_case" "schedule"."availability_exception_span_case",
    "resource_id" TEXT NOT NULL,
    "window_id" TEXT,
    "date_range_id" TEXT,

    CONSTRAINT "availability_exceptions_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "schedule"."schedules" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "etag" TEXT,
    "buffers_id" TEXT,
    "stay_constraints_id" TEXT,
    "cancellation_policy_id" TEXT,

    CONSTRAINT "schedules_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "schedule"."recurring_rules" (
    "id" TEXT NOT NULL,
    "rrule" TEXT NOT NULL,
    "opens" TEXT,
    "closes" TEXT,
    "schedule_id" TEXT NOT NULL,

    CONSTRAINT "recurring_rules_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "schedule"."buffer_settings" (
    "id" TEXT NOT NULL,
    "start_delta" TEXT,
    "end_delta" TEXT,
    "min_notice" TEXT,
    "max_advance" TEXT,
    "gap" TEXT,

    CONSTRAINT "buffer_settings_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "schedule"."stay_constraints" (
    "id" TEXT NOT NULL,
    "min_nights" INTEGER,
    "max_nights" INTEGER,
    "checkin_weekdays" TEXT[],
    "checkout_weekdays" TEXT[],
    "advance_min_days" INTEGER,
    "advance_max_days" INTEGER,

    CONSTRAINT "stay_constraints_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "schedule"."cancellation_policies" (
    "id" TEXT NOT NULL,

    CONSTRAINT "cancellation_policies_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "schedule"."refund_tiers" (
    "id" TEXT NOT NULL,
    "cutoff" TEXT NOT NULL,
    "refund_percent" INTEGER NOT NULL,
    "cancellation_policy_id" TEXT NOT NULL,

    CONSTRAINT "refund_tiers_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "schedule"."schedule_exceptions" (
    "id" TEXT NOT NULL,
    "schedule_id" TEXT NOT NULL,
    "availability_exception_id" TEXT NOT NULL,

    CONSTRAINT "schedule_exceptions_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "booking"."contacts" (
    "id" TEXT NOT NULL,
    "display_name" TEXT,
    "email" TEXT,
    "phone_number" TEXT,

    CONSTRAINT "contacts_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "booking"."time_windows" (
    "id" TEXT NOT NULL,
    "start_time" TIMESTAMP(3) NOT NULL,
    "end_time" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "time_windows_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "booking"."price_components" (
    "id" TEXT NOT NULL,
    "type" "booking"."type",
    "code" TEXT,
    "display_name" TEXT,
    "amount" JSONB,
    "booking_id" TEXT NOT NULL,

    CONSTRAINT "price_components_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "schedule"."date_ranges" (
    "id" TEXT NOT NULL,
    "start_date" TIMESTAMP(3) NOT NULL,
    "end_date" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "date_ranges_pkey" PRIMARY KEY ("id")
);

-- CreateIndex
CREATE UNIQUE INDEX "bookings_name_key" ON "booking"."bookings"("name");

-- CreateIndex
CREATE INDEX "bookings_resource_idx" ON "booking"."bookings"("resource");

-- CreateIndex
CREATE INDEX "bookings_offering_idx" ON "booking"."bookings"("offering");

-- CreateIndex
CREATE INDEX "bookings_customer_idx" ON "booking"."bookings"("customer");

-- CreateIndex
CREATE INDEX "bookings_promo_code_idx" ON "booking"."bookings"("promo_code");

-- CreateIndex
CREATE INDEX "bookings_contact_id_idx" ON "booking"."bookings"("contact_id");

-- CreateIndex
CREATE INDEX "bookings_window_id_idx" ON "booking"."bookings"("window_id");

-- CreateIndex
CREATE UNIQUE INDEX "users_name_key" ON "identity"."users"("name");

-- CreateIndex
CREATE INDEX "membership_summaries_organisation_idx" ON "identity"."membership_summaries"("organisation");

-- CreateIndex
CREATE INDEX "membership_summaries_user_id_idx" ON "identity"."membership_summaries"("user_id");

-- CreateIndex
CREATE UNIQUE INDEX "organisations_name_key" ON "organisation"."organisations"("name");

-- CreateIndex
CREATE UNIQUE INDEX "members_name_key" ON "organisation"."members"("name");

-- CreateIndex
CREATE INDEX "members_user_idx" ON "organisation"."members"("user");

-- CreateIndex
CREATE INDEX "members_inviter_idx" ON "organisation"."members"("inviter");

-- CreateIndex
CREATE INDEX "members_organisation_id_idx" ON "organisation"."members"("organisation_id");

-- CreateIndex
CREATE UNIQUE INDEX "promo_codes_name_key" ON "promocode"."promo_codes"("name");

-- CreateIndex
CREATE INDEX "promo_code_applicable_resources_promo_code_id_idx" ON "promocode"."promo_code_applicable_resources"("promo_code_id");

-- CreateIndex
CREATE INDEX "promo_code_applicable_resources_resource_id_idx" ON "promocode"."promo_code_applicable_resources"("resource_id");

-- CreateIndex
CREATE UNIQUE INDEX "promo_code_applicable_resources_promo_code_id_resource_id_key" ON "promocode"."promo_code_applicable_resources"("promo_code_id", "resource_id");

-- CreateIndex
CREATE INDEX "promo_code_applicable_offerings_promo_code_id_idx" ON "promocode"."promo_code_applicable_offerings"("promo_code_id");

-- CreateIndex
CREATE INDEX "promo_code_applicable_offerings_offering_id_idx" ON "promocode"."promo_code_applicable_offerings"("offering_id");

-- CreateIndex
CREATE UNIQUE INDEX "promo_code_applicable_offerings_promo_code_id_offering_id_key" ON "promocode"."promo_code_applicable_offerings"("promo_code_id", "offering_id");

-- CreateIndex
CREATE UNIQUE INDEX "resources_name_key" ON "resource"."resources"("name");

-- CreateIndex
CREATE UNIQUE INDEX "offerings_name_key" ON "resource"."offerings"("name");

-- CreateIndex
CREATE INDEX "offerings_resource_id_idx" ON "resource"."offerings"("resource_id");

-- CreateIndex
CREATE INDEX "rate_overrides_offering_id_idx" ON "resource"."rate_overrides"("offering_id");

-- CreateIndex
CREATE INDEX "rate_overrides_date_range_id_idx" ON "resource"."rate_overrides"("date_range_id");

-- CreateIndex
CREATE INDEX "los_discounts_offering_id_idx" ON "resource"."los_discounts"("offering_id");

-- CreateIndex
CREATE INDEX "fees_offering_id_idx" ON "resource"."fees"("offering_id");

-- CreateIndex
CREATE INDEX "taxes_offering_id_idx" ON "resource"."taxes"("offering_id");

-- CreateIndex
CREATE INDEX "resource_offerings_resource_id_idx" ON "resource"."resource_offerings"("resource_id");

-- CreateIndex
CREATE INDEX "resource_offerings_offering_id_idx" ON "resource"."resource_offerings"("offering_id");

-- CreateIndex
CREATE UNIQUE INDEX "resource_offerings_resource_id_offering_id_key" ON "resource"."resource_offerings"("resource_id", "offering_id");

-- CreateIndex
CREATE UNIQUE INDEX "availability_exceptions_name_key" ON "schedule"."availability_exceptions"("name");

-- CreateIndex
CREATE INDEX "availability_exceptions_resource_id_idx" ON "schedule"."availability_exceptions"("resource_id");

-- CreateIndex
CREATE INDEX "availability_exceptions_window_id_idx" ON "schedule"."availability_exceptions"("window_id");

-- CreateIndex
CREATE INDEX "availability_exceptions_date_range_id_idx" ON "schedule"."availability_exceptions"("date_range_id");

-- CreateIndex
CREATE UNIQUE INDEX "schedules_name_key" ON "schedule"."schedules"("name");

-- CreateIndex
CREATE INDEX "schedules_buffers_id_idx" ON "schedule"."schedules"("buffers_id");

-- CreateIndex
CREATE INDEX "schedules_stay_constraints_id_idx" ON "schedule"."schedules"("stay_constraints_id");

-- CreateIndex
CREATE INDEX "schedules_cancellation_policy_id_idx" ON "schedule"."schedules"("cancellation_policy_id");

-- CreateIndex
CREATE INDEX "recurring_rules_schedule_id_idx" ON "schedule"."recurring_rules"("schedule_id");

-- CreateIndex
CREATE INDEX "refund_tiers_cancellation_policy_id_idx" ON "schedule"."refund_tiers"("cancellation_policy_id");

-- CreateIndex
CREATE INDEX "schedule_exceptions_schedule_id_idx" ON "schedule"."schedule_exceptions"("schedule_id");

-- CreateIndex
CREATE INDEX "schedule_exceptions_availability_exception_id_idx" ON "schedule"."schedule_exceptions"("availability_exception_id");

-- CreateIndex
CREATE UNIQUE INDEX "schedule_exceptions_schedule_id_availability_exception_id_key" ON "schedule"."schedule_exceptions"("schedule_id", "availability_exception_id");

-- CreateIndex
CREATE INDEX "price_components_booking_id_idx" ON "booking"."price_components"("booking_id");

-- AddForeignKey
ALTER TABLE "booking"."bookings" ADD CONSTRAINT "bookings_resource_fkey" FOREIGN KEY ("resource") REFERENCES "resource"."resources"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "booking"."bookings" ADD CONSTRAINT "bookings_offering_fkey" FOREIGN KEY ("offering") REFERENCES "resource"."offerings"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "booking"."bookings" ADD CONSTRAINT "bookings_customer_fkey" FOREIGN KEY ("customer") REFERENCES "identity"."users"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "booking"."bookings" ADD CONSTRAINT "bookings_promo_code_fkey" FOREIGN KEY ("promo_code") REFERENCES "promocode"."promo_codes"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "booking"."bookings" ADD CONSTRAINT "bookings_contact_id_fkey" FOREIGN KEY ("contact_id") REFERENCES "booking"."contacts"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "booking"."bookings" ADD CONSTRAINT "bookings_window_id_fkey" FOREIGN KEY ("window_id") REFERENCES "booking"."time_windows"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "identity"."membership_summaries" ADD CONSTRAINT "membership_summaries_organisation_fkey" FOREIGN KEY ("organisation") REFERENCES "organisation"."organisations"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "identity"."membership_summaries" ADD CONSTRAINT "membership_summaries_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "identity"."users"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "organisation"."members" ADD CONSTRAINT "members_user_fkey" FOREIGN KEY ("user") REFERENCES "identity"."users"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "organisation"."members" ADD CONSTRAINT "members_inviter_fkey" FOREIGN KEY ("inviter") REFERENCES "identity"."users"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "organisation"."members" ADD CONSTRAINT "members_organisation_id_fkey" FOREIGN KEY ("organisation_id") REFERENCES "organisation"."organisations"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "promocode"."promo_code_applicable_resources" ADD CONSTRAINT "promo_code_applicable_resources_promo_code_id_fkey" FOREIGN KEY ("promo_code_id") REFERENCES "promocode"."promo_codes"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "promocode"."promo_code_applicable_resources" ADD CONSTRAINT "promo_code_applicable_resources_resource_id_fkey" FOREIGN KEY ("resource_id") REFERENCES "resource"."resources"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "promocode"."promo_code_applicable_offerings" ADD CONSTRAINT "promo_code_applicable_offerings_promo_code_id_fkey" FOREIGN KEY ("promo_code_id") REFERENCES "promocode"."promo_codes"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "promocode"."promo_code_applicable_offerings" ADD CONSTRAINT "promo_code_applicable_offerings_offering_id_fkey" FOREIGN KEY ("offering_id") REFERENCES "resource"."offerings"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "resource"."offerings" ADD CONSTRAINT "offerings_resource_id_fkey" FOREIGN KEY ("resource_id") REFERENCES "resource"."resources"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "resource"."rate_overrides" ADD CONSTRAINT "rate_overrides_offering_id_fkey" FOREIGN KEY ("offering_id") REFERENCES "resource"."offerings"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "resource"."rate_overrides" ADD CONSTRAINT "rate_overrides_date_range_id_fkey" FOREIGN KEY ("date_range_id") REFERENCES "schedule"."date_ranges"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "resource"."los_discounts" ADD CONSTRAINT "los_discounts_offering_id_fkey" FOREIGN KEY ("offering_id") REFERENCES "resource"."offerings"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "resource"."fees" ADD CONSTRAINT "fees_offering_id_fkey" FOREIGN KEY ("offering_id") REFERENCES "resource"."offerings"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "resource"."taxes" ADD CONSTRAINT "taxes_offering_id_fkey" FOREIGN KEY ("offering_id") REFERENCES "resource"."offerings"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "resource"."resource_offerings" ADD CONSTRAINT "resource_offerings_resource_id_fkey" FOREIGN KEY ("resource_id") REFERENCES "resource"."resources"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "resource"."resource_offerings" ADD CONSTRAINT "resource_offerings_offering_id_fkey" FOREIGN KEY ("offering_id") REFERENCES "resource"."offerings"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "schedule"."availability_exceptions" ADD CONSTRAINT "availability_exceptions_resource_id_fkey" FOREIGN KEY ("resource_id") REFERENCES "resource"."resources"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "schedule"."availability_exceptions" ADD CONSTRAINT "availability_exceptions_window_id_fkey" FOREIGN KEY ("window_id") REFERENCES "booking"."time_windows"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "schedule"."availability_exceptions" ADD CONSTRAINT "availability_exceptions_date_range_id_fkey" FOREIGN KEY ("date_range_id") REFERENCES "schedule"."date_ranges"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "schedule"."schedules" ADD CONSTRAINT "schedules_buffers_id_fkey" FOREIGN KEY ("buffers_id") REFERENCES "schedule"."buffer_settings"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "schedule"."schedules" ADD CONSTRAINT "schedules_stay_constraints_id_fkey" FOREIGN KEY ("stay_constraints_id") REFERENCES "schedule"."stay_constraints"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "schedule"."schedules" ADD CONSTRAINT "schedules_cancellation_policy_id_fkey" FOREIGN KEY ("cancellation_policy_id") REFERENCES "schedule"."cancellation_policies"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "schedule"."recurring_rules" ADD CONSTRAINT "recurring_rules_schedule_id_fkey" FOREIGN KEY ("schedule_id") REFERENCES "schedule"."schedules"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "schedule"."refund_tiers" ADD CONSTRAINT "refund_tiers_cancellation_policy_id_fkey" FOREIGN KEY ("cancellation_policy_id") REFERENCES "schedule"."cancellation_policies"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "schedule"."schedule_exceptions" ADD CONSTRAINT "schedule_exceptions_schedule_id_fkey" FOREIGN KEY ("schedule_id") REFERENCES "schedule"."schedules"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "schedule"."schedule_exceptions" ADD CONSTRAINT "schedule_exceptions_availability_exception_id_fkey" FOREIGN KEY ("availability_exception_id") REFERENCES "schedule"."availability_exceptions"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "booking"."price_components" ADD CONSTRAINT "price_components_booking_id_fkey" FOREIGN KEY ("booking_id") REFERENCES "booking"."bookings"("id") ON DELETE CASCADE ON UPDATE CASCADE;
