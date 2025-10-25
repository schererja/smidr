const api = {
  async listBuilds() {
    const r = await fetch("/api/builds");
    return r.json();
  },
  async startBuild(payload) {
    const r = await fetch("/api/start", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    return r.json();
  },
  async getStatus(id) {
    const r = await fetch(`/api/status/${encodeURIComponent(id)}`);
    return r.json();
  },
  async cancel(id) {
    const r = await fetch(`/api/cancel/${encodeURIComponent(id)}`, {
      method: "POST",
    });
    return r.json();
  },
  async listArtifacts(id) {
    const r = await fetch(`/api/artifacts/${encodeURIComponent(id)}`);
    return r.json();
  },
  streamLogs(id, onLine) {
    const ev = new EventSource(`/api/logs/${encodeURIComponent(id)}`);
    ev.addEventListener("log", (e) => {
      try {
        onLine(JSON.parse(e.data));
      } catch {}
    });
    ev.addEventListener("end", () => ev.close());
    ev.addEventListener("error", () => ev.close());
    return ev;
  },
};

const els = {
  daemon: document.getElementById("daemonAddr"),
  config: document.getElementById("config"),
  target: document.getElementById("target"),
  customer: document.getElementById("customer"),
  forceClean: document.getElementById("forceClean"),
  forceImage: document.getElementById("forceImage"),
  startBtn: document.getElementById("startBtn"),
  startResult: document.getElementById("startResult"),
  refreshBuilds: document.getElementById("refreshBuilds"),
  buildsTable: document.getElementById("buildsTable").querySelector("tbody"),
  logs: document.getElementById("logs"),
  logsTitle: document.getElementById("logsTitle"),
  artTitle: document.getElementById("artTitle"),
  artifactsTable: document
    .getElementById("artifactsTable")
    .querySelector("tbody"),
};

let currentLogStream = null;

function fmtUnixSeconds(s) {
  if (!s) return "";
  const d = new Date(Number(s) * 1000);
  return isNaN(d.getTime()) ? "" : d.toLocaleString();
}
function setLogsTitle(id) {
  els.logsTitle.textContent = id || "none";
}
function setArtTitle(id) {
  els.artTitle.textContent = id || "none";
}
function appendLog(line) {
  const s = `[${line.stream || "stdout"}] ${line.content || ""}\n`;
  els.logs.textContent += s;
  els.logs.scrollTop = els.logs.scrollHeight;
}
function renderBuilds(list) {
  els.buildsTable.innerHTML = "";
  (list.builds || []).forEach((b) => {
    const tr = document.createElement("tr");
    tr.innerHTML = `
      <td class="mono">${b.build_id}</td>
      <td class="state-${b.state}">${b.state}</td>
      <td>${b.target || ""}</td>
      <td>${fmtUnixSeconds(b.started_at)}</td>
      <td>${b.exit_code ?? ""}</td>
      <td>
        <button data-act="logs" data-id="${b.build_id}">Logs</button>
        <button data-act="status" data-id="${b.build_id}">Status</button>
        <button data-act="artifacts" data-id="${b.build_id}">Artifacts</button>
        <button data-act="cancel" data-id="${b.build_id}">Cancel</button>
      </td>
    `;
    els.buildsTable.appendChild(tr);
  });
}
async function refreshBuilds() {
  const data = await api.listBuilds();
  renderBuilds(data);
}
async function showArtifacts(id) {
  setArtTitle(id);
  const data = await api.listArtifacts(id);
  els.artifactsTable.innerHTML = "";
  (data.artifacts || []).forEach((a) => {
    const tr = document.createElement("tr");
    tr.innerHTML = `
      <td class="mono">${a.name}</td>
      <td>${a.size || 0}</td>
      <td class="mono">${a.checksum || ""}</td>
      <td class="mono">${a.path || ""}</td>
    `;
    els.artifactsTable.appendChild(tr);
  });
}

function bind() {
  els.startBtn.addEventListener("click", async () => {
    const payload = {
      config: els.config.value.trim(),
      target: els.target.value.trim(),
      customer: els.customer.value.trim(),
      forceClean: els.forceClean.checked,
      forceImage: els.forceImage.checked,
    };
    const resp = await api.startBuild(payload);
    els.startResult.textContent = JSON.stringify(resp, null, 2);
    if (resp.build_id) {
      setLogsTitle(resp.build_id);
      if (currentLogStream) currentLogStream.close();
      currentLogStream = api.streamLogs(resp.build_id, appendLog);
      await refreshBuilds();
    }
  });

  els.refreshBuilds.addEventListener("click", refreshBuilds);

  els.buildsTable.addEventListener("click", async (e) => {
    const btn = e.target.closest("button");
    if (!btn) return;
    const id = btn.getAttribute("data-id");
    const act = btn.getAttribute("data-act");
    if (act === "logs") {
      setLogsTitle(id);
      els.logs.textContent = "";
      if (currentLogStream) currentLogStream.close();
      currentLogStream = api.streamLogs(id, appendLog);
    } else if (act === "status") {
      const s = await api.getStatus(id);
      alert(JSON.stringify(s, null, 2));
    } else if (act === "artifacts") {
      await showArtifacts(id);
    } else if (act === "cancel") {
      await api.cancel(id);
      await refreshBuilds();
    }
  });
}

(async function init() {
  const daemon = window.SMIDR_DAEMON_ADDR || "localhost:50051";
  document.getElementById("daemonAddr").textContent = daemon;
  await refreshBuilds();
  bind();
})();
