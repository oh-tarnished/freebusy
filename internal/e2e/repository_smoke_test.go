// Package e2e holds tests that exercise assembled layers against live
// backends. This file smoke-tests the GENERATED repository layer (the
// target=repository output) against a real Postgres: name minting, converter
// completeness, reference mapping, etag optimistic concurrency, field-mask
// updates, and filterx-driven lists — the whole generated write/read path with
// zero hand-written persistence code in the loop.
//
// Gated like the other live suites:
//
//	FREEBUSY_TEST_POSTGRES_DSN="postgresql://postgres:postgrespassword@localhost:5432/freebusydb" \
//	  go test ./internal/e2e/ -run RepositorySmoke -v
//
// (run `just migrate` first so the schemas exist).
package e2e

import (
	"context"
	"errors"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql"
	"github.com/oh-tarnished/freebusy/internal/database/repository/freebusy/identity"
	"github.com/oh-tarnished/freebusy/internal/database/repository/freebusy/organisation"
	"github.com/oh-tarnished/freebusy/internal/database/repository/repox"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/identity/v1/identitypbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/organisation/v1/orgpbv1"
	"google.golang.org/protobuf/types/known/structpb"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestRepositorySmoke_OrganisationGorm(t *testing.T) {
	dsn := os.Getenv("FREEBUSY_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("FREEBUSY_TEST_POSTGRES_DSN not set — live repository smoke test skipped")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:         logger.Default.LogMode(logger.Silent),
		TranslateError: true, // sentinel errors (ErrDuplicatedKey) like database.Open
	})
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	orgLifecycle(t, organisation.NewGorm(db), identity.NewGorm(db))
}

// TestRepositorySmoke_OrganisationGraphQL runs the exact same lifecycle
// through the GraphQL adapters against a live DDN engine — one behavior, two
// generated backends.
func TestRepositorySmoke_OrganisationGraphQL(t *testing.T) {
	raw := os.Getenv("FREEBUSY_TEST_GRAPHQL_URL")
	if raw == "" {
		t.Skip("FREEBUSY_TEST_GRAPHQL_URL not set — live repository smoke test skipped")
	}
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse %s: %v", raw, err)
	}
	svc, err := freebusyql.Connect(u)
	if err != nil {
		t.Fatalf("connect %s: %v", raw, err)
	}
	orgLifecycle(t, organisation.NewGraphQL(svc), identity.NewGraphQL(svc))
}

// orgLifecycle drives the full organisation + member lifecycle through
// whichever adapter set it is handed.
func orgLifecycle(t *testing.T, repos organisation.Repositories, idRepos identity.Repositories) {
	t.Helper()
	ctx := context.Background()

	// Create: server-minted name, etag set, stored record returned.
	created, err := repos.Organisations.Create(ctx, &orgpbv1.Organisation{
		DisplayName: "smoke-org",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if !strings.HasPrefix(created.GetName(), "organisations/") {
		t.Fatalf("created name = %q, want organisations/*", created.GetName())
	}
	if created.GetEtag() == "" {
		t.Fatal("created etag is empty")
	}
	if created.GetDisplayName() != "smoke-org" {
		t.Fatalf("display_name = %q", created.GetDisplayName())
	}
	defer repos.Organisations.Delete(ctx, created.GetName())

	// Get round-trips.
	got, err := repos.Organisations.Get(ctx, created.GetName())
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.GetDisplayName() != "smoke-org" {
		t.Fatalf("Get display_name = %q", got.GetDisplayName())
	}

	// List with a generated filter finds it.
	page, _, err := repos.Organisations.List(ctx, repox.ListInput{
		Filter: `display_name = "smoke-org"`,
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(page) == 0 {
		t.Fatal("List: created organisation not found by filter")
	}

	// Masked update touches only the masked field; message field via mask too.
	settings, _ := structpb.NewStruct(map[string]any{"theme": "dark"})
	updated, err := repos.Organisations.Update(ctx, &orgpbv1.Organisation{
		Name:        created.GetName(),
		DisplayName: "smoke-org-2",
		Settings:    settings,
	}, []string{"display_name", "settings"})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.GetDisplayName() != "smoke-org-2" {
		t.Fatalf("updated display_name = %q", updated.GetDisplayName())
	}
	if updated.GetSettings().GetFields()["theme"].GetStringValue() != "dark" {
		t.Fatalf("updated settings = %v", updated.GetSettings())
	}
	if updated.GetEtag() == created.GetEtag() {
		t.Fatal("update must bump the etag")
	}

	// Stale etag conflicts.
	if _, err := repos.Organisations.Update(ctx, &orgpbv1.Organisation{
		Name:        created.GetName(),
		DisplayName: "stale-write",
		Etag:        created.GetEtag(),
	}, []string{"display_name"}); !errors.Is(err, repox.ErrConflict) {
		t.Fatalf("stale-etag update err = %v, want repox.ErrConflict", err)
	}

	// Parented child: member under the organisation, with a real user
	// reference (the FK is enforced) created through the generated identity
	// repository — two schemas' generated adapters cooperating.
	user, err := idRepos.Users.Create(ctx, &identitypbv1.User{DisplayName: "smoke-user"})
	if err != nil {
		t.Fatalf("Create user: %v", err)
	}
	defer idRepos.Users.Delete(ctx, user.GetName())
	member, err := repos.Members.Create(ctx, created.GetName(), &orgpbv1.Member{
		Email: "smoke@example.com",
		User:  user.GetName(),
		Role:  orgpbv1.OrganisationRole_ORGANISATION_ROLE_MEMBER,
	})
	if err != nil {
		t.Fatalf("Create member: %v", err)
	}
	if !strings.HasPrefix(member.GetName(), created.GetName()+"/members/") {
		t.Fatalf("member name = %q, want under %q", member.GetName(), created.GetName())
	}
	if member.GetUser() != user.GetName() {
		t.Fatalf("member user ref = %q, want %q (ref round-trip broken)", member.GetUser(), user.GetName())
	}
	members, _, err := repos.Members.List(ctx, created.GetName(), repox.ListInput{})
	if err != nil {
		t.Fatalf("List members: %v", err)
	}
	if len(members) != 1 {
		t.Fatalf("member list = %d items, want 1", len(members))
	}
	if err := repos.Members.Delete(ctx, member.GetName()); err != nil {
		t.Fatalf("Delete member: %v", err)
	}

	// Delete + NotFound.
	if err := repos.Organisations.Delete(ctx, created.GetName()); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := repos.Organisations.Get(ctx, created.GetName()); !errors.Is(err, repox.ErrNotFound) {
		t.Fatalf("Get after delete err = %v, want repox.ErrNotFound", err)
	}
}
