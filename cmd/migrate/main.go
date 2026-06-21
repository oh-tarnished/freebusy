// Command migrate is a DEV-ONLY helper: it creates the Postgres schemas and runs
// gorm AutoMigrate for the models the PromoCode service needs, so `just run` has
// tables to talk to. It is not a production migration tool — use real, versioned
// migrations there. Configuration is the same FREEBUSY_DATABASE_DSN the server reads.
//
// Scope: the booking and promocode schemas only. The resource and schedule models
// carry plain []string columns (e.g. Tags, Weekdays, CheckinWeekdays) that gorm
// AutoMigrate cannot map without a serializer/array tag, so they are intentionally
// excluded until the generator emits proper column types for those fields.
package main

import (
	"log"
	"os"

	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/booking"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/promocode"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// schemas are the Postgres schemas the migrated models live in (their TableName
// values are "<schema>.<table>"). gorm AutoMigrate does not create schemas, so
// they must exist first.
var schemas = []string{"booking", "promocode"}

// migrated returns one struct per table the PromoCode service touches: booking.moneys
// (backs the amount_off / min_subtotal value-objects) and the three promocode tables
// (the resource plus its two applicable-* join tables). The rest of the booking
// domain is intentionally excluded — e.g. booking.resource (Booking) carries a
// json.RawMessage "attributes" column that gorm AutoMigrate tries to coerce to
// bytea, which Postgres rejects when the column already exists as jsonb.
func migrated() []any {
	return []any{
		&booking.Money{},
		&promocode.PromoCode{}, &promocode.PromoCodeApplicableResources{}, &promocode.PromoCodeApplicableOfferings{},
	}
}

func main() {
	dsn := os.Getenv("FREEBUSY_DATABASE_DSN")
	if dsn == "" {
		log.Fatal("FREEBUSY_DATABASE_DSN is required (e.g. host=127.0.0.1 port=5432 user=postgres password=... dbname=freebusydb sslmode=disable)")
	}

	// Foreign-key constraints are disabled during migration so cross-schema
	// references (e.g. promocode.resource -> booking.moneys) don't impose an
	// ordering requirement on table creation.
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		log.Fatalf("open database: %v", err)
	}

	for _, schema := range schemas {
		if err := db.Exec("CREATE SCHEMA IF NOT EXISTS " + schema).Error; err != nil {
			log.Fatalf("create schema %q: %v", schema, err)
		}
		log.Printf("schema ready: %s", schema)
	}

	models := migrated()
	if err := db.AutoMigrate(models...); err != nil {
		log.Fatalf("auto-migrate: %v", err)
	}
	log.Printf("migrated %d tables across %d schemas", len(models), len(schemas))
}
