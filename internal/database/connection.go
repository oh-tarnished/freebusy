// Package database owns the database connection plumbing: it opens the configured
// backend and holds the live handles (GORM and/or Hasura), and it selects the
// provider once from the loaded config ([database].provider; GORM by default,
// Hasura opt-in). It is domain-agnostic — each service builds its own
// provider-specific repositories from a Connection (see
// internal/service/<svc>/db), so swapping backends is a configuration change.
package database

import (
	"strings"

	"github.com/oh-tarnished/freebusy/config"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql"
	"gorm.io/gorm"
)

// Provider identifies a database backend implementation.
type Provider string

const (
	// ProviderGorm is the default backend: GORM over the relational database.
	ProviderGorm Provider = "gorm"
	// ProviderHasura is the opt-in backend: Hasura GraphQL.
	ProviderHasura Provider = "hasura"
)

// Connection carries the live backend handles. Only the handle for the selected
// provider needs to be set (PgSQLConn for GORM, Hasura for Hasura).
type Connection struct {
	PgSQLConn *gorm.DB
	Hasura    *freebusyql.Service
}

// ProviderFromConfig reports the provider selected by [database].provider in the
// loaded config. Service db factories use it to pick which handle to build their
// repository over; the bootstrap uses it to decide which backend to open.
func ProviderFromConfig() Provider { return providerFromConfig() }

// providerFromConfig resolves the configured provider, defaulting to GORM for an
// empty or unrecognized value.
func providerFromConfig() Provider {
	if testBackend != nil {
		return testBackend.provider
	}
	switch strings.ToLower(strings.TrimSpace(config.Get().Database.Provider)) {
	case string(ProviderHasura):
		return ProviderHasura
	default:
		return ProviderGorm
	}
}

// testBackend, when set, overrides both Open and providerFromConfig: the e2e
// suites inject an explicit connection and provider so the fully assembled
// server runs against the test backend without touching the loaded config.
var testBackend *testOverride

type testOverride struct {
	conn     *Connection
	provider Provider
}

// SetTestBackend forces Open to return conn and ProviderFromConfig to report
// provider until the returned restore func runs. Test seam only — it mutates
// package state, so don't call it concurrently with Open or with itself.
func SetTestBackend(conn *Connection, p Provider) (restore func()) {
	prev := testBackend
	testBackend = &testOverride{conn: conn, provider: p}
	return func() { testBackend = prev }
}
