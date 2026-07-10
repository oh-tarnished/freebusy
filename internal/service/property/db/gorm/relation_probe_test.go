package gorm

import (
	"sync"
	"testing"

	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/property"
	"gorm.io/gorm/schema"
)

func TestUnitLicencesRelationParses(t *testing.T) {
	s, err := schema.Parse(&property.Unit{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatalf("parse Unit schema: %v", err)
	}
	rel, ok := s.Relationships.Relations["Licences"]
	if !ok {
		t.Fatal("Unit has no Licences relation")
	}
	if len(rel.References) != 1 {
		t.Fatalf("Licences relation has %d references, want 1", len(rel.References))
	}
	ref := rel.References[0]
	if ref.ForeignKey.DBName != "unit" {
		t.Fatalf("Licences FK column = %q, want %q", ref.ForeignKey.DBName, "unit")
	}
}
