// Command migrate is a DEV-ONLY helper: it creates the Postgres schemas and runs
// gorm AutoMigrate for every generated freebusy model (all domains), so `just run`
// has tables to talk to. It is not a production migration tool — use real,
// versioned migrations there. Configuration is the same FREEBUSY_DATABASE_DSN the
// server reads.
//
// It delegates to the generated freebusy.Default registry (protorm emits it with
// every model + EnsureSchemas), so new entities are covered automatically on the
// next `just orm`.
package main

import (
	"log"
	"os"

	freebusy "github.com/oh-tarnished/freebusy/protobuf/generated/gorm/freebusy"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := os.Getenv("FREEBUSY_DATABASE_DSN")
	if dsn == "" {
		log.Fatal("FREEBUSY_DATABASE_DSN is required (e.g. host=127.0.0.1 port=5432 user=postgres password=... dbname=freebusydb sslmode=disable)")
	}

	// FK constraints are disabled during migration so cross-schema references
	// (e.g. promocode.resource -> booking.moneys) don't impose a table-creation
	// ordering.
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		log.Fatalf("open database: %v", err)
	}

	if err := freebusy.Default.EnsureSchemas(db); err != nil {
		log.Fatalf("ensure schemas: %v", err)
	}
	if err := freebusy.Default.Migrate(db); err != nil {
		log.Fatalf("auto-migrate: %v", err)
	}
	log.Printf("migrated %d models across all freebusy schemas", len(freebusy.Default.Models()))
}
