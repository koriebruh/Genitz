package generator

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds genitz's persistent user-level defaults — set via
// `genitz config set <key> <value>`, applied by runInitFlags when the
// corresponding flag isn't passed.
type Config struct {
	License      string `json:"license"`
	Author       string `json:"author"`
	ModulePrefix string `json:"modulePrefix"`
}

const configPathEnvVar = "GENITZ_CONFIG_OVERRIDE"

func configPath() string {
	if p := os.Getenv(configPathEnvVar); p != "" {
		return p
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "genitz", "config.json")
}

// LoadConfig reads the persisted config, returning a zero Config (not an
// error) when no file exists yet — an unconfigured genitz is the default
// state, not a failure.
func LoadConfig() (Config, error) {
	path := configPath()
	if path == "" {
		return Config{}, nil
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return Config{}, nil
	}
	var cfg Config
	if err := json.Unmarshal(content, &cfg); err != nil {
		return Config{}, fmt.Errorf("invalid config at %s: %w", path, err)
	}
	return cfg, nil
}

// Save persists c to disk, creating the config directory if needed.
func (c Config) Save() error {
	path := configPath()
	if path == "" {
		return errors.New("could not determine user config directory")
	}
	content, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, content, 0o644)
}

// Get returns the value for a known key, or "" for an unknown one.
func (c Config) Get(key string) string {
	switch key {
	case "license":
		return c.License
	case "author":
		return c.Author
	case "modulePrefix":
		return c.ModulePrefix
	default:
		return ""
	}
}

// Set validates and assigns value to key, failing on an unknown key or an
// invalid license value rather than silently accepting garbage.
func (c *Config) Set(key, value string) error {
	switch key {
	case "license":
		if !ValidLicenseKind(value) {
			return fmt.Errorf("invalid license %q — expected \"mit\" or \"apache-2.0\"", value)
		}
		c.License = value
	case "author":
		c.Author = value
	case "modulePrefix":
		c.ModulePrefix = value
	default:
		return fmt.Errorf("unknown config key %q — expected license, author, or modulePrefix", key)
	}
	return nil
}

// ConfigKeys are the recognized `genitz config set`/`get` keys, in display
// order — the single source of truth backing both Get/Set's validation and
// runConfig's default listing.
var ConfigKeys = []string{"license", "author", "modulePrefix"}
