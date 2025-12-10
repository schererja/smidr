# Documentation

Smidr streamlines Yocto/BitBake builds with containerized execution, shared caches, and clear configuration.

## Key topics

- Caches and directories
  - `directories.layers` — clone cache of layer repositories; Smidr scans for `conf/layer.conf` and auto-includes compatible sublayers
  - `directories.downloads` — BitBake `DL_DIR`
  - `directories.sstate` — shared state cache; in containers Smidr prefers `SSTATE_MIRRORS` to a bind mount and maps to `/home/builder/sstate-cache`
  - `directories.tmp` — `TMPDIR`; set to a host-writable path to avoid container permission issues
  - `directories.deploy` — artifacts output

- Container environment
  - Base image: `crops/yocto:ubuntu-22.04-builder` (configurable)
  - Host caches are bind-mounted into `/home/builder/...` and referenced in generated `local.conf`
  - Parallelism is controlled via `BB_NUMBER_THREADS` and `PARALLEL_MAKE`

- CI workflows and tiers
  - Fast validation tiers restore from sstate mirrors and pre-restored downloads
  - Artifact tiers perform cleansstate/full builds to produce fresh images and SDKs
  - Nightly jobs can publish sstate/downloads to accelerate PR runs

- Layer discovery and filtering
  - Smidr excludes test/example layers (e.g., `meta-selftest`, `meta-skeleton`, `/tests/`, `/testdata/`)
  - Optional `yocto_series` restricts layers via `LAYERSERIES_COMPAT_*`

## References

- [Cache & Source Management](cache.md)
- [Concurrent Builds](concurrent-builds.md) — Running multiple builds with shared caches
- [Container Backend Design](container-backend-design.md)
- [Daemon Architecture](daemon.md)
- [Troubleshooting](troubleshooting.md)

- [Smidr Daemon (gRPC Server)](daemon.md)
