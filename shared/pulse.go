package shared

import (
	"sync"

	"github.com/machanirobotics/pulse/pulse-go"
)

var (
	Pulse *pulse.Pulse // once ensures that the pulse client is only initialized once
	once  sync.Once    // sync.Once is used to ensure that the pulse client is only initialized once
)

func init() {
	once.Do(func() {
		// Automatically loads pulse.toml
		p, err := pulse.New().WithConfig("./pulse.toml").Build()
		if err != nil {
			panic(err)
		}
		Pulse = p
	})
}

// Close should be called by the main application on shutdown
func Close() error {
	if Pulse != nil {
		Pulse.Close()
		Pulse = nil
	}
	return nil
}
