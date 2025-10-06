# Cache & Source Management

Smidr's source management subsystem provides persistent caching and robust download behavior.

Key points:

- Persistent cache locations:
  - `sourcesDir`: git layer clones are stored here.
  - `downloadDir`: downloaded archives and files are stored here.
- Metadata:
  - For every cached object, Smidr writes a companion metadata file named `<object>.smidr_meta.json` that contains a `last_access` timestamp.
  - Metadata is used to make eviction decisions and for observability.
- Eviction:
  - Use the `EvictOldCache(ttl time.Duration)` API to remove objects not accessed within `ttl`.
  - By design, entries without a metadata file are skipped to avoid unintended deletions.
- Locking:
  - Per-repo lockfiles prevent concurrent fetches from corrupting clones when multiple processes run simultaneously.
- Mirrors & Retries:
  - Downloaders accept a list of mirrors and will try mirrors in order.
  - Each mirror is retried with exponential backoff before falling back to the next mirror.

Configuration snippet:

```yaml
layers:
  - name: meta-mycompany
    git: https://git.example.com/meta-mycompany
    mirrors:
      - https://mirror1.example.com/meta-mycompany.tar.gz
      - https://mirror2.example.com/meta-mycompany.tar.gz
    branch: main
```

Developer notes:

- Metadata file format: JSON `{ "last_access": "<RFC3339 timestamp>" }`.
- Metadata helpers live in `internal/source/cachemeta.go`.
- Logs include per-attempt messages for retries and mirror fallbacks.
