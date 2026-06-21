package repository

import (
	"strings"
	"testing"
)

func TestPromoCodeNameRoundTrip(t *testing.T) {
	const id = "01HZX9SUMMER25CODEXYZ"
	name, err := PromoCodeName(id)
	if err != nil {
		t.Fatalf("PromoCodeName: %v", err)
	}
	if want := "promoCodes/" + id; name != want {
		t.Fatalf("name = %q, want %q", name, want)
	}
	got, err := PromoCodeID(name)
	if err != nil {
		t.Fatalf("PromoCodeID: %v", err)
	}
	if got != id {
		t.Fatalf("id round-trip = %q, want %q", got, id)
	}
}

func TestPromoCodeIDRejectsMalformed(t *testing.T) {
	if _, err := PromoCodeID("not-a-promo-code-name"); err == nil {
		t.Fatal("expected error for malformed resource name, got nil")
	}
}

func TestResolvePromoCodeName(t *testing.T) {
	// Empty input generates a fresh id and a matching name.
	id, name, err := ResolvePromoCodeName("")
	if err != nil {
		t.Fatalf("ResolvePromoCodeName(\"\"): %v", err)
	}
	if id == "" {
		t.Fatal("expected a generated id, got empty")
	}
	if name != "promoCodes/"+id {
		t.Fatalf("name = %q, want promoCodes/%s", name, id)
	}

	// A supplied name is parsed, not regenerated.
	gotID, gotName, err := ResolvePromoCodeName("promoCodes/ABC")
	if err != nil {
		t.Fatalf("ResolvePromoCodeName(name): %v", err)
	}
	if gotID != "ABC" || gotName != "promoCodes/ABC" {
		t.Fatalf("resolve = (%q,%q), want (ABC, promoCodes/ABC)", gotID, gotName)
	}
}

func TestOfferingName(t *testing.T) {
	name, err := OfferingName("room-1", "night")
	if err != nil {
		t.Fatalf("OfferingName: %v", err)
	}
	if !strings.HasSuffix(name, "resources/room-1/offerings/night") {
		t.Fatalf("offering name = %q", name)
	}
	got, err := OfferingID(name)
	if err != nil {
		t.Fatalf("OfferingID: %v", err)
	}
	if got != "night" {
		t.Fatalf("offering id = %q, want night", got)
	}
}
