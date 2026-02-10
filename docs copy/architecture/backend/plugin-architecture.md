# Plugin Architecture

## Overview

Yggdrasil is designed to be highly extensible through a plugin architecture. This document defines the plugin system for both **agent plugins** (running on managed systems) and **control plane plugins** (extending backend functionality).

## Design Philosophy

### Core Principles

1. **Security First**: Plugins must not compromise system security
2. **Isolation**: Plugin failures should not crash the core system
3. **Versioning**: Multiple plugin versions can coexist
4. **Discoverability**: Plugins expose capabilities through metadata
5. **Ease of Development**: Simple API for plugin authors

### Plugin Types

| Type                 | Location         | Purpose                                | Language   |
| -------------------- | ---------------- | -------------------------------------- | ---------- |
| Agent Plugin         | Agent (Go)       | System monitoring, task execution      | Go         |
| Control Plane Plugin | Backend (Go)     | Business logic extension, integrations | Go         |
| UI Plugin            | Frontend (React) | Custom dashboards, views (future)      | TypeScript |

**v0.5 Scope**: Agent plugins only (compiled-in). Control plane plugins in v1.0+.

---

## Agent Plugin Architecture

### v0.5: Compiled-In Plugins (Path 1)

For MVP, plugins are compiled directly into the agent binary. This provides:

- ✅ Simple development and deployment
- ✅ Cross-platform compatibility (Windows, Linux, macOS)
- ✅ No dynamic loading complexity
- ❌ Requires agent recompilation to add plugins
- ❌ Not true "plugin" architecture yet

**Migration Path**: Design plugin interface from day one to match Hashicorp go-plugin API. When we migrate to dynamic loading in v1.0, plugin code changes will be minimal.

### Plugin Interface

```go
// packages/agent-sdk/plugin.go
package sdk

import (
    "context"
)

// Plugin is the base interface all agent plugins must implement
type Plugin interface {
    // Name returns the unique plugin identifier
    Name() string

    // Version returns the plugin version (semver)
    Version() string

    // Description returns human-readable description
    Description() string

    // Initialize is called once when the agent starts
    // config contains plugin-specific configuration from control plane
    Initialize(ctx context.Context, config map[string]interface{}) error

    // Shutdown is called when the agent is stopping
    Shutdown(ctx context.Context) error
}

// MetricsPlugin extends Plugin for metric collection
type MetricsPlugin interface {
    Plugin

    // CollectMetrics gathers metrics and returns them
    // Called periodically by the agent scheduler
    CollectMetrics(ctx context.Context) ([]Metric, error)
}

// TaskPlugin extends Plugin for task execution (v1.0+)
type TaskPlugin interface {
    Plugin

    // SupportedTaskTypes returns list of task types this plugin can handle
    SupportedTaskTypes() []string

    // ExecuteTask runs a task and returns the result
    ExecuteTask(ctx context.Context, task Task) (TaskResult, error)
}

// Metric represents a single metric data point
type Metric struct {
    Time       time.Time
    MetricName string
    Value      float64
    Labels     map[string]string
}

// Task represents work to be executed (v1.0+)
type Task struct {
    ID      string
    Type    string
    Payload map[string]interface{}
}

// TaskResult is returned after task execution
type TaskResult struct {
    Success bool
    Output  map[string]interface{}
    Error   string
}
```

### Built-In Plugins (v0.5)

#### 1. System Metrics Plugin

Collects basic system metrics (CPU, memory, disk, network).

```go
// internal/agent/plugins/system_metrics/plugin.go
package system_metrics

import (
    "context"
    "time"

    "github.com/shirou/gopsutil/v3/cpu"
    "github.com/shirou/gopsutil/v3/mem"
    "github.com/shirou/gopsutil/v3/disk"

    "github.com/intrik8-labs/yggdrasil/packages/agent-sdk"
)

type SystemMetricsPlugin struct {
    interval time.Duration
}

func New() *SystemMetricsPlugin {
    return &SystemMetricsPlugin{}
}

func (p *SystemMetricsPlugin) Name() string {
    return "system_metrics"
}

func (p *SystemMetricsPlugin) Version() string {
    return "0.5.0"
}

func (p *SystemMetricsPlugin) Description() string {
    return "Collects system-level metrics (CPU, memory, disk, network)"
}

func (p *SystemMetricsPlugin) Initialize(ctx context.Context, config map[string]interface{}) error {
    // Parse interval from config, default to 60s
    intervalSecs := 60
    if val, ok := config["interval_seconds"].(float64); ok {
        intervalSecs = int(val)
    }
    p.interval = time.Duration(intervalSecs) * time.Second
    return nil
}

func (p *SystemMetricsPlugin) Shutdown(ctx context.Context) error {
    return nil
}

func (p *SystemMetricsPlugin) CollectMetrics(ctx context.Context) ([]sdk.Metric, error) {
    metrics := []sdk.Metric{}
    now := time.Now()

    // CPU usage
    cpuPercent, err := cpu.PercentWithContext(ctx, 0, false)
    if err == nil && len(cpuPercent) > 0 {
        metrics = append(metrics, sdk.Metric{
            Time:       now,
            MetricName: "cpu_usage",
            Value:      cpuPercent[0],
            Labels:     map[string]string{},
        })
    }

    // Memory usage
    vmStat, err := mem.VirtualMemoryWithContext(ctx)
    if err == nil {
        metrics = append(metrics, sdk.Metric{
            Time:       now,
            MetricName: "memory_usage",
            Value:      vmStat.UsedPercent,
            Labels:     map[string]string{},
        })
    }

    // Disk usage (all partitions)
    partitions, err := disk.PartitionsWithContext(ctx, false)
    if err == nil {
        for _, partition := range partitions {
            usage, err := disk.UsageWithContext(ctx, partition.Mountpoint)
            if err == nil {
                metrics = append(metrics, sdk.Metric{
                    Time:       now,
                    MetricName: "disk_usage",
                    Value:      usage.UsedPercent,
                    Labels: map[string]string{
                        "device":     partition.Device,
                        "mountpoint": partition.Mountpoint,
                        "fstype":     partition.Fstype,
                    },
                })
            }
        }
    }

    // Network I/O (future enhancement)
    // ...

    return metrics, nil
}
```

### Plugin Registry

The agent maintains a registry of all available plugins.

```go
// internal/agent/registry/registry.go
package registry

import (
    "fmt"
    "sync"

    "github.com/intrik8-labs/yggdrasil/packages/agent-sdk"
)

type Registry struct {
    plugins map[string]sdk.Plugin
    mu      sync.RWMutex
}

func NewRegistry() *Registry {
    return &Registry{
        plugins: make(map[string]sdk.Plugin),
    }
}

func (r *Registry) Register(plugin sdk.Plugin) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    name := plugin.Name()
    if _, exists := r.plugins[name]; exists {
        return fmt.Errorf("plugin %s already registered", name)
    }

    r.plugins[name] = plugin
    return nil
}

func (r *Registry) Get(name string) (sdk.Plugin, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    plugin, ok := r.plugins[name]
    return plugin, ok
}

func (r *Registry) List() []sdk.Plugin {
    r.mu.RLock()
    defer r.mu.RUnlock()

    plugins := make([]sdk.Plugin, 0, len(r.plugins))
    for _, plugin := range r.plugins {
        plugins = append(plugins, plugin)
    }
    return plugins
}

func (r *Registry) GetMetricsPlugins() []sdk.MetricsPlugin {
    r.mu.RLock()
    defer r.mu.RUnlock()

    var metricsPlugins []sdk.MetricsPlugin
    for _, plugin := range r.plugins {
        if mp, ok := plugin.(sdk.MetricsPlugin); ok {
            metricsPlugins = append(metricsPlugins, mp)
        }
    }
    return metricsPlugins
}
```

### Agent Initialization with Plugins

```go
// cmd/agent/main.go
package main

import (
    "context"
    "log"

    "github.com/intrik8-labs/yggdrasil/internal/agent/registry"
    "github.com/intrik8-labs/yggdrasil/internal/agent/plugins/system_metrics"
)

func main() {
    ctx := context.Background()

    // Create plugin registry
    pluginRegistry := registry.NewRegistry()

    // Register built-in plugins
    if err := pluginRegistry.Register(system_metrics.New()); err != nil {
        log.Fatalf("Failed to register system_metrics plugin: %v", err)
    }

    // Initialize all plugins
    for _, plugin := range pluginRegistry.List() {
        config := fetchPluginConfig(plugin.Name()) // From control plane
        if err := plugin.Initialize(ctx, config); err != nil {
            log.Printf("Failed to initialize plugin %s: %v", plugin.Name(), err)
            continue
        }
        log.Printf("Initialized plugin: %s v%s", plugin.Name(), plugin.Version())
    }

    // Start metric collection scheduler
    startMetricsScheduler(ctx, pluginRegistry)

    // ... rest of agent initialization
}

func startMetricsScheduler(ctx context.Context, registry *registry.Registry) {
    ticker := time.NewTicker(60 * time.Second)
    defer ticker.Stop()

    go func() {
        for {
            select {
            case <-ticker.C:
                collectAndSendMetrics(ctx, registry)
            case <-ctx.Done():
                return
            }
        }
    }()
}

func collectAndSendMetrics(ctx context.Context, registry *registry.Registry) {
    metricsPlugins := registry.GetMetricsPlugins()

    var allMetrics []sdk.Metric
    for _, plugin := range metricsPlugins {
        metrics, err := plugin.CollectMetrics(ctx)
        if err != nil {
            log.Printf("Plugin %s failed to collect metrics: %v", plugin.Name(), err)
            continue
        }
        allMetrics = append(allMetrics, metrics...)
    }

    // Send to control plane via gRPC
    if err := sendMetricsToControlPlane(ctx, allMetrics); err != nil {
        log.Printf("Failed to send metrics: %v", err)
    }
}
```

---

## v1.0+: Dynamic Plugin Loading (Hashicorp go-plugin)

### Migration to go-plugin

When we migrate to dynamic loading in v1.0, the plugin interface stays the same, but the loading mechanism changes.

**Advantages**:

- ✅ Plugins are separate binaries
- ✅ Plugin crashes don't kill agent
- ✅ Plugins can be updated without agent restart
- ✅ Support for external/third-party plugins

**Architecture**:

```bash
Agent (Main Process)
  ↓
  Launches Plugin as Subprocess
  ↓
gRPC over localhost (127.0.0.1:random_port)
  ↓
Plugin Process (isolated)
```

**Example Plugin Loading** (v1.0):

```go
import "github.com/hashicorp/go-plugin"

// Plugin handshake
var handshakeConfig = plugin.HandshakeConfig{
    ProtocolVersion:  1,
    MagicCookieKey:   "YGGDRASIL_PLUGIN",
    MagicCookieValue: "yggdrasil",
}

// Load plugin from binary
client := plugin.NewClient(&plugin.ClientConfig{
    HandshakeConfig: handshakeConfig,
    Plugins: map[string]plugin.Plugin{
        "metrics": &MetricsPluginRPC{},
    },
    Cmd: exec.Command("./plugins/system_metrics"),
})
defer client.Kill()

rpcClient, err := client.Client()
raw, err := rpcClient.Dispense("metrics")
metricsPlugin := raw.(sdk.MetricsPlugin)

// Use plugin normally
metrics, err := metricsPlugin.CollectMetrics(ctx)
```

**Migration Path**:

1. v0.5: Plugins compiled-in, using sdk.Plugin interface
2. v1.0: Same interface, but loaded via go-plugin
3. Plugin code requires minimal changes (add RPC wrappers)
4. Internal plugins shipped as binaries alongside agent
5. External plugins distributed separately

---

## Control Plane Plugins (v1.0+)

### Go Plugin Interface

Control plane plugins extend backend functionality (integrations, custom business logic, webhooks).

```go
// packages/control-plane-sdk/plugin.go
package sdk

import "context"

type Plugin interface {
    // Name returns the unique plugin identifier
    Name() string

    // Version returns the plugin version (semver)
    Version() string

    // Description returns a human-readable description
    Description() string

// Initialize is called when plugin is loaded
    Initialize(ctx context.Context, config map[string]interface{}) error

    // Shutdown is called when plugin is unloaded
    Shutdown(ctx context.Context) error

type IntegrationPlugin interface {
    Plugin
    HandleEvent(ctx context.Context, eventType string, payload map[string]interface{}) error
}

type WorkflowPlugin interface {
    Plugin
    ExecuteWorkflow(ctx context.Context, workflowID string, context map[string]interface{}) (map[string]interface{}, error)
}
```

**Example: Slack Integration Plugin**:

```go
package main

import (
    "context"
    "fmt"

    "github.com/intrik8-labs/yggdrasil/pkg/sdk"
)

type SlackPlugin struct {
    client         *slack.Client
    defaultChannel string
}

func (p *SlackPlugin) Name() string {
    return "slack_integration"
}

func (p *SlackPlugin) Version() string {
    return "1.0.0"
}

func (p *SlackPlugin) Description() string {
    return "Send notifications to Slack channels"
}

func (p *SlackPlugin) Initialize(ctx context.Context, config map[string]interface{}) error {
    token := config["slack_token"].(string)
    p.client = slack.New(token)
    p.defaultChannel = config["default_channel"].(string)
    if p.defaultChannel == "" {
        p.defaultChannel = "#alerts"
    }
    return nil
}

func (p *SlackPlugin) HandleEvent(ctx context.Context, eventType string, payload map[string]interface{}) error {
    switch eventType {
    case "ticket.created":
        return p.sendTicketNotification(ctx, payload)
    case "agent.offline":
        return p.sendAgentAlert(ctx, payload)
    }
    return nil
}

func (p *SlackPlugin) sendTicketNotification(ctx context.Context, ticket map[string]interface{}) error {
    ticketNumber := ticket["ticket_number"].(string)
    title := ticket["title"].(string)
    message := fmt.Sprintf("🎫 New Ticket #%s: %s", ticketNumber, title)

    _, _, err := p.client.PostMessageContext(
        ctx,
        p.defaultChannel,
        slack.MsgOptionText(message, false),
    )
    return err
}
```

### Plugin Discovery and Loading

```go
// internal/platform/plugins/loader.go
package plugins

import (
    "fmt"
    "path/filepath"
    "plugin"
)

// PluginLoader manages loading and registering control plane plugins
type PluginLoader struct {
    pluginDir string
    plugins   map[string]*plugin.Plugin
}

func NewPluginLoader(pluginDir string) *PluginLoader {
    return &PluginLoader{
        pluginDir: pluginDir,
        plugins:   make(map[string]*plugin.Plugin),
    }
}

func (l *PluginLoader) LoadPlugins() error {
    matches, err := filepath.Glob(filepath.Join(l.pluginDir, "*.so"))
    if err != nil {
        return fmt.Errorf("failed to glob plugins: %w", err)
    }

    for _, path := range matches {
        p, err := plugin.Open(path)
        if err != nil {
            return fmt.Errorf("failed to load plugin %s: %w", path, err)
        }
        l.plugins[filepath.Base(path)] = p
    }

    return nil
}
```
        pluginDir: string
        plugins:   map[string]*plugin.Plugin
    }

    func NewPluginLoader(pluginDir string) *PluginLoader {
        return &PluginLoader{
            pluginDir: pluginDir,
            plugins:   make(map[string]*plugin.Plugin),
        }
    }

    func (l *PluginLoader) LoadPlugins() error {
        matches, err := filepath.Glob(filepath.Join(l.pluginDir, "*.so"))
        if err != nil {
            return fmt.Errorf("failed to glob plugins: %w", err)
        }

        for _, path := range matches {
            p, err := plugin.Open(path)
            if err != nil {
                return fmt.Errorf("failed to load plugin %s: %w", path, err)
            }
            l.plugins[filepath.Base(path)] = p
        }

        return nil
    }

    func (l *PluginLoader) GetPlugin(name string) *plugin.Plugin {
        if p, ok := l.plugins[name]; ok {
            return p
        }
        return nil
    }
```

---

## Plugin Marketplace (v2.0+)

### External Plugin Distribution

**Marketplace Features**:

- Plugin discovery and browsing
- Version management
- Security scanning and verification
- User ratings and reviews
- Installation and updates via CLI

**Plugin Manifest** (`plugin.yaml`):

```yaml
name: slack_integration
version: 1.0.0
type: control_plane # or "agent"
author: Intrik8 Labs
description: Send notifications to Slack
repository: https://github.com/intrik8-labs/yggdrasil-plugin-slack
license: MIT

# For agent plugins
platform:
  - linux-amd64
  - linux-arm64
  - windows-amd64
  - darwin-amd64
  - darwin-arm64

# Dependencies
dependencies:
  agent_version: ">=1.0.0"
  control_plane_version: ">=1.0.0"

# Configuration schema
config_schema:
  slack_token:
    type: string
    required: true
    description: Slack Bot Token
  default_channel:
    type: string
    required: false
    default: "#alerts"
    description: Default Slack channel for notifications

# Permissions required
permissions:
  - tickets:read
  - agents:read
  - notifications:write
```

### Plugin Security Model

**Code Signing** (marketplace plugins):

- All marketplace plugins must be signed by Intrik8 Labs or trusted authors
- Signature verification before installation
- Revocation mechanism for compromised plugins

**Sandboxing** (external plugins):

- Separate processes (go-plugin for agents)
- Limited system access (no direct database access)
- API-only communication with core system
- Resource limits (CPU, memory, network)

**Permission System**:

```yaml
# Plugin declares required permissions in manifest
permissions:
  - tickets:read
  - tickets:write
  - clients:read

# Runtime permission checks in Go
func requirePluginPermission(permission string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            plugin := getPluginFromContext(r.Context())
            if !plugin.HasPermission(permission) {
                http.Error(w, "Forbidden", http.StatusForbidden)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

func createTicketFromPlugin(plugin sdk.Plugin, ticketData TicketCreate) error {
    // Plugin can only create tickets if it has tickets:write permission
    // Handler logic here
    return nil
}
```

---

## Plugin Development Guidelines

### Best Practices

1. **Error Handling**: Always handle errors gracefully, never crash the agent/control plane
2. **Logging**: Use structured logging with context
3. **Configuration**: Support configuration via manifest
4. **Testing**: Include unit and integration tests
5. **Documentation**: Provide clear README and usage examples
6. **Versioning**: Follow semantic versioning (semver)

### Testing Plugins

**Agent Plugin Test**:

1. **Logging**: Use structured logging with context
2. **Configuration**: Support configuration via manifest
3. **Testing**: Include unit and integration tests
4. **Documentation**: Provide clear README and usage examples
5. **Versioning**: Follow semantic versioning (semver)

### Testing Plugins examples

**Agent Plugin Test**:

```go
func TestSystemMetricsPlugin(t *testing.T) {
    ctx := context.Background()
    plugin := system_metrics.New()

    // Initialize
    config := map[string]interface{}{
        "interval_seconds": 60,
    }
    err := plugin.Initialize(ctx, config)
    assert.NoError(t, err)

    // Collect metrics
    metrics, err := plugin.CollectMetrics(ctx)
    assert.NoError(t, err)
    assert.NotEmpty(t, metrics)

    // Verify CPU metric exists
    found := false
    for _, m := range metrics {
        if m.MetricName == "cpu_usage" {
            found = true
            assert.GreaterOrEqual(t, m.Value, 0.0)
            assert.LessOrEqual(t, m.Value, 100.0)
        }
    }
    assert.True(t, found, "CPU metric should be collected")

    // Shutdown
    err = plugin.Shutdown(ctx)
    assert.NoError(t, err)
}
```

---

## Future Enhancements

### UI Plugins (v2.0+)

- Custom dashboard widgets
- Custom views and pages
- React component-based
- Loaded dynamically via module federation

### Plugin API Versioning

- Plugins declare compatible SDK version
- SDK follows semver for breaking changes
- Deprecation warnings for old plugin API usage

### Plugin Analytics

- Track plugin usage and performance
- Metrics on plugin adoption
- Error reporting and telemetry

---

## Next Steps

1. Review [Implementation Roadmap](../platform/implementation-roadmap.md) for plugin development timeline
2. Review [Project Structure](./06-project-structure.md) for plugin code organization
3. Review [Testing Strategy](./08-testing-strategy.md) for plugin testing approach
