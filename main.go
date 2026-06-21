// Command freebusy starts the freebusy API server: a HybridServer that exposes
// the PromoCode service over gRPC and an HTTP/JSON gateway, backed by either the
// GORM or Hasura database provider (selected by FREEBUSY_DB_PROVIDER).
package main

import (
	"context"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/oh-tarnished/freebusy/internal/database"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql"
	"github.com/oh-tarnished/freebusy/internal/database/repository"
	"github.com/oh-tarnished/freebusy/internal/service"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/promocode/v1/promocodepbv1"
	"github.com/oh-tarnished/freebusy/shared"
	"github.com/oh-tarnished/runtime-go/config"
	"github.com/oh-tarnished/runtime-go/grpc"
	"github.com/oh-tarnished/runtime-go/grpc/options"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	defer func() { _ = shared.Close() }()

	cfg := loadConfig()
	conn, err := openConnection(cfg)
	if err != nil {
		fatal("database connection failed: %v", err)
	}

	factory := database.NewFactory(conn)
	promoServer := service.NewPromoCodeServer(factory.PromoCodes())
	shared.Pulse.Logger.Infof("freebusy starting with %q database provider", factory.Provider())

	srv := grpc.NewHybridServer(
		options.Options{
			ServiceName:  "freebusy",
			Environment:  options.Development,
			GRPC:         options.GRPCOptions{Host: cfg.String("grpc.host"), Port: cfg.Int("grpc.port")},
			HTTP:         options.HTTPOptions{Host: cfg.String("http.host"), Port: cfg.Int("http.port")},
			EnableHTTP:   true,
			EnableHealth: true,
		},
		grpc.WithGRPCServers(func(s *grpc.GRPCServer) {
			promocodepbv1.RegisterPromoCodeServiceServer(s, promoServer)
		}),
		grpc.WithHTTPGateways(func(mux *grpc.ServeMux, endpoint string, opts []grpc.DialOption) error {
			return promocodepbv1.RegisterPromoCodeServiceHandlerFromEndpoint(context.Background(), mux, endpoint, opts)
		}),
	)

	if err := srv.Start(); err != nil {
		fatal("server failed to start: %v", err)
	}

	// Start() returns once the listeners are up — it does not block — so wait here
	// for a termination signal, then shut the servers down gracefully.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	shared.Pulse.Logger.Info("shutdown signal received, stopping servers")
	if err := srv.Stop(); err != nil {
		shared.Pulse.Logger.Errorf("graceful stop failed: %v", err)
	}
}

// appConfig declares the configuration shape and its baked-in defaults. Values
// are overridden by FREEBUSY_*-prefixed environment variables, e.g.
// FREEBUSY_GRPC_PORT, FREEBUSY_DATABASE_DSN, FREEBUSY_HASURA_URL.
type appConfig struct {
	GRPC     endpoint `koanf:"grpc"`
	HTTP     endpoint `koanf:"http"`
	Database struct {
		DSN string `koanf:"dsn"`
	} `koanf:"database"`
	Hasura struct {
		URL string `koanf:"url"`
	} `koanf:"hasura"`
}

type endpoint struct {
	Host string `koanf:"host"`
	Port int    `koanf:"port"`
}

// loadConfig builds the application config from baked-in defaults overlaid with
// FREEBUSY_*-prefixed environment variables.
func loadConfig() *config.Config {
	cfg, err := config.New(config.Options{ServiceName: "freebusy"})
	if err != nil {
		fatal("config init failed: %v", err)
	}
	defaults := appConfig{GRPC: endpoint{Host: "0.0.0.0", Port: 50051}, HTTP: endpoint{Host: "0.0.0.0", Port: 8080}}
	if err := cfg.LoadDefaults(defaults); err != nil {
		fatal("config defaults failed: %v", err)
	}
	if err := cfg.LoadEnv("FREEBUSY_"); err != nil {
		fatal("config env failed: %v", err)
	}
	return cfg
}

// openConnection opens a database handle for the provider selected by
// FREEBUSY_DB_PROVIDER. Only the chosen provider's handle is created.
func openConnection(cfg *config.Config) (*database.Connection, error) {
	if database.ProviderFromEnv() == repository.ProviderHasura {
		u, err := url.Parse(cfg.String("hasura.url"))
		if err != nil {
			return nil, err
		}
		svc, err := freebusyql.Connect(u)
		if err != nil {
			return nil, err
		}
		return &database.Connection{Hasura: svc}, nil
	}
	// NowFunc forces gorm's auto create/update timestamps to UTC so they round-trip
	// correctly even through "timestamp without time zone" columns.
	db, err := gorm.Open(postgres.Open(cfg.String("database.dsn")), &gorm.Config{
		NowFunc: func() time.Time { return time.Now().UTC() },
	})
	if err != nil {
		return nil, err
	}
	return &database.Connection{PgSQLConn: db}, nil
}

// fatal logs a fatal error, flushes telemetry, and exits non-zero. It is used
// instead of Logger.Fatalf because that calls os.Exit directly, skipping the
// deferred shared.Close() and dropping any buffered logs/traces/metrics.
func fatal(format string, args ...any) {
	_ = shared.Pulse.Logger.Errorf(format, args...)
	_ = shared.Close()
	os.Exit(1)
}
