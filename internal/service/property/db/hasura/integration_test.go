package hasura

// Live-DDN integration tests for the licence surface: the first real
// verification of the Hasura licence provider (everything else is
// compile-only). They exercise the attachment (bytea) round-trip, the
// expiry_date renewal-reminder filter, masked updates, and the DeleteUnit
// force guard against a running engine.
//
// Skipped unless FREEBUSY_TEST_GRAPHQL_URL points at a live engine:
//
//	FREEBUSY_TEST_GRAPHQL_URL=http://localhost:3280/graphql \
//	  go test ./internal/service/property/db/hasura/ -run Live -v
//
// Each run seeds a fresh organisation → property → unit chain under new ULIDs,
// so runs never collide; rows are left behind (local dev database).

import (
	"bytes"
	"context"
	"errors"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql"
	orgresourceql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/organisationql/resourceql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/propertiesql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/unitsql"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/property/v1/propertypbv1"
	sharedpbv1 "github.com/oh-tarnished/freebusy/protobuf/generated/go/shared/v1/sharedpbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// liveService connects to the engine named by FREEBUSY_TEST_GRAPHQL_URL, or
// skips the test when unset.
func liveService(t *testing.T) *freebusyql.Service {
	t.Helper()
	raw := os.Getenv("FREEBUSY_TEST_GRAPHQL_URL")
	if raw == "" {
		t.Skip("FREEBUSY_TEST_GRAPHQL_URL not set — live DDN integration tests skipped")
	}
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse %s: %v", raw, err)
	}
	svc, err := freebusyql.Connect(u)
	if err != nil {
		t.Fatalf("connect %s: %v", raw, err)
	}
	return svc
}

// seedUnit inserts a fresh organisation → property → unit chain and returns
// the property and unit resource names.
func seedUnit(t *testing.T, svc *freebusyql.Service) (propertyName, unitName string) {
	t.Helper()
	ctx := context.Background()
	now := tsToStr(timestamppb.New(time.Now().UTC()))
	orgID, propID, unitID := ulid.GenerateString(), ulid.GenerateString(), ulid.GenerateString()

	if _, err := svc.Mutation.Organisation.Resource.Create(ctx, orgresourceql.CreateInput{
		Id:          orgID,
		Name:        "organisations/" + orgID,
		DisplayName: "it-org",
		UpdateTime:  now,
	}); err != nil {
		t.Fatalf("seed organisation: %v", err)
	}
	propertyName = "properties/" + propID
	if _, err := svc.Mutation.Property.Properties.Create(ctx, propertiesql.CreateInput{
		Id:           propID,
		Name:         propertyName,
		DisplayName:  "it-property",
		Organisation: orgID,
		TimeZone:     "UTC",
		UpdateTime:   now,
	}); err != nil {
		t.Fatalf("seed property: %v", err)
	}
	unitName = propertyName + "/units/" + unitID
	if _, err := svc.Mutation.Property.Units.Create(ctx, unitsql.CreateInput{
		Id:           unitID,
		Name:         unitName,
		DisplayName:  "it-room",
		PropertyId:   propID,
		TimeZone:     "UTC",
		Capacity:     1,
		MaxOccupancy: 2,
		BookingMode:  "NIGHTLY",
		UpdateTime:   now,
	}); err != nil {
		t.Fatalf("seed unit: %v", err)
	}
	return propertyName, unitName
}

// TestLicenceLifecycleLive walks a property licence through create (with an
// inline attachment) → get → expiry_date filter → masked update → delete, and
// a unit licence through create → the DeleteUnit force guard.
func TestLicenceLifecycleLive(t *testing.T) {
	svc := liveService(t)
	repo := NewPropertyRepository(svc)
	ctx := context.Background()
	propertyName, unitName := seedUnit(t, svc)

	content := []byte("%PDF-1.4 fake fire-safety certificate\x00\x01\x02")
	created, err := repo.CreatePropertyLicence(ctx, propertyName, &propertypbv1.PropertyLicence{
		Type:             propertypbv1.LicenceType_LICENCE_TYPE_FIRE_SAFETY,
		LicenceNumber:    "FS-2026-001",
		IssuingAuthority: "it-authority",
		IssueDate:        &date.Date{Year: 2025, Month: 7, Day: 1},
		ExpiryDate:       &date.Date{Year: 2026, Month: 8, Day: 1},
		Attachment: &sharedpbv1.Attachment{
			Filename:  "fire-noc.pdf",
			MimeType:  "application/pdf",
			SizeBytes: int64(len(content)),
			Content:   content,
		},
	})
	if err != nil {
		t.Fatalf("CreatePropertyLicence: %v", err)
	}
	if created.GetState() != propertypbv1.LicenceState_LICENCE_STATE_ACTIVE {
		t.Fatalf("created state = %v, want ACTIVE", created.GetState())
	}

	got, err := repo.GetPropertyLicence(ctx, created.GetName())
	if err != nil {
		t.Fatalf("GetPropertyLicence: %v", err)
	}
	if !bytes.Equal(got.GetAttachment().GetContent(), content) {
		t.Fatalf("attachment content did not round-trip: got %d bytes %q, want %d bytes",
			len(got.GetAttachment().GetContent()), got.GetAttachment().GetContent(), len(content))
	}
	if got.GetExpiryDate().GetYear() != 2026 || got.GetExpiryDate().GetMonth() != 8 {
		t.Fatalf("expiry_date = %v, want 2026-08-01", got.GetExpiryDate())
	}

	// The renewal-reminder query: due on/before the horizon → included.
	due, _, err := repo.ListPropertyLicences(ctx, propertyName, listParams(t, "expiry_date <= 2026-09-01"))
	if err != nil {
		t.Fatalf("ListPropertyLicences (due): %v", err)
	}
	if len(due) != 1 || due[0].GetName() != created.GetName() {
		t.Fatalf("due list = %d items, want the created licence", len(due))
	}
	// Not yet due at an earlier horizon → excluded.
	notDue, _, err := repo.ListPropertyLicences(ctx, propertyName, listParams(t, "expiry_date <= 2026-07-01"))
	if err != nil {
		t.Fatalf("ListPropertyLicences (not due): %v", err)
	}
	if len(notDue) != 0 {
		t.Fatalf("not-due list = %d items, want 0", len(notDue))
	}

	// Masked update: renew the number without touching the attachment.
	updated, err := repo.UpdatePropertyLicence(ctx, &propertypbv1.PropertyLicence{
		Name:          created.GetName(),
		LicenceNumber: "FS-2026-002",
	}, (&fieldmaskpb.FieldMask{Paths: []string{"licence_number"}}).GetPaths())
	if err != nil {
		t.Fatalf("UpdatePropertyLicence: %v", err)
	}
	if updated.GetLicenceNumber() != "FS-2026-002" {
		t.Fatalf("licence_number = %q, want FS-2026-002", updated.GetLicenceNumber())
	}
	if !bytes.Equal(updated.GetAttachment().GetContent(), content) {
		t.Fatal("masked update must preserve the attachment")
	}

	// Unit licence + the DeleteUnit force guard.
	unitLic, err := repo.CreateUnitLicence(ctx, unitName, &propertypbv1.UnitLicence{
		Type:       propertypbv1.LicenceType_LICENCE_TYPE_LIQUOR,
		ExpiryDate: &date.Date{Year: 2026, Month: 12, Day: 31},
	})
	if err != nil {
		t.Fatalf("CreateUnitLicence: %v", err)
	}
	unitDue, _, err := repo.ListUnitLicences(ctx, unitName, listParams(t, "type = LIQUOR"))
	if err != nil {
		t.Fatalf("ListUnitLicences: %v", err)
	}
	if len(unitDue) != 1 || unitDue[0].GetName() != unitLic.GetName() {
		t.Fatalf("unit licence list = %d items, want the created licence", len(unitDue))
	}
	if err := repo.DeleteUnit(ctx, unitName, false); !errors.Is(err, types.ErrInvalidArgument) {
		t.Fatalf("DeleteUnit without force = %v, want ErrInvalidArgument", err)
	}
	if err := repo.DeleteUnit(ctx, unitName, true); err != nil {
		t.Fatalf("DeleteUnit with force: %v", err)
	}
	if _, err := repo.GetUnitLicence(ctx, unitLic.GetName()); !errors.Is(err, types.ErrNotFound) {
		t.Fatalf("unit licence after forced unit delete = %v, want ErrNotFound", err)
	}

	// Delete the property licence; its attachment row goes with it.
	if err := repo.DeletePropertyLicence(ctx, created.GetName()); err != nil {
		t.Fatalf("DeletePropertyLicence: %v", err)
	}
	if _, err := repo.GetPropertyLicence(ctx, created.GetName()); !errors.Is(err, types.ErrNotFound) {
		t.Fatalf("get after delete = %v, want ErrNotFound", err)
	}
}

func listParams(t *testing.T, filter string) types.ListParams {
	t.Helper()
	conds, err := types.ParseFilter(filter)
	if err != nil {
		t.Fatalf("parse filter %q: %v", filter, err)
	}
	return types.ListParams{Filter: conds}
}
