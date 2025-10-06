# ⚒️ Smidr

**The Digital Forge for Embedded Linux**

Smidr is a modern, container-based build system for embedded Linux that eliminates the pain points of traditional Yocto/BitBake workflows. Built by developers, for developers who are tired of cryptic errors, build state corruption, and massive storage requirements.

---

## 🔥 The Problem

Traditional embedded Linux development with Yocto/BitBake suffers from fundamental issues:

- **Massive storage waste**: 120GB+ per build environment, multiplied by every image variant
- **Build state corruption**: Mysterious failures requiring full rebuilds
- **Cryptic error messages**: Hours spent debugging parser failures and dependency hell
- **No parallelization**: Can't work on multiple images without massive resource overhead
- **Fragile environments**: One misconfigured layer breaks everything

## ⚡ The Smidr Solution

Smidr reimagines embedded Linux builds using modern container technology:

### 🏗️ Container-Native Architecture

- **Isolated builds**: Every build runs in a clean container environment
- **Shared caching**: Intelligent layer sharing dramatically reduces storage requirements
- **Parallel execution**: Build multiple images simultaneously without conflicts
- **Reproducible results**: Same input always produces the same output

### 💾 Intelligent Storage Management

- **Deduplicated storage**: Common base layers (Toradex, BSPs) shared across all builds
- **90%+ space savings**: Turn 360GB of duplicated builds into 40GB of smart cache
- **Incremental builds**: Only rebuild what actually changed

### 🛠️ Developer Experience First

- **Clear error messages**: Actionable feedback instead of cryptic stack traces
- **Simple configuration**: YAML-based build definitions that make sense
- **Fast iteration**: Quick builds for development, thorough builds for release

---

## 🚀 Quick Start

```bash
# Initialize a new project
smidr init my-embedded-project

# Configure your build (edit smidr.yaml)
vim smidr.yaml

# Build your image
smidr build

# Check build status
smidr status

# Get your artifacts
smidr artifacts list
```

### Example Configuration

```yaml
name: toradex-custom-image
description: "Custom Toradex image with our application stack"

base:
  provider: toradex
  machine: verdin-imx8mp
  distro: tdx-xwayland
  version: "6.0.0"

layers:
  - name: meta-toradex-bsp-common
    git: https://git.toradex.com/meta-toradex-bsp-common
    branch: kirkstone-6.x.y

  - name: meta-mycompany
    path: ./layers/meta-mycompany

custom:
  image: core-image-weston
  extra_packages:
    - python3
    - docker
    - nodejs

artifacts:
  - "*.wic"
  - "*.tar.bz2"
  - "*-sdk-*.sh"
```

---

## 🏛️ Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   smidr.yaml    │───▶│  Smidr Engine    │───▶│  Build Results  │
│  Configuration  │    │   (Go Runtime)   │    │   & Artifacts   │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │ Container Pool  │
                       │ ┌─────────────┐ │
                       │ │ Build Env 1 │ │
                       │ ├─────────────┤ │
                       │ │ Build Env 2 │ │
                       │ ├─────────────┤ │
                       │ │ Build Env N │ │
                       │ └─────────────┘ │
                       └─────────────────┘
                                │
                                ▼
                    ┌──────────────────────┐
                    │   Shared Cache       │
                    │ ┌─────────────────┐  │
                    │ │ Source Downloads│  │
                    │ ├─────────────────┤  │
                    │ │ sstate-cache    │  │
                    │ ├─────────────────┤  │
                    │ │ Layer Cache     │  │
                    │ └─────────────────┘  │
                    └──────────────────────┘
```

---

## 🎯 Project Status

**Current Phase**: MVP Development

### ✅ Completed

- [x] Project architecture design
- [x] Core concept validation
- [x] Technology stack selection

### 🚧 In Progress

- [ ] CLI framework implementation
- [ ] Container orchestration system
- [ ] Configuration parser
- [ ] Build execution engine

### 📋 Planned Features

**MVP (v0.1.0)**

- Container-based build isolation
- YAML configuration system
- CLI tools for common operations
- Toradex BSP support
- Basic artifact management

**Post-MVP**

- Web UI for build management
- Advanced caching strategies
- CI/CD integration
- Multi-vendor BSP support
- Cloud build service

---

## 🏢 Built by Jason Scherer

Smidr is developed by Jason Scherer, a software engineer focused on solving real problems in embedded systems development and DevOps automation.

**Team Background:**

- 20+ years in embedded systems and automation
- Real-world experience with Yocto's pain points

---

## 📖 Documentation

- [Installation Guide](docs/installation.md)
- [Configuration Reference](docs/configuration.md)
- [CLI Commands](docs/cli-reference.md)
- [Architecture Overview](docs/architecture.md)
- [Contributing Guide](CONTRIBUTING.md)
 - [Cache & Source Management](docs/cache.md)

---

## 🤝 Contributing

We welcome contributions! Smidr is built by embedded developers, for embedded developers.

1. Check our [Contributing Guide](CONTRIBUTING.md)
2. Look at [open issues](https://github.com/schereja/smidr/issues)
3. Join the discussion in [GitHub Discussions](https://github.com/schereja/smidr/discussions)

---

## 📄 License

Smidr is licensed under the [MIT License](LICENSE).

---

## 🔗 Links

- **Website**: [smidr.dev](https://smidr.dev)
- **Documentation**: [docs.smidr.dev](https://docs.smidr.dev)
- **Issues**: [GitHub Issues](https://github.com/schereja/smidr/issues)

---

<div align="center">
<strong>Transform your embedded Linux builds from painful to powerful.</strong>
<br><br>
⚒️ <strong>Smidr - Forge Better Builds</strong> ⚒️
</div>
