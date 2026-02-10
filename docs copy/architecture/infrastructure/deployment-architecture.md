# CI/CD Pipeline

## Overview

Yggdrasil uses **GitHub Actions** for continuous integration and deployment. This document defines the CI/CD workflows for all components.

---

## CI Workflow Goals

1. **Fast Feedback**: Fail fast on obvious issues (linting, type checking)
2. **Comprehensive Testing**: Run all tests before merge
3. **Security**: Scan for vulnerabilities
4. **Build Verification**: Ensure Docker images and binaries build successfully
5. **Deployment Automation**: Deploy to staging/production (v1.0+)

---

## Workflow Structure

### Pull Request Workflow

Triggered on: `pull_request` to `main` or `develop`

**Jobs**:

1. Lint and Type Check (fast, runs first)
2. Unit Tests (parallel per component)
3. Integration Tests
4. Build Verification
5. Security Scan

### Main Branch Workflow

Triggered on: `push` to `main`

**Jobs**:

1. All PR checks (lint, test, build)
2. Build and push Docker images
3. Create GitHub Release (if tagged)
4. Deploy to staging (v1.0+)

### Release Workflow

Triggered on: `push` tag matching `v*.*.*`

**Jobs**:

1. Build release artifacts
2. Create GitHub Release with binaries
3. Deploy to production (v1.0+)

---

## GitHub Actions Workflows

### 1. PR Check Workflow

`.github/workflows/pr-check.yml`:

```yaml
name: PR Check

on:
  pull_request:
    branches: [main, develop]

jobs:
  lint:
    name: Lint and Type Check
    runs-on: ubuntu-latest
    strategy:
      matrix:
        component: [control-plane, web, agent]

    steps:
      - uses: actions/checkout@v4

      # Control Plane (Go)
      - name: Setup Go
        if: matrix.component == 'control-plane'
        uses: actions/setup-go@v5
        with:
          go-version: "1.22"
          cache-dependency-path: ./go.sum

      - name: Install dependencies
        if: matrix.component == 'control-plane'
        working-directory: .
        run: go mod download

      - name: Lint
        if: matrix.component == 'control-plane'
        working-directory: .
        run: |
          go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
          golangci-lint run ./...

      - name: Format check
        if: matrix.component == 'control-plane'
        working-directory: .
        run: |
          if [ -n "$(gofmt -l .)" ]; then
            echo "Code is not formatted"
            exit 1
          fi

      # Web (React/TypeScript)
      - name: Setup Node.js
        if: matrix.component == 'web'
        uses: actions/setup-node@v4
        with:
          node-version: "20"
          cache: "npm"
          cache-dependency-path: web/package-lock.json

      - name: Install dependencies
        if: matrix.component == 'web'
        working-directory: web
        run: npm ci

      - name: ESLint
        if: matrix.component == 'web'
        working-directory: web
        run: npm run lint

      - name: TypeScript check
        if: matrix.component == 'web'
        working-directory: web
        run: npx tsc --noEmit

      # Agent (Go)
      - name: Setup Go
        if: matrix.component == 'agent'
        uses: actions/setup-go@v5
        with:
          go-version: "1.22"
          cache-dependency-path: internal/agent/go.sum

      - name: golangci-lint
        if: matrix.component == 'agent'
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest
          working-directory: internal/agent

  test:
    name: Unit Tests
    runs-on: ubuntu-latest
    needs: lint
    strategy:
      matrix:
        component: [control-plane, web, agent]

    services:
      postgres:
        image: timescale/timescaledb:latest-pg15
        env:
          POSTGRES_DB: yggdrasil_test
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v4

      # Control Plane Tests
      - name: Setup Go
        if: matrix.component == 'control-plane'
        uses: actions/setup-go@v5
        with:
          go-version: "1.22"
          cache-dependency-path: ./go.sum

      - name: Install dependencies
        if: matrix.component == 'control-plane'
        working-directory: .
        run: go mod download

      - name: Run tests
        if: matrix.component == 'control-plane'
        working-directory: .
        env:
          DATABASE_URL: postgresql://test:test@localhost:5432/yggdrasil_test?sslmode=disable
        run: go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

      - name: Upload coverage
        if: matrix.component == 'control-plane'
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
          flags: control-plane

      # Web Tests
      - name: Setup Node.js
        if: matrix.component == 'web'
        uses: actions/setup-node@v4
        with:
          node-version: "20"
          cache: "npm"
          cache-dependency-path: web/package-lock.json

      - name: Install dependencies
        if: matrix.component == 'web'
        working-directory: web
        run: npm ci

      - name: Run tests
        if: matrix.component == 'web'
        working-directory: web
        run: npm test -- --coverage

      - name: Upload coverage
        if: matrix.component == 'web'
        uses: codecov/codecov-action@v3
        with:
          files: web/coverage/coverage-final.json
          flags: web

      # Agent Tests
      - name: Setup Go
        if: matrix.component == 'agent'
        uses: actions/setup-go@v5
        with:
          go-version: "1.22"
          cache-dependency-path: internal/agent/go.sum

      - name: Run tests
        if: matrix.component == 'agent'
        working-directory: internal/agent
        run: go test ./... -race -coverprofile=coverage.out -covermode=atomic

      - name: Upload coverage
        if: matrix.component == 'agent'
        uses: codecov/codecov-action@v3
        with:
          files: internal/agent/coverage.out
          flags: agent

  integration-tests:
    name: Integration Tests
    runs-on: ubuntu-latest
    needs: test

    services:
      postgres:
        image: timescale/timescaledb:latest-pg15
        env:
          POSTGRES_DB: yggdrasil_test
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
        ports:
          - 5432:5432

      nats:
        image: nats:latest
        ports:
          - 4222:4222

    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: "1.22"

      - name: Install dependencies
        working-directory: .
        run: go mod download

      - name: Run integration tests
        working-directory: .
        env:
          DATABASE_URL: postgresql://test:test@localhost:5432/yggdrasil_test?sslmode=disable
          NATS_URL: nats://localhost:4222
        run: go test -v -tags=integration ./test/integration/...

  build:
    name: Build Verification
    runs-on: ubuntu-latest
    needs: lint
    strategy:
      matrix:
        component: [control-plane, web, agent]

    steps:
      - uses: actions/checkout@v4

      # Control Plane Docker Build
      - name: Build control-plane Docker image
        if: matrix.component == 'control-plane'
        working-directory: .
        run: docker build -t yggdrasil-control-plane:test .

      # Web Build
      - name: Setup Node.js
        if: matrix.component == 'web'
        uses: actions/setup-node@v4
        with:
          node-version: "20"
          cache: "npm"
          cache-dependency-path: web/package-lock.json

      - name: Install dependencies
        if: matrix.component == 'web'
        working-directory: web
        run: npm ci

      - name: Build web
        if: matrix.component == 'web'
        working-directory: web
        run: npm run build

      # Agent Binary Build
      - name: Setup Go
        if: matrix.component == 'agent'
        uses: actions/setup-go@v5
        with:
          go-version: "1.22"

      - name: Build agent
        if: matrix.component == 'agent'
        working-directory: internal/agent
        run: go build -o dist/agent cmd/agent/main.go

  security:
    name: Security Scan
    runs-on: ubuntu-latest
    needs: lint

    steps:
      - uses: actions/checkout@v4

      # Go control plane dependencies
      - name: Run Snyk (Go control plane)
        uses: snyk/actions/golang@master
        continue-on-error: true
        env:
          SNYK_TOKEN: ${{ secrets.SNYK_TOKEN }}
        with:
          args: --file=./go.mod

      # Node.js dependencies
      - name: Run Snyk (npm)
        uses: snyk/actions/node@master
        continue-on-error: true
        env:
          SNYK_TOKEN: ${{ secrets.SNYK_TOKEN }}
        with:
          args: --file=web/package.json

      # Go dependencies
      - name: Run Snyk (Go)
        uses: snyk/actions/golang@master
        continue-on-error: true
        env:
          SNYK_TOKEN: ${{ secrets.SNYK_TOKEN }}
        with:
          args: --file=internal/agent/go.mod
```

---

### 2. Main Branch Workflow

`.github/workflows/main.yml`:

```yaml
name: Main Branch

on:
  push:
    branches: [main]

jobs:
  # Reuse all PR checks
  pr-checks:
    uses: ./.github/workflows/pr-check.yml

  build-and-push:
    name: Build and Push Docker Images
    runs-on: ubuntu-latest
    needs: pr-checks

    steps:
      - uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Login to Docker Hub
        uses: docker/login-action@v3
        with:
          username: ${{ secrets.DOCKER_USERNAME }}
          password: ${{ secrets.DOCKER_PASSWORD }}

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: intrik8labs/yggdrasil-control-plane
          tags: |
            type=ref,event=branch
            type=sha,prefix={{branch}}-

      - name: Build and push control-plane
        uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

      # Similar for web image
      - name: Build and push web
        uses: docker/build-push-action@v5
        with:
          context: web
          push: true
          tags: intrik8labs/yggdrasil-web:${{ github.sha }}
```

---

### 3. Release Workflow

`.github/workflows/release.yml`:

```yaml
name: Release

on:
  push:
    tags:
      - "v*.*.*"

jobs:
  create-release:
    name: Create GitHub Release
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Generate changelog
        id: changelog
        run: |
          echo "## What's Changed" > changelog.txt
          git log $(git describe --tags --abbrev=0 HEAD^)..HEAD --pretty=format:"* %s (%h)" >> changelog.txt

      - name: Create Release
        uses: actions/create-release@v1
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          tag_name: ${{ github.ref_name }}
          release_name: Release ${{ github.ref_name }}
          body_path: changelog.txt
          draft: false
          prerelease: false

  build-agent-binaries:
    name: Build Agent Binaries
    runs-on: ubuntu-latest
    needs: create-release

    strategy:
      matrix:
        goos: [linux, windows, darwin]
        goarch: [amd64, arm64]
        exclude:
          - goos: windows
            goarch: arm64

    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: "1.22"

      - name: Install GoReleaser
        uses: goreleaser/goreleaser-action@v5
        with:
          version: latest
          install-only: true

      - name: Build binaries
        working-directory: internal/agent
        run: goreleaser build --snapshot --clean

      - name: Upload binaries to release
        uses: actions/upload-release-asset@v1
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          upload_url: ${{ needs.create-release.outputs.upload_url }}
          asset_path: internal/agent/dist/agent_${{ matrix.goos }}_${{ matrix.goarch }}
          asset_name: yggdrasil-agent-${{ matrix.goos }}-${{ matrix.goarch }}
          asset_content_type: application/octet-stream

  build-docker-images:
    name: Build Release Docker Images
    runs-on: ubuntu-latest
    needs: create-release

    steps:
      - uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Login to Docker Hub
        uses: docker/login-action@v3
        with:
          username: ${{ secrets.DOCKER_USERNAME }}
          password: ${{ secrets.DOCKER_PASSWORD }}

      - name: Extract version
        id: version
        run: echo "VERSION=${GITHUB_REF#refs/tags/v}" >> $GITHUB_OUTPUT

      - name: Build and push control-plane
        uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: |
            intrik8labs/yggdrasil-control-plane:${{ steps.version.outputs.VERSION }}
            intrik8labs/yggdrasil-control-plane:latest

      - name: Build and push web
        uses: docker/build-push-action@v5
        with:
          context: web
          push: true
          tags: |
            intrik8labs/yggdrasil-web:${{ steps.version.outputs.VERSION }}
            intrik8labs/yggdrasil-web:latest
```

---

## Dependabot Configuration

`.github/dependabot.yml`:

```yaml
version: 2
updates:
  # Control Plane (Go)
  - package-ecosystem: "gomod"
    directory: "/."
    schedule:
      interval: "weekly"
    open-pull-requests-limit: 5

  # Web (npm)
  - package-ecosystem: "npm"
    directory: "/web"
    schedule:
      interval: "weekly"
    open-pull-requests-limit: 5

  # Agent (Go)
  - package-ecosystem: "gomod"
    directory: "/agent"
    schedule:
      interval: "weekly"
    open-pull-requests-limit: 5

  # GitHub Actions
  - package-ecosystem: "github-actions"
    directory: "/"
    schedule:
      interval: "weekly"
```

---

## Branch Protection Rules

**For `main` branch**:

- ✅ Require pull request before merging
- ✅ Require approvals: 1 (increase when team grows)
- ✅ Require status checks to pass:
  - `lint (control-plane)`
  - `lint (web)`
  - `lint (agent)`
  - `test (control-plane)`
  - `test (web)`
  - `test (agent)`
  - `integration-tests`
  - `build (control-plane)`
  - `build (web)`
  - `build (agent)`
- ✅ Require branches to be up to date
- ✅ Require conversation resolution
- ❌ Allow force pushes (dangerous!)
- ❌ Allow deletions

---

## Secrets Configuration

**GitHub Repository Secrets** (Settings → Secrets):

- `DOCKER_USERNAME` - Docker Hub username
- `DOCKER_PASSWORD` - Docker Hub password or access token
- `SNYK_TOKEN` - Snyk API token for security scanning
- `CODECOV_TOKEN` - Codecov token for coverage reports (optional, Codecov works without for public repos)

---

## Performance Optimizations

### Caching

**Go dependencies** (control plane):

```yaml
uses: actions/setup-go@v5
with:
  go-version: "1.22"
  cache-dependency-path: ./go.sum
```

**Node.js dependencies**:

```yaml
uses: actions/setup-node@v4
with:
  cache: "npm"
  cache-dependency-path: web/package-lock.json
```

**Go dependencies** (agent):

```yaml
uses: actions/setup-go@v5
with:
  go-version: "1.22"
  cache-dependency-path: internal/agent/go.sum
```

### Parallel Execution

Use `strategy.matrix` to run jobs in parallel:

```yaml
strategy:
  matrix:
    component: [control-plane, web, agent]
```

### Conditional Steps

Skip unnecessary steps:

```yaml
- name: Run Go tests
  if: matrix.component == 'control-plane'
  working-directory: .
  run: go test -v ./...
```

---

## Deployment (v1.0+)

### Staging Deployment

Triggered on push to `develop` branch:

```yaml
deploy-staging:
  name: Deploy to Staging
  runs-on: ubuntu-latest
  environment: staging

  steps:
    - name: Deploy to staging server
      uses: appleboy/ssh-action@master
      with:
        host: ${{ secrets.STAGING_HOST }}
        username: ${{ secrets.STAGING_USER }}
        key: ${{ secrets.STAGING_SSH_KEY }}
        script: |
          cd /opt/yggdrasil
          docker-compose pull
          docker-compose up -d
```

### Production Deployment

Triggered on release tag:

```yaml
deploy-production:
  name: Deploy to Production
  runs-on: ubuntu-latest
  environment: production
  needs: build-docker-images

  steps:
    - name: Deploy to production
      # Similar to staging, but with production secrets
```

---

## Monitoring CI/CD

### GitHub Actions Insights

- View workflow runs: `https://github.com/intrik8-labs/yggdrasil/actions`
- Monitor run times and identify slow jobs
- Set up notifications for failed workflows

### Codecov Dashboard

- Coverage trends over time
- Diff coverage for PRs
- Identify untested code

---

## Best Practices

1. **Keep workflows fast**: Target <10 minutes for PR checks
2. **Fail fast**: Run linting/type checks first
3. **Use matrix builds**: Parallelize where possible
4. **Cache dependencies**: Speed up install steps
5. **Require status checks**: Enforce quality gates
6. **Auto-merge Dependabot**: For minor/patch updates (with good test coverage)
7. **Monitor workflow costs**: GitHub Actions has usage limits

---

## Troubleshooting

### Workflow fails intermittently

**Cause**: Flaky tests or network issues

**Solution**:

- Identify flaky tests (review test history)
- Add retries for network-dependent tests
- Increase timeouts if needed

### Workflow is too slow

**Cause**: Missing caches, redundant steps

**Solution**:

- Ensure all dependencies are cached
- Use matrix builds for parallelization
- Profile workflow steps to find bottlenecks

### Tests pass locally but fail in CI

**Cause**: Environment differences

**Solution**:

- Check environment variables
- Ensure database versions match
- Use Docker for consistent environments

---

## Next Steps

1. Set up GitHub repository
2. Add secrets to repository settings
3. Create initial workflows
4. Test workflows with dummy PRs
5. Enable branch protection rules
6. Integrate Codecov and Snyk

## Complete! 🎉

You now have a comprehensive architectural design for Yggdrasil v0.5. All 9 architecture documents are ready. Start building!
