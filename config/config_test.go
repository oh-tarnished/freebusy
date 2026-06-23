package config

import (
	"strings"
	"testing"
)

// TestConfigLoadsReleaseDefaults verifies the embedded release config parses and
// validates. During tests the working directory is the package dir, so the dev
// overlay (config/freebusy.dev.toml, resolved relative to the repo root) is not
// loaded — these assertions reflect the embedded release defaults.
func TestConfigLoadsReleaseDefaults(t *testing.T) {
	cfg := Get()
	if cfg.Meta.Name != "freebusy" {
		t.Errorf("Meta.Name = %q, want %q", cfg.Meta.Name, "freebusy")
	}
	if cfg.Meta.Version == "" {
		t.Error("Meta.Version is empty")
	}
	if cfg.Database.Provider != "gorm" {
		t.Errorf("Database.Provider = %q, want %q", cfg.Database.Provider, "gorm")
	}
	if got := cfg.Database.Postgres.DSN(); !strings.Contains(got, "dbname=freebusydb") {
		t.Errorf("DSN() = %q, want it to contain dbname=freebusydb", got)
	}
	if cfg.Server.Environment != "production" {
		t.Errorf("Server.Environment = %q, want %q", cfg.Server.Environment, "production")
	}
	if cfg.Server.GRPC.Port != 50051 {
		t.Errorf("Server.GRPC.Port = %d, want %d", cfg.Server.GRPC.Port, 50051)
	}
	if !cfg.Server.EnableMCP || cfg.Server.MCP.Transport != "streamable-http" {
		t.Errorf("Server.MCP = %+v, want EnableMCP=true transport=streamable-http", cfg.Server.MCP)
	}
}

// TestGetGRPCOptions verifies the config maps cleanly onto the runtime options.
func TestGetGRPCOptions(t *testing.T) {
	opts := GetGRPCOptions()
	if opts.ServiceName != "freebusy" {
		t.Errorf("ServiceName = %q, want freebusy", opts.ServiceName)
	}
	if opts.GRPC.Host != "0.0.0.0" || opts.GRPC.Port != 50051 {
		t.Errorf("GRPC = %+v, want host=0.0.0.0 port=50051", opts.GRPC)
	}
	if !opts.EnableHTTP || !opts.EnableHealth || !opts.EnableMCP {
		t.Errorf("enable flags = http:%v health:%v mcp:%v, want all true",
			opts.EnableHTTP, opts.EnableHealth, opts.EnableMCP)
	}
	if string(opts.Environment) != "production" {
		t.Errorf("Environment = %q, want production", opts.Environment)
	}
}
