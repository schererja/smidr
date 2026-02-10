# Yggdrasil Plugin System

The Yggdrasil agent uses [go-plugin](https://github.com/hashicorp/go-plugin) to provide a flexible plugin architecture. Plugins run as separate processes and communicate with the agent via gRPC.

## Quick Start

```bash
# Build and run the agent with plugins
make run-agent

# Or manually
make build
./build/agent
```

You should see output like:
```
Agent started successfully
Loaded and started plugin: metrics
Agent running with plugins...
[metrics] goroutines=14, memory_alloc_mb=0.88, memory_sys_mb=8.52, gc_count=0, timestamp=1769536686
```

Metrics are displayed every 15 seconds showing the agent's system resource usage.

## Architecture

The plugin system consists of:

1. **Plugin Interface** (`pkg/plugin/interface.go`) - Defines the contract all plugins must implement
2. **gRPC Protocol** (`pkg/plugin/grpc.go`, `pkg/plugin/plugin.proto`) - Communication layer between agent and plugins
3. **Agent Plugin Manager** (`cmd/agent/main.go`) - Loads, manages, and monitors plugins
4. **Plugins** (`cmd/plugins/*/`) - Individual plugin implementations

## Plugin Interface

All plugins must implement:

```go
type Plugin interface {
    Name() string                                 // Plugin identifier
    Start(ctx context.Context) error              // Initialize plugin
    Stop(ctx context.Context) error               // Graceful shutdown
    Health(ctx context.Context) (bool, error)     // Health check
    GetMetrics(ctx context.Context) (map[string]interface{}, error) // Get plugin metrics
}
```

## Available Plugins

### Metrics Plugin

Located in `cmd/plugins/metrics/`, this plugin collects and reports system metrics:

- **Goroutine count** - Number of active goroutines
- **Memory allocation** - Current heap memory usage (MB)
- **System memory** - Total memory obtained from OS (MB)  
- **GC statistics** - Number of garbage collections performed
- **Timestamp** - Unix timestamp of the measurement

Metrics are collected internally every 10 seconds. The agent queries and displays them every 15 seconds via the `GetMetrics()` interface.

## Building Plugins

Build all plugins:
```bash
make build-plugins
```

Build a specific plugin:
```bash
go build -o build/plugins/metrics ./cmd/plugins/metrics
```

## Running the Agent with Plugins

The agent automatically discovers and loads plugins from `build/plugins/`:

```bash
# Build and run everything
make run-agent

# Or step by step
make build          # Builds agent + all plugins
./build/agent       # Runs agent

# For a timed test (35 seconds with metrics display)
./scripts/test-agent-long.sh
```

**Expected output:**
```
Agent started successfully
Loaded and started plugin: metrics
Agent running with plugins...
[metrics] goroutines=14, memory_alloc_mb=0.88, memory_sys_mb=8.52, gc_count=0, timestamp=1769536686
[metrics] goroutines=14, memory_alloc_mb=0.90, memory_sys_mb=8.52, gc_count=0, timestamp=1769536701
...
```

Press `Ctrl+C` to gracefully shutdown the agent and all plugins.

## Creating a New Plugin

1. Create a new directory: `cmd/plugins/your-plugin/`

2. Implement the Plugin interface:

```go
package main

import (
    "context"
    "github.com/hashicorp/go-plugin"
    yggplugin "github.com/intrik8-labs/yggdrasil/pkg/plugin"
)

type YourPlugin struct {
    // your fields
}

func (p *YourPlugin) Name() string {
    return "your-plugin"
}

func (p *YourPlugin) Start(ctx context.Context) error {
    // Initialize your plugin
    return nil
}

func (p *YourPlugin) Stop(ctx context.Context) error {
    // Cleanup
    return nil
}

func (p *YourPlugin) Health(ctx context.Context) (bool, error) {
    // Return health status
    return true, nil
}

func (p *YourPlugin) GetMetrics(ctx context.Context) (map[string]interface{}, error) {
    // Return plugin metrics (optional - return empty map if not applicable)
    return map[string]interface{}{
        "requests_handled": 42,
        "errors": 0,
    }, nil
}

func main() {
    plugin.Serve(&plugin.ServeConfig{
        HandshakeConfig: yggplugin.HandshakeConfig,
        Plugins: map[string]plugin.Plugin{
            "yggdrasil_plugin": &yggplugin.PluginGRPC{
                Impl: &YourPlugin{},
            },
        },
        GRPCServer: plugin.DefaultGRPCServer,
    })
}
```

3. Build your plugin:
```bash
go build -o build/plugins/your-plugin ./cmd/plugins/your-plugin
```

4. The agent will automatically load it on startup

## Plugin Lifecycle

1. **Discovery** - Agent scans `build/plugins/` directory
2. **Load** - Each plugin binary is executed as a subprocess
3. **Connect** - gRPC connection established via handshake
4. **Start** - Plugin's `Start()` method called
5. **Monitor** - Health checks and metrics queries every 15 seconds
6. **Stop** - On shutdown, `Stop()` called for graceful cleanup
7. **Kill** - Plugin process terminated

## Metrics Display

The agent periodically queries each plugin's metrics via `GetMetrics()` and displays them:

```
[plugin-name] metric1=value1, metric2=value2, ...
```

Metrics display interval: **15 seconds**  
Plugin collection interval: **10 seconds** (metrics plugin)

## Plugin Communication

Plugins communicate with the agent via gRPC using the protocol defined in `pkg/plugin/plugin.proto`. The handshake ensures compatibility:

```go
HandshakeConfig = plugin.HandshakeConfig{
    ProtocolVersion:  1,
    MagicCookieKey:   "YGGDRASIL_PLUGIN",
    MagicCookieValue: "yggdrasil",
}
```

## Future Enhancements

- Plugin configuration via YAML
- Hot-reload plugin support
- Plugin marketplace/registry
- Inter-plugin communication
- Plugin versioning and compatibility checks
