// Live-suite scaffolding: service connect and seed data.
package hasura

import (
	"context"
	"github.com/oh-tarnished/freebusy/internal/service/dbutil"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/organisationql/resourceql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/propertiesql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/unitsql"
	"github.com/oh-tarnished/freebusy/internal/database/repository/repox"
	"github.com/oh-tarnished/runtime-go/ulid"
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
	now := dbutil.TsToStr(timestamppb.New(time.Now().UTC()))
	orgID, propID, unitID := ulid.GenerateString(), ulid.GenerateString(), ulid.GenerateString()

	if _, err := svc.Mutation.Organisation.Resource.Create(ctx, resourceql.CreateInput{
		Id:          orgID,
		Name:        "organisations/" + orgID,
		DisplayName: "it-org",
		CreateTime:  now,
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
		CreateTime:   now,
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
		CreateTime:   now,
		UpdateTime:   now,
	}); err != nil {
		t.Fatalf("seed unit: %v", err)
	}
	return propertyName, unitName
}

func listParams(t *testing.T, filter string) repox.ListInput {
	t.Helper()
	return repox.ListInput{Filter: filter}
}
