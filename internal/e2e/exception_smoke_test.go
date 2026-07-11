// Availability-exception smoke: the generated value-object lifecycle (oneof span create, hydration, mask-switch, cleanup) on both providers.
package e2e

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/repository/freebusy/organisation"
	"github.com/oh-tarnished/freebusy/internal/database/repository/freebusy/property"
	"github.com/oh-tarnished/freebusy/internal/database/repository/freebusy/schedule"
	"github.com/oh-tarnished/freebusy/internal/database/repository/repox"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/organisation/v1/orgpbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/property/v1/propertypbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/schedule/v1/schedulepbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/shared/v1/sharedpbv1"
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TestRepositorySmoke_ScheduleExceptionGorm proves the generated value-object
// path live on Postgres: the exception's oneof span (TimeWindow / DateRange)
// is created, preloaded, mask-switched, and cleaned up entirely by generated
// code — including the property_id/unit_id ancestor wiring the FKs enforce.
func TestRepositorySmoke_ScheduleExceptionGorm(t *testing.T) {
	db := openTestGorm(t)
	exceptionLifecycle(t, organisation.NewGorm(db), property.NewGorm(db), schedule.NewGorm(db))
}

// TestRepositorySmoke_ScheduleExceptionGraphQL runs the identical lifecycle
// through the GraphQL adapters (insert-then-reference, follow-up hydration,
// patch replacement) against a live DDN engine.
func TestRepositorySmoke_ScheduleExceptionGraphQL(t *testing.T) {
	svc := connectTestGraphQL(t)
	exceptionLifecycle(t, organisation.NewGraphQL(svc), property.NewGraphQL(svc), schedule.NewGraphQL(svc))
}

// exceptionLifecycle drives org → property → unit → availability exception,
// asserting the span value object round-trips, switches oneof arms under a
// field mask, and disappears with its owner.
func exceptionLifecycle(t *testing.T, orgRepos organisation.Repositories, propRepos property.Repositories, schedRepos schedule.Repositories) {
	t.Helper()
	ctx := context.Background()

	org, err := orgRepos.Organisations.Create(ctx, &orgpbv1.Organisation{DisplayName: "smoke-vo-org"})
	if err != nil {
		t.Fatalf("Create org: %v", err)
	}
	defer orgRepos.Organisations.Delete(ctx, org.GetName())

	prop, err := propRepos.Properties.Create(ctx, &propertypbv1.Property{
		Organisation: org.GetName(),
		DisplayName:  "smoke-vo-prop",
		TimeZone:     "UTC",
	})
	if err != nil {
		t.Fatalf("Create property: %v", err)
	}
	defer propRepos.Properties.Delete(ctx, prop.GetName())

	unit, err := propRepos.Units.Create(ctx, prop.GetName(), &propertypbv1.Unit{
		DisplayName: "smoke-vo-unit",
		Type:        propertypbv1.UnitType_UNIT_TYPE_ROOM,
		BookingMode: sharedpbv1.BookingMode_BOOKING_MODE_NIGHTLY,
		TimeZone:    "UTC",
	})
	if err != nil {
		t.Fatalf("Create unit: %v", err)
	}
	defer propRepos.Units.Delete(ctx, unit.GetName())

	start := time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC)
	end := start.Add(4 * time.Hour)
	exc, err := schedRepos.AvailabilityExceptions.Create(ctx, unit.GetName(), &schedulepbv1.AvailabilityException{
		Kind:   schedulepbv1.ExceptionKind_EXCEPTION_KIND_CLOSURE,
		Reason: "maintenance",
		Span: &schedulepbv1.AvailabilityException_Window{Window: &sharedpbv1.TimeWindow{
			StartTime: timestamppb.New(start),
			EndTime:   timestamppb.New(end),
		}},
	})
	if err != nil {
		t.Fatalf("Create exception: %v", err)
	}
	defer schedRepos.AvailabilityExceptions.Delete(ctx, exc.GetName())
	if !strings.HasPrefix(exc.GetName(), unit.GetName()+"/availabilityExceptions/") {
		t.Fatalf("exception name = %q, want under %q", exc.GetName(), unit.GetName())
	}
	if w := exc.GetWindow(); w == nil || !w.GetStartTime().AsTime().Equal(start) || !w.GetEndTime().AsTime().Equal(end) {
		t.Fatalf("created span window = %v, want [%v, %v]", exc.GetWindow(), start, end)
	}

	// Get re-hydrates the span from its own row.
	got, err := schedRepos.AvailabilityExceptions.Get(ctx, exc.GetName())
	if err != nil {
		t.Fatalf("Get exception: %v", err)
	}
	if w := got.GetWindow(); w == nil || !w.GetStartTime().AsTime().Equal(start) {
		t.Fatalf("Get span window = %v, want start %v", got.GetWindow(), start)
	}

	// A masked update switches the oneof arm: the old window row is replaced
	// by a fresh date-range row.
	updated, err := schedRepos.AvailabilityExceptions.Update(ctx, &schedulepbv1.AvailabilityException{
		Name: exc.GetName(),
		Span: &schedulepbv1.AvailabilityException_DateRange{DateRange: &sharedpbv1.DateRange{
			StartDate: &date.Date{Year: 2026, Month: 8, Day: 1},
			EndDate:   &date.Date{Year: 2026, Month: 8, Day: 5},
		}},
	}, []string{"date_range"})
	if err != nil {
		t.Fatalf("Update exception span: %v", err)
	}
	if updated.GetWindow() != nil {
		t.Fatalf("updated span still carries window: %v", updated.GetWindow())
	}
	if dr := updated.GetDateRange(); dr == nil || dr.GetStartDate().GetDay() != 1 || dr.GetEndDate().GetDay() != 5 {
		t.Fatalf("updated span date_range = %v, want Aug 1–5", updated.GetDateRange())
	}

	// List under the unit finds it.
	page, _, err := schedRepos.AvailabilityExceptions.List(ctx, unit.GetName(), repox.ListInput{})
	if err != nil {
		t.Fatalf("List exceptions: %v", err)
	}
	if len(page) != 1 {
		t.Fatalf("List exceptions = %d items, want 1", len(page))
	}

	// Delete removes the exception and its span row.
	if err := schedRepos.AvailabilityExceptions.Delete(ctx, exc.GetName()); err != nil {
		t.Fatalf("Delete exception: %v", err)
	}
	if _, err := schedRepos.AvailabilityExceptions.Get(ctx, exc.GetName()); !errors.Is(err, repox.ErrNotFound) {
		t.Fatalf("Get after delete err = %v, want repox.ErrNotFound", err)
	}
}
