const statusEl = document.getElementById("status");
const scanBtn = document.getElementById("scan-btn");
const scanScopeSelect = document.getElementById("scan-scope");
const schedulerStateEl = document.getElementById("scheduler-state");
const toastEl = document.getElementById("toast");
const statusHintEl = document.getElementById("status-hint");
const configForm = document.getElementById("config-form");
const schedulerEnabledInput = document.getElementById("scheduler-enabled");
const scanIntervalInput = document.getElementById("scan-interval");
const globalPolicySelect = document.getElementById("global-policy");
const discordInput = document.getElementById("discord-url");
const discordEnabledInput = document.getElementById("discord-enabled");
const discordStartupNotifyInput = document.getElementById("discord-startup-notify");
const discordUpdateDetectedInput = document.getElementById("discord-update-detected");
const discordContainerUpdatedInput = document.getElementById("discord-container-updated");
const updateStoppedInput = document.getElementById("update-stopped");
const pruneDanglingInput = document.getElementById("prune-dangling");
const localForm = document.getElementById("local-form");
const localNameInput = document.getElementById("local-name");
const localSocketInput = document.getElementById("local-socket");
const remoteForm = document.getElementById("remote-form");
const remoteNameInput = document.getElementById("remote-name");
const remoteUrlInput = document.getElementById("remote-url");
const remoteTokenInput = document.getElementById("remote-token");
const generateTokenBtn = document.getElementById("generate-token");
const serversListEl = document.getElementById("servers-list");
const sidebar = document.getElementById("sidebar");
const sidebarSearch = document.getElementById("sidebar-search");
const themeToggleBtn = document.getElementById("theme-toggle");
const themeLabel = document.getElementById("theme-label");
const viewStatusEl = document.getElementById("view-status");
const viewSettingsEl = document.getElementById("view-settings");
const viewServersEl = document.getElementById("view-servers");
const viewLogsEl = document.getElementById("view-logs");
const refreshLogsBtn = document.getElementById("refresh-logs");
const clearLogsBtn = document.getElementById("clear-logs");
const logsLevelSelect = document.getElementById("logs-level");
const logsAutoToggle = document.getElementById("logs-auto");
const logsListEl = document.getElementById("logs-list");
const appVersionEl = document.getElementById("app-version");
const testWebhookBtn = document.getElementById("test-webhook");
const saveIntervalBtn = document.getElementById("save-interval");
const detailsModal = document.getElementById("details-modal");
const detailsTitle = document.getElementById("details-title");
const detailsBody = document.getElementById("details-body");
const detailsCloseBtn = document.getElementById("details-close");

let currentScanController = null;
let currentView = "status";
let updateInProgress = false;
let logsRefreshTimer = null;
let currentConfig = null;
let cachedLocals = [];
let cachedRemotes = [];
let cachedServerInfo = {};
let scanPollingTimer = null;
let scanActive = false;
let lastStatusResults = [];
let scanStateOverrides = {};
let selectedScanScope = "all";
let restartingServers = {};

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

function closeDetailsModal() {
  if (!detailsModal) return;
  detailsModal.classList.add("hidden");
  detailsModal.setAttribute("aria-hidden", "true");
  if (detailsBody) {
    detailsBody.innerHTML = "";
  }
}

function buildContainerRow(container, result, canUpdateStopped) {
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

  const isSelfContainer = result.local && container.name === "contiwatch";
  const canUpdate =
    container.update_available &&
    !container.updated &&
    !container.paused &&
    !currentScanController &&
    !updateInProgress &&
    (container.running || canUpdateStopped) &&
    !isSelfContainer;

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
  } else if (container.update_available) {
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

    if (isSelfContainer) {
      updateBtn.textContent = "Self-update disabled";
      updateBtn.disabled = true;
    }

    const updateState = document.createElement("span");
    updateState.className = "tag hidden";

    const infoPanel = document.createElement("div");
    infoPanel.className = "info-panel hidden";
    const currentImageID = container.image_id || "unknown";
    const newImageID = container.new_image_id || "unknown";
    infoPanel.textContent = `Current image: ${currentImageID}\nNew image: ${newImageID}`;

    updateBtn.addEventListener("click", async () => {
      if (updateInProgress || currentScanController) return;
      const serverKey = result.server_name || (result.local ? "local" : "remote");
      updateInProgress = true;
      scanBtn.disabled = true;
      if (scanScopeSelect) {
        scanScopeSelect.disabled = true;
      }
      updateBtn.disabled = true;
      updateState.classList.remove("hidden");
      updateState.textContent = "Updating…";
      let updateResult = null;
      let restartHint = false;
      try {
        const scope = result.local
          ? `local:${result.server_name || "local"}`
          : `remote:${result.server_name || "remote"}`;
        const serverParam = encodeURIComponent(scope);
        updateResult = await fetchJSON(`/api/update/${encodeURIComponent(container.id)}?server=${serverParam}`, { method: "POST" });
        if (updateResult.updated) {
          updateState.textContent = `Updated (${updateResult.previous_state} → ${updateResult.current_state})`;
        } else {
          if (updateResult.message && /agent restarting|self update scheduled/i.test(updateResult.message)) {
            restartHint = true;
            updateState.textContent = "Agent restarting — recheck in 30s";
          } else {
            updateState.textContent = updateResult.message || "Not updated";
          }
        }
      } catch (err) {
        updateState.textContent = `Error: ${err.message}`;
      } finally {
        updateInProgress = false;
        if (restartHint) {
          restartingServers[serverKey] = Date.now() + 30000;
          window.setTimeout(async () => {
            delete restartingServers[serverKey];
            await refreshStatus();
          }, 30000);
        }
        scanBtn.disabled = Boolean(currentScanController);
        if (scanScopeSelect) {
          scanScopeSelect.disabled = Boolean(currentScanController);
        }
        updateBtn.disabled = false;
        setTimeout(() => updateState.classList.add("hidden"), 8000);
        await refreshStatus();
        await refreshLogs();
        if (updateResult) {
          if (updateResult.updated) {
            showToast(`Updated ${updateResult.name} (${updateResult.previous_state} → ${updateResult.current_state})`);
          } else if (restartHint) {
            showToast("Agent restarting — recheck in 30s.");
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
}

function buildDetailsContent(result, canUpdateStopped) {
  const wrapper = document.createElement("div");
  const checkedAt = result.checked_at ? new Date(result.checked_at) : null;
  const hasCheckedAt = checkedAt && !Number.isNaN(checkedAt.getTime()) && checkedAt.getTime() > 0;
  const isOffline = Boolean(result.error);
  const isRestarting = Boolean(restartingServers[result.server_name] && restartingServers[result.server_name] > Date.now());
  const scanState = String(result.scan_state || "").toLowerCase();
  const cancelError = result.error && /cancelled|canceled|context canceled/i.test(result.error);
  const isScanningActive = scanState === "scanning";
  const isCancelled = scanState === "cancelled" || cancelError;

  if (result.error) {
    const errorTag = document.createElement("p");
    errorTag.className = "hint";
    errorTag.textContent = `Error: ${result.error}`;
    wrapper.appendChild(errorTag);
  }

  if (!hasCheckedAt) {
    const empty = document.createElement("p");
    if (isRestarting) {
      empty.textContent = "Agent restarting...";
    } else if (isCancelled) {
      empty.textContent = "Scan cancelled.";
    } else {
      empty.textContent = isScanningActive ? "Scanning in progress..." : "Scan has not been run yet.";
    }
    wrapper.appendChild(empty);
    return wrapper;
  }

  if (!result.containers || result.containers.length === 0) {
    const empty = document.createElement("p");
    empty.textContent = "No containers found.";
    wrapper.appendChild(empty);
    return wrapper;
  }

  const updateRequired = result.containers.filter((container) => container.update_available);
  const upToDate = result.containers.filter((container) => !container.update_available);

  const grid = document.createElement("div");
  grid.className = "results-grid";

  const leftCol = document.createElement("div");
  leftCol.className = "results-col";
  const leftTitle = document.createElement("h4");
  leftTitle.textContent = "Updates available";
  leftCol.appendChild(leftTitle);

  const rightCol = document.createElement("div");
  rightCol.className = "results-col";
  const rightTitle = document.createElement("h4");
  rightTitle.textContent = "Scanned containers";
  rightCol.appendChild(rightTitle);

  if (updateRequired.length === 0) {
    const empty = document.createElement("p");
    empty.textContent = "No updates required.";
    leftCol.appendChild(empty);
  } else {
    updateRequired.forEach((container) => leftCol.appendChild(buildContainerRow(container, result, canUpdateStopped)));
  }

  if (upToDate.length === 0) {
    const empty = document.createElement("p");
    empty.textContent = "No containers.";
    rightCol.appendChild(empty);
  } else {
    upToDate.forEach((container) => rightCol.appendChild(buildContainerRow(container, result, canUpdateStopped)));
  }

  grid.appendChild(leftCol);
  grid.appendChild(rightCol);
  wrapper.appendChild(grid);
  return wrapper;
}

function openDetailsModal(result, canUpdateStopped) {
  if (!detailsModal || !detailsBody) return;
  detailsBody.innerHTML = "";
  detailsBody.appendChild(buildDetailsContent(result, canUpdateStopped));
  if (detailsTitle) {
    detailsTitle.textContent = `Server details: ${result.server_name || "server"}`;
  }
  detailsModal.classList.remove("hidden");
  detailsModal.setAttribute("aria-hidden", "false");
  if (detailsCloseBtn) {
    detailsCloseBtn.focus();
  }
}

function renderStatus(results) {
  statusEl.innerHTML = "";
  if (!results || results.length === 0) {
    statusEl.textContent = "No scan data yet.";
    return;
  }

  const canUpdateStopped = Boolean(currentConfig && currentConfig.update_stopped_containers);
  statusHintEl.classList.add("hidden");

  const visibleResults = selectedScanScope && selectedScanScope !== "all"
    ? results.filter((result) => {
      const prefix = result.local ? "local:" : "remote:";
      return `${prefix}${result.server_name}` === selectedScanScope;
    })
    : results;

  if (!visibleResults.length) {
    statusEl.textContent = "No matching servers for selected target.";
    return;
  }

  visibleResults.forEach((result) => {
    const card = document.createElement("div");
    card.className = "server-card";
    const searchTokens = [
      result.server_name,
      result.server_url,
      ...(result.containers || []).map((container) => `${container.name} ${container.image}`),
    ].filter(Boolean);
    card.dataset.search = searchTokens.join(" ");

    const checkedAt = result.checked_at ? new Date(result.checked_at) : null;
    const hasCheckedAt = checkedAt && !Number.isNaN(checkedAt.getTime()) && checkedAt.getTime() > 0;
    const isOffline = Boolean(result.error);
    const key = `${result.local ? "local" : "remote"}:${result.server_name || "unknown"}`;
    const backendState = String(result.scan_state || "").toLowerCase();
    const overrideState = scanStateOverrides[key];
    const scanState = backendState && backendState !== "idle"
      ? backendState
      : (overrideState || backendState);
    const cancelError = result.error && /cancelled|canceled|context canceled/i.test(result.error);
    const isPending = scanState === "pending";
    const isScanningActive = scanState === "scanning";
    const isCancelled = scanState === "cancelled" || cancelError;
    const hasScanErrorState = scanState === "error";
    const isRestarting = Boolean(restartingServers[result.server_name] && restartingServers[result.server_name] > Date.now());
    const isOfflineForStatus = isOffline && !isCancelled;
    const statusLabel = isRestarting ? "restarting" : (isOfflineForStatus ? "offline" : "online");
    let scanLabelText = "";
    let scanLabelClass = "";
    if (!isRestarting) {
      if (isScanningActive) {
        scanLabelText = "scanning";
        scanLabelClass = "scanning";
      } else if (isPending) {
        scanLabelText = "pending scan";
        scanLabelClass = "pending";
      } else if (isCancelled) {
        scanLabelText = "cancelled";
        scanLabelClass = "cancelled";
      } else if (hasScanErrorState || (isOffline && !isCancelled)) {
        scanLabelText = "scan error";
        scanLabelClass = "error";
      }
    }

    const head = document.createElement("div");
    head.className = "status-card-head";

    const headLeft = document.createElement("div");
    headLeft.className = "status-card-name";

    const nameWrap = document.createElement("div");
    nameWrap.className = "status-card-title";
    const nameIcon = document.createElementNS("http://www.w3.org/2000/svg", "svg");
    nameIcon.setAttribute("class", `status-card-icon status-icon status-icon-${statusLabel} icon icon-tabler icons-tabler-outline icon-tabler-server-bolt`);
    nameIcon.setAttribute("width", "24");
    nameIcon.setAttribute("height", "24");
    nameIcon.setAttribute("viewBox", "0 0 24 24");
    nameIcon.setAttribute("fill", "none");
    nameIcon.setAttribute("stroke", "currentColor");
    nameIcon.setAttribute("stroke-width", "2");
    nameIcon.setAttribute("stroke-linecap", "round");
    nameIcon.setAttribute("stroke-linejoin", "round");
    nameIcon.setAttribute("aria-hidden", "true");
    nameIcon.setAttribute("title", `Status: ${statusLabel}`);
    nameIcon.innerHTML = '<path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M3 7a3 3 0 0 1 3 -3h12a3 3 0 0 1 3 3v2a3 3 0 0 1 -3 3h-12a3 3 0 0 1 -3 -3" /><path d="M15 20h-9a3 3 0 0 1 -3 -3v-2a3 3 0 0 1 3 -3h12" /><path d="M7 8v.01" /><path d="M7 16v.01" /><path d="M20 15l-2 3h3l-2 3" />';
    const name = document.createElement("strong");
    name.textContent = result.server_name || "unknown";
    nameWrap.appendChild(nameIcon);
    nameWrap.appendChild(name);
    headLeft.appendChild(nameWrap);

    const typeTag = document.createElement("span");
    typeTag.className = "tag";
    typeTag.textContent = result.local ? "LOCAL" : "REMOTE";
    headLeft.appendChild(typeTag);

    // Status indicator moved to the server icon.
    if (scanLabelText) {
      const scanBadge = document.createElement("span");
      scanBadge.className = `status-badge status-${scanLabelClass}`;
      scanBadge.textContent = scanLabelText;
      headLeft.appendChild(scanBadge);
    }

    const summary = document.createElement("div");
    summary.className = "server-summary";
    if (isCancelled) {
      summary.textContent = "Updates: cancelled.";
    } else if (!hasCheckedAt) {
      summary.textContent = "Updates: pending.";
    } else if (!result.containers || result.containers.length === 0) {
      summary.textContent = "Updates: none.";
    } else {
      const total = result.containers.length;
      const updates = result.containers.filter((container) => container.update_available).length;
      summary.textContent = updates === 0
        ? "Updates: none."
        : `Updates: ${updates} / ${total} scanned.`;
    }

    const meta = document.createElement("div");
    meta.className = "status-card-meta";
    if (isCancelled) {
      meta.textContent = "Last scan: cancelled.";
    } else {
      meta.textContent = hasCheckedAt ? `Last scan: ${checkedAt.toLocaleString()}` : "Last scan: not run yet.";
    }

    headLeft.appendChild(summary);
    headLeft.appendChild(meta);

    const headRight = document.createElement("div");
    headRight.className = "status-card-actions";
    const detailsBtn = document.createElement("button");
    detailsBtn.type = "button";
    detailsBtn.className = "secondary btn-small";
    detailsBtn.textContent = "Details";
    detailsBtn.addEventListener("click", () => {
      openDetailsModal(result, canUpdateStopped);
    });
    headRight.appendChild(detailsBtn);

    head.appendChild(headLeft);
    head.appendChild(headRight);
    card.appendChild(head);

    if (result.error && !isCancelled) {
      const errorTag = document.createElement("p");
      errorTag.className = "hint";
      errorTag.textContent = `Error: ${result.error}`;
      card.appendChild(errorTag);
    }

    if (!hasCheckedAt && !isOffline && isRestarting) {
      const stateLine = document.createElement("div");
      stateLine.className = "status-card-meta";
      stateLine.textContent = "Agent restarting...";
      card.appendChild(stateLine);
    }

    statusEl.appendChild(card);
  });

  applySidebarFilter();
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
    const infoKey = `${item.type}:${item.server.name}`;
    const info = cachedServerInfo[infoKey] || {};
    const li = document.createElement("li");
    li.className = "server-item";

    const infoBlock = document.createElement("div");
    infoBlock.className = "server-item-info";

    const nameRow = document.createElement("div");
    nameRow.className = "server-item-title";
    nameRow.textContent = item.server.name;

    const typeTag = document.createElement("span");
    typeTag.className = "tag";
    typeTag.textContent = item.type.toUpperCase();
    nameRow.appendChild(typeTag);
    infoBlock.appendChild(nameRow);

    const addressRow = document.createElement("div");
    addressRow.className = "server-item-meta";
    const addressValue = item.type === "local" ? item.server.socket : item.server.url;
    addressRow.textContent = addressValue || "unknown";
    infoBlock.appendChild(addressRow);

    const versionRow = document.createElement("div");
    versionRow.className = "server-item-meta";
    const versionValue = info.version ? formatVersion(info.version) : "unknown";
    versionRow.textContent = `Release: ${versionValue}`;
    infoBlock.appendChild(versionRow);

    li.appendChild(infoBlock);

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

    const actions = document.createElement("div");
    actions.className = "server-item-actions";
    actions.appendChild(editBtn);
    actions.appendChild(removeBtn);
    li.appendChild(actions);
    serversListEl.appendChild(li);
  });

  applySidebarFilter();
}

function formatVersion(raw) {
  const version = String(raw || "").trim();
  if (!version) return "unknown";
  if (version.startsWith("v")) return version;
  return `v${version}`;
}

function normalizeQuery(value) {
  return String(value || "").trim().toLowerCase();
}

function applySidebarFilter(queryOverride) {
  if (!sidebarSearch) return;
  const query = normalizeQuery(queryOverride ?? sidebarSearch.value);
  const cards = Array.from(statusEl ? statusEl.querySelectorAll(".server-card") : []);
  const listItems = Array.from(serversListEl ? serversListEl.querySelectorAll(".server-item") : []);
  const targets = cards.concat(listItems);
  targets.forEach((el) => {
    const source = el.dataset && el.dataset.search ? el.dataset.search : el.textContent;
    const haystack = normalizeQuery(source);
    const match = !query || haystack.includes(query);
    el.classList.toggle("filtered-out", !match);
  });
}

const themeModes = ["light", "dark"];
const themeStorageKey = "contiwatch_theme";

function applyTheme(mode) {
  if (!themeModes.includes(mode)) return;
  document.documentElement.setAttribute("data-theme", mode);
  localStorage.setItem(themeStorageKey, mode);
  if (themeToggleBtn) {
    const nextMode = mode === "dark" ? "light" : "dark";
    const label = nextMode[0].toUpperCase() + nextMode.slice(1);
    themeToggleBtn.dataset.mode = mode;
    themeToggleBtn.setAttribute("aria-label", `Switch to ${label} theme`);
    themeToggleBtn.setAttribute("title", `Switch to ${label} theme`);
  }
  if (themeLabel) {
    themeLabel.textContent = mode[0].toUpperCase() + mode.slice(1);
  }
}

function initThemeToggle() {
  const saved = localStorage.getItem(themeStorageKey);
  applyTheme(saved && themeModes.includes(saved) ? saved : "light");
  if (!themeToggleBtn) return;
  themeToggleBtn.addEventListener("click", () => {
    const current = localStorage.getItem(themeStorageKey) || "light";
    const next = current === "dark" ? "light" : "dark";
    applyTheme(next);
  });
}

function updateScanScopeOptions() {
  if (!scanScopeSelect) return;
  const currentValue = scanScopeSelect.value || selectedScanScope || "all";
  const options = [
    { value: "all", label: "All servers" },
  ];

  cachedLocals.forEach((local) => {
    options.push({ value: `local:${local.name}`, label: local.name });
  });

  cachedRemotes.forEach((remote) => {
    options.push({ value: `remote:${remote.name}`, label: remote.name });
  });

  scanScopeSelect.innerHTML = "";
  options.forEach((option) => {
    const el = document.createElement("option");
    el.value = option.value;
    el.textContent = option.label;
    scanScopeSelect.appendChild(el);
  });

  if (options.some((option) => option.value === currentValue)) {
    scanScopeSelect.value = currentValue;
  }
  selectedScanScope = scanScopeSelect.value;
}

if (scanScopeSelect) {
  scanScopeSelect.addEventListener("change", () => {
    selectedScanScope = scanScopeSelect.value || "all";
    refreshStatus();
  });
}

async function refreshConfig() {
  const cfg = await fetchJSON("/api/config");
  currentConfig = cfg;
  schedulerEnabledInput.checked = Boolean(cfg.scheduler_enabled);
  scanIntervalInput.value = Math.max(1, Math.round(cfg.scan_interval_sec / 60));
  globalPolicySelect.value = cfg.global_policy;
  discordInput.value = cfg.discord_webhook_url || "";
  discordEnabledInput.checked = cfg.discord_notifications_enabled !== false;
  if (discordStartupNotifyInput) {
    discordStartupNotifyInput.checked = cfg.discord_notify_on_start !== false;
  }
  if (discordUpdateDetectedInput) {
    discordUpdateDetectedInput.checked = cfg.discord_notify_on_update_detected !== false;
  }
  if (discordContainerUpdatedInput) {
    discordContainerUpdatedInput.checked = cfg.discord_notify_on_container_updated !== false;
  }
  updateStoppedInput.checked = Boolean(cfg.update_stopped_containers);
  pruneDanglingInput.checked = Boolean(cfg.prune_dangling_images);
  const enabled = Boolean(cfg.scheduler_enabled);
  if (schedulerStateEl) {
    schedulerStateEl.dataset.enabled = enabled ? "true" : "false";
    const tooltip = enabled ? "Auto Scheduler is ON" : "Auto Scheduler is OFF";
    schedulerStateEl.setAttribute("aria-label", tooltip);
    schedulerStateEl.setAttribute("title", tooltip);
  }
}

function buildConfigPayload(options = {}) {
  const intervalMinutes = Number(scanIntervalInput.value);
  const currentInterval = currentConfig ? Number(currentConfig.scan_interval_sec) : 0;
  const nextInterval = Number.isFinite(intervalMinutes) && intervalMinutes > 0
    ? intervalMinutes * 60
    : (currentInterval > 0 ? currentInterval : 0);
  const useWebhookInput = Boolean(options.useWebhookInput);
  const webhookValue = useWebhookInput
    ? discordInput.value.trim()
    : (currentConfig ? currentConfig.discord_webhook_url || "" : discordInput.value.trim());
  return {
    scheduler_enabled: schedulerEnabledInput.checked,
    scan_interval_sec: nextInterval,
    global_policy: globalPolicySelect.value,
    discord_webhook_url: webhookValue,
    discord_notifications_enabled: discordEnabledInput.checked,
    discord_notify_on_start: discordStartupNotifyInput ? discordStartupNotifyInput.checked : true,
    discord_notify_on_update_detected: discordUpdateDetectedInput ? discordUpdateDetectedInput.checked : true,
    discord_notify_on_container_updated: discordContainerUpdatedInput ? discordContainerUpdatedInput.checked : true,
    update_stopped_containers: updateStoppedInput.checked,
    prune_dangling_images: pruneDanglingInput.checked,
  };
}

async function saveConfig(options = {}) {
  const payload = buildConfigPayload(options);
  const updated = await fetchJSON("/api/config", {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  currentConfig = updated;
  return updated;
}

async function refreshServers() {
  const [locals, remotes, info] = await Promise.all([
    fetchJSON("/api/locals"),
    fetchJSON("/api/servers"),
    fetchJSON("/api/servers/info").catch(() => []),
  ]);
  cachedLocals = locals || [];
  cachedRemotes = remotes || [];
  cachedServerInfo = {};
  if (Array.isArray(info)) {
    info.forEach((item) => {
      if (!item || !item.name || !item.type) return;
      const key = `${item.type}:${item.name}`;
      cachedServerInfo[key] = item;
    });
  }
  renderServers(cachedLocals, cachedRemotes);
  updateScanScopeOptions();
  statusHintEl.classList.add("hidden");
}

async function refreshStatus() {
  const results = await fetchJSON("/api/aggregate");
  if (Array.isArray(results)) {
    results.sort((a, b) => {
      if (a.local === b.local) return 0;
      return a.local ? -1 : 1;
    });
  }
  lastStatusResults = Array.isArray(results) ? results : [];
  if (lastStatusResults.length > 0) {
    const nextOverrides = { ...scanStateOverrides };
    lastStatusResults.forEach((result) => {
      const key = `${result.local ? "local" : "remote"}:${result.server_name || "unknown"}`;
      const state = String(result.scan_state || "").toLowerCase();
      if (state && state !== "idle") {
        delete nextOverrides[key];
      }
    });
    scanStateOverrides = nextOverrides;
  }
  renderStatus(results);
  updateScanPolling(results);
  return results;
}

async function refreshVersion() {
  if (!appVersionEl) return;
  try {
    const payload = await fetchJSON("/api/version");
    const raw = (payload && payload.version) ? String(payload.version) : "dev";
    const normalized = raw.startsWith("v") ? raw : `v${raw}`;
    const display = raw.startsWith("dev") ? `v${raw}` : normalized;
    appVersionEl.textContent = `Release: ${display}`;
  } catch (err) {
    appVersionEl.textContent = "Release: vdev";
  }
}

async function refreshLogs() {
  if (!logsListEl) return;
  if (viewLogsEl.classList.contains("hidden")) return;
  try {
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
      return;
    }
    filtered.forEach((entry) => {
      const row = document.createElement("div");
      row.className = "log-row";
      const ts = new Date(entry.timestamp).toLocaleString();
      const level = String(entry.level || "info").toLowerCase();
      const levelClass = `log-level log-level-${level}`;
      row.innerHTML = `
        <span class="log-ts">${ts}</span>
        <span class="${levelClass}">${level}</span>
        <span class="log-msg">${entry.message}</span>
      `;
    logsListEl.appendChild(row);
    });
  } catch (err) {
    logsListEl.innerHTML = "";
    const empty = document.createElement("p");
    empty.textContent = "Logs unavailable. Add a server in Servers to enable operational logs.";
    logsListEl.appendChild(empty);
  }
}

function isLogsAutoEnabled() {
  if (!logsAutoToggle) return false;
  return logsAutoToggle.dataset.enabled === "true";
}

function setLogsAutoEnabled(enabled) {
  if (!logsAutoToggle) return;
  const state = enabled ? "Disable" : "Enable";
  logsAutoToggle.dataset.enabled = enabled ? "true" : "false";
  logsAutoToggle.setAttribute("aria-pressed", enabled ? "true" : "false");
  logsAutoToggle.setAttribute("aria-label", `${state} logs auto refresh`);
  logsAutoToggle.setAttribute("title", `${state} logs auto refresh`);
}

function startLogsAutoRefresh() {
  stopLogsAutoRefresh();
  refreshLogs();
  if (logsAutoToggle && !isLogsAutoEnabled()) {
    return;
  }
  logsRefreshTimer = window.setInterval(refreshLogs, 5000);
}

function stopLogsAutoRefresh() {
  if (logsRefreshTimer) {
    window.clearInterval(logsRefreshTimer);
    logsRefreshTimer = null;
  }
}

async function logClientEvent(level, message) {
  try {
    await fetchJSON("/api/logs", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ level, message }),
    });
  } catch {
    // Ignore logging failures from UI events.
  }
}

function flashButton(btn, text, className, durationMs = 2000) {
  if (!btn) return;
  const prevText = btn.textContent;
  const prevClass = btn.className;
  btn.textContent = text;
  if (className) {
    btn.classList.add(className);
  }
  btn.disabled = true;
  window.setTimeout(() => {
    btn.textContent = prevText;
    btn.className = prevClass;
    btn.disabled = false;
  }, durationMs);
}


function setScanningUI(isScanning) {
  const isActive = isScanning || scanActive;
  scanBtn.disabled = updateInProgress;
  if (scanScopeSelect) {
    scanScopeSelect.disabled = isActive || updateInProgress;
  }
  scanBtn.classList.toggle("secondary", !isActive);
  scanBtn.classList.toggle("btn-warning", isActive);
  scanBtn.textContent = isActive ? "Stop scan" : "Run scan";
}

function updateScanPolling(results) {
  const active = Array.isArray(results)
    ? results.some((result) => ["pending", "scanning"].includes(String(result.scan_state || "")))
    : false;
  scanActive = active;
  setScanningUI(false);
  if (active && !scanPollingTimer) {
    scanPollingTimer = window.setInterval(async () => {
      await refreshStatus();
    }, 2000);
  } else if (!active && scanPollingTimer) {
    window.clearInterval(scanPollingTimer);
    scanPollingTimer = null;
  }
}

function applyOptimisticScanState(scope) {
  const nextOverrides = {};
  const base = Array.isArray(lastStatusResults) && lastStatusResults.length > 0
    ? lastStatusResults
    : cachedLocals.map((server) => ({
      server_name: server.name,
      local: true,
    })).concat(cachedRemotes.map((server) => ({
      server_name: server.name,
      local: false,
    })));
  const items = Array.isArray(base) ? base : [];
  if (scope && scope !== "all") {
    const key = scope;
    nextOverrides[key] = "scanning";
  } else if (items.length > 0) {
    items.forEach((item, index) => {
      const key = `${item.local ? "local" : "remote"}:${item.server_name || "unknown"}`;
      nextOverrides[key] = index === 0 ? "scanning" : "pending";
    });
  }
  scanStateOverrides = nextOverrides;
  if (lastStatusResults.length > 0) {
    renderStatus(lastStatusResults);
  }
}

function startScanPolling() {
  scanActive = true;
  setScanningUI(false);
  if (!scanPollingTimer) {
    scanPollingTimer = window.setInterval(async () => {
      await refreshStatus();
    }, 2000);
  }
}

function showToast(message, timeoutMs = 8000) {
  if (!toastEl) return;
  if (!message) return;
  toastEl.textContent = message;
  toastEl.classList.remove("hidden");
  window.clearTimeout(showToast._t);
  showToast._t = window.setTimeout(() => toastEl.classList.add("hidden"), timeoutMs);
}

function refreshViewData(view) {
  const tasks = [];
  tasks.push(refreshConfig().catch(() => {}));
  if (view === "status") {
    tasks.push(refreshStatus().catch(() => {}));
  }
  if (view === "servers") {
    tasks.push(refreshServers().catch(() => {}));
  }
  if (view === "logs") {
    tasks.push(refreshLogs().catch(() => {}));
  }
  void Promise.all(tasks);
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

  refreshViewData(nextView);
}

function goToStatus() {
  setView("status");
}

scanBtn.addEventListener("click", async () => {
  if (currentScanController || scanActive) {
    try {
      await fetchJSON("/api/scan/stop", { method: "POST" });
    } catch (err) {
      showToast(err.message || "Failed to stop scan.");
    }
    if (currentScanController) {
      currentScanController.abort();
      currentScanController = null;
    }
    scanStateOverrides = {};
    await refreshStatus();
    return;
  }
  currentScanController = new AbortController();
  selectedScanScope = scanScopeSelect ? scanScopeSelect.value : "all";
  setScanningUI(true);
  startScanPolling();
  applyOptimisticScanState(selectedScanScope);
  refreshStatus().catch(() => {});
  try {
    const scope = scanScopeSelect ? scanScopeSelect.value : "all";
    const url = scope && scope !== "all"
      ? `/api/scan?server=${encodeURIComponent(scope)}`
      : "/api/scan?server=all";
    await fetchJSON(url, { method: "POST", signal: currentScanController.signal });
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
if (logsAutoToggle) {
  logsAutoToggle.addEventListener("click", async () => {
    const nextEnabled = !isLogsAutoEnabled();
    setLogsAutoEnabled(nextEnabled);
    if (viewLogsEl.classList.contains("hidden")) return;
    if (nextEnabled) {
      startLogsAutoRefresh();
      await logClientEvent("info", "logs auto refresh enabled");
    } else {
      stopLogsAutoRefresh();
      await logClientEvent("info", "logs auto refresh disabled");
    }
    await refreshLogs();
  });
}

function attachImmediateSave(input) {
  if (!input) return;
  input.addEventListener("change", async () => {
    try {
      await saveConfig({ useWebhookInput: false });
      currentConfig = await fetchJSON("/api/config");
    } catch (err) {
      showToast(err.message || "Failed to save setting.");
    }
  });
}

attachImmediateSave(schedulerEnabledInput);
attachImmediateSave(updateStoppedInput);
attachImmediateSave(pruneDanglingInput);
attachImmediateSave(discordEnabledInput);
attachImmediateSave(discordStartupNotifyInput);
attachImmediateSave(discordUpdateDetectedInput);
attachImmediateSave(discordContainerUpdatedInput);
globalPolicySelect.addEventListener("change", async () => {
  try {
    await saveConfig({ useWebhookInput: false });
    currentConfig = await fetchJSON("/api/config");
  } catch (err) {
    showToast(err.message || "Failed to save policy.");
  }
});

if (testWebhookBtn) {
  testWebhookBtn.addEventListener("click", async () => {
    const url = discordInput.value.trim();
    if (!url) {
      showToast("Discord webhook URL is required.");
      return;
    }
    try {
      await fetchJSON("/api/notifications/test", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ webhook_url: url }),
      });
      await saveConfig({ useWebhookInput: true });
      await refreshConfig();
      flashButton(testWebhookBtn, "Test success", "btn-success");
    } catch (err) {
      flashButton(testWebhookBtn, "Test failed", "btn-danger");
    }
  });
}

configForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  try {
    await saveConfig({ useWebhookInput: false });
    await refreshConfig();
    flashButton(saveIntervalBtn, "Saved", "btn-success");
  } catch (err) {
    alert(err.message);
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
  sidebar.setAttribute("aria-hidden", "false");

  if (sidebarSearch) {
    sidebarSearch.addEventListener("input", (event) => {
      applySidebarFilter(event.target.value);
    });
  }

  sidebar.querySelectorAll("[data-view]").forEach((btn) => {
    btn.addEventListener("click", () => {
      const nextView = btn.getAttribute("data-view");
      if (nextView) {
        setView(nextView);
      }
    });
  });

  if (detailsCloseBtn) {
    detailsCloseBtn.addEventListener("click", closeDetailsModal);
  }
  if (detailsModal) {
    detailsModal.addEventListener("click", (event) => {
      if (event.target && event.target.dataset && event.target.dataset.close) {
        closeDetailsModal();
      }
    });
  }
  document.addEventListener("keydown", (event) => {
    if (event.key === "Escape" && detailsModal && !detailsModal.classList.contains("hidden")) {
      closeDetailsModal();
    }
  });

  initThemeToggle();
  setView("status");
  setScanningUI(false);
  await refreshConfig();
  await refreshServers();
  await refreshStatus();
  await refreshVersion();
  if (localSocketInput && !localSocketInput.value) {
    localSocketInput.value = "/var/run/docker.sock";
  }
  statusHintEl.classList.toggle("hidden", cachedLocals.length > 0);
}

init().catch((err) => {
  statusEl.textContent = `Error: ${err.message}`;
});
