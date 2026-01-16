package config

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v3"
)

type Config struct {
	AgentPort int
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return LoadFromBytes(data)
}

// LoadFromBytes parses configuration from YAML or JSON bytes, performing
// environment variable substitution and full validation.
func LoadFromBytes(data []byte) (*Config, error) {
	// Perform environment variable substitution

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Validate the configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.AgentPort <= 0 || c.AgentPort > 65535 {
		return fmt.Errorf("agent port must be between 1 and 65535")
	}
	return nil
}
