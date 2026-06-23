// Package shared exposes process-wide singletons used across the freebusy service.
package shared

import (
	_ "embed"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"

	"github.com/machanirobotics/pulse/pulse-go"
	"github.com/machanirobotics/pulse/pulse-go/options"
)

//go:embed pulse.toml
var pulseConfigData []byte

var (
	Pulse *pulse.Pulse // Pulse is the singleton Pulse client initialised at package load time.
	once  sync.Once    // once is used to ensure that the Pulse client is initialised only once.
)

func init() {
	once.Do(func() {
		tmpPath, err := writePulseConfig()
		if err != nil {
			panic(fmt.Sprintf("failed to write embedded pulse config: %v", err))
		}
		defer os.Remove(tmpPath) //nolint:errcheck

		b := pulse.New().WithConfig(tmpPath)
		applyDeploymentOverrides(b)

		p, err := b.Build()
		if err != nil {
			fmt.Printf("ERROR: Failed to create Pulse: %v\n", err)
			panic(err)
		}

		Pulse = p
	})
}

// applyDeploymentOverrides layers per-deployment values over the embedded
// pulse.toml defaults, so one binary serves every environment and hotel
// property without a rebuild. Each override is optional — an unset variable
// leaves the embedded default in place.
//
//	FREEBUSY_TELEMETRY_ENVIRONMENT  deployment environment (development|staging|production)
//	FREEBUSY_OTLP_ENDPOINT          OTLP collector as host:port
//	FREEBUSY_PROPERTY               property label slug (e.g. "doubletree-del-mar-san-diego")
//	FREEBUSY_PROPERTY_NAME          human-readable property name
func applyDeploymentOverrides(b *pulse.Builder) {
	if env := os.Getenv("FREEBUSY_TELEMETRY_ENVIRONMENT"); env != "" {
		b.WithEnvironment(options.Environment(env))
	}
	if endpoint := os.Getenv("FREEBUSY_OTLP_ENDPOINT"); endpoint != "" {
		if host, port, ok := splitHostPort(endpoint); ok {
			b.WithOTLP(host, port)
		} else {
			fmt.Printf("WARNING: ignoring malformed FREEBUSY_OTLP_ENDPOINT %q (want host:port)\n", endpoint)
		}
	}
	labels := map[string]string{}
	if v := os.Getenv("FREEBUSY_PROPERTY"); v != "" {
		labels["property"] = v
	}
	if v := os.Getenv("FREEBUSY_PROPERTY_NAME"); v != "" {
		labels["property_name"] = v
	}
	if len(labels) > 0 {
		b.WithLabels(labels)
	}
}

// splitHostPort splits a "host:port" OTLP endpoint into its parts.
func splitHostPort(endpoint string) (host string, port int, ok bool) {
	h, portStr, err := net.SplitHostPort(endpoint)
	if err != nil {
		return "", 0, false
	}
	p, err := strconv.Atoi(portStr)
	if err != nil {
		return "", 0, false
	}
	return h, p, true
}

// writePulseConfig writes the embedded pulse.toml to a temporary file
// and returns its path. The caller is responsible for cleanup.
func writePulseConfig() (string, error) {
	f, err := os.CreateTemp("", "pulse-*.toml")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	if _, err := f.Write(pulseConfigData); err != nil {
		f.Close()
		os.Remove(f.Name())
		return "", fmt.Errorf("write pulse config: %w", err)
	}
	if err := f.Close(); err != nil {
		os.Remove(f.Name())
		return "", fmt.Errorf("close temp file: %w", err)
	}
	return f.Name(), nil
}

// Close releases resources held by the Pulse client.
// It is safe to call even when Pulse was never initialised.
// The caller (typically main) should invoke this on application shutdown.
func Close() error {
	if Pulse != nil {
		return Pulse.Close()
	}
	return nil
}
