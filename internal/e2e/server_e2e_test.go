// This file end-to-end tests the ASSEMBLED server: the real *internal.Service
// (every domain server on its provider-selected repository) behind the real
// protovalidate interceptor chain, served over an in-memory bufconn listener
// and driven through the generated gRPC clients — exactly the stack a
// production client talks to, minus the TCP socket.
//
// Two env-gated matrices mirror the live-suite conventions:
//
//	FREEBUSY_TEST_POSTGRES_DSN="host=... dbname=freebusydb ..." go test ./internal/e2e/ -run TestE2E_Gorm -v
//	FREEBUSY_TEST_GRAPHQL_URL="http://localhost:3280/graphql"  go test ./internal/e2e/ -run TestE2E_Hasura -v
//
// (run `just migrate` / `just hasura regen` first so the backends are ready).
package e2e

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/oh-tarnished/freebusy/internal"
	"github.com/oh-tarnished/freebusy/internal/database"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql"
	propertyrepo "github.com/oh-tarnished/freebusy/internal/database/repository/freebusy/property"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/availability/v1/availabilitypbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/booking/v1/bookingpbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/identity/v1/identitypbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/organisation/v1/orgpbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/promocode/v1/promocodepbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/property/v1/propertypbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/schedule/v1/schedulepbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/shared/v1/sharedpbv1"
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/genproto/googleapis/type/money"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestE2E_Gorm(t *testing.T) {
	dsn := os.Getenv("FREEBUSY_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("FREEBUSY_TEST_POSTGRES_DSN not set — live server e2e skipped")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:         logger.Default.LogMode(logger.Silent),
		TranslateError: true, // sentinel errors (ErrDuplicatedKey) like database.Open
	})
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	cc := dialServer(t, &database.Connection{PgSQLConn: db}, database.ProviderGorm)
	serverLifecycle(t, cc, database.ProviderGorm, propertyrepo.NewGorm(db))
}

// TestE2E_Hasura runs the identical client-visible lifecycle with the server
// assembled on the Hasura DDN backend — one behavior, two providers.
func TestE2E_Hasura(t *testing.T) {
	raw := os.Getenv("FREEBUSY_TEST_GRAPHQL_URL")
	if raw == "" {
		t.Skip("FREEBUSY_TEST_GRAPHQL_URL not set — live server e2e skipped")
	}
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse %s: %v", raw, err)
	}
	svc, err := freebusyql.Connect(u)
	if err != nil {
		t.Fatalf("connect %s: %v", raw, err)
	}
	cc := dialServer(t, &database.Connection{Hasura: svc}, database.ProviderHasura)
	serverLifecycle(t, cc, database.ProviderHasura, propertyrepo.NewGraphQL(svc))
}

// dialServer assembles the full server on conn/provider, serves it over
// bufconn, and returns a connected client conn. Everything is torn down with
// the test.
func dialServer(t *testing.T, conn *database.Connection, provider database.Provider) *grpc.ClientConn {
	t.Helper()
	restore := database.SetTestBackend(conn, provider)
	t.Cleanup(restore)

	srv, _, err := internal.NewGRPCServer()
	if err != nil {
		t.Fatalf("assemble server: %v", err)
	}
	lis := bufconn.Listen(1 << 20)
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)

	cc, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("dial bufconn: %v", err)
	}
	t.Cleanup(func() { _ = cc.Close() })
	return cc
}

// wantCode asserts err carries the given gRPC status code.
func wantCode(t *testing.T, err error, want codes.Code, what string) {
	t.Helper()
	if status.Code(err) != want {
		t.Fatalf("%s: got %v (err %v), want %v", what, status.Code(err), err, want)
	}
}

// serverLifecycle drives every service through the wire: interceptor
// rejections, full CRUD with masks and etags, the booking hold flow, promo
// validation, and the provider-visible filter divergence on derived state.
// propRepos is the generated repository set for the same backend, used only to
// clean up the property row — PropertyService archives rather than deletes.
func serverLifecycle(t *testing.T, cc *grpc.ClientConn, provider database.Provider, propRepos propertyrepo.Repositories) {
	t.Helper()
	ctx := context.Background()
	suffix := fmt.Sprintf("%d", time.Now().UnixNano()%1_000_000_000)

	orgs := orgpbv1.NewOrganisationServiceClient(cc)
	props := propertypbv1.NewPropertyServiceClient(cc)
	licences := propertypbv1.NewLicenceServiceClient(cc)
	schedules := schedulepbv1.NewScheduleServiceClient(cc)
	bookings := bookingpbv1.NewBookingServiceClient(cc)
	promos := promocodepbv1.NewPromoCodeServiceClient(cc)
	avail := availabilitypbv1.NewAvailabilityServiceClient(cc)
	identity := identitypbv1.NewIdentityServiceClient(cc)

	// --- Interceptor rejections reach the wire as InvalidArgument ------------
	_, err := orgs.CreateOrganisation(ctx, &orgpbv1.CreateOrganisationRequest{})
	wantCode(t, err, codes.InvalidArgument, "create org without body")
	_, err = orgs.GetOrganisation(ctx, &orgpbv1.GetOrganisationRequest{Name: "orgs/x"})
	wantCode(t, err, codes.InvalidArgument, "get org with bad name")
	_, err = avail.ComputeAvailability(ctx, &availabilitypbv1.ComputeAvailabilityRequest{
		Unit: "properties/p/units/u",
	})
	wantCode(t, err, codes.InvalidArgument, "compute availability without period")

	// --- Organisation + member lifecycle -------------------------------------
	org, err := orgs.CreateOrganisation(ctx, &orgpbv1.CreateOrganisationRequest{
		Organisation: &orgpbv1.Organisation{DisplayName: "e2e-org-" + suffix},
	})
	if err != nil {
		t.Fatalf("CreateOrganisation: %v", err)
	}
	if org.GetState() != orgpbv1.OrganisationState_ORGANISATION_STATE_ACTIVE {
		t.Fatalf("org state = %v, want ACTIVE (database default)", org.GetState())
	}
	defer orgs.DeleteOrganisation(ctx, &orgpbv1.DeleteOrganisationRequest{Name: org.GetName(), Force: true})

	page, err := orgs.ListOrganisations(ctx, &orgpbv1.ListOrganisationsRequest{
		Filter: fmt.Sprintf("display_name = %q", org.GetDisplayName()),
	})
	if err != nil || len(page.GetOrganisations()) != 1 {
		t.Fatalf("ListOrganisations filter: err=%v n=%d, want 1", err, len(page.GetOrganisations()))
	}

	renamed, err := orgs.UpdateOrganisation(ctx, &orgpbv1.UpdateOrganisationRequest{
		Organisation: &orgpbv1.Organisation{Name: org.GetName(), DisplayName: "e2e-org-renamed-" + suffix},
		UpdateMask:   &fieldmaskpb.FieldMask{Paths: []string{"display_name"}},
	})
	if err != nil || renamed.GetDisplayName() != "e2e-org-renamed-"+suffix {
		t.Fatalf("UpdateOrganisation: %v (display_name %q)", err, renamed.GetDisplayName())
	}
	// The pre-rename etag is now stale: optimistic concurrency must reject it.
	_, err = orgs.UpdateOrganisation(ctx, &orgpbv1.UpdateOrganisationRequest{
		Organisation: &orgpbv1.Organisation{Name: org.GetName(), DisplayName: "x", Etag: org.GetEtag()},
		UpdateMask:   &fieldmaskpb.FieldMask{Paths: []string{"display_name"}},
	})
	wantCode(t, err, codes.Aborted, "update org with stale etag")

	invited, err := orgs.InviteMember(ctx, &orgpbv1.InviteMemberRequest{
		Parent: org.GetName(),
		Email:  "e2e-" + suffix + "@example.com",
		Role:   orgpbv1.OrganisationRole_ORGANISATION_ROLE_ADMIN,
	})
	if err != nil {
		t.Fatalf("InviteMember: %v", err)
	}
	if invited.GetMember().GetState() != orgpbv1.MemberState_MEMBER_STATE_INVITED {
		t.Fatalf("member state = %v, want INVITED (database default)", invited.GetMember().GetState())
	}
	// The force-delete guard: an organisation with members refuses a plain delete.
	_, err = orgs.DeleteOrganisation(ctx, &orgpbv1.DeleteOrganisationRequest{Name: org.GetName()})
	wantCode(t, err, codes.Aborted, "delete org with members without force")
	if _, err := orgs.DeleteMember(ctx, &orgpbv1.DeleteMemberRequest{Name: invited.GetMember().GetName()}); err != nil {
		t.Fatalf("DeleteMember: %v", err)
	}

	// --- Property / unit / licence -------------------------------------------
	prop, err := props.CreateProperty(ctx, &propertypbv1.CreatePropertyRequest{
		Property: &propertypbv1.Property{
			Organisation: org.GetName(),
			DisplayName:  "e2e-prop-" + suffix,
			TimeZone:     "UTC",
		},
	})
	if err != nil {
		t.Fatalf("CreateProperty: %v", err)
	}
	defer propRepos.Properties.Delete(ctx, prop.GetName())

	unit, err := props.CreateUnit(ctx, &propertypbv1.CreateUnitRequest{
		Parent: prop.GetName(),
		Unit: &propertypbv1.Unit{
			DisplayName: "e2e-unit-" + suffix,
			Type:        propertypbv1.UnitType_UNIT_TYPE_ROOM,
			BookingMode: sharedpbv1.BookingMode_BOOKING_MODE_NIGHTLY,
			TimeZone:    "UTC",
			MaxOccupancy: 2,
			Price:       &money.Money{CurrencyCode: "USD", Units: 100},
		},
	})
	if err != nil {
		t.Fatalf("CreateUnit: %v", err)
	}
	defer props.DeleteUnit(ctx, &propertypbv1.DeleteUnitRequest{Name: unit.GetName(), Force: true})

	// A licence naming a unit of a DIFFERENT property is rejected before any I/O.
	_, err = licences.CreateLicence(ctx, &propertypbv1.CreateLicenceRequest{
		Parent: prop.GetName(),
		Licence: &propertypbv1.Licence{
			Type: propertypbv1.LicenceType_LICENCE_TYPE_FIRE_SAFETY,
			Unit: "properties/someone-else/units/u1",
		},
	})
	wantCode(t, err, codes.InvalidArgument, "create licence with foreign unit")

	lic, err := licences.CreateLicence(ctx, &propertypbv1.CreateLicenceRequest{
		Parent: prop.GetName(),
		Licence: &propertypbv1.Licence{
			Type:       propertypbv1.LicenceType_LICENCE_TYPE_FIRE_SAFETY,
			Unit:       unit.GetName(),
			ExpiryDate: &date.Date{Year: 2027, Month: 6, Day: 30},
		},
	})
	if err != nil {
		t.Fatalf("CreateLicence: %v", err)
	}
	if lic.GetTarget() != propertypbv1.LicenceTarget_LICENCE_TARGET_UNIT {
		t.Fatalf("licence target = %v, want UNIT (derived from unit)", lic.GetTarget())
	}
	// The compliance query: licences due for renewal before a cutoff.
	licPage, err := licences.ListLicences(ctx, &propertypbv1.ListLicencesRequest{
		Parent: prop.GetName(),
		Filter: `expiry_date <= 2027-12-31`,
	})
	if err != nil || len(licPage.GetLicences()) != 1 {
		t.Fatalf("ListLicences expiry filter: err=%v n=%d, want 1", err, len(licPage.GetLicences()))
	}
	if _, err := licences.DeleteLicence(ctx, &propertypbv1.DeleteLicenceRequest{Name: lic.GetName()}); err != nil {
		t.Fatalf("DeleteLicence: %v", err)
	}

	// --- Schedule exception over the wire ------------------------------------
	_, err = schedules.CreateAvailabilityException(ctx, &schedulepbv1.CreateAvailabilityExceptionRequest{
		Parent: unit.GetName(),
		AvailabilityException: &schedulepbv1.AvailabilityException{
			Kind: schedulepbv1.ExceptionKind_EXCEPTION_KIND_CLOSURE,
		},
	})
	wantCode(t, err, codes.InvalidArgument, "create exception without span")

	exc, err := schedules.CreateAvailabilityException(ctx, &schedulepbv1.CreateAvailabilityExceptionRequest{
		Parent: unit.GetName(),
		AvailabilityException: &schedulepbv1.AvailabilityException{
			Kind:   schedulepbv1.ExceptionKind_EXCEPTION_KIND_CLOSURE,
			Reason: "e2e closure",
			Span: &schedulepbv1.AvailabilityException_DateRange{DateRange: &sharedpbv1.DateRange{
				StartDate: &date.Date{Year: 2027, Month: 12, Day: 24},
				EndDate:   &date.Date{Year: 2027, Month: 12, Day: 26},
			}},
		},
	})
	if err != nil {
		t.Fatalf("CreateAvailabilityException: %v", err)
	}
	excPage, err := schedules.ListAvailabilityExceptions(ctx, &schedulepbv1.ListAvailabilityExceptionsRequest{Parent: unit.GetName()})
	if err != nil || len(excPage.GetAvailabilityExceptions()) != 1 {
		t.Fatalf("ListAvailabilityExceptions: err=%v n=%d, want 1", err, len(excPage.GetAvailabilityExceptions()))
	}
	if _, err := schedules.DeleteAvailabilityException(ctx, &schedulepbv1.DeleteAvailabilityExceptionRequest{Name: exc.GetName()}); err != nil {
		t.Fatalf("DeleteAvailabilityException: %v", err)
	}

	// --- Promo code + validation ---------------------------------------------
	code := "E2E" + suffix
	promo, err := promos.CreatePromoCode(ctx, &promocodepbv1.CreatePromoCodeRequest{
		PromoCode: &promocodepbv1.PromoCode{
			Code:     code,
			Discount: &promocodepbv1.Discount{Amount: &promocodepbv1.Discount_PercentOff{PercentOff: 20}},
		},
	})
	if err != nil {
		t.Fatalf("CreatePromoCode: %v", err)
	}
	defer promos.DeletePromoCode(ctx, &promocodepbv1.DeletePromoCodeRequest{Name: promo.GetName(), Force: true})
	if promo.GetState() != promocodepbv1.PromoCodeState_PROMO_CODE_STATE_ACTIVE {
		t.Fatalf("promo state = %v, want derived ACTIVE", promo.GetState())
	}

	verdict, err := promos.ValidatePromoCode(ctx, &promocodepbv1.ValidatePromoCodeRequest{
		Code:     code,
		Subtotal: &money.Money{CurrencyCode: "USD", Units: 100},
	})
	if err != nil {
		t.Fatalf("ValidatePromoCode: %v", err)
	}
	if !verdict.GetValid() || verdict.GetDiscountAmount().GetUnits() != 20 {
		t.Fatalf("ValidatePromoCode: valid=%t discount=%v, want valid with 20 USD off", verdict.GetValid(), verdict.GetDiscountAmount())
	}

	// Derived-state filtering is the one deliberate provider divergence: the
	// gorm engine answers `state = ACTIVE` through a hand-written override; the
	// hasura engine rejects the non-stored column.
	_, err = promos.ListPromoCodes(ctx, &promocodepbv1.ListPromoCodesRequest{Filter: "state = ACTIVE"})
	if provider == database.ProviderGorm {
		if err != nil {
			t.Fatalf("ListPromoCodes state filter (gorm): %v", err)
		}
	} else {
		wantCode(t, err, codes.InvalidArgument, "ListPromoCodes state filter (hasura)")
	}

	// --- Booking hold flow -----------------------------------------------------
	_, err = bookings.RescheduleBooking(ctx, &bookingpbv1.RescheduleBookingRequest{Name: "bookings/nope"})
	wantCode(t, err, codes.InvalidArgument, "reschedule without window")

	start := time.Now().UTC().AddDate(0, 0, 30).Truncate(24 * time.Hour)
	booking, err := bookings.CreateBooking(ctx, &bookingpbv1.CreateBookingRequest{
		Booking: &bookingpbv1.Booking{
			Unit: unit.GetName(),
			Window: &sharedpbv1.TimeWindow{
				StartTime: timestamppb.New(start),
				EndTime:   timestamppb.New(start.AddDate(0, 0, 2)),
			},
			Contact: &sharedpbv1.Contact{DisplayName: "E2E Guest", Email: "guest-" + suffix + "@example.com"},
		},
	})
	if err != nil {
		t.Fatalf("CreateBooking: %v", err)
	}
	if booking.GetState() != bookingpbv1.BookingState_BOOKING_STATE_PENDING_HOLD {
		t.Fatalf("booking state = %v, want PENDING_HOLD", booking.GetState())
	}

	confirmed, err := bookings.ConfirmBooking(ctx, &bookingpbv1.ConfirmBookingRequest{Name: booking.GetName()})
	if err != nil || confirmed.GetState() != bookingpbv1.BookingState_BOOKING_STATE_CONFIRMED {
		t.Fatalf("ConfirmBooking: %v (state %v)", err, confirmed.GetState())
	}

	// The confirmed stay occupies the span: the same window is no longer bookable.
	check, err := avail.CheckAvailability(ctx, &availabilitypbv1.CheckAvailabilityRequest{
		Unit: unit.GetName(),
		Period: &availabilitypbv1.CheckAvailabilityRequest_Window{Window: &sharedpbv1.TimeWindow{
			StartTime: timestamppb.New(start),
			EndTime:   timestamppb.New(start.AddDate(0, 0, 2)),
		}},
	})
	if err != nil {
		t.Fatalf("CheckAvailability: %v", err)
	}
	if check.GetBookable() {
		t.Fatal("CheckAvailability: confirmed span still reports bookable")
	}

	if _, err := bookings.PreviewCancellation(ctx, &bookingpbv1.PreviewCancellationRequest{Name: booking.GetName()}); err != nil {
		t.Fatalf("PreviewCancellation: %v", err)
	}
	cancelled, err := bookings.CancelBooking(ctx, &bookingpbv1.CancelBookingRequest{Name: booking.GetName()})
	if err != nil || cancelled.GetState() != bookingpbv1.BookingState_BOOKING_STATE_CANCELLED {
		t.Fatalf("CancelBooking: %v (state %v)", err, cancelled.GetState())
	}

	// --- Identity -----------------------------------------------------------
	if _, err := identity.ListUsers(ctx, &identitypbv1.ListUsersRequest{}); err != nil {
		t.Fatalf("ListUsers: %v", err)
	}
	// "users/me" needs an authenticated caller; the bufconn client sends none.
	_, err = identity.GetUser(ctx, &identitypbv1.GetUserRequest{Name: "users/me"})
	wantCode(t, err, codes.Unauthenticated, "get users/me without caller identity")

	// Teardown in dependency order (the deferred deletes are the safety net).
	if _, err := props.DeleteUnit(ctx, &propertypbv1.DeleteUnitRequest{Name: unit.GetName(), Force: true}); err != nil {
		t.Logf("DeleteUnit: %v", err)
	}
	if err := propRepos.Properties.Delete(ctx, prop.GetName()); err != nil {
		t.Logf("delete property row: %v", err)
	}
	if _, err := orgs.DeleteOrganisation(ctx, &orgpbv1.DeleteOrganisationRequest{Name: org.GetName(), Force: true}); err != nil {
		t.Logf("DeleteOrganisation: %v", err)
	}
}
