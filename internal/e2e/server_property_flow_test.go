// Property, unit, licence, and schedule-exception flows of the server e2e
// suite.
package e2e

import (
	"context"
	"testing"

	"github.com/oh-tarnished/freebusy/protobuf/generated/go/property/v1/propertypbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/schedule/v1/schedulepbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/shared/v1/sharedpbv1"
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/genproto/googleapis/type/money"
	"google.golang.org/grpc/codes"
)

// propertyFlow creates the property + unit the later flows book against, and
// walks the licence lifecycle: the cross-field CEL rejection, target
// derivation, and the expiry_date compliance filter. Returns the unit name.
// Cleanup force-deletes the unit and removes the property row (no delete RPC —
// properties archive) via the generated repositories, LIFO before orgFlow's
// organisation delete.
func propertyFlow(t *testing.T, c *e2eClients, org string) string {
	t.Helper()
	ctx := context.Background()

	prop, err := c.props.CreateProperty(ctx, &propertypbv1.CreatePropertyRequest{
		Property: &propertypbv1.Property{
			Organisation: org,
			DisplayName:  "e2e-prop-" + c.suffix,
			TimeZone:     "UTC",
		},
	})
	if err != nil {
		t.Fatalf("CreateProperty: %v", err)
	}
	t.Cleanup(func() {
		if err := c.propRepos.Properties.Delete(ctx, prop.GetName()); err != nil {
			t.Logf("delete property row: %v", err)
		}
	})

	unit, err := c.props.CreateUnit(ctx, &propertypbv1.CreateUnitRequest{
		Parent: prop.GetName(),
		Unit: &propertypbv1.Unit{
			DisplayName:  "e2e-unit-" + c.suffix,
			Type:         propertypbv1.UnitType_UNIT_TYPE_ROOM,
			BookingMode:  sharedpbv1.BookingMode_BOOKING_MODE_NIGHTLY,
			TimeZone:     "UTC",
			MaxOccupancy: 2,
			Price:        &money.Money{CurrencyCode: "USD", Units: 100},
		},
	})
	if err != nil {
		t.Fatalf("CreateUnit: %v", err)
	}
	t.Cleanup(func() {
		if _, err := c.props.DeleteUnit(ctx, &propertypbv1.DeleteUnitRequest{Name: unit.GetName(), Force: true}); err != nil {
			t.Logf("DeleteUnit: %v", err)
		}
	})

	licenceFlow(t, c, prop.GetName(), unit.GetName())
	return unit.GetName()
}

// licenceFlow: foreign-unit CEL rejection, per-unit create with derived
// target, the expiry compliance filter, and delete.
func licenceFlow(t *testing.T, c *e2eClients, prop, unit string) {
	t.Helper()
	ctx := context.Background()

	// A licence naming a unit of a DIFFERENT property is rejected before any I/O.
	_, err := c.licences.CreateLicence(ctx, &propertypbv1.CreateLicenceRequest{
		Parent: prop,
		Licence: &propertypbv1.Licence{
			Type: propertypbv1.LicenceType_LICENCE_TYPE_FIRE_SAFETY,
			Unit: "properties/someone-else/units/u1",
		},
	})
	wantCode(t, err, codes.InvalidArgument, "create licence with foreign unit")

	lic, err := c.licences.CreateLicence(ctx, &propertypbv1.CreateLicenceRequest{
		Parent: prop,
		Licence: &propertypbv1.Licence{
			Type:       propertypbv1.LicenceType_LICENCE_TYPE_FIRE_SAFETY,
			Unit:       unit,
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
	page, err := c.licences.ListLicences(ctx, &propertypbv1.ListLicencesRequest{
		Parent: prop,
		Filter: `expiry_date <= 2027-12-31`,
	})
	if err != nil || len(page.GetLicences()) != 1 {
		t.Fatalf("ListLicences expiry filter: err=%v n=%d, want 1", err, len(page.GetLicences()))
	}
	if _, err := c.licences.DeleteLicence(ctx, &propertypbv1.DeleteLicenceRequest{Name: lic.GetName()}); err != nil {
		t.Fatalf("DeleteLicence: %v", err)
	}
}

// scheduleFlow: exception span validation, create with a date_range arm, list,
// delete.
func scheduleFlow(t *testing.T, c *e2eClients, unit string) {
	t.Helper()
	ctx := context.Background()

	_, err := c.schedules.CreateAvailabilityException(ctx, &schedulepbv1.CreateAvailabilityExceptionRequest{
		Parent: unit,
		AvailabilityException: &schedulepbv1.AvailabilityException{
			Kind: schedulepbv1.ExceptionKind_EXCEPTION_KIND_CLOSURE,
		},
	})
	wantCode(t, err, codes.InvalidArgument, "create exception without span")

	exc, err := c.schedules.CreateAvailabilityException(ctx, &schedulepbv1.CreateAvailabilityExceptionRequest{
		Parent: unit,
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
	page, err := c.schedules.ListAvailabilityExceptions(ctx, &schedulepbv1.ListAvailabilityExceptionsRequest{Parent: unit})
	if err != nil || len(page.GetAvailabilityExceptions()) != 1 {
		t.Fatalf("ListAvailabilityExceptions: err=%v n=%d, want 1", err, len(page.GetAvailabilityExceptions()))
	}
	if _, err := c.schedules.DeleteAvailabilityException(ctx, &schedulepbv1.DeleteAvailabilityExceptionRequest{Name: exc.GetName()}); err != nil {
		t.Fatalf("DeleteAvailabilityException: %v", err)
	}
}
