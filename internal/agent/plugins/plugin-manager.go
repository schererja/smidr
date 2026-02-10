package plugins

import (
	"errors"

	config "github.com/schererja/smidr/internal/config/agent"
	sdk "github.com/schererja/smidr/pkg/agent-sdk"
	"github.com/schererja/smidr/pkg/logger"
)

type PluginManager struct {
	config  *config.AgentConfig
	plugins map[string]sdk.Plugin
	log     *logger.Logger
}

// NewPluginManager creates a new PluginManager with the given configuration.
func NewPluginManager(cfg *config.AgentConfig, log *logger.Logger) *PluginManager {
	return &PluginManager{
		config:  cfg,
		plugins: make(map[string]sdk.Plugin),
		log:     log.WithComponent("plugin-manager"),
	}
}

// Register adds a built-in plugin to the manager.
func (pm *PluginManager) Register(plugin sdk.Plugin) error {
	if plugin == nil {
		return errors.New("plugin cannot be nil")
	}
}
