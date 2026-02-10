package config

type Config struct {
	Name        string      `mapstructure:"name"`
	Description string      `mapstructure:"description"`
	Version     string      `mapstructure:"version"`
	AgentConfig AgentConfig `mapstructure:"agentConfig"`
}

type AgentConfig struct {
	ListenAddress       string `mapstructure:"listenAddress"`
	ControlPlaneAddress string `mapstructure:"controlPlaneAddress"`
	PluginsDirectory    string `mapstructure:"pluginsDirectory"`
}
