// Minimal REST/SSE bridge to the gRPC daemon plus static UI
const path = require("path");
const express = require("express");
const cors = require("cors");
const grpc = require("@grpc/grpc-js");
const protoLoader = require("@grpc/proto-loader");

const app = express();
app.use(cors());
app.use(express.json());
app.use(express.static(path.join(__dirname, "public")));

const DAEMON_ADDR = process.env.SMIDR_DAEMON_ADDR || "localhost:50051";
const PROTO_PATH = path.join(__dirname, "..", "api", "proto", "smidr.proto");

const pkgDef = protoLoader.loadSync(PROTO_PATH, {
  keepCase: true,
  longs: String,
  enums: String,
  defaults: true,
  oneofs: true,
});
const smidrProto = grpc.loadPackageDefinition(pkgDef).smidr.v1;
const client = new smidrProto.Smidr(
  DAEMON_ADDR,
  grpc.credentials.createInsecure()
);

// API: list builds
app.get("/api/builds", (req, res) => {
  client.ListBuilds({}, (err, resp) => {
    if (err) return res.status(500).json({ error: err.message });
    res.json(resp || { builds: [] });
  });
});

// API: start build
app.post("/api/start", (req, res) => {
  const { config, target, customer, forceClean, forceImage } = req.body || {};
  client.StartBuild(
    {
      config: config || "",
      target: target || "",
      customer: customer || "",
      force_clean: !!forceClean,
      force_image: !!forceImage,
    },
    (err, resp) => {
      if (err) return res.status(500).json({ error: err.message });
      res.json(resp || {});
    }
  );
});

// API: status
app.get("/api/status/:id", (req, res) => {
  client.GetBuildStatus({ build_id: req.params.id }, (err, resp) => {
    if (err) return res.status(500).json({ error: err.message });
    res.json(resp || {});
  });
});

// API: cancel
app.post("/api/cancel/:id", (req, res) => {
  client.CancelBuild({ build_id: req.params.id }, (err, resp) => {
    if (err) return res.status(500).json({ error: err.message });
    res.json(resp || {});
  });
});

// API: artifacts
app.get("/api/artifacts/:id", (req, res) => {
  client.ListArtifacts({ build_id: req.params.id }, (err, resp) => {
    if (err) return res.status(500).json({ error: err.message });
    res.json(resp || { artifacts: [] });
  });
});

// SSE: logs stream
app.get("/api/logs/:id", (req, res) => {
  res.set({
    "Content-Type": "text/event-stream",
    "Cache-Control": "no-cache",
    Connection: "keep-alive",
    "X-Accel-Buffering": "no",
  });
  res.flushHeaders();

  const stream = client.StreamLogs({ build_id: req.params.id, follow: true });
  const keepalive = setInterval(() => res.write(":\n\n"), 15000);

  stream.on("data", (msg) => {
    try {
      const data = JSON.stringify(msg);
      res.write("event: log\n");
      res.write(`data: ${data}\n\n`);
    } catch (e) {
      // ignore
    }
  });
  stream.on("error", (err) => {
    res.write("event: error\n");
    res.write(`data: ${JSON.stringify({ error: err.message })}\n\n`);
    clearInterval(keepalive);
    res.end();
  });
  stream.on("end", () => {
    res.write("event: end\n");
    res.write("data: {}\n\n");
    clearInterval(keepalive);
    res.end();
  });

  req.on("close", () => {
    try {
      stream.cancel();
    } catch (_) {}
    clearInterval(keepalive);
  });
});

const PORT = process.env.PORT || 5173;
app.listen(PORT, () => {
  console.log(
    `[webui] listening on http://localhost:${PORT} (daemon: ${DAEMON_ADDR})`
  );
});
