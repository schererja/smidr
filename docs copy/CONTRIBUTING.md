# Contributing to Yggdrasil

Thank you for your interest in contributing to Yggdrasil! This guide will help you understand how to contribute effectively to our MSP/CRM/ERP platform.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Development Workflow](#development-workflow)
3. [Code Standards](#code-standards)
4. [Testing Requirements](#testing-requirements)
5. [Documentation](#documentation)
6. [Pull Request Process](#pull-request-process)
7. [Issue Reporting](#issue-reporting)
8. [Community Guidelines](#community-guidelines)

## Getting Started

### Prerequisites

- **Node.js** 20+
- **Go** 1.22+
- **Docker** + Docker Compose
- **Required tools**: `sqlc`, `buf` (protobuf), `migrate`, `make`

### Understanding the License

Before contributing, please review our mixed licensing approach:

- **Core Platform** (`cmd/`, `internal/`) - BUSL 1.1
- **Plugins & SDKs** (`packages/`, `web/`) - MIT License

See [docs/LICENSING.md](LICENSING.md) for detailed information.

### Initial Setup

1. **Fork and Clone**

   ```bash
   git clone https://github.com/your-username/yggdrasil.git
   cd yggdrasil
   git remote add upstream https://github.com/intrik8-labs/yggdrasil.git
   ```

2. **Environment Setup**

   ```bash
   cp .env.example .env
   # Edit .env with your local configuration
   ```

3. **Start Development Services**

   ```bash
   docker compose up -d postgres nats
   ```

4. **Install Dependencies**

   ```bash
   # Go dependencies (control plane & agent)
   go mod download

   # Generate sqlc code
   sqlc generate

   # Web App (TypeScript)
   cd web
   npm ci
   cd ..
   ```

## Development Workflow

### Branch Strategy

- **`main`**: Production-ready code
- **`develop`**: Integration branch for features
- **`feature/*`**: New features
- **`bugfix/*`**: Bug fixes
- **`hotfix/*`**: Critical production fixes

### Creating a Feature Branch

```bash
git checkout develop
git pull upstream develop
git checkout -b feature/your-feature-name
```

### Making Changes

1. **Small, focused commits** with clear messages
2. **Follow the style guide** for your language/framework
3. **Write tests** for new functionality
4. **Update documentation** as needed

### Commit Message Format

Use conventional commits:

```bash
type(scope): description

[optional body]

[optional footer]
```

**Types**: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

**Examples**:

- `feat(control-plane): add multi-tenant ticket filtering`
- `fix(web): resolve dashboard data loading issue`
- `docs(agent): update installation instructions`

## Code Standards

### Go (Control Plane & Agent)

```bash
# Linting
go vet ./...
golangci-lint run

# Formatting
gofmt -w .
goimports -w .

# Testing
go test ./...

# Generate code
sqlc generate
```

**Guidelines**:

- Use `golangci-lint` configuration
- Follow Go idioms and domain-based package structure
- Exported functions must have comments
- Use domain-based vertical slices in `internal/`
- Repository pattern for database access
- Use sqlc for type-safe SQL

## Testing Requirements

### Coverage Targets

- **Control Plane & Agent (Go)**: 80%+ coverage
- **Web App (TypeScript)**: 70%+ coverage

### Running Tests

```bash
# Control Plane & Agent (Go)
go test ./... -cover
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Web App
cd web
npm test -- --coverage

# Agent
# Run from root
go test -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Test Types

- **Unit Tests**: Test individual functions/components
- **Integration Tests**: Test component interactions
- **E2E Tests**: Test complete user workflows
- **Performance Tests**: For critical paths

## Documentation

### Types of Documentation

1. **API Documentation**: Auto-generated from OpenAPI specs
2. **Architecture Docs**: In `docs/architecture/` (keep in sync with code)
3. **Code Comments**: For complex business logic
4. **README Files**: For each major component

### Documentation Standards

- Use Markdown for all documentation
- Include code examples for common tasks
- Update diagrams when architecture changes
- Document breaking changes clearly

### Documentation Updates

When submitting PRs:

- Update relevant architecture docs
- Add/remove examples as needed
- Update API documentation if endpoints change
- Review existing docs for accuracy

## Pull Request Process

### Before Submitting

1. **Sync with upstream**

   ```bash
   git checkout develop
   git pull upstream develop
   git checkout your-branch
   git rebase develop
   ```

2. **Run all checks**

   ```bash
   # From project root
   make lint
   make test
   make build
   ```

3. **Ensure tests pass** and coverage meets requirements

### PR Template

Use the PR template and include:

- **Clear title** following commit message format
- **Description** of changes and their purpose
- **Testing done** and results
- **Breaking changes** (if any)
- **Screenshots** for UI changes
- **Links to related issues**

### Review Process

1. **Automated Checks**: CI/CD pipeline runs lint, tests, security scans
2. **Code Review**: At least one maintainer approval required
3. **Integration Tests**: Run on PR merge to `develop`
4. **Documentation Review**: Ensure docs are updated

### Merge Requirements

- All status checks must pass
- At least one maintainer approval
- No merge conflicts
- Documentation updated (if needed)
- Tests passing with adequate coverage

## Issue Reporting

### Bug Reports

Use the bug report template and include:

- **Environment details** (OS, versions, etc.)
- **Steps to reproduce** the issue
- **Expected vs actual behavior**
- **Error messages** and logs
- **Potential solutions** (if known)

### Feature Requests

Use the feature request template and include:

- **Problem statement** the feature solves
- **Proposed solution** with implementation ideas
- **Alternatives considered**
- **Acceptance criteria**

### Security Issues

Do NOT report security issues publicly. Email <security@yggdrasil.dev> with details.

## Community Guidelines

### Communication Channels

- **GitHub Issues/PRs**: For code-related discussions
- **Discord**: For general community chat and support
- **Email**: For private matters and security issues

### Getting Help

- Check existing documentation and issues first
- Use appropriate channels for your question
- Be patient and respectful when asking for help
- Help others when you can

### Recognition

Contributors are recognized through:

- GitHub contributor statistics
- Release notes acknowledgments
- Community spotlight in communications
- Invitation to core contributor Discord channels

## Release Process

### Versioning

Follow Semantic Versioning (SemVer):

- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

### Release Checklist

1. **Update version numbers** in all affected packages
2. **Update CHANGELOG** with all changes
3. **Tag release** in Git
4. **Build and publish** Docker images and binaries
5. **Update documentation** with new features
6. **Announce release** to community

## Additional Resources

- [Architecture Documentation](../architecture/README.md)
- [Style Guide](STYLE_GUIDE.md)
- [API Documentation](../architecture/03-api-contract.md)
- [Testing Strategy](../architecture/08-testing-strategy.md)
- [Licensing Information](LICENSING.md)
- [Code of Conduct](CODE_OF_CONDUCT.md)

## Questions?

If you have questions about contributing:

- Check existing issues and discussions
- Read the documentation thoroughly
- Ask in Discord community channels
- Open a "question" issue if needed

Thank you for contributing to Yggdrasil! Your contributions help make this project better for everyone. 🚀
