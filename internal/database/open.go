package database

import (
	"fmt"
	"net/url"

	"github.com/oh-tarnished/freebusy/config"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql"
	"github.com/oh-tarnished/freebusy/internal/database/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Open opens the database backend selected by config ([database].provider) and
// returns a Connection carrying the live handle for that provider. The caller
// passes the Connection to NewFactory.
func Open() (*Connection, error) {
	switch providerFromConfig() {
	case repository.ProviderHasura:
		return openHasura()
	default:
		return openGorm()
	}
}

// openGorm dials Postgres with the libpq DSN rendered from config.
func openGorm() (*Connection, error) {
	dsn := config.Get().Database.Postgres.DSN()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	return &Connection{PgSQLConn: db}, nil
}

// openHasura connects the typed GraphQL client to the configured endpoint,
// sending the admin secret as the x-hasura-admin-secret header when set.
func openHasura() (*Connection, error) {
	h := config.Get().Database.Hasura
	u, err := url.Parse(h.URL)
	if err != nil {
		return nil, fmt.Errorf("parse hasura url %q: %w", h.URL, err)
	}
	var svc *freebusyql.Service
	if h.AdminSecret != "" {
		svc, err = freebusyql.Connect(u, map[string]string{"x-hasura-admin-secret": h.AdminSecret})
	} else {
		svc, err = freebusyql.Connect(u)
	}
	if err != nil {
		return nil, fmt.Errorf("connect hasura: %w", err)
	}
	return &Connection{Hasura: svc}, nil
}
