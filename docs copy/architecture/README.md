# Yggdrasil Architecture Documentation

## Overview

This directory contains comprehensive architectural documentation for Yggdrasil, a modern MSP/CRM/ERP platform.

## Quick Navigation

| Document                  | Audience           | Focus                       | Dependencies    |
| ------------------------- | ------------------ | --------------------------- | --------------- |
| **System Overview**       | All                | High-level architecture     | None            |
| **Platform Design**       | Architects         | Core platform design        | System Overview |
| **Data Model**            | Backend/Data Teams | Database design             | System Overview |
| **API Reference**         | Developers         | API specifications          | Data Model      |
| **Security Architecture** | Security Teams     | Security design             | System Overview |
| **Infrastructure**        | DevOps/SRE         | Deployment & infrastructure | All             |
| **Development**           | Developers         | Development workflows       | All             |

## Document Structure

```
architecture/
├── README.md                 # This file - navigation and overview
├── platform/
│   ├── system-architecture.md    # Core system architecture (moved from 01-system-architecture.md)
│   ├── implementation-roadmap.md # Implementation roadmap (moved from 05-implementation-roadmap.md)
│   └── system-overview.md        # High-level platform overview
├── backend/
│   ├── data-model.md         # Database design and entity relationships
│   ├── api-contract.md       # API specifications and contracts
│   ├── plugin-architecture.md # Plugin system design and SDK
│   └── database-migrations.md # Database migration strategy (moved from 07-database-migrations.md)
├── infrastructure/
│   └── deployment-architecture.md # Deployment patterns and infrastructure
├── development/
│   ├── testing-strategy.md    # Testing methodology and strategy
│   └── workflows.md         # Development processes and workflows
└── guides/
    └── development-workflow.md  # Development workflow guide
```

## Getting Started

1. **New to Yggdrasil**: Start with [System Architecture](./platform/system-architecture.md)
2. **Developers**: Read [API Reference](./backend/api-contract.md) and [Data Model](./backend/data-model.md)
3. **Architects**: Start with [System Overview](./platform/system-overview.md)
4. **DevOps**: Review [Deployment Architecture](./infrastructure/deployment-architecture.md)
5. **Development Team**: See [Development Workflow](./guides/development-workflow.md)

## Document Descriptions

### Platform Architecture

- **system-architecture.md**: High-level system design, component interactions, technology stack
- **implementation-roadmap.md**: Implementation phases, milestones, and timeline
- **system-overview.md**: High-level platform overview and design principles

### Backend Architecture

- **data-model.md**: Entity relationships, database design, data flows
- **api-contract.md**: REST/gRPC API specifications, authentication, versioning
- **plugin-architecture.md**: Plugin system design, SDK architecture, extension points
- **database-migrations.md**: Database migration strategy and versioning

### Infrastructure Architecture

- **deployment-architecture.md**: Deployment patterns, environment configuration, scaling strategies

### Development Architecture

- **testing-strategy.md**: Testing methodology, test types, automation
- **workflows.md**: Development processes, code review, release management

### Development Guides

- **development-workflow.md**: Step-by-step development workflow guide

## Benefits of Current Structure

1. **Clear Separation of Concerns**: Platform vs Backend vs Infrastructure
2. **Logical Organization**: Related documents grouped together
3. **Better Navigation**: Clearer progression through architecture concepts
4. **Enhanced Discoverability**: Easier to find relevant information
5. **Maintainable Structure**: Logical grouping for future updates

---

For the complete architecture documentation, start with the navigation guide above and follow the logical progression through each domain.
