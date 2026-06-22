package gorm

import (
	"errors"
	"testing"

	"github.com/oh-tarnished/freebusy/internal/types"
)

func TestOrderClauseAllowlisted(t *testing.T) {
	got, err := orderClause("create_time desc, code")
	if err != nil {
		t.Fatalf("orderClause: %v", err)
	}
	if want := "create_time DESC, code ASC"; got != want {
		t.Fatalf("orderClause = %q, want %q", got, want)
	}
}

func TestOrderClauseRejectsInjection(t *testing.T) {
	// Fields outside the allowlist (including SQL-injection attempts) must be
	// rejected, never passed through to the ORDER BY clause.
	for _, in := range []string{
		"id",                 // real column, but not allowlisted
		"code; DROP TABLE x", // injection via extra tokens
		"(CASE WHEN 1=1 THEN name END)",
	} {
		if _, err := orderClause(in); !errors.Is(err, types.ErrInvalidArgument) {
			t.Fatalf("orderClause(%q) err = %v, want ErrInvalidArgument", in, err)
		}
	}
}

func TestOrderClauseEmpty(t *testing.T) {
	got, err := orderClause("")
	if err != nil || got != "" {
		t.Fatalf("orderClause(\"\") = (%q, %v), want (\"\", nil)", got, err)
	}
}
