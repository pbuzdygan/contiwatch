const statusEl = document.getElementById("status");
const lastScanEl = document.getElementById("last-scan");
const scanBtn = document.getElementById("scan-btn");
const stopBtn = document.getElementById("stop-btn");
const scanStateEl = document.getElementById("scan-state");
const schedulerStateEl = document.getElementById("scheduler-state");
const toastEl = document.getElementById("toast");
const statusHintEl = document.getElementById("status-hint");
const configForm = document.getElementById("config-form");
const schedulerEnabledInput = document.getElementById("scheduler-enabled");
const scanIntervalInput = document.getElementById("scan-interval");
const globalPolicySelect = document.getElementById("global-policy");
const discordInput = document.getElementById("discord-url");
const discordEnabledInput = document.getElementById("discord-enabled");
const updateStoppedInput = document.getElementById("update-stopped");
const localForm = document.getElementById("local-form");
const localNameInput = document.getElementById("local-name");
const localSocketInput = document.getElementById("local-socket");
const remoteForm = document.getElementById("remote-form");
const remoteNameInput = document.getElementById("remote-name");
const remoteUrlInput = document.getElementById("remote-url");
const remoteTokenInput = document.getElementById("remote-token");
const generateTokenBtn = document.getElementById("generate-token");
const serversListEl = document.getElementById("servers-list");
const menuBtn = document.getElementById("menu-btn");
const sidebar = document.getElementById("sidebar");
const sidebarCloseBtn = document.getElementById("sidebar-close");
const overlay = document.getElementById("overlay");
const sidebarPin = document.getElementById("sidebar-pin");
const sidebarSchedulerInput = document.getElementById("sidebar-scheduler");
const viewStatusEl = document.getElementById("view-status");
const viewSettingsEl = document.getElementById("view-settings");
const viewServersEl = document.getElementById("view-servers");
const viewLogsEl = document.getElementById("view-logs");
const backFromSettingsBtn = document.getElementById("back-from-settings");
const backFromServersBtn = document.getElementById("back-from-servers");
const backFromLogsBtn = document.getElementById("back-from-logs");
const refreshLogsBtn = document.getElementById("refresh-logs");
const clearLogsBtn = document.getElementById("clear-logs");
const logsLevelSelect = document.getElementById("logs-level");
const logsStatusEl = document.getElementById("logs-status");
const logsListEl = document.getElementById("logs-list");

let currentScanController = null;
let currentView = "status";
let updateInProgress = false;
let logsRefreshTimer = null;
let currentConfig = null;
let cachedLocals = [];
let cachedRemotes = [];

async function fetchJSON(url, options = {}) {
  const res = await fetch(url, options);
  if (!res.ok) {
    const payload = await res.json().catch(() => ({}));
    const message = payload.error || `Request failed: ${res.status}`;
    throw new Error(message);
  }
  return res.json();
}

function displayImage(imageRef) {
  if (!imageRef) return "";
  if (imageRef.startsWith("sha256:")) return "";
  return imageRef.split("@")[0];
}

function generateToken(length = 32) {
  const bytes = new Uint8Array(length / 2);
  crypto.getRandomValues(bytes);
  return Array.from(bytes, (b) => b.toString(16).padStart(2, "0")).join("");
}

function maskToken(token) {
  if (!token) return "";
  if (token.length <= 8) return "********";
  return `${token.slice(0, 4)}…${token.slice(-4)}`;
}

function renderStatus(results) {
  statusEl.innerHTML = "";
  if (!results || results.length === 0) {
    statusEl.textContent = "No scan data yet.";
    return;
  }

  const canUpdateStopped = Boolean(currentConfig && currentConfig.update_stopped_containers);
  const hasLocalServers = Array.isArray(cachedLocals) && cachedLocals.length > 0;
  if (!hasLocalServers) {
    statusHintEl.textContent = "No local servers configured. Add a Docker socket in Servers to use Run scan.";
    statusHintEl.classList.remove("hidden");
  } else {
    statusHintEl.classList.add("hidden");
  }

  results.forEach((result) => {
    const card = document.createElement("div");
    card.className = "server-card";

    const title = document.createElement("h3");
    const statusLabel = result.error ? `offline (${result.error})` : "online";
    title.textContent = `${result.server_name} (${result.local ? "local" : "remote"})`;
    card.appendChild(title);

    const meta = document.createElement("div");
    meta.className = "tag";
    const checkedAt = result.checked_at ? new Date(result.checked_at) : null;
    const hasCheckedAt = checkedAt && !Number.isNaN(checkedAt.getTime()) && checkedAt.getTime() > 0;
    meta.textContent = hasCheckedAt ? checkedAt.toLocaleString() : "Scan not run yet";
    card.appendChild(meta);

    if (!result.local || result.error) {
      const statusTag = document.createElement("span");
      statusTag.className = "tag";
      statusTag.textContent = `status: ${statusLabel}`;
      card.appendChild(statusTag);
    }

    if (!hasCheckedAt) {
      const empty = document.createElement("p");
      empty.textContent = "Scan has not been run yet.";
      card.appendChild(empty);
    } else if (!result.containers || result.containers.length === 0) {
      const empty = document.createElement("p");
      empty.textContent = "No containers found.";
      card.appendChild(empty);
    } else {
      const updateRequired = result.containers.filter((container) => container.update_available);
      const upToDate = result.containers.filter((container) => !container.update_available);

      const grid = document.createElement("div");
      grid.className = "results-grid";

      const leftCol = document.createElement("div");
      leftCol.className = "results-col";
      const leftTitle = document.createElement("h4");
      leftTitle.textContent = "Update required";
      leftCol.appendChild(leftTitle);

      const rightCol = document.createElement("div");
      rightCol.className = "results-col";
      const rightTitle = document.createElement("h4");
      rightTitle.textContent = "Up to date";
      rightCol.appendChild(rightTitle);

      const renderContainerRow = (container) => {
        const row = document.createElement("div");
        row.className = "container-row";
        const status = container.update_available
          ? container.updated
            ? "updated"
            : "update available"
          : "up to date";
        const image = displayImage(container.image);
        row.innerHTML = `
          <strong>${container.name}</strong>
        `;
        if (image) {
          const imageSpan = document.createElement("span");
          imageSpan.textContent = image;
          row.appendChild(imageSpan);
        }
        const statusSpan = document.createElement("span");
        statusSpan.className = "tag";
        statusSpan.textContent = status;
        row.appendChild(statusSpan);
        const policySpan = document.createElement("span");
        policySpan.className = "tag";
        policySpan.textContent = `policy: ${container.policy}`;
        row.appendChild(policySpan);

        const canUpdate =
          result.local &&
          container.update_available &&
          !container.updated &&
          !container.paused &&
          !currentScanController &&
          !updateInProgress &&
          (container.running || canUpdateStopped);

        if (container.updated) {
          const actions = document.createElement("div");
          actions.className = "row-actions";
          const updatedBtn = document.createElement("button");
          updatedBtn.type = "button";
          updatedBtn.className = "btn-small btn-success";
          updatedBtn.textContent = "Updated";
          updatedBtn.disabled = true;
          actions.appendChild(updatedBtn);
          row.appendChild(actions);
        } else if (result.local && container.update_available) {
          const actions = document.createElement("div");
          actions.className = "row-actions";

          const updateBtn = document.createElement("button");
          updateBtn.type = "button";
          updateBtn.className = "btn-small";
          updateBtn.textContent = "Update";
          updateBtn.disabled = !canUpdate;

          const infoBtn = document.createElement("button");
          infoBtn.type = "button";
          infoBtn.className = "btn-small secondary";
          infoBtn.textContent = "Info";

          const updateState = document.createElement("span");
          updateState.className = "tag hidden";

          const infoPanel = document.createElement("div");
          infoPanel.className = "info-panel hidden";
          const currentImageID = container.image_id || "unknown";
          const newImageID = container.new_image_id || "unknown";
          infoPanel.textContent = `Current image: ${currentImageID}\nNew image: ${newImageID}`;

          updateBtn.addEventListener("click", async () => {
            if (updateInProgress || currentScanController) return;
            updateInProgress = true;
            scanBtn.disabled = true;
            updateBtn.disabled = true;
            updateState.classList.remove("hidden");
            updateState.textContent = "Updating…";
            let updateResult = null;
            try {
              const serverParam = encodeURIComponent(result.server_name || "local");
              updateResult = await fetchJSON(`/api/update/${encodeURIComponent(container.id)}?server=${serverParam}`, { method: "POST" });
              if (updateResult.updated) {
                updateState.textContent = `Updated (${updateResult.previous_state} → ${updateResult.current_state})`;
              } else {
                updateState.textContent = updateResult.message || "Not updated";
              }
            } catch (err) {
              updateState.textContent = `Error: ${err.message}`;
            } finally {
              updateInProgress = false;
              scanBtn.disabled = Boolean(currentScanController);
              updateBtn.disabled = false;
              setTimeout(() => updateState.classList.add("hidden"), 8000);
              await refreshStatus();
              await refreshLogs();
              if (updateResult) {
                if (updateResult.updated) {
                  showToast(`Updated ${updateResult.name} (${updateResult.previous_state} → ${updateResult.current_state})`);
                } else {
                  showToast(`${updateResult.name}: ${updateResult.message || "Not updated"}`);
                }
              }
            }
          });

          infoBtn.addEventListener("click", () => {
            infoPanel.classList.toggle("hidden");
          });

          actions.appendChild(updateBtn);
          actions.appendChild(infoBtn);
          actions.appendChild(updateState);
          row.appendChild(actions);
          row.appendChild(infoPanel);
        }

        return row;
      };

      if (updateRequired.length === 0) {
        const empty = document.createElement("p");
        empty.textContent = "No updates required.";
        leftCol.appendChild(empty);
      } else {
        updateRequired.forEach((container) => leftCol.appendChild(renderContainerRow(container)));
      }

      if (upToDate.length === 0) {
        const empty = document.createElement("p");
        empty.textContent = "No containers.";
        rightCol.appendChild(empty);
      } else {
        upToDate.forEach((container) => rightCol.appendChild(renderContainerRow(container)));
      }

      grid.appendChild(leftCol);
      grid.appendChild(rightCol);
      card.appendChild(grid);
    }

    statusEl.appendChild(card);
  });
}

function renderServers(localServers, remoteServers) {
  serversListEl.innerHTML = "";
  const items = [
    ...localServers.map((server) => ({ type: "local", server })),
    ...remoteServers.map((server) => ({ type: "remote", server })),
  ];

  if (items.length === 0) {
    const empty = document.createElement("li");
    empty.textContent = "No servers added.";
    serversListEl.appendChild(empty);
    return;
  }

  items.forEach((item) => {
    const li = document.createElement("li");
    const tokenLabel = item.type === "remote"
      ? (item.server.token ? ` (${maskToken(item.server.token)})` : " (token missing)")
      : "";
    const label = item.type === "local"
      ? `${item.server.name} - ${item.server.socket}`
      : `${item.server.name} - ${item.server.url}${tokenLabel}`;
    li.textContent = `${item.type.toUpperCase()}: ${label}`;

    const editBtn = document.createElement("button");
    editBtn.textContent = "Edit";
    editBtn.className = "secondary";
    editBtn.addEventListener("click", () => {
      if (item.type === "local") {
        localNameInput.value = item.server.name;
        localSocketInput.value = item.server.socket;
      } else {
        remoteNameInput.value = item.server.name;
        remoteUrlInput.value = item.server.url;
        remoteTokenInput.value = item.server.token || "";
      }
      showToast(`Editing ${item.type} server: ${item.server.name}`);
    });

    const removeBtn = document.createElement("button");
    removeBtn.textContent = "Remove";
    removeBtn.addEventListener("click", async () => {
      const path = item.type === "local" ? "/api/locals/" : "/api/servers/";
      await fetchJSON(`${path}${encodeURIComponent(item.server.name)}`, { method: "DELETE" });
      await refreshServers();
      await refreshStatus();
    });

    li.appendChild(editBtn);
    li.appendChild(removeBtn);
    serversListEl.appendChild(li);
  });
}

async function refreshConfig() {
  const cfg = await fetchJSON("/api/config");
  currentConfig = cfg;
  schedulerEnabledInput.checked = Boolean(cfg.scheduler_enabled);
  scanIntervalInput.value = cfg.scan_interval_sec;
  globalPolicySelect.value = cfg.global_policy;
  discordInput.value = cfg.discord_webhook_url || "";
  discordEnabledInput.checked = cfg.discord_notifications_enabled !== false;
  updateStoppedInput.checked = Boolean(cfg.update_stopped_containers);
  sidebarSchedulerInput.checked = Boolean(cfg.scheduler_enabled);

  const enabled = Boolean(cfg.scheduler_enabled);
  const interval = Number(cfg.scan_interval_sec);
  if (enabled) {
    schedulerStateEl.textContent = Number.isFinite(interval) && interval > 0
      ? `Scheduler: on (${interval}s)`
      : "Scheduler: on";
  } else {
    schedulerStateEl.textContent = "Scheduler: off";
  }
}

async function refreshServers() {
  const [locals, remotes] = await Promise.all([
    fetchJSON("/api/locals"),
    fetchJSON("/api/servers"),
  ]);
  cachedLocals = locals || [];
  cachedRemotes = remotes || [];
  renderServers(cachedLocals, cachedRemotes);
  if (!cachedLocals.length) {
    statusHintEl.textContent = "No local servers configured. Add a Docker socket in Servers to use Run scan.";
    statusHintEl.classList.remove("hidden");
  } else {
    statusHintEl.classList.add("hidden");
  }
}

async function refreshStatus() {
  const results = await fetchJSON("/api/aggregate");
  if (Array.isArray(results)) {
    results.sort((a, b) => {
      if (a.local === b.local) return 0;
      return a.local ? -1 : 1;
    });
  }
  const firstCheckedAt = results && results[0] ? results[0].checked_at : "";
  if (firstCheckedAt) {
    const ts = new Date(firstCheckedAt);
    if (!Number.isNaN(ts.getTime()) && ts.getTime() > 0) {
      lastScanEl.textContent = `Last scan: ${ts.toLocaleString()}`;
    } else {
      lastScanEl.textContent = "Last scan: not run yet";
    }
  } else {
    lastScanEl.textContent = "Last scan: not run yet";
  }
  renderStatus(results);
}

async function refreshLogs() {
  if (!logsListEl) return;
  if (viewLogsEl.classList.contains("hidden")) return;
  logsStatusEl.textContent = "Loading...";
  try {
    if (!cachedLocals || cachedLocals.length === 0) {
      logsListEl.innerHTML = "";
      const empty = document.createElement("p");
      empty.textContent = "No local servers configured. Add a Docker socket to see operational logs.";
      logsListEl.appendChild(empty);
      logsStatusEl.textContent = "Idle";
      return;
    }

    const logs = await fetchJSON("/api/logs");
    const level = logsLevelSelect ? logsLevelSelect.value : "all";
    const list = Array.isArray(logs) ? logs : [];
    const filtered = level === "all" ? list : list.filter((entry) => entry.level === level);
    filtered.sort((a, b) => {
      if (typeof a.seq === "number" && typeof b.seq === "number") {
        return b.seq - a.seq;
      }
      return new Date(b.timestamp) - new Date(a.timestamp);
    });
    logsListEl.innerHTML = "";
    if (!filtered || filtered.length === 0) {
      const empty = document.createElement("p");
      empty.textContent = "No logs yet.";
      logsListEl.appendChild(empty);
      logsStatusEl.textContent = "Idle";
      return;
    }
    filtered.forEach((entry) => {
      const row = document.createElement("div");
      row.className = "log-row";
      const ts = new Date(entry.timestamp).toLocaleString();
      row.innerHTML = `
        <span class="tag">${ts}</span>
        <span class="level">${entry.level}</span>
        <span>${entry.message}</span>
      `;
      logsListEl.appendChild(row);
    });
    logsStatusEl.textContent = "Idle";
  } catch (err) {
    logsStatusEl.textContent = "Unavailable";
    logsListEl.innerHTML = "";
    const empty = document.createElement("p");
    empty.textContent = "Logs unavailable. Add a server in Servers to enable operational logs.";
    logsListEl.appendChild(empty);
  }
}

function startLogsAutoRefresh() {
  stopLogsAutoRefresh();
  refreshLogs();
  logsRefreshTimer = window.setInterval(refreshLogs, 5000);
}

function stopLogsAutoRefresh() {
  if (logsRefreshTimer) {
    window.clearInterval(logsRefreshTimer);
    logsRefreshTimer = null;
  }
}

function setScanningUI(isScanning) {
  scanBtn.disabled = isScanning || updateInProgress;
  scanStateEl.classList.toggle("hidden", !isScanning);
  stopBtn.classList.toggle("hidden", !isScanning);
}

function showToast(message, timeoutMs = 8000) {
  if (!toastEl) return;
  if (!message) return;
  toastEl.textContent = message;
  toastEl.classList.remove("hidden");
  window.clearTimeout(showToast._t);
  showToast._t = window.setTimeout(() => toastEl.classList.add("hidden"), timeoutMs);
}

function setView(nextView) {
  currentView = nextView;
  viewStatusEl.classList.toggle("hidden", nextView !== "status");
  viewSettingsEl.classList.toggle("hidden", nextView !== "settings");
  viewServersEl.classList.toggle("hidden", nextView !== "servers");
  viewLogsEl.classList.toggle("hidden", nextView !== "logs");

  sidebar.querySelectorAll("[data-view]").forEach((btn) => {
    btn.classList.toggle("active", btn.getAttribute("data-view") === nextView);
  });

  if (nextView === "logs") {
    startLogsAutoRefresh();
  } else {
    stopLogsAutoRefresh();
  }
}

function isSidebarPinned() {
  return sidebar.classList.contains("pinned");
}

function isSidebarOpen() {
  return sidebar.classList.contains("open");
}

function openSidebar() {
  sidebar.classList.add("open");
  sidebar.setAttribute("aria-hidden", "false");
  overlay.classList.toggle("hidden", isSidebarPinned());
}

function closeSidebar() {
  if (isSidebarPinned()) return;
  sidebar.classList.remove("open");
  sidebar.setAttribute("aria-hidden", "true");
  overlay.classList.add("hidden");
}

function goToStatus() {
  if (isSidebarPinned()) {
    setView("status");
    return;
  }
  setView("status");
  closeSidebar();
}

scanBtn.addEventListener("click", async () => {
  if (currentScanController) return;
  currentScanController = new AbortController();
  setScanningUI(true);
  try {
    await fetchJSON("/api/scan", { method: "POST", signal: currentScanController.signal });
  } catch (err) {
    if (err.name === "AbortError" || /aborted|canceled|cancelled/i.test(err.message)) {
      alert("Scan cancelled.");
    } else {
      alert(err.message);
    }
  } finally {
    currentScanController = null;
    setScanningUI(false);
    await refreshStatus();
    await refreshLogs();
  }
});

stopBtn.addEventListener("click", () => {
  if (currentScanController) {
    currentScanController.abort();
  }
});

backFromSettingsBtn.addEventListener("click", goToStatus);
backFromServersBtn.addEventListener("click", goToStatus);
backFromLogsBtn.addEventListener("click", goToStatus);
refreshLogsBtn.addEventListener("click", refreshLogs);
clearLogsBtn.addEventListener("click", async () => {
  try {
    await fetchJSON("/api/logs", { method: "DELETE" });
    await refreshLogs();
  } catch (err) {
    showToast(`Clear logs failed: ${err.message}`);
  }
});
logsLevelSelect.addEventListener("change", refreshLogs);

configForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  const payload = {
    scheduler_enabled: schedulerEnabledInput.checked,
    scan_interval_sec: Number(scanIntervalInput.value),
    global_policy: globalPolicySelect.value,
    discord_webhook_url: discordInput.value.trim(),
    discord_notifications_enabled: discordEnabledInput.checked,
    update_stopped_containers: updateStoppedInput.checked,
  };
  try {
    await fetchJSON("/api/config", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    alert("Saved");
  } catch (err) {
    alert(err.message);
  }
});

sidebarSchedulerInput.addEventListener("change", async () => {
  try {
    const cfg = await fetchJSON("/api/config");
    cfg.scheduler_enabled = sidebarSchedulerInput.checked;
    await fetchJSON("/api/config", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(cfg),
    });
    await refreshConfig();
  } catch (err) {
    alert(err.message);
    sidebarSchedulerInput.checked = !sidebarSchedulerInput.checked;
  }
});

remoteForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  const payload = {
    name: remoteNameInput.value.trim(),
    url: remoteUrlInput.value.trim(),
    token: remoteTokenInput.value.trim(),
  };
  try {
    await fetchJSON("/api/servers", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    remoteNameInput.value = "";
    remoteUrlInput.value = "";
    remoteTokenInput.value = "";
    await refreshServers();
    await refreshStatus();
  } catch (err) {
    alert(err.message);
  }
});

generateTokenBtn.addEventListener("click", () => {
  remoteTokenInput.value = generateToken(32);
});

localForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  const payload = {
    name: localNameInput.value.trim(),
    socket: localSocketInput.value.trim() || "/var/run/docker.sock",
  };
  try {
    await fetchJSON("/api/locals", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    localNameInput.value = "";
    localSocketInput.value = "/var/run/docker.sock";
    await refreshServers();
    await refreshStatus();
  } catch (err) {
    alert(err.message);
  }
});

async function init() {
  // Sidebar state
  const pinned = localStorage.getItem("sidebarPinned") === "true";
  sidebarPin.checked = pinned;
  sidebar.classList.toggle("pinned", pinned);
  sidebar.classList.toggle("open", pinned);
  sidebar.setAttribute("aria-hidden", pinned ? "false" : "true");
  overlay.classList.toggle("hidden", true);
  document.body.classList.toggle("sidebar-pinned", pinned);

  menuBtn.addEventListener("click", () => {
    if (isSidebarPinned()) return;
    if (isSidebarOpen()) {
      closeSidebar();
    } else {
      openSidebar();
    }
  });
  sidebarCloseBtn.addEventListener("click", closeSidebar);
  overlay.addEventListener("click", closeSidebar);
  sidebarPin.addEventListener("change", () => {
    const isPinned = sidebarPin.checked;
    localStorage.setItem("sidebarPinned", String(isPinned));
    sidebar.classList.toggle("pinned", isPinned);
    document.body.classList.toggle("sidebar-pinned", isPinned);
    if (isPinned) {
      openSidebar();
    } else {
      closeSidebar();
    }
  });

  sidebar.querySelectorAll("[data-view]").forEach((btn) => {
    btn.addEventListener("click", () => {
      const nextView = btn.getAttribute("data-view");
      if (nextView) {
        setView(nextView);
      }
      if (!sidebar.classList.contains("pinned") && nextView === "status") {
        closeSidebar();
      }
    });
  });

  document.addEventListener("keydown", (event) => {
    if (event.key !== "Escape") return;
    if (isSidebarPinned()) return;
    if (!isSidebarOpen()) return;
    closeSidebar();
  });

  setView("status");
  setScanningUI(false);
  await refreshConfig();
  await refreshServers();
  await refreshStatus();
  if (localSocketInput && !localSocketInput.value) {
    localSocketInput.value = "/var/run/docker.sock";
  }
  statusHintEl.classList.toggle("hidden", cachedLocals.length > 0);
}

init().catch((err) => {
  statusEl.textContent = `Error: ${err.message}`;
});
