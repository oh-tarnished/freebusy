// Package database owns the database connection and the provider factory: it
// holds the live backend handles (GORM and/or Hasura) and builds the
// provider-agnostic repositories the service layer depends on. The provider is
// selected once from FREEBUSY_DB_PROVIDER (GORM by default; Hasura opt-in), so
// swapping backends is a configuration change, not a code change.
package database

import (
	"os"
	"strings"

	gormrepo "github.com/oh-tarnished/freebusy/internal/database/gorm"
	hasurarepo "github.com/oh-tarnished/freebusy/internal/database/hasura"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql"
	"github.com/oh-tarnished/freebusy/internal/database/repository"
	"gorm.io/gorm"
)

// providerEnvVar names the environment variable that selects the database
// backend. It defaults to GORM; set it to "hasura" to use the Hasura backend.
const providerEnvVar = "FREEBUSY_DB_PROVIDER"

// Connection carries the live backend handles. Only the handle for the selected
// provider needs to be set (PgSQLConn for GORM, Hasura for Hasura).
type Connection struct {
	PgSQLConn *gorm.DB
	Hasura    *freebusyql.Service
}

// Factory builds provider-agnostic repositories over a Connection. The backend is
// resolved once, at construction, from FREEBUSY_DB_PROVIDER. Callers depend only
// on the repository interfaces.
type Factory struct {
	conn     *Connection
	provider repository.Provider
}

// NewFactory returns a Factory bound to conn, resolving the provider from the
// environment.
func NewFactory(conn *Connection) *Factory {
	return &Factory{conn: conn, provider: providerFromEnv()}
}

// Provider reports the backend this factory builds repositories for.
func (f *Factory) Provider() repository.Provider { return f.provider }

// PromoCodes returns the PromoCodeRepository for the selected provider.
func (f *Factory) PromoCodes() repository.PromoCodeRepository {
	if f.provider == repository.ProviderHasura {
		return hasurarepo.NewPromoCodeRepository(f.conn.Hasura)
	}
	return gormrepo.NewPromoCodeRepository(f.conn.PgSQLConn)
}

// ProviderFromEnv reports the provider selected by FREEBUSY_DB_PROVIDER. The
// bootstrap uses it to decide which backend connection to open before
// constructing the factory.
func ProviderFromEnv() repository.Provider { return providerFromEnv() }

// providerFromEnv resolves the configured provider, defaulting to GORM for an
// empty or unrecognized value.
func providerFromEnv() repository.Provider {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(providerEnvVar))) {
	case string(repository.ProviderHasura):
		return repository.ProviderHasura
	default:
		return repository.ProviderGorm
	}
}
