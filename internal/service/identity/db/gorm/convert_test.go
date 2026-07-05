package gorm

import (
	"testing"
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/identity"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/identity/v1/identitypbv1"
)

func TestUserFromModel(t *testing.T) {
	m := &identity.User{
		Name:        "users/u1",
		Email:       ptr("asha@example.com"),
		DisplayName: ptr("Asha"),
		AvatarURL:   ptr("https://cdn/a.png"),
		Locale:      ptr("en-IN"),
		TimeZone:    ptr("Asia/Kolkata"),
		CreateTime:  time.Now(),
		Etag:        ptr("etag-1"),
	}
	out := userFromModel(m)
	if out.GetName() != "users/u1" || out.GetEmail() != "asha@example.com" || out.GetDisplayName() != "Asha" {
		t.Fatalf("core fields not preserved: %+v", out)
	}
	if out.GetLocale() != "en-IN" || out.GetTimeZone() != "Asia/Kolkata" || out.GetEtag() != "etag-1" {
		t.Fatalf("profile fields not preserved: %+v", out)
	}
}

// A masked update touches only the selected profile field; email (IdP-owned) is
// never writable and other profile fields are left intact.
func TestApplyUserMaskPartial(t *testing.T) {
	m := &identity.User{
		Name:        "users/u1",
		Email:       ptr("asha@example.com"),
		DisplayName: ptr("Asha"),
		TimeZone:    ptr("Asia/Kolkata"),
	}
	// Client sends a new display name and (illegally) a new email; mask selects
	// only display_name.
	in := &identitypbv1.User{DisplayName: "Asha K", Email: "hacker@evil.com", TimeZone: "America/New_York"}
	applyUserMask(m, in, []string{"display_name"})

	if deref(m.DisplayName) != "Asha K" {
		t.Fatalf("display_name = %q, want Asha K", deref(m.DisplayName))
	}
	if deref(m.Email) != "asha@example.com" {
		t.Fatalf("email must not change on update, got %q", deref(m.Email))
	}
	if deref(m.TimeZone) != "Asia/Kolkata" {
		t.Fatalf("time_zone not in mask, must stay Asia/Kolkata, got %q", deref(m.TimeZone))
	}
}

// An empty mask replaces every mutable profile field (and still never email).
func TestApplyUserMaskFullReplace(t *testing.T) {
	m := &identity.User{Name: "users/u1", Email: ptr("asha@example.com"), DisplayName: ptr("Old"), Locale: ptr("en-IN")}
	in := &identitypbv1.User{DisplayName: "New", TimeZone: "UTC"} // locale cleared, tz set
	applyUserMask(m, in, nil)

	if deref(m.DisplayName) != "New" || deref(m.TimeZone) != "UTC" {
		t.Fatalf("full replace not applied: %+v", m)
	}
	if m.Locale != nil {
		t.Fatalf("locale should be cleared on full replace, got %q", deref(m.Locale))
	}
	if deref(m.Email) != "asha@example.com" {
		t.Fatalf("email must never change, got %q", deref(m.Email))
	}
}
