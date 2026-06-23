// Package database owns the database connection and the provider factory: it
// holds the live backend handles (GORM and/or Hasura) and builds the
// provider-agnostic repositories the service layer depends on. The provider is
// selected once from the loaded config ([database].provider; GORM by default,
// Hasura opt-in), so swapping backends is a configuration change, not a code
// change.
package database

import (
	"strings"

	"github.com/oh-tarnished/freebusy/config"
	"github.com/oh-tarnished/freebusy/internal/database/gorm"
	"github.com/oh-tarnished/freebusy/internal/database/hasura"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql"
	"github.com/oh-tarnished/freebusy/internal/database/repository"
)

// Connection carries the live backend handles. Only the handle for the selected
// provider needs to be set (PgSQLConn for GORM, Hasura for Hasura).
type Connection struct {
	PgSQLConn *gorm.DB
	Hasura    *freebusyql.Service
}

// Factory builds provider-agnostic repositories over a Connection. The backend is
// resolved once, at construction, from [database].provider in config. Callers
// depend only on the repository interfaces.
type Factory struct {
	conn     *Connection
	provider repository.Provider
}

// NewFactory returns a Factory bound to conn, resolving the provider from config.
func NewFactory(conn *Connection) *Factory {
	return &Factory{conn: conn, provider: providerFromConfig()}
}

// Provider reports the backend this factory builds repositories for.
func (f *Factory) Provider() repository.Provider { return f.provider }

// PromoCodes returns the PromoCodeRepository for the selected provider.
func (f *Factory) PromoCodes() repository.PromoCodeRepository {
	if f.provider == repository.ProviderHasura {
		return hasura.NewPromoCodeRepository(f.conn.Hasura)
	}
	return gorm.NewPromoCodeRepository(f.conn.PgSQLConn)
}

// ProviderFromConfig reports the provider selected by [database].provider in the
// loaded config. The bootstrap uses it to decide which backend connection to
// open before constructing the factory.
func ProviderFromConfig() repository.Provider { return providerFromConfig() }

// providerFromConfig resolves the configured provider, defaulting to GORM for an
// empty or unrecognized value.
func providerFromConfig() repository.Provider {
	switch strings.ToLower(strings.TrimSpace(config.Get().Database.Provider)) {
	case string(repository.ProviderHasura):
		return repository.ProviderHasura
	default:
		return repository.ProviderGorm
	}
}
