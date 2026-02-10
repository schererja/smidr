# Yggdrasil Documentation

Welcome to the Yggdrasil MSP/CRM/ERP platform documentation.

## Quick Navigation

### 🏗️ Architecture & Development

- [Architecture Overview](architecture/README.md) - System design and components
- [Development Workflow](architecture/10-development-workflow.md) - How we build software
- [Style Guide](STYLE_GUIDE.md) - Coding standards and conventions

### 🚀 Deployment & Operations

- [Deployment Guide](deployment/DEPLOYMENT.md) - Production deployment methods
- [Security Guide](security/SECURITY.md) - Security hardening and best practices
- [Infrastructure Requirements](deployment/INFRASTRUCTURE.md) - Hardware and software needs

### 🔌 API & Integration

- [API Quick Start](api/QUICK_START.md) - Getting started with APIs

### 👥 User Guides

- [Administrator Guide](user-guide/ADMIN.md) - System administration
- [User Guide](user-guide/USER_GUIDE.md) - General user documentation

### 🤝 Community

- [Contributing Guide](CONTRIBUTING.md) - How to contribute
- [Code of Conduct](CODE_OF_CONDUCT.md) - Community guidelines
- [Licensing Information](LICENSING.md) - License details

## Documentation Structure

```bash
docs/
├── README.md                    # This file - main navigation
├── STYLE_GUIDE.md             # Coding standards and conventions
├── CONTRIBUTING.md             # Contribution guidelines
├── CODE_OF_CONDUCT.md         # Community guidelines
├── LICENSING.md               # License information
├── DEVELOPMENT_SETUP.md       # Developer onboarding guide
├── architecture/              # System architecture docs
│   ├── README.md              # Architecture navigation
│   ├── platform/              # Platform-level architecture
│   │   ├── system-overview.md # System architecture (moved from 01-*)
│   │   ├── implementation-roadmap.md # Roadmap (moved from 05-*)
│   │   └── system-overview.md # High-level platform design
│   ├── backend/               # Backend architecture
│   │   ├── data-model.md      # Database design and entities
│   │   ├── api-contract.md    # API specifications
│   │   ├── plugin-architecture.md # Plugin system design
│   │   └── database-migrations.md # Database migrations (moved from 07-*)
│   ├── infrastructure/        # Infrastructure and deployment
│   │   └── deployment-architecture.md # Deployment patterns
│   ├── development/           # Development workflows
│   │   ├── testing-strategy.md # Testing methodology
│   │   └── workflows.md       # Development processes
│   └── guides/                # Development guides
│       └── development-workflow.md # Development workflow
├── deployment/               # Deployment and operations
│   └── DEPLOYMENT.md          # Main deployment guide
├── security/                 # Security documentation
│   └── SECURITY.md           # Security overview and best practices
├── api/                     # API documentation
│   └── QUICK_START.md        # API getting started guide
└── user-guide/              # End-user documentation
    ├── ADMIN.md             # Administrator guide
    └── USER_GUIDE.md         # General user guide
```

## Getting Started

1. **New to Yggdrasil?** Start with [Architecture Overview](architecture/README.md)
2. **Want to deploy?** Read the [Deployment Guide](deployment/DEPLOYMENT.md)
3. **Developing plugins?** Check the [Plugin Architecture](architecture/backend/plugin-architecture.md)
4. **Integrating APIs?** See the [API Quick Start](api/QUICK_START.md)

## Contributing to Documentation

We welcome documentation improvements! Please:

1. Follow the [Style Guide](STYLE_GUIDE.md) for consistency
2. Test your instructions if possible
3. Include examples and troubleshooting sections
4. Update related files when making changes

## Support

- **Questions?** Open a GitHub issue with the "question" label
- **Documentation issues?** Open an issue with the "docs" label
- **Security concerns?** Email <security@yggdrasil.dev>

---

_This documentation is continuously evolving. Check back often for updates!_
