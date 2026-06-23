package shared

import (
	"log"
	"sync"

	"github.com/oh-tarnished/runtime-go/config"
)

var (
	// The global instance is unexported to prevent other packages from modifying it directly.
	globalconfig *config.Config
	// Once ensures the initialization function is called exactly one time.
	configOnce sync.Once
)

// Getconfig returns the single, shared instance of the config session.
// It initializes the session on the first call. If initialization fails,
// the program will exit with a fatal error.
func Getconfig() *config.Config {
	// once.Do will execute the function only on the very first call to Getconfig.
	// All subsequent calls will skip the function but still get the result.
	configOnce.Do(func() {
		// Initialize the session with a default configuration.
		// You can customize the BasePath or other options here.
		sy, err := config.New(config.Options{
			BasePath:   ".",
			YamlParser: config.KoanfYamlParser,
			JsonParser: config.KoanfJsonParser,
			TomlParser: config.KoanfTomlParser,
		})
		if err != nil {
			// Using log.Fatalf is preferable to panic for initialization errors.
			log.Fatalf("FATAL: Failed to initialize global config session: %v", err)
		}
		globalconfig = sy
	})

	return globalconfig
}

// LoadTomlBytes loads raw TOML bytes into the shared config koanf state
// by writing them to a temporary file and using config's Toml.Load pipeline.
// This avoids importing koanf sub-packages directly.
func LoadTomlBytes(data []byte) error {
	const tmpFile = ".embedded_config.toml"
	sy := Getconfig()
	if err := sy.IO.WriteFile(tmpFile, data); err != nil {
		return err
	}
	defer sy.IO.DeleteFile(tmpFile) //nolint:errcheck
	return sy.Toml.Load(tmpFile)
}
