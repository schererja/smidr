package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
	"go.yaml.in/yaml/v3"
)

type Config struct {
	AgentConfig  AgentConfig        `yaml:"agent" mapstructure:"agent"`
	ControlPlane ControlPlaneConfig `yaml:"control_plane" mapstructure:"control_plane"`
	Runtime      RuntimeConfig      `yaml:"runtime" mapstructure:"runtime"`
}

type AgentConfig struct {
	ID                string `yaml:"id" mapstructure:"id"`
	IDFile            string `yaml:"id_file" mapstructure:"id_file"`
	Demo              bool   `yaml:"demo" mapstructure:"demo"`
	HeartBeatInterval int    `yaml:"heartbeat_interval" mapstructure:"heartbeat_interval"`
}

type ControlPlaneConfig struct {
	URI string `yaml:"uri" mapstructure:"uri"`
}

type RuntimeConfig struct {
	WorkDir string `yaml:"work_dir" mapstructure:"work_dir"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	// Implementation for loading and parsing the config file goes here
	return LoadFromBytes(data)
}
func LoadFromBytes(data []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	// Implementation for loading and parsing the config from bytes goes here
	return &cfg, nil
}

func LoadFromViper() (*Config, error) {
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if cfg.AgentConfig.ID == "" {
		// TODO: Consider auto-generating an ID if not set also need to load from id_file if there
		return nil, fmt.Errorf("agent.id must be set in the configuration")
	}
	if cfg.ControlPlane.URI == "" {
		// TODO: Consider providing a default URI or handling this case differently
		return nil, fmt.Errorf("control_plane.uri must be set in the configuration")
	}

	// Setting defaults
	if cfg.AgentConfig.HeartBeatInterval == 0 {
		cfg.AgentConfig.HeartBeatInterval = 30 // Default to 30 seconds
	}
	if cfg.Runtime.WorkDir == "" {
		//TODO: Use a more appropriate default or temp directory based on OS
		cfg.Runtime.WorkDir = "/tmp/smidr" // Default work directory
	}
	return &cfg, nil
}
