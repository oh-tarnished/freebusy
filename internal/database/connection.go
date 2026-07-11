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

// Connection carries the live backend handles plus the provider that selected
// them. Only the handle for the selected provider needs to be set (PgSQLConn
// for GORM, Hasura for Hasura); Provider tells every consumer — the per-domain
// repository factories — which one to build on, so provider selection travels
// with the connection instead of being re-read from global config.
type Connection struct {
	PgSQLConn *gorm.DB
	Hasura    *freebusyql.Service
	Provider  Provider
}

// providerFromConfig resolves the configured provider, defaulting to GORM for an
// empty or unrecognized value. Only Open consults config; everything downstream
// reads Connection.Provider.
func providerFromConfig() Provider {
	switch strings.ToLower(strings.TrimSpace(config.Get().Database.Provider)) {
	case string(ProviderHasura):
		return ProviderHasura
	default:
		return ProviderGorm
	}
}
