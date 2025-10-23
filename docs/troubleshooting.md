# Troubleshooting

Common issues and how to resolve them when using Smidr with Yocto/BitBake.

## Duplicate BBFILE_COLLECTIONS or meta-layer collisions

- Symptom: Parser errors like "BBFILE_COLLECTIONS: layer XYZ is already defined".
- Cause: Test/example layers auto-added into BBLAYERS, often from nested repos.
- Fix:
  - Smidr filters `/tests/`, `/testdata/`, `meta-selftest`, and `meta-skeleton` automatically during layer discovery.
  - If duplicates persist, remove any hand-edited `conf/bblayers.conf` and let Smidr regenerate it.
  - Clean up stale or duplicated repos under your `directories.layers` cache.

## EULA required for NXP/Freescale components

- Symptom: Errors referencing i.MX GPU/VPU recipes or Freescale/NXP EULA.
- Fix: Set this in your config to explicitly accept the EULA for affected builds:

  ```yaml
  advanced:
    accept_fsl_eula: true
  ```

  Smidr writes `ACCEPT_FSL_EULA = "1"` into `local.conf`. This is always opt-in.

## Threads stuck at 2 in local.conf

- Symptom: `BB_NUMBER_THREADS` and `PARALLEL_MAKE` show 2 in `local.conf`.
- Causes:
  - The intended config file wasn’t loaded.
  - Parallelism not set in the config.
- Fixes:
  - Run with an explicit config file: `smidr --config smidr-ci.yaml build`.
  - Set `build.bb_number_threads` and `build.parallel_make` in your config.
  - Smidr prints the parsed values during container setup for easy verification.

## Permission errors under TMPDIR or DEPLOY in containers

- Symptom: BitBake fails writing to TMPDIR or deploy dirs due to permissions.
- Fix:
  - Set `directories.tmp` and `directories.deploy` to host-writable paths (e.g., under `~/.smidr/...`).
  - Use the env overrides from the CI configs: `YOCTO_TMP_DIR`, `YOCTO_DEPLOY_DIR`.
  - Smidr maps these into the container (`/home/builder/tmp`, `.../deploy`) and writes `TMPDIR` accordingly.

## Sstate-only validation produces no new artifacts

- Symptom: CI validation tier completes quickly but doesn’t produce new images/SDKs.
- Explanation: Validation tiers are optimized to restore from `SSTATE_MIRRORS` and verify correctness quickly.
- Options to force artifact generation:
  - Disable/change `advanced.sstate_mirrors` for that job.
  - Run a cleansstate/full build tier.
  - Configure a nightly job to publish sstate/downloads, leaving PR tiers fast.

## Meta-layer compatibility with Yocto series

- Symptom: Layer compatibility warnings or parse failures.
- Fix:
  - Set `yocto_series` (e.g., `kirkstone`) in your config to restrict included layers to those declaring compatibility via `LAYERSERIES_COMPAT_*`.
  - Verify you’ve checked out branches for your BSP that match your selected series.
