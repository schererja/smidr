package config

import (
	"os"

	"go.yaml.in/yaml/v3"
)

type Config struct {
	AgentConfig  AgentConfig        `yaml:"agent"`
	ControlPlane ControlPlaneConfig `yaml:"control_plane"`
	Runtime      RuntimeConfig      `yaml:"runtime"`
}

type AgentConfig struct {
	ID     string `yaml:"id"`
	IDFile string `yaml:"id_file"`
	Demo   bool   `yaml:"demo"`
}

type ControlPlaneConfig struct {
	URI string `yaml:"uri"`
}

type RuntimeConfig struct {
	WorkDir string `yaml:"work_dir"`
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
