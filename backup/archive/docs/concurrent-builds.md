# Concurrent BitBake Builds with Shared Caches

Smidr is designed from the ground up to support **multiple concurrent BitBake builds** while **maximizing disk efficiency** through shared caches. This document explains how it works and how to configure it correctly.

---

## 🎯 Design Goals

1. **Run multiple builds in parallel** up to CPU/memory limits
2. **Share downloads and sstate-cache** across all builds (saves GB of disk + bandwidth)
3. **Isolate TMPDIR** per build to avoid lock contention and corruption
4. **Prevent BitBake server conflicts** using unique workspace paths

---

## ✅ Safe Directory Sharing

| Directory                         | Shared?           | Notes                                                                                                                      |
| --------------------------------- | ----------------- | -------------------------------------------------------------------------------------------------------------------------- |
| `DL_DIR` (downloads)              | ✅ Yes             | All builds safely share downloaded source tarballs. BitBake handles locking.                                               |
| `SSTATE_DIR` (shared state cache) | ✅ Yes             | Can be shared across builds; BitBake handles per-object locking. Store on fast SSD for best performance.                  |
| `TMPDIR` (build output)           | ❌ No              | **NEVER share tmp/** — contains build artifacts and locks unique per build. Each container gets its own isolated TMPDIR.  |
| `LAYERS` (meta-*, poky, etc.)     | ✅ Read-only       | Mounted read-only into all containers.                                                                                     |
| `DEPLOY_DIR`                      | ⚠️ Per-build      | Each build should have its own deploy directory to avoid artifact collisions.                                              |

---

## 🏗️ How Smidr Implements This

### 1. Shared Caches (configured once in `smidr.yaml`)

```yaml
directories:
  downloads: ~/.smidr/downloads     # Shared DL_DIR for all builds
  sstate: ~/.smidr/sstate-cache     # Shared SSTATE_DIR for all builds
  layers: ~/.smidr/layers           # Shared layer checkout (read-only)
```

### 2. Isolated TMPDIR (automatic per-build)

When the daemon starts a build, it:
- Creates a unique workspace: `/home/builder/build-{buildID}` inside the container
- Lets `oe-init-build-env` set `TMPDIR` under that workspace (defaults to `tmp/`)
- **Does NOT export host TMPDIR** into the container env (critical!)

Each build gets its own:
- BitBake lock files (`bitbake.lock`, `bitbake.sock`)
- BitBake server process (isolated by workspace path)
- `tmp/work`, `tmp/deploy`, `tmp/log` directories

### 3. BitBake Server Isolation

Each build runs with:
```bash
unset BBSERVER  # Prevents connecting to external servers
bitbake {target}
```

BitBake automatically creates a unique server instance per workspace (based on `BUILDDIR`). Since each build has a unique workspace path inside the container, their servers never conflict.

---

## 📝 Example Configuration

### smidr.yaml (per-customer config)
```yaml
name: my-project
description: Custom Yocto build

yocto_series: kirkstone

base:
  provider: poky
  machine: qemux86-64
  distro: poky
  version: "4.0.20"

# Shared across all builds (high-level daemon config)
directories:
  downloads: ~/.smidr/downloads         # Shared
  sstate: ~/.smidr/sstate-cache         # Shared
  layers: ~/.smidr/layers               # Shared
  # build, tmp, deploy are set per-build by daemon

layers:
  - name: poky
    git: https://git.yoctoproject.org/poky
    branch: kirkstone

build:
  image: core-image-minimal
  machine: qemux86-64
  parallel_make: -j8
  bb_number_threads: 8

container:
  base_image: crops/yocto:ubuntu-22.04-builder
  memory: 8g
  cpu_count: 4
```

---

## 🚀 Running Concurrent Builds

### Via Daemon (Recommended)

The daemon automatically handles isolation:

```bash
# Start daemon (uses smidr.yaml config)
smidr daemon start

# Trigger multiple builds (they'll run concurrently up to CPU limit)
curl -X POST http://localhost:50051/start_build \
  -d '{"customer": "acme", "target": "core-image-minimal", "config": "acme-config.yaml"}'

curl -X POST http://localhost:50051/start_build \
  -d '{"customer": "widgetco", "target": "core-image-sato", "config": "widgetco-config.yaml"}'
```

### Via docker-compose (Manual)

For local testing or CI:

```yaml
version: "3.8"

volumes:
  yocto_downloads: {}
  yocto_sstate: {}

services:
  builder:
    image: crops/yocto:ubuntu-22.04-builder
    environment:
      HOME: /home/builder
      USER: builder
      BB_SERVER_TIMEOUT: "600"
      BB_HEARTBEAT_EVENT: "60"
      # CRITICAL: Do NOT set TMPDIR here!
    volumes:
      - yocto_downloads:/yocto/downloads
      - yocto_sstate:/yocto/sstate-cache
      - ./layers:/yocto/layers:ro
      - ./workspaces:/workspaces
    command: |
      bash -lc '
        BUILD_ID="${BUILD_ID:-$(date +%s)-$RANDOM}";
        BUILD_DIR="/workspaces/build-${BUILD_ID}";
        mkdir -p "${BUILD_DIR}";
        source /yocto/layers/poky/oe-init-build-env "${BUILD_DIR}";
        
        # Wire shared caches in local.conf
        cat >> conf/local.conf <<EOF
DL_DIR = "/yocto/downloads"
SSTATE_DIR = "/yocto/sstate-cache"
EOF
        
        unset BBSERVER;
        bitbake -c fetch core-image-minimal;
        bitbake core-image-minimal;
      '
    deploy:
      resources:
        limits:
          cpus: "4"
          memory: "8g"

# Scale to run multiple builds
# docker compose up --scale builder=2
```

---

## 🔍 Verifying Concurrent Builds

### 1. Check running containers
```bash
docker ps --filter name=smidr-build-
```
You should see multiple containers with unique names like:
- `smidr-build-acme-abc123`
- `smidr-build-widgetco-def456`

### 2. Verify isolated TMPDIR
```bash
docker exec smidr-build-acme-abc123 bash -c 'echo $BUILDDIR; ls -d tmp/'
# Should show: /home/builder/build-acme-abc123 and tmp/ exists
```

### 3. Check BitBake server processes
```bash
docker exec smidr-build-acme-abc123 ps aux | grep bitbake
# Each container should have its own bitbake cooker process
```

### 4. Monitor shared cache usage
```bash
# Downloads (should grow as new sources are fetched)
du -sh ~/.smidr/downloads

# Sstate cache (should grow significantly, shared across all builds)
du -sh ~/.smidr/sstate-cache
```

---

## ⚡ Performance Tuning

### CPU/Memory Limits (per build)
```yaml
container:
  cpu_count: 4       # Cores per build
  memory: 8g         # RAM per build

build:
  parallel_make: -j8          # Parallel compilation jobs
  bb_number_threads: 8        # Parallel BitBake tasks
```

**Rule of thumb:**
- CPU: Allocate 4-8 cores per build
- Memory: 6-8GB per build minimum
- Total concurrent builds = `floor(total_cpus / cpu_per_build)`

### Storage Optimization
- **SSD/NVMe for sstate-cache** — Critical for performance
- **Network storage OK for downloads** — Less I/O intensive
- **Clean old sstate periodically:**
  ```bash
  # Remove sstate objects older than 30 days
  find ~/.smidr/sstate-cache -type f -mtime +30 -delete
  ```

### Build Scheduling
The daemon uses a simple FIFO queue per customer:
- Builds for different customers run concurrently (up to CPU limit)
- Builds for the same customer queue serially (prevents resource thrashing)

To override, set `SMIDR_MAX_CONCURRENT_BUILDS`:
```bash
export SMIDR_MAX_CONCURRENT_BUILDS=4
smidr daemon start
```

---

## 🐛 Troubleshooting

### "Bitbake server is busy"
**Cause:** Two builds tried to use the same workspace path.  
**Fix:** Ensure each build gets a unique `BuildID`. The daemon does this automatically.

### "No active tasks" / stalled build
**Cause:** TMPDIR was exported from host into container, pointing to invalid path.  
**Fix:** Verified in v0.x.x — we no longer export TMPDIR to containers.

### Disk space issues
**Cause:** Multiple TMPDIR instances accumulate.  
**Fix:** 
- Enable `--clean` flag for one-off builds
- Periodically clean old build workspaces:
  ```bash
  find ~/.smidr/builds -type d -name "build-*" -mtime +7 -exec rm -rf {} +
  ```

### Slow sstate reuse
**Cause:** Sstate on slow disk or network storage.  
**Fix:** Move `~/.smidr/sstate-cache` to local SSD.

---

## 📊 Expected Space Usage

For reference, here's typical disk usage for concurrent Yocto builds:

| Component       | Size (per build) | Shared? | Notes                                           |
| --------------- | ---------------- | ------- | ----------------------------------------------- |
| TMPDIR          | 15-50 GB         | ❌       | Per-build, cleaned after completion             |
| SSTATE_DIR      | 20-80 GB         | ✅       | Grows over time, reused across all builds       |
| DL_DIR          | 5-15 GB          | ✅       | Grows as new sources are fetched                |
| DEPLOY_DIR      | 500 MB - 2 GB    | ❌       | Per-build, extracted to artifacts               |
| Layers          | 1-5 GB           | ✅       | Read-only, minimal overhead                     |

**Total for 3 concurrent builds:**
- Without shared caches: ~150-250 GB (3× everything)
- With shared caches: ~70-130 GB (1× shared + 3× TMPDIR)

**Savings: ~50-60% disk usage**

---

## 🎓 Best Practices

1. **Always use unique build IDs** — The daemon does this automatically
2. **Share downloads and sstate** — Configure once in smidr.yaml
3. **Never share TMPDIR** — One per build, always
4. **Mount layers read-only** — Prevents accidental corruption
5. **Use fast storage for sstate** — SSD/NVMe strongly recommended
6. **Set resource limits** — Prevent one build from starving others
7. **Clean old builds** — Automate cleanup of stale TMPDIR instances
8. **Monitor disk usage** — Set up alerts for >80% usage

---

## 🔗 Related Documentation

- [Cache Design](./cache.md) — Deep dive into sstate and download management
- [Daemon Architecture](./daemon.md) — How the daemon schedules builds
- [Container Backend](./container-backend-design.md) — Container isolation details

---

**Summary:** Smidr's concurrent build support is production-ready. Just configure shared `downloads` and `sstate` directories once, and the daemon handles the rest — unique workspaces, isolated TMPDIR, and no server conflicts.
