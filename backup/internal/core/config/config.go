package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
	"go.yaml.in/yaml/v3"
)

// Config holds the configuration for smidr-core
type Config struct {
	AgentServer ServerConfig `yaml:"agent_server" mapstructure:"agent_server"`
	WebServer   ServerConfig `yaml:"web_server" mapstructure:"web_server"`
	Store       StoreConfig  `yaml:"store" mapstructure:"store"`
	Database    DBConfig     `yaml:"database" mapstructure:"database"`
	LogLevel    string       `yaml:"log_level" mapstructure:"log_level"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host string `yaml:"host" mapstructure:"host"`
	Port int    `yaml:"port" mapstructure:"port"`
}

// StoreConfig holds store configuration
type StoreConfig struct {
	Type       string `yaml:"type" mapstructure:"type"` // memory, sqlite, postgres
	SQLitePath string `yaml:"sqlite_path" mapstructure:"sqlite_path"`
}

// DBConfig holds database configuration
type DBConfig struct {
	Type     string `yaml:"type" mapstructure:"type"` // postgres, sqlite, etc
	Host     string `yaml:"host" mapstructure:"host"`
	Port     int    `yaml:"port" mapstructure:"port"`
	Database string `yaml:"database" mapstructure:"database"`
	User     string `yaml:"user" mapstructure:"user"`
	Password string `yaml:"password" mapstructure:"password"`
	SSLMode  string `yaml:"ssl_mode" mapstructure:"ssl_mode"`
}

// Address returns the formatted address for a server
func (s ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadFromBytes(data)
}

func LoadFromBytes(data []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func LoadFromViper() (*Config, error) {
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return &cfg, nil
}
