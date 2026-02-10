package smidr

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// PluginManager manages the lifecycle of plugins
type PluginManager struct {
	config  *AgentConfig
	plugins map[string]yggplugin.Plugin
}

func NewPluginManager(config *AgentConfig) *PluginManager {
	return &PluginManager{
		config:  config,
		plugins: make(map[string]yggplugin.Plugin),
	}
}

// LoadPlugins discovers and loads all plugins from the plugins directory
func (pm *PluginManager) LoadPlugins(ctx context.Context) error {
	if _, err := os.Stat(pm.config.PluginsDir); os.IsNotExist(err) {
		log.Printf("Plugins directory does not exist: %s", pm.config.PluginsDir)
		return nil
	}

	entries, err := os.ReadDir(pm.config.PluginsDir)
	if err != nil {
		return fmt.Errorf("failed to read plugins directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		pluginPath := filepath.Join(pm.config.PluginsDir, entry.Name())
		if err := pm.loadPlugin(ctx, pluginPath); err != nil {
			log.Printf("Failed to load plugin %s: %v", entry.Name(), err)
		}
	}

	return nil
}

// loadPlugin loads a single plugin
func (pm *PluginManager) loadPlugin(ctx context.Context, path string) error {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "plugin",
		Output: os.Stdout,
		Level:  hclog.Info,
	})

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: yggplugin.HandshakeConfig,
		Plugins:         yggplugin.PluginMap,
		Cmd:             exec.Command(path),
		Logger:          logger,
		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolGRPC,
		},
	})

	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return fmt.Errorf("failed to get RPC client: %w", err)
	}

	raw, err := rpcClient.Dispense("yggdrasil_plugin")
	if err != nil {
		client.Kill()
		return fmt.Errorf("failed to dispense plugin: %w", err)
	}

	plug := raw.(yggplugin.Plugin)
	name := plug.Name()

	// Start the plugin
	if err := plug.Start(ctx); err != nil {
		client.Kill()
		return fmt.Errorf("failed to start plugin: %w", err)
	}

	pm.plugins[name] = plug

	fmt.Printf("Loaded and started plugin: %s\n", name)

	// Display initial metrics
	time.Sleep(1 * time.Second) // Give plugin time to collect first metrics
	if metrics, err := plug.GetMetrics(ctx); err == nil && len(metrics) > 0 {
		fmt.Printf("[%s] Initial metrics: %s\n", name, formatMetrics(metrics))
	}

	// Monitor plugin health
	go pm.monitorHealth(ctx, name, plug)

	return nil
}

// monitorHealth periodically checks plugin health and displays metrics
func (pm *PluginManager) monitorHealth(ctx context.Context, name string, plug yggplugin.Plugin) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			healthy, err := plug.Health(ctx)
			if err != nil {
				log.Printf("[%s] Health check failed: %v", name, err)
			} else if !healthy {
				log.Printf("[%s] Plugin is unhealthy", name)
			}

			// Query and display metrics
			metrics, err := plug.GetMetrics(ctx)
			if err != nil {
				log.Printf("[%s] Failed to get metrics: %v", name, err)
			} else if len(metrics) > 0 {
				fmt.Printf("[%s] %s\n", name, formatMetrics(metrics))
			}
		}
	}
}

// formatMetrics formats metrics for display
func formatMetrics(metrics map[string]interface{}) string {
	if len(metrics) == 0 {
		return "none"
	}

	result := ""
	for k, v := range metrics {
		if result != "" {
			result += ", "
		}
		switch val := v.(type) {
		case float64:
			result += fmt.Sprintf("%s=%.2f", k, val)
		default:
			result += fmt.Sprintf("%s=%v", k, val)
		}
	}
	return result
}

// StopAll stops all running plugins
func (pm *PluginManager) StopAll(ctx context.Context) {
	for name, plug := range pm.plugins {
		fmt.Printf("Stopping plugin: %s\n", name)
		if err := plug.Stop(ctx); err != nil {
			log.Printf("Error stopping plugin %s: %v", name, err)
		}
	}

	// for name, client := range pm.clients {
	// 	fmt.Printf("Killing plugin client: %s\n", name)
	// 	client.Kill()
	// }
}
