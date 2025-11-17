# GitHub Workflows

Streamlined CI/CD workflows for the Smidr project.

## Workflows

### `ci.yml` - Main CI Pipeline

**Triggers:** Push to main, Pull requests
**Purpose:** Fast feedback for code changes

- ✅ Go build and unit tests
- ✅ Integration tests (on main branch only)
- ✅ Yocto smoke test (PR only)

### `generate-sdks.yml` - Auto-generate SDK Code

**Triggers:** Proto file changes
**Purpose:** Keep SDK clients in sync

- Runs `buf generate` when `.proto` files change
- Creates PR with generated code
- Requires manual review before merge

### `publish-sdks.yml` - Release SDKs

**Triggers:** Version tags (`v*`)
**Purpose:** Publish to package managers

- Publishes C# SDK to NuGet
- Publishes TypeScript SDK to npm
- Updates separate SDK repositories

### `yocto-builds.yml` - Full Yocto Builds

**Triggers:** Manual, Weekly schedule
**Purpose:** Test complete builds for all BSPs

- Supports multiple BSP configurations (Poky, Toradex, Raspberry Pi, Intel)
- Optional force clean rebuild
- Uploads build artifacts
- Runs weekly on Mondays

## Removed Workflows

The following workflows were consolidated:

- ❌ `go.yml` → merged into `ci.yml`
- ❌ `yocto-ci.yml` → simplified to smoke test in `ci.yml`
- ❌ `multi-layer-test.yml` → consolidated into `yocto-builds.yml`
- ❌ `yocto-toradex.yml` → consolidated into `yocto-builds.yml`
- ❌ `publish-csharp.yml` → merged into `publish-sdks.yml`
- ❌ `publish-ts.yml` → merged into `publish-sdks.yml`

## Benefits

### Before: 7 workflows

- Duplicate Go setup code
- Separate files for each BSP
- Split SDK publishing

### After: 4 workflows

- ✅ Single source of truth for CI
- ✅ Matrix strategy for BSP testing
- ✅ Combined SDK publishing
- ✅ Reduced maintenance overhead
- ✅ Faster to understand

## Usage

### Run smoke test on PR

Happens automatically when you open a PR.

### Test specific BSP manually

```bash
gh workflow run yocto-builds.yml -f config=toradex
```

### Publish new SDK version

```bash
git tag v1.2.3
git push origin v1.2.3
```

### Generate SDKs from proto changes

Happens automatically when you push proto file changes.
