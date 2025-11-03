# SMIDR GUI (dev/testing)

A minimal web UI and REST/SSE bridge for the SMIDR gRPC daemon. This is intended for local testing and tinkering, not production.

## Prerequisites

- Node.js 18+
- SMIDR daemon running (default: `localhost:50051`)

## Quick start

```bash
cd gui
npm install
SMIDR_DAEMON_ADDR=localhost:50051 npm start
# open http://localhost:5173
```

If your daemon runs elsewhere, set the environment variable accordingly:

```bash
SMIDR_DAEMON_ADDR=smidr-server.ik8labs.local:50051 npm start
```

## What it exposes

- REST endpoints bridging to gRPC:
  - `GET /api/builds`
  - `POST /api/start` { config, target, customer, forceClean, forceImage }
  - `GET /api/status/:id`
  - `POST /api/cancel/:id`
  - `GET /api/artifacts/:id`
- SSE stream for logs:
  - `GET /api/logs/:id`

## Notes

- Uses `@grpc/grpc-js` + `@grpc/proto-loader` to load `api/proto/smidr.proto` on the fly.
- Field names in JSON use proto snake_case (e.g., `build_id`, `started_at`).
- Logs are streamed as Server-Sent Events (SSE).
- Static UI is served from `gui/public`.
