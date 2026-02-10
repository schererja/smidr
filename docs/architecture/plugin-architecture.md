# Plugin Architecture (Smidr)

## Overview

Smidr is the build/agent + control-plane component of the larger Yggdrasil project. This repository will evolve into Yggdrasil; other Norse‑named components and the business application layer (C# WebAPI + Next.js) are planned to live alongside Smidr in the same repo and integrate via internal APIs.

This document defines the Smidr plugin system for **agent plugins** and, later, **control‑plane plugins**.

## Design Philosophy

1. **Security First**: Plugins must not compromise system security
2. **Isolation**: Plugin failures should not crash the core system
3. **Versioning**: Multiple plugin versions can coexist
4. **Discoverability**: Plugins expose capabilities through metadata
5. **Ease of Development**: Simple API for plugin authors

## Plugin Types

| Type                 | Location      | Purpose                           | Language   |
| -------------------- | ------------- | --------------------------------- | ---------- |
| Agent Plugin         | Agent (Go)    | System monitoring, task execution | Go         |
| Control Plane Plugin | Control Plane | Build automation, integrations    | Go         |
| UI Plugin            | Web (Next.js) | Custom dashboards, views (future) | TypeScript |

**MVP Scope**: Agent plugins only (compiled‑in). Control‑plane plugins in v1.0+.

---

## Agent Plugin Architecture

### MVP: Compiled‑In Plugins

For MVP, plugins are compiled directly into the agent binary. This provides:

- ✅ Simple development and deployment
- ✅ Cross‑platform compatibility (Windows, Linux, macOS)
- ✅ No dynamic loading complexity
- ❌ Requires agent recompilation to add plugins
- ❌ Not true process isolation yet

**Migration Path**: Design plugin interfaces from day one to align with Hashicorp go‑plugin. When we migrate to dynamic loading in v1.0, plugin code changes should be minimal.

### Plugin Interface

```go
// packages/agent-sdk/plugin.go
package sdk

import "context"

// Plugin is the base interface all agent plugins must implement
type Plugin interface {
    Name() string
    Version() string
    Description() string
    Initialize(ctx context.Context, config map[string]interface{}) error
    Shutdown(ctx context.Context) error
}

// MetricsPlugin extends Plugin for metric collection
type MetricsPlugin interface {
    Plugin
    CollectMetrics(ctx context.Context) ([]Metric, error)
}

// TaskPlugin extends Plugin for task execution (v1.0+)
type TaskPlugin interface {
    Plugin
    SupportedTaskTypes() []string
    ExecuteTask(ctx context.Context, task Task) (TaskResult, error)
}
```

---

## v1.0+: Dynamic Plugin Loading (Hashicorp go‑plugin)

When we migrate to dynamic loading in v1.0, the plugin interface stays the same, but the loading mechanism changes.

**Advantages**:

- ✅ Plugins are separate binaries
- ✅ Plugin crashes don't kill agent
- ✅ Plugins can be updated without agent restart
- ✅ Support for external/third‑party plugins

---

## Control‑Plane Plugins (v1.0+)

Control‑plane plugins extend build automation and backend functionality. These are intentionally deferred until after the MVP control‑plane stabilizes.
