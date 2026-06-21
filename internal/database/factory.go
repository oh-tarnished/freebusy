package database

import (
	"os"
	"strings"

	gormrepo "github.com/oh-tarnished/freebusy/internal/database/gorm"
	hasurarepo "github.com/oh-tarnished/freebusy/internal/database/hasura"
	"github.com/oh-tarnished/freebusy/internal/database/repository"
)

// providerEnvVar names the environment variable that selects the database
// backend. It defaults to GORM; set it to "hasura" to use the Hasura backend.
const providerEnvVar = "FREEBUSY_DB_PROVIDER"

// Factory builds provider-agnostic repositories over a Connection. The backend is
// resolved once, at construction, from the FREEBUSY_DB_PROVIDER environment
// variable (GORM by default; Hasura opt-in). An unrecognized value falls back to
// GORM. Callers depend only on the repository interfaces, so swapping providers
// is a configuration change, not a code change.
type Factory struct {
	conn     *Connection
	provider repository.Provider
}

// NewFactory returns a Factory bound to conn, resolving the provider from the
// environment. conn must carry a live handle for the selected provider
// (PgSQLConn for GORM, Hasura for Hasura).
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

// ProviderFromEnv reports the provider selected by FREEBUSY_DB_PROVIDER. Callers
// (e.g. the application bootstrap) use it to decide which backend connection to
// open before constructing the factory.
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
