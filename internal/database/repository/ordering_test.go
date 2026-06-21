package repository

import (
	"errors"
	"testing"
)

func TestParseOrderBy(t *testing.T) {
	got, err := ParseOrderBy("create_time desc, code")
	if err != nil {
		t.Fatalf("ParseOrderBy: %v", err)
	}
	want := []OrderTerm{{Field: "create_time", Desc: true}, {Field: "code", Desc: false}}
	if len(got) != len(want) {
		t.Fatalf("got %d terms, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("term %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestParseOrderByEmpty(t *testing.T) {
	got, err := ParseOrderBy("   ")
	if err != nil || got != nil {
		t.Fatalf("empty order_by = (%v, %v), want (nil, nil)", got, err)
	}
}

func TestParseOrderByRejectsMalformed(t *testing.T) {
	for _, in := range []string{"code asc desc", "code sideways", "name; DROP TABLE x"} {
		if _, err := ParseOrderBy(in); !errors.Is(err, ErrInvalidArgument) {
			t.Fatalf("ParseOrderBy(%q) err = %v, want ErrInvalidArgument", in, err)
		}
	}
}
