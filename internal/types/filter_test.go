package types

import (
	"errors"
	"testing"
)

func TestParseFilter(t *testing.T) {
	cases := []struct {
		name   string
		filter string
		want   []FilterCondition
	}{
		{"empty", "", nil},
		{"whitespace only", "   ", nil},
		{"equality", "state = ACTIVE", []FilterCondition{{Field: "state", Op: FilterEq, Value: "ACTIVE"}}},
		{"inequality", "state != EXPIRED", []FilterCondition{{Field: "state", Op: FilterNeq, Value: "EXPIRED"}}},
		{"has substring", "code : SUM", []FilterCondition{{Field: "code", Op: FilterHas, Value: "SUM"}}},
		{"quoted value", `display_name : "Summer Sale"`, []FilterCondition{{Field: "display_name", Op: FilterHas, Value: "Summer Sale"}}},
		{"free text bareword", "SUMMER25", []FilterCondition{{Op: FilterHas, Value: "SUMMER25"}}},
		{"free text quoted", `"big sale"`, []FilterCondition{{Op: FilterHas, Value: "big sale"}}},
		{"two terms with AND", "state = ACTIVE AND code : SUM", []FilterCondition{
			{Field: "state", Op: FilterEq, Value: "ACTIVE"},
			{Field: "code", Op: FilterHas, Value: "SUM"},
		}},
		{"two terms space separated", "disabled = false code : X", []FilterCondition{
			{Field: "disabled", Op: FilterEq, Value: "false"},
			{Field: "code", Op: FilterHas, Value: "X"},
		}},
		{"no space around operator", "state=ACTIVE", []FilterCondition{{Field: "state", Op: FilterEq, Value: "ACTIVE"}}},
		{"less than or equal", "expiry_date <= 2026-08-01", []FilterCondition{{Field: "expiry_date", Op: FilterLte, Value: "2026-08-01"}}},
		{"greater than or equal", "expiry_date >= 2026-08-01", []FilterCondition{{Field: "expiry_date", Op: FilterGte, Value: "2026-08-01"}}},
		{"no space around lte", "expiry_date<=2026-08-01", []FilterCondition{{Field: "expiry_date", Op: FilterLte, Value: "2026-08-01"}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseFilter(tc.filter)
			if err != nil {
				t.Fatalf("ParseFilter(%q): %v", tc.filter, err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("got %d conditions %+v, want %d %+v", len(got), got, len(tc.want), tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("condition %d = %+v, want %+v", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestParseFilterErrors(t *testing.T) {
	for _, filter := range []string{
		"= ACTIVE",                 // leading operator
		"state =",                  // missing value
		"state = = ACTIVE",         // operator where value expected
		`code : "unbalanced`,       // unterminated quote
		"code ! ACTIVE",            // lone bang
		"expiry_date < 2026-08-01", // lone less-than
	} {
		t.Run(filter, func(t *testing.T) {
			if _, err := ParseFilter(filter); !errors.Is(err, ErrInvalidArgument) {
				t.Fatalf("ParseFilter(%q) error = %v, want ErrInvalidArgument", filter, err)
			}
		})
	}
}
