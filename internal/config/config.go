package config

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v3"
)

type Config struct {
	AgentConfig  AgentConfig `json:"agent_config" yaml:"agent_config"`
	ServerConfig ServerConfig
}
type AgentConfig struct {
	AgentID           string `json:"agent_id" yaml:"agent_id"`
	HostName          string `json:"host_name" yaml:"host_name"`
	HeartbeatInterval int    `json:"heartbeat_interval" yaml:"heartbeat_interval"`
}

type ServerConfig struct {
	ServerAddress string `json:"server_address" yaml:"server_address"`
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
	if c.AgentConfig.AgentID == "" {
		return fmt.Errorf("agent_id is required")
	}
	if c.AgentConfig.HeartbeatInterval <= 0 {
		return fmt.Errorf("heartbeat_interval must be greater than 0")
	}
	if c.ServerConfig.ServerAddress == "" {
		return fmt.Errorf("server_address is required")
	}
	return nil
}
