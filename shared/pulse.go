package shared

import (
	"sync"

	"github.com/machanirobotics/pulse/pulse-go"
)

var (
	Pulse *pulse.Pulse
	once  sync.Once
)

func init() {
	once.Do(func() {
		// Automatically loads pulse.toml
		p, err := pulse.New().Build()
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
	}
	return nil
}
