const statusEl = document.getElementById("status");
const scanBtn = document.getElementById("scan-btn");
const statusSelectiveToggleBtn = document.getElementById("status-selective-toggle");
const refreshStatusBtn = document.getElementById("refresh-status");
const schedulerStateEl = document.getElementById("scheduler-state");
const schedulerNextRunEl = document.getElementById("scheduler-next-run");
const tooltipEl = document.createElement("div");
tooltipEl.id = "app-tooltip";
document.body.appendChild(tooltipEl);
const toastEl = document.getElementById("toast");
const statusHintEl = document.getElementById("status-hint");
const configForm = document.getElementById("config-form");
const schedulerEnabledInput = document.getElementById("scheduler-enabled");
const scanIntervalInput = document.getElementById("scan-interval");
const globalPolicySelect = document.getElementById("global-policy");
const globalPolicyInfoBtn = document.getElementById("global-policy-info");
const discordInput = document.getElementById("discord-url");
const discordEnabledInput = document.getElementById("discord-enabled");
const discordStartupNotifyInput = document.getElementById("discord-startup-notify");
const discordUpdateDetectedInput = document.getElementById("discord-update-detected");
const discordContainerUpdatedInput = document.getElementById("discord-container-updated");
const updateStoppedInput = document.getElementById("update-stopped");
const pruneDanglingInput = document.getElementById("prune-dangling");
const addServerBtn = document.getElementById("add-server");
const refreshServersBtn = document.getElementById("refresh-servers");
const serversTableWrap = document.getElementById("servers-table-wrap");
const serversTableBody = document.getElementById("servers-table-body");
const serversCardsEl = document.getElementById("servers-cards");
const serversViewToggleBtn = document.getElementById("servers-view-toggle");
const serversFilterSelect = document.getElementById("servers-filter");
const serversFilterResetBtn = document.getElementById("servers-filter-reset");
const localModalForm = document.getElementById("local-modal-form");
const localModalNameInput = document.getElementById("local-modal-name");
const localModalSocketInput = document.getElementById("local-modal-socket");
const localModalSave = document.getElementById("local-modal-save");
const localModalCancel = document.getElementById("local-modal-cancel");
const remoteModal = document.getElementById("remote-modal");
const remoteModalTitle = document.getElementById("remote-modal-title");
const remoteModalClose = document.getElementById("remote-modal-close");
const serverTabRemote = document.getElementById("server-tab-remote");
const serverTabLocal = document.getElementById("server-tab-local");
const serverPanelRemote = document.getElementById("server-panel-remote");
const serverPanelLocal = document.getElementById("server-panel-local");
const remoteModalNameInput = document.getElementById("remote-modal-name");
const remoteModalHostInput = document.getElementById("remote-modal-host");
const remoteModalPortInput = document.getElementById("remote-modal-port");
const remoteModalTokenInput = document.getElementById("remote-modal-token");
const remoteModalTokenCopy = document.getElementById("remote-modal-token-copy");
const remoteModalCompose = document.getElementById("remote-modal-compose");
const remoteModalComposeCopy = document.getElementById("remote-modal-compose-copy");
const remoteModalSave = document.getElementById("remote-modal-save");
const remoteModalCancel = document.getElementById("remote-modal-cancel");
const sidebar = document.getElementById("sidebar");
const sidebarSearch = document.getElementById("sidebar-search");
const themeToggleBtn = document.getElementById("theme-toggle");
const themeLabel = document.getElementById("theme-label");
const topbarEl = document.querySelector(".topbar");
const topbarStatusEl = document.getElementById("topbar-status");
const topbarSettingsEl = document.getElementById("topbar-settings");
const topbarServersEl = document.getElementById("topbar-servers");
const topbarLogsEl = document.getElementById("topbar-logs");
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
const policyModal = document.getElementById("policy-modal");
const policyCloseBtn = document.getElementById("policy-close");

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
let scanRequestActive = false;
let lastStatusResults = [];
let scanStateOverrides = {};
let currentScanStartedAtMs = null;
let selectedScanScope = "all";
let restartingServers = {};
let currentDetailsServerKey = null;
let selectedScanServers = new Set();
let serversRemoveConfirming = new Set();
let serversFilterMode = "all";
let serversSearchQuery = "";
let editingLocalServer = null;
let editingRemoteServer = null;
const schedulerAnchorKey = "contiwatch_scheduler_anchor";
const serversViewStorageKey = "contiwatch_servers_view";
let serversViewMode = "table";
let lastSchedulerEnabled = null;
let lastSchedulerIntervalSec = null;
let serverStream = null;
let checkingStartMsByServerKey = {};
let checkingDebounceTimersByServerKey = {};
let pendingSelfUpdates = {};
let pendingSelfUpdateTimers = {};

function pendingSelfUpdateKey(serverKey, containerName) {
  return `${serverKey}::${containerName}`;
}

function markPendingSelfUpdate(serverKey, containerName, ttlMs = 60000) {
  if (!serverKey || !containerName) return;
  const key = pendingSelfUpdateKey(serverKey, containerName);
  const expiresAt = Date.now() + ttlMs;
  pendingSelfUpdates[key] = expiresAt;
  if (pendingSelfUpdateTimers[key]) {
    window.clearTimeout(pendingSelfUpdateTimers[key]);
  }
  pendingSelfUpdateTimers[key] = window.setTimeout(() => {
    delete pendingSelfUpdates[key];
    delete pendingSelfUpdateTimers[key];
  }, ttlMs + 1500);
}

function clearPendingSelfUpdate(serverKey, containerName) {
  const key = pendingSelfUpdateKey(serverKey, containerName);
  delete pendingSelfUpdates[key];
  if (pendingSelfUpdateTimers[key]) {
    window.clearTimeout(pendingSelfUpdateTimers[key]);
    delete pendingSelfUpdateTimers[key];
  }
}

function isSelfUpdateResolved(serverKey, containerName) {
  if (!serverKey || !containerName) return false;
  const result = lastStatusResults.find((entry) => serverScopeKey(entry) === serverKey);
  if (!result || !Array.isArray(result.containers)) return false;
  const container = result.containers.find((item) => item && item.name === containerName);
  if (!container) return false;
  return Boolean(container.updated) || !Boolean(container.update_available);
}

function schedulePostSelfUpdateRefresh(serverKey, containerName) {
  markPendingSelfUpdate(serverKey, containerName);
  restartingServers[serverKey] = Date.now() + 60000;

  const startedAt = Date.now();
  const deadlineMs = 60000;
  let attempt = 0;

  const tick = async () => {
    attempt++;
    if (isSelfUpdateResolved(serverKey, containerName)) {
      clearPendingSelfUpdate(serverKey, containerName);
      delete restartingServers[serverKey];
      showToast("Agent back online — status refreshed.");
      return;
    }
    if (Date.now() - startedAt > deadlineMs) {
      clearPendingSelfUpdate(serverKey, containerName);
      delete restartingServers[serverKey];
      showToast("Agent update pending — use Refresh status in a moment.");
      return;
    }
    try {
      await triggerServersRefresh();
    } catch {
      // ignore
    }
    try {
      await triggerStatusRefresh();
    } catch {
      // ignore
    }
    try {
      await refreshStatus();
    } catch {
      // ignore
    }

    const nextDelay = attempt <= 4 ? 2500 : 5000;
    window.setTimeout(tick, nextDelay);
  };

  window.setTimeout(tick, 1500);
}

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

function isContiwatchImage(value) {
  return /contiwatch/i.test(String(value || ""));
}

function serverScopeKey(result) {
  if (!result) return "unknown:unknown";
  const scope = result.local ? "local" : "remote";
  const name = result.server_name || "unknown";
  return `${scope}:${name}`;
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

function normalizeRemoteUrl(host, port) {
  const trimmedHost = String(host || "").trim();
  const trimmedPort = String(port || "").trim();
  if (!trimmedHost) return "";
  try {
    const base = /^https?:\/\//i.test(trimmedHost) ? trimmedHost : `http://${trimmedHost}`;
    const parsed = new URL(base);
    parsed.port = trimmedPort || parsed.port || "8080";
    return parsed.toString().replace(/\/$/, "");
  } catch {
    const cleanedHost = trimmedHost.replace(/\/+$/, "");
    return `http://${cleanedHost}:${trimmedPort || "8080"}`;
  }
}

function parseRemoteUrl(url) {
  const raw = String(url || "").trim();
  if (!raw) {
    return { host: "", port: "8080" };
  }
  try {
    const parsed = new URL(/^https?:\/\//i.test(raw) ? raw : `http://${raw}`);
    return {
      host: parsed.hostname || raw,
      port: parsed.port || "8080",
    };
  } catch {
    return { host: raw, port: "8080" };
  }
}

function buildAgentCompose(token, port) {
  const tokenValue = token || "PUT_32_CHAR_TOKEN_HERE";
  const hostPort = port || "8080";
  return `services:
  contiwatch-agent:
    image: ghcr.io/pbuzdygan/contiwatch:latest
    container_name: contiwatch-agent
    ports:
      - "${hostPort}:8080"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - contiwatch_agent_data:/data
    environment:
      CONTIWATCH_AGENT: "true"
      CONTIWATCH_AGENT_TOKEN: "${tokenValue}"
      CONTIWATCH_ADDR: ":8080"
      CONTIWATCH_CONFIG: "/data/config.json"
      TZ: Europe/Warsaw
    restart: unless-stopped

volumes:
  contiwatch_agent_data:
`;
}

function updateRemoteComposePreview() {
  if (!remoteModalCompose) return;
  const token = remoteModalTokenInput ? remoteModalTokenInput.value.trim() : "";
  const port = remoteModalPortInput ? remoteModalPortInput.value.trim() : "8080";
  remoteModalCompose.textContent = buildAgentCompose(token, port);
}

function setServerModalTab(tab) {
  const next = tab === "local" ? "local" : "remote";
  if (serverPanelLocal) {
    serverPanelLocal.classList.toggle("hidden", next !== "local");
  }
  if (serverPanelRemote) {
    serverPanelRemote.classList.toggle("hidden", next !== "remote");
  }
  if (serverTabLocal) {
    serverTabLocal.classList.toggle("is-active", next === "local");
    serverTabLocal.setAttribute("aria-selected", next === "local" ? "true" : "false");
  }
  if (serverTabRemote) {
    serverTabRemote.classList.toggle("is-active", next === "remote");
    serverTabRemote.setAttribute("aria-selected", next === "remote" ? "true" : "false");
  }
}

function openLocalModal(mode, server) {
  if (!remoteModal) return;
  const isEdit = mode === "edit";
  editingLocalServer = isEdit ? server : null;
  editingRemoteServer = null;
  setServerModalTab("local");
  if (remoteModalTitle) {
    remoteModalTitle.textContent = isEdit ? "Edit local server" : "Add local server";
  }
  if (localModalNameInput) {
    localModalNameInput.value = server ? server.name || "" : "";
  }
  if (localModalSocketInput) {
    localModalSocketInput.value = server ? server.socket || "/var/run/docker.sock" : "/var/run/docker.sock";
  }
  remoteModal.classList.remove("hidden");
  remoteModal.setAttribute("aria-hidden", "false");
  if (localModalNameInput) {
    localModalNameInput.focus();
  }
}

function openRemoteModal(mode, server) {
  if (!remoteModal) return;
  const isEdit = mode === "edit";
  editingRemoteServer = isEdit ? server : null;
  editingLocalServer = null;
  setServerModalTab("remote");
  if (remoteModalTitle) {
    remoteModalTitle.textContent = isEdit ? "Edit remote server" : "Add remote server";
  }
  const parsed = server ? parseRemoteUrl(server.url) : { host: "", port: "8080" };
  if (remoteModalNameInput) {
    remoteModalNameInput.value = server ? server.name || "" : "";
  }
  if (remoteModalHostInput) {
    remoteModalHostInput.value = parsed.host || "";
  }
  if (remoteModalPortInput) {
    remoteModalPortInput.value = parsed.port || "8080";
  }
  if (remoteModalTokenInput) {
    if (server && server.token) {
      remoteModalTokenInput.value = server.token;
    } else {
      remoteModalTokenInput.value = generateToken(32);
    }
  }
  updateRemoteComposePreview();
  remoteModal.classList.remove("hidden");
  remoteModal.setAttribute("aria-hidden", "false");
  if (remoteModalNameInput) {
    remoteModalNameInput.focus();
  }
}

function closeRemoteModal() {
  if (!remoteModal) return;
  remoteModal.classList.add("hidden");
  remoteModal.setAttribute("aria-hidden", "true");
  editingRemoteServer = null;
  editingLocalServer = null;
}

async function copyToClipboard(value, label) {
  if (!value) return;
  try {
    if (navigator.clipboard && navigator.clipboard.writeText) {
      await navigator.clipboard.writeText(value);
    } else {
      const temp = document.createElement("textarea");
      temp.value = value;
      temp.setAttribute("readonly", "true");
      temp.style.position = "absolute";
      temp.style.left = "-9999px";
      document.body.appendChild(temp);
      temp.select();
      document.execCommand("copy");
      document.body.removeChild(temp);
    }
    showToast(`${label || "Copied"} to clipboard.`);
  } catch {
    showToast("Copy failed.");
  }
}

async function toggleMaintenance(type, server) {
  if (!server || !type) return;
  const payload = { ...server, maintenance: !server.maintenance };
  const path = type === "local" ? "/api/locals" : "/api/servers";
  try {
    await fetchJSON(path, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    await refreshServers();
    await refreshStatus();
  } catch (err) {
    showToast(err.message || "Failed to update maintenance.");
  }
}

function closeDetailsModal() {
  if (!detailsModal) return;
  detailsModal.classList.add("hidden");
  detailsModal.setAttribute("aria-hidden", "true");
  currentDetailsServerKey = null;
  if (detailsBody) {
    detailsBody.innerHTML = "";
  }
  refreshStatus().catch(() => {});
}

function openPolicyModal() {
  if (!policyModal) return;
  policyModal.classList.remove("hidden");
  policyModal.setAttribute("aria-hidden", "false");
  if (policyCloseBtn) {
    policyCloseBtn.focus();
  }
}

function closePolicyModal() {
  if (!policyModal) return;
  policyModal.classList.add("hidden");
  policyModal.setAttribute("aria-hidden", "true");
}

function buildDetailsTitle(name) {
  const frag = document.createDocumentFragment();
  const iconWrap = document.createElement("span");
  iconWrap.className = "details-title-icon";
  const svg = document.createElementNS("http://www.w3.org/2000/svg", "svg");
  svg.setAttribute("class", "icon icon-tabler icons-tabler-outline icon-tabler-server-bolt");
  svg.setAttribute("width", "24");
  svg.setAttribute("height", "24");
  svg.setAttribute("viewBox", "0 0 24 24");
  svg.setAttribute("fill", "none");
  svg.setAttribute("stroke", "currentColor");
  svg.setAttribute("stroke-width", "2");
  svg.setAttribute("stroke-linecap", "round");
  svg.setAttribute("stroke-linejoin", "round");
  svg.setAttribute("aria-hidden", "true");
  svg.innerHTML = '<path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M3 7a3 3 0 0 1 3 -3h12a3 3 0 0 1 3 3v2a3 3 0 0 1 -3 3h-12a3 3 0 0 1 -3 -3" /><path d="M15 20h-9a3 3 0 0 1 -3 -3v-2a3 3 0 0 1 3 -3h12" /><path d="M7 8v.01" /><path d="M7 16v.01" /><path d="M20 15l-2 3h3l-2 3" />';
  iconWrap.appendChild(svg);
  const label = document.createElement("span");
  label.textContent = name;
  frag.append(iconWrap, label);
  return frag;
}

function buildContainerCard(container, result, canUpdateStopped, variant) {
  const card = document.createElement("div");
  card.className = "details-card";

  const name = document.createElement("strong");
  name.className = "details-card-title";
  const titleIcon = document.createElementNS("http://www.w3.org/2000/svg", "svg");
  titleIcon.setAttribute("class", "details-card-title-icon icon icon-tabler icons-tabler-outline icon-tabler-brand-docker");
  titleIcon.setAttribute("width", "24");
  titleIcon.setAttribute("height", "24");
  titleIcon.setAttribute("viewBox", "0 0 24 24");
  titleIcon.setAttribute("fill", "none");
  titleIcon.setAttribute("stroke", "currentColor");
  titleIcon.setAttribute("stroke-width", "2");
  titleIcon.setAttribute("stroke-linecap", "round");
  titleIcon.setAttribute("stroke-linejoin", "round");
  titleIcon.setAttribute("aria-hidden", "true");
  titleIcon.innerHTML = '<path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M22 12.54c-1.804 -.345 -2.701 -1.08 -3.523 -2.94c-.487 .696 -1.102 1.568 -.92 2.4c.028 .238 -.32 1 -.557 1h-14c0 5.208 3.164 7 6.196 7c4.124 .022 7.828 -1.376 9.854 -5c1.146 -.101 2.296 -1.505 2.95 -2.46" /><path d="M5 10h3v3h-3l0 -3" /><path d="M8 10h3v3h-3l0 -3" /><path d="M11 10h3v3h-3l0 -3" /><path d="M8 7h3v3h-3l0 -3" /><path d="M11 7h3v3h-3l0 -3" /><path d="M11 4h3v3h-3l0 -3" /><path d="M4.571 18c1.5 0 2.047 -.074 2.958 -.78" /><path d="M10 16l0 .01" />';
  const titleText = document.createElement("span");
  titleText.textContent = container.name;
  name.append(titleIcon, titleText);
  card.appendChild(name);

  const statusLine = document.createElement("div");
  let statusLabel = "Up to date";
  let statusClass = "details-card-status success";
  let statusIcon = '<svg xmlns="http://www.w3.org/2000/svg" class="details-card-status-icon icon icon-tabler icons-tabler-outline icon-tabler-progress-check" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M10 20.777a8.942 8.942 0 0 1 -2.48 -.969" /><path d="M14 3.223a9.003 9.003 0 0 1 0 17.554" /><path d="M4.579 17.093a8.961 8.961 0 0 1 -1.227 -2.592" /><path d="M3.124 10.5c.16 -.95 .468 -1.85 .9 -2.675l.169 -.305" /><path d="M6.907 4.579a8.954 8.954 0 0 1 3.093 -1.356" /><path d="M9 12l2 2l4 -4" /></svg>';
  if (variant === "updates") {
    statusLabel = "Update available";
    statusClass = "details-card-status warning";
    statusIcon = '<svg xmlns="http://www.w3.org/2000/svg" class="details-card-status-icon icon icon-tabler icons-tabler-outline icon-tabler-progress-down" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M10 20.777a8.942 8.942 0 0 1 -2.48 -.969" /><path d="M14 3.223a9.003 9.003 0 0 1 0 17.554" /><path d="M4.579 17.093a8.961 8.961 0 0 1 -1.227 -2.592" /><path d="M3.124 10.5c.16 -.95 .468 -1.85 .9 -2.675l.169 -.305" /><path d="M6.907 4.579a8.954 8.954 0 0 1 3.093 -1.356" /><path d="M12 9v6" /><path d="M15 12l-3 3l-3 -3" /></svg>';
  } else if (variant === "updated") {
    statusLabel = "Updated";
    statusClass = "details-card-status success";
    statusIcon = '<svg xmlns="http://www.w3.org/2000/svg" class="details-card-status-icon icon icon-tabler icons-tabler-outline icon-tabler-circle-check" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M3 12a9 9 0 1 0 18 0a9 9 0 1 0 -18 0" /><path d="M9 12l2 2l4 -4" /></svg>';
  } else if (variant === "skipped") {
    statusLabel = "Skipped";
    statusClass = "details-card-status warning";
    statusIcon = '<svg xmlns="http://www.w3.org/2000/svg" class="details-card-status-icon icon icon-tabler icons-tabler-outline icon-tabler-chevrons-right" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M7 7l5 5l-5 5" /><path d="M13 7l5 5l-5 5" /></svg>';
  }
  statusLine.className = statusClass;
  statusLine.innerHTML = `${statusIcon}<span>${statusLabel}</span>`;
  card.appendChild(statusLine);

  const isSelfContainer = result.local && container.name === "contiwatch";
  const serverKey = serverScopeKey(result);
  const pendingKey = pendingSelfUpdateKey(serverKey, container.name);
  const isPendingSelfUpdate = Boolean(pendingSelfUpdates[pendingKey] && pendingSelfUpdates[pendingKey] > Date.now());
  const updateStoppedEnabled = Boolean(currentConfig && currentConfig.update_stopped_containers);
  const canUpdate =
    container.update_available &&
    !container.updated &&
    !container.paused &&
    !currentScanController &&
    !updateInProgress &&
    (container.running || canUpdateStopped) &&
    !isSelfContainer &&
    !isPendingSelfUpdate;
  const blockedByStoppedSetting =
    container.update_available &&
    !container.updated &&
    !container.paused &&
    !currentScanController &&
    !updateInProgress &&
    !container.running &&
    !updateStoppedEnabled &&
    !isSelfContainer &&
    !isPendingSelfUpdate;

  const actions = document.createElement("div");
  actions.className = "details-card-actions";

  const updateBtn = document.createElement("button");
  updateBtn.type = "button";
  updateBtn.className = "btn-small";
  updateBtn.textContent = container.updated ? "Updated" : "Update";
  updateBtn.disabled = container.updated || (!canUpdate && !blockedByStoppedSetting);

  const infoBtn = document.createElement("button");
  infoBtn.type = "button";
  infoBtn.className = "btn-small secondary";
  infoBtn.textContent = "Info";

  if (isSelfContainer) {
    updateBtn.textContent = "Self-update disabled";
    updateBtn.disabled = true;
  }

  const updateState = document.createElement("span");
  updateState.className = "details-card-state hidden";

  if (isPendingSelfUpdate) {
    updateBtn.textContent = "Waiting…";
    updateBtn.disabled = true;
    updateState.classList.remove("hidden");
    updateState.textContent = "Agent restarting — waiting for reconnect…";
  }

  const infoPanel = document.createElement("div");
  infoPanel.className = "info-panel hidden";
  const currentImageID = container.image_id || "unknown";
  const newImageID = container.new_image_id || "unknown";
  infoPanel.textContent = `Current image: ${currentImageID}\nNew image: ${newImageID}`;

  updateBtn.addEventListener("click", async () => {
    if (blockedByStoppedSetting) {
      showToast(
        "Cannot update: container is stopped. Change settings to update stopped containers."
      );
      return;
    }
    if (updateInProgress || currentScanController || updateBtn.disabled) return;
    updateInProgress = true;
    scanBtn.disabled = true;
    if (statusSelectiveToggleBtn) {
      statusSelectiveToggleBtn.disabled = true;
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
      let updateURL = `/api/update/${encodeURIComponent(container.id)}?server=${serverParam}`;
      if (!result.local && isContiwatchImage(container.image)) {
        updateURL += "&self_update=1";
      }
      updateResult = await fetchJSON(updateURL, { method: "POST" });
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
        schedulePostSelfUpdateRefresh(serverKey, container.name);
      }
      scanBtn.disabled = Boolean(currentScanController);
      if (statusSelectiveToggleBtn) {
        statusSelectiveToggleBtn.disabled = Boolean(currentScanController);
      }
      const hideDelay = restartHint ? 30000 : 8000;
      setTimeout(() => updateState.classList.add("hidden"), hideDelay);
      await refreshStatus();
      await refreshLogs();
      if (updateResult) {
        if (updateResult.updated) {
          showToast(`Updated ${updateResult.name} (${updateResult.previous_state} → ${updateResult.current_state})`);
        } else if (restartHint) {
          showToast("Agent restarting — waiting for reconnect.");
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
  card.appendChild(actions);
  card.appendChild(infoPanel);

  return card;
}

function buildDetailsContent(result, canUpdateStopped) {
  const wrapper = document.createElement("div");
  const checkedAt = result.checked_at ? new Date(result.checked_at) : null;
  const hasCheckedAt = checkedAt && !Number.isNaN(checkedAt.getTime()) && checkedAt.getTime() > 0;
  const isOffline = Boolean(result.error);
  const restartKey = serverScopeKey(result);
  const isRestarting = Boolean(restartingServers[restartKey] && restartingServers[restartKey] > Date.now());
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

  const skippedContainers = result.containers
    .filter((container) => String(container.error || "").startsWith("skipped:"))
    .sort((a, b) => a.name.localeCompare(b.name));
  const updatedContainers = result.containers
    .filter((container) => container.updated)
    .sort((a, b) => a.name.localeCompare(b.name));
  const updateRequired = result.containers
    .filter((container) => container.update_available && !container.updated && !String(container.error || "").startsWith("skipped:"))
    .sort((a, b) => a.name.localeCompare(b.name));
  const upToDate = result.containers
    .filter((container) => !container.update_available && !container.updated && !String(container.error || "").startsWith("skipped:"))
    .sort((a, b) => a.name.localeCompare(b.name));

  const updatesGroup = document.createElement("details");
  updatesGroup.className = "details-group";
  updatesGroup.dataset.group = "updates";
  updatesGroup.open = true;
  const updatesSummary = document.createElement("summary");
  updatesSummary.className = "details-group-header";
  updatesSummary.innerHTML = `
    <span class="details-group-label">
      <svg xmlns="http://www.w3.org/2000/svg" class="details-group-icon icon icon-tabler icons-tabler-outline icon-tabler-progress-down" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M10 20.777a8.942 8.942 0 0 1 -2.48 -.969" /><path d="M14 3.223a9.003 9.003 0 0 1 0 17.554" /><path d="M4.579 17.093a8.961 8.961 0 0 1 -1.227 -2.592" /><path d="M3.124 10.5c.16 -.95 .468 -1.85 .9 -2.675l.169 -.305" /><path d="M6.907 4.579a8.954 8.954 0 0 1 3.093 -1.356" /><path d="M12 9v6" /><path d="M15 12l-3 3l-3 -3" /></svg>
      <span>Updates available</span>
      <span class="details-count">${updateRequired.length}</span>
    </span>
  `;
  updatesGroup.appendChild(updatesSummary);
  const updatesGrid = document.createElement("div");
  updatesGrid.className = "details-card-grid";
  if (updateRequired.length === 0) {
    const empty = document.createElement("p");
    empty.className = "details-empty";
    empty.textContent = "No updates required.";
    updatesGrid.appendChild(empty);
  } else {
    updateRequired.forEach((container) => updatesGrid.appendChild(buildContainerCard(container, result, canUpdateStopped, "updates")));
  }
  updatesGroup.appendChild(updatesGrid);
  wrapper.appendChild(updatesGroup);

  const updatedGroup = document.createElement("details");
  updatedGroup.className = "details-group";
  updatedGroup.dataset.group = "updated";
  updatedGroup.open = false;
  const updatedSummary = document.createElement("summary");
  updatedSummary.className = "details-group-header";
  updatedSummary.innerHTML = `
    <span class="details-group-label">
      <svg xmlns="http://www.w3.org/2000/svg" class="details-group-icon icon icon-tabler icons-tabler-outline icon-tabler-circle-check" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M3 12a9 9 0 1 0 18 0a9 9 0 1 0 -18 0" /><path d="M9 12l2 2l4 -4" /></svg>
      <span>Updated</span>
      <span class="details-count">${updatedContainers.length}</span>
    </span>
  `;
  updatedGroup.appendChild(updatedSummary);
  const updatedGrid = document.createElement("div");
  updatedGrid.className = "details-card-grid";
  if (updatedContainers.length === 0) {
    const empty = document.createElement("p");
    empty.className = "details-empty";
    empty.textContent = "No updated containers.";
    updatedGrid.appendChild(empty);
  } else {
    updatedContainers.forEach((container) => updatedGrid.appendChild(buildContainerCard(container, result, canUpdateStopped, "updated")));
  }
  updatedGroup.appendChild(updatedGrid);
  wrapper.appendChild(updatedGroup);

  const skippedGroup = document.createElement("details");
  skippedGroup.className = "details-group";
  skippedGroup.dataset.group = "skipped";
  skippedGroup.open = false;
  const skippedSummary = document.createElement("summary");
  skippedSummary.className = "details-group-header";
  skippedSummary.innerHTML = `
    <span class="details-group-label">
      <svg xmlns="http://www.w3.org/2000/svg" class="details-group-icon icon icon-tabler icons-tabler-outline icon-tabler-chevrons-right" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M7 7l5 5l-5 5" /><path d="M13 7l5 5l-5 5" /></svg>
      <span>Skipped</span>
      <span class="details-count">${skippedContainers.length}</span>
    </span>
  `;
  skippedGroup.appendChild(skippedSummary);
  const skippedGrid = document.createElement("div");
  skippedGrid.className = "details-card-grid";
  if (skippedContainers.length === 0) {
    const empty = document.createElement("p");
    empty.className = "details-empty";
    empty.textContent = "No skipped containers.";
    skippedGrid.appendChild(empty);
  } else {
    skippedContainers.forEach((container) => skippedGrid.appendChild(buildContainerCard(container, result, canUpdateStopped, "skipped")));
  }
  skippedGroup.appendChild(skippedGrid);
  wrapper.appendChild(skippedGroup);

  const scannedGroup = document.createElement("details");
  scannedGroup.className = "details-group";
  scannedGroup.dataset.group = "scanned";
  scannedGroup.open = false;
  const scannedSummary = document.createElement("summary");
  scannedSummary.className = "details-group-header";
  scannedSummary.innerHTML = `
    <span class="details-group-label">
      <svg xmlns="http://www.w3.org/2000/svg" class="details-group-icon icon icon-tabler icons-tabler-outline icon-tabler-circle-dot" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M11 12a1 1 0 1 0 2 0a1 1 0 1 0 -2 0" /><path d="M3 12a9 9 0 1 0 18 0a9 9 0 1 0 -18 0" /></svg>
      <span>Scanned containers</span>
      <span class="details-count">${upToDate.length}</span>
    </span>
  `;
  scannedGroup.appendChild(scannedSummary);
  const scannedGrid = document.createElement("div");
  scannedGrid.className = "details-card-grid";
  if (upToDate.length === 0) {
    const empty = document.createElement("p");
    empty.className = "details-empty";
    empty.textContent = "No containers.";
    scannedGrid.appendChild(empty);
  } else {
    upToDate.forEach((container) => scannedGrid.appendChild(buildContainerCard(container, result, canUpdateStopped, "scanned")));
  }
  scannedGroup.appendChild(scannedGrid);
  wrapper.appendChild(scannedGroup);
  return wrapper;
}

function openDetailsModal(result, canUpdateStopped) {
  if (!detailsModal || !detailsBody) return;
  currentDetailsServerKey = serverScopeKey(result);
  detailsBody.innerHTML = "";
  detailsBody.appendChild(buildDetailsContent(result, canUpdateStopped));
  if (detailsTitle) {
    detailsTitle.textContent = "";
    detailsTitle.appendChild(buildDetailsTitle(result.server_name || "server"));
  }
  detailsModal.classList.remove("hidden");
  detailsModal.setAttribute("aria-hidden", "false");
  if (detailsCloseBtn) {
    detailsCloseBtn.focus();
  }
}

function updateDetailsModal(result, canUpdateStopped) {
  if (!detailsModal || !detailsBody) return;
  const groupStates = new Map();
  detailsBody.querySelectorAll("details[data-group]").forEach((group) => {
    groupStates.set(group.dataset.group, group.open);
  });
  detailsBody.innerHTML = "";
  detailsBody.appendChild(buildDetailsContent(result, canUpdateStopped));
  detailsBody.querySelectorAll("details[data-group]").forEach((group) => {
    if (groupStates.has(group.dataset.group)) {
      group.open = groupStates.get(group.dataset.group);
    }
  });
  if (detailsTitle) {
    detailsTitle.textContent = "";
    detailsTitle.appendChild(buildDetailsTitle(result.server_name || "server"));
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

  const visibleResults = selectedScanScope && selectedScanScope !== "all" && selectedScanScope !== "selective"
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
    const key = serverScopeKey(result);
    const infoKey = `${result.local ? "local" : "remote"}:${result.server_name || ""}`;
    const infoEntry = cachedServerInfo[infoKey] || {};
    const infoStatusRaw = String(infoEntry.status || "");
    const infoChecking = debouncedCheckingVisible(infoKey, infoEntry);
    const infoCheckedAt = infoEntry.checked_at;
    const infoStatus = infoStatusRaw.toLowerCase();
    const serverConfig = result.local
      ? cachedLocals.find((local) => local.name === result.server_name)
      : cachedRemotes.find((remote) => remote.name === result.server_name);
    const isMaintenance = Boolean(serverConfig && serverConfig.maintenance);
    const backendState = String(result.scan_state || "").toLowerCase();
    const overrideState = scanStateOverrides[key];
    let scanState = backendState && backendState !== "idle"
      ? backendState
      : (overrideState || backendState);
    const cancelError = result.error && /cancelled|canceled|context canceled/i.test(result.error);
    const isPending = scanState === "pending";
    const isScanningActive = scanState === "scanning";
    const isUpdatingActive = scanState === "updating";
    const isCancelled = scanState === "cancelled" || cancelError;
    const hasScanErrorState = scanState === "error";
    const restartKey = serverScopeKey(result);
    const isRestarting = Boolean(restartingServers[restartKey] && restartingServers[restartKey] > Date.now());
    const isOfflineForStatus = !isMaintenance && ((isOffline && !isCancelled) || infoStatus === "offline");
    const isRestartingForStatus = !isMaintenance && (isRestarting || infoStatus === "restarting");
    let statusLabel = "online";
    if (isMaintenance) {
      statusLabel = "maintenance";
    } else if (isRestartingForStatus) {
      statusLabel = "restarting";
    } else if (isOfflineForStatus) {
      statusLabel = "offline";
    } else if (infoChecking) {
      statusLabel = "checking";
    }
    if (!hasCheckedAt && !infoStatus && !result.error && !isMaintenance && !infoChecking) {
      statusLabel = "unknown";
    }
    let scanLabelText = "";
    let scanLabelClass = "";
    if (isMaintenance) {
      scanLabelText = "maintenance";
      scanLabelClass = "maintenance";
    } else if (!isRestarting) {
      if (isUpdatingActive) {
        scanLabelText = "updating";
        scanLabelClass = "updating";
      } else if (isScanningActive) {
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
    if (!scanLabelText && infoChecking) {
      scanLabelText = "checking";
      scanLabelClass = "checking";
    }
    if (scanLabelText === "checking" && (scanActive || scanRequestActive || isScanningActive || isUpdatingActive || isPending)) {
      scanLabelText = "";
      scanLabelClass = "";
    }
    if (!scanActive && !scanRequestActive && !isScanningActive && !isUpdatingActive && !isPending && infoStatus === "offline") {
      scanLabelText = "";
      scanLabelClass = "";
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
    nameIcon.innerHTML = '<path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M3 7a3 3 0 0 1 3 -3h12a3 3 0 0 1 3 3v2a3 3 0 0 1 -3 3h-12a3 3 0 0 1 -3 -3" /><path d="M15 20h-9a3 3 0 0 1 -3 -3v-2a3 3 0 0 1 3 -3h12" /><path d="M7 8v.01" /><path d="M7 16v.01" /><path d="M20 15l-2 3h3l-2 3" />';
    const nameIconWrap = document.createElement("span");
    nameIconWrap.className = "status-card-icon-wrap has-tooltip";
    nameIconWrap.dataset.tooltipServerStatus = "true";
    nameIconWrap.dataset.statusLabel = statusLabel;
    nameIconWrap.dataset.checkedAt = infoCheckedAt || "";
    nameIconWrap.setAttribute("data-tooltip", `Status: ${statusLabel}`);
    nameIconWrap.setAttribute("aria-label", `Status: ${statusLabel}`);
    nameIconWrap.appendChild(nameIcon);
    const name = document.createElement("strong");
    name.textContent = result.server_name || "unknown";
    nameWrap.appendChild(nameIconWrap);
    nameWrap.appendChild(name);
    headLeft.appendChild(nameWrap);

    const typeIcon = document.createElementNS("http://www.w3.org/2000/svg", "svg");
    typeIcon.setAttribute(
      "class",
      `status-card-type-icon icon icon-tabler icons-tabler-outline icon-tabler-${result.local ? "plug-connected" : "network"}`
    );
    typeIcon.setAttribute("width", "24");
    typeIcon.setAttribute("height", "24");
    typeIcon.setAttribute("viewBox", "0 0 24 24");
    typeIcon.setAttribute("fill", "none");
    typeIcon.setAttribute("stroke", "currentColor");
    typeIcon.setAttribute("stroke-width", "2");
    typeIcon.setAttribute("stroke-linecap", "round");
    typeIcon.setAttribute("stroke-linejoin", "round");
    typeIcon.setAttribute("aria-hidden", "true");
    const typeWrap = document.createElement("span");
    typeWrap.className = "status-card-type has-tooltip";
    if (result.local) {
      typeWrap.setAttribute("data-tooltip", "Local socket");
      typeWrap.setAttribute("aria-label", "Local socket");
      typeIcon.innerHTML = '<path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M7 12l5 5l-1.5 1.5a3.536 3.536 0 1 1 -5 -5l1.5 -1.5" /><path d="M17 12l-5 -5l1.5 -1.5a3.536 3.536 0 1 1 5 5l-1.5 1.5" /><path d="M3 21l2.5 -2.5" /><path d="M18.5 5.5l2.5 -2.5" /><path d="M10 11l-2 2" /><path d="M13 14l-2 2" />';
    } else {
      typeWrap.setAttribute("data-tooltip", "Remote agent");
      typeWrap.setAttribute("aria-label", "Remote agent");
      typeIcon.innerHTML = '<path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M6 9a6 6 0 1 0 12 0a6 6 0 0 0 -12 0" /><path d="M12 3c1.333 .333 2 2.333 2 6s-.667 5.667 -2 6" /><path d="M12 3c-1.333 .333 -2 2.333 -2 6s.667 5.667 2 6" /><path d="M6 9h12" /><path d="M3 20h7" /><path d="M14 20h7" /><path d="M10 20a2 2 0 1 0 4 0a2 2 0 0 0 -4 0" /><path d="M12 15v3" />';
    }
    typeWrap.appendChild(typeIcon);
    headLeft.appendChild(typeWrap);

    if (currentConfig && currentConfig.scheduler_enabled) {
      const schedulerWrap = document.createElement("span");
      schedulerWrap.className = "status-card-scheduler-wrap has-tooltip";
      const tooltip = formatSchedulerNextRun(currentConfig);
      schedulerWrap.setAttribute("data-tooltip", tooltip);
      schedulerWrap.setAttribute("aria-label", tooltip);
      schedulerWrap.dataset.tooltipScheduler = "true";

      const schedulerIcon = document.createElementNS("http://www.w3.org/2000/svg", "svg");
      schedulerIcon.setAttribute(
        "class",
        "status-card-scheduler icon icon-tabler icons-tabler-outline icon-tabler-calendar-check"
      );
      schedulerIcon.setAttribute("width", "24");
      schedulerIcon.setAttribute("height", "24");
      schedulerIcon.setAttribute("viewBox", "0 0 24 24");
      schedulerIcon.setAttribute("fill", "none");
      schedulerIcon.setAttribute("stroke", "currentColor");
      schedulerIcon.setAttribute("stroke-width", "2");
      schedulerIcon.setAttribute("stroke-linecap", "round");
      schedulerIcon.setAttribute("stroke-linejoin", "round");
      schedulerIcon.setAttribute("aria-hidden", "true");
      schedulerIcon.innerHTML = '<path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M11 19h-6a2 2 0 0 1 -2 -2v-11a2 2 0 0 1 2 -2h11a2 2 0 0 1 2 2v5" /><path d="M16 3v4" /><path d="M8 3v4" /><path d="M4 11h12" /><path d="M15 19l2 2l4 -4" />';
      schedulerWrap.appendChild(schedulerIcon);
      headLeft.appendChild(schedulerWrap);
    }

    // Status indicator moved to the server icon.
    if (scanLabelText) {
      const scanBadge = document.createElement("span");
      scanBadge.className = `status-badge status-${scanLabelClass}`;
      scanBadge.textContent = scanLabelText;
      headLeft.appendChild(scanBadge);
    }

    const summary = document.createElement("div");
    summary.className = "server-summary";
    const summaryTop = document.createElement("div");
    summaryTop.className = "server-summary-metrics";
    const totalScanned = hasCheckedAt && Array.isArray(result.containers) ? result.containers.length : 0;
    const updates = hasCheckedAt && Array.isArray(result.containers)
      ? result.containers.filter((container) => container.update_available).length
      : 0;
    const updated = hasCheckedAt && Array.isArray(result.containers)
      ? result.containers.filter((container) => container.updated).length
      : 0;
    const skipped = hasCheckedAt && Array.isArray(result.containers)
      ? result.containers.filter((container) => String(container.error || "").startsWith("skipped:")).length
      : 0;
    const upToDate = totalScanned > 0 ? Math.max(0, totalScanned - updates - skipped) : 0;

    const upToDateEl = document.createElement("span");
    upToDateEl.className = `summary-metric${hasCheckedAt && upToDate > 0 ? " metric-success" : ""}`;
    upToDateEl.innerHTML =
      '<svg xmlns="http://www.w3.org/2000/svg" class="summary-icon icon icon-tabler icons-tabler-outline icon-tabler-progress-check" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M10 20.777a8.942 8.942 0 0 1 -2.48 -.969" /><path d="M14 3.223a9.003 9.003 0 0 1 0 17.554" /><path d="M4.579 17.093a8.961 8.961 0 0 1 -1.227 -2.592" /><path d="M3.124 10.5c.16 -.95 .468 -1.85 .9 -2.675l.169 -.305" /><path d="M6.907 4.579a8.954 8.954 0 0 1 3.093 -1.356" /><path d="M9 12l2 2l4 -4" /></svg>' +
      `<span>${upToDate}</span><span class="summary-label">Latest</span>`;

    const updatesEl = document.createElement("span");
    updatesEl.className = `summary-metric${hasCheckedAt && updates > 0 ? " metric-warning" : ""}`;
    updatesEl.innerHTML =
      '<svg xmlns="http://www.w3.org/2000/svg" class="summary-icon icon icon-tabler icons-tabler-outline icon-tabler-progress-down" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M10 20.777a8.942 8.942 0 0 1 -2.48 -.969" /><path d="M14 3.223a9.003 9.003 0 0 1 0 17.554" /><path d="M4.579 17.093a8.961 8.961 0 0 1 -1.227 -2.592" /><path d="M3.124 10.5c.16 -.95 .468 -1.85 .9 -2.675l.169 -.305" /><path d="M6.907 4.579a8.954 8.954 0 0 1 3.093 -1.356" /><path d="M12 9v6" /><path d="M15 12l-3 3l-3 -3" /></svg>' +
      `<span>${updates}</span><span class="summary-label">Outdated</span>`;

    const scannedEl = document.createElement("span");
    scannedEl.className = "summary-metric";
    scannedEl.innerHTML =
      '<svg xmlns="http://www.w3.org/2000/svg" class="summary-icon icon icon-tabler icons-tabler-outline icon-tabler-circle-dot" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M11 12a1 1 0 1 0 2 0a1 1 0 1 0 -2 0" /><path d="M3 12a9 9 0 1 0 18 0a9 9 0 1 0 -18 0" /></svg>' +
      `<span>${totalScanned}</span><span class="summary-label">Scanned</span>`;

    const updatedEl = document.createElement("span");
    updatedEl.className = `summary-metric${hasCheckedAt && updated > 0 ? " metric-success" : ""}`;
    updatedEl.innerHTML =
      '<svg xmlns="http://www.w3.org/2000/svg" class="summary-icon icon icon-tabler icons-tabler-outline icon-tabler-circle-check" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M3 12a9 9 0 1 0 18 0a9 9 0 1 0 -18 0" /><path d="M9 12l2 2l4 -4" /></svg>' +
      `<span>${updated}</span><span class="summary-label">Updated</span>`;

    const skippedEl = document.createElement("span");
    skippedEl.className = "summary-metric";
    skippedEl.innerHTML =
      '<svg xmlns="http://www.w3.org/2000/svg" class="summary-icon icon icon-tabler icons-tabler-outline icon-tabler-chevrons-right" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M7 7l5 5l-5 5" /><path d="M13 7l5 5l-5 5" /></svg>' +
      `<span>${skipped}</span><span class="summary-label">Skipped</span>`;

    summaryTop.appendChild(upToDateEl);
    summaryTop.appendChild(updatesEl);
    summaryTop.appendChild(scannedEl);
    summary.appendChild(summaryTop);

    const summaryBottom = document.createElement("div");
    summaryBottom.className = "server-summary-metrics";
    summaryBottom.appendChild(updatedEl);
    summaryBottom.appendChild(skippedEl);
    summary.appendChild(summaryBottom);

    const meta = document.createElement("div");
    meta.className = "status-card-meta";
    if (isCancelled) {
      meta.textContent = "Last scan: cancelled.";
    } else {
      meta.textContent = hasCheckedAt
        ? `Last scan: ${checkedAt.toLocaleString([], {
          year: "numeric",
          month: "short",
          day: "2-digit",
          hour: "2-digit",
          minute: "2-digit",
        })}`
        : "Last scan: not run yet.";
    }

    headLeft.appendChild(summary);
    headLeft.appendChild(meta);

	    const headRight = document.createElement("div");
	    headRight.className = "status-card-actions";
	    const detailsBtn = document.createElement("button");
	    detailsBtn.type = "button";
	    detailsBtn.className = "secondary btn-small icon-action-btn has-tooltip";
	    const serverKey = serverScopeKey(result);
	    if (selectedScanScope === "selective") {
	      detailsBtn.classList.add("server-select-toggle");
	      detailsBtn.setAttribute("aria-pressed", selectedScanServers.has(serverKey) ? "true" : "false");
	      const checkbox = document.createElement("input");
      checkbox.type = "checkbox";
      checkbox.className = "server-select-checkbox";
      checkbox.checked = selectedScanServers.has(serverKey);
	      checkbox.setAttribute("aria-label", `Select ${result.server_name || "server"}`);
	      const label = document.createElement("span");
	      label.className = "server-select-label";
	      label.textContent = "Details";
	      detailsBtn.append(checkbox, label);
      const setChecked = (checked) => {
        if (checked) {
          selectedScanServers.add(serverKey);
        } else {
          selectedScanServers.delete(serverKey);
        }
        detailsBtn.setAttribute("aria-pressed", checked ? "true" : "false");
        checkbox.checked = checked;
      };
      checkbox.addEventListener("click", (event) => {
        event.stopPropagation();
      });
      detailsBtn.addEventListener("click", (event) => {
        event.preventDefault();
        if (detailsBtn.disabled) return;
        setChecked(!checkbox.checked);
      });
      checkbox.addEventListener("change", () => {
        setChecked(checkbox.checked);
      });
	      detailsBtn.disabled = updateInProgress || scanActive || isMaintenance;
	    } else {
	      detailsBtn.setAttribute("aria-label", "Details");
	      detailsBtn.setAttribute("data-tooltip", "Details");
	      const icon = document.createElement("span");
	      icon.className = "icon-action icon-details";
	      icon.setAttribute("aria-hidden", "true");
	      detailsBtn.appendChild(icon);
	      detailsBtn.addEventListener("click", () => {
	        openDetailsModal(result, canUpdateStopped);
	      });
	      detailsBtn.disabled = updateInProgress || scanActive;
	    }
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

function buildServerActions(item) {
  const actions = document.createElement("div");
  actions.className = "servers-row-actions";
  const isMaintenance = Boolean(item.server.maintenance);
  const confirmKey = `${item.type}:${item.server.name}`;

  function makeActionIcon(className) {
    const icon = document.createElement("span");
    icon.className = `icon-action ${className}`;
    icon.setAttribute("aria-hidden", "true");
    return icon;
  }

  const editBtn = document.createElement("button");
  editBtn.type = "button";
  editBtn.className = "secondary icon-action-btn has-tooltip";
  editBtn.setAttribute("aria-label", "Edit server");
  editBtn.setAttribute("data-tooltip", "Edit server");
  editBtn.appendChild(makeActionIcon("icon-edit"));
  editBtn.addEventListener("click", () => {
    if (item.type === "local") {
      openLocalModal("edit", item.server);
    } else {
      openRemoteModal("edit", item.server);
    }
  });

  const maintenanceBtn = document.createElement("button");
  maintenanceBtn.type = "button";
  maintenanceBtn.className = "secondary icon-action-btn has-tooltip";
  maintenanceBtn.setAttribute("aria-label", isMaintenance ? "End Maintenance" : "Maintenance");
  maintenanceBtn.setAttribute("data-tooltip", isMaintenance ? "End Maintenance" : "Maintenance");
  maintenanceBtn.appendChild(makeActionIcon(isMaintenance ? "icon-barrier-block-off" : "icon-barrier-block"));
  maintenanceBtn.addEventListener("click", async () => {
    await toggleMaintenance(item.type, item.server);
  });

  const removeBtn = document.createElement("button");
  removeBtn.type = "button";
  removeBtn.className = "secondary icon-action-btn has-tooltip servers-remove-btn";
  removeBtn.setAttribute("aria-label", "Remove server");
  removeBtn.setAttribute("data-tooltip", "Remove server");
  removeBtn.appendChild(makeActionIcon("icon-trash"));

  const confirmBtn = document.createElement("button");
  confirmBtn.type = "button";
  confirmBtn.className = "btn-danger icon-action-btn has-tooltip hidden servers-confirm-btn";
  confirmBtn.setAttribute("aria-label", "Confirm remove");
  confirmBtn.setAttribute("data-tooltip", "Confirm remove");
  confirmBtn.appendChild(makeActionIcon("icon-trash-x"));

  function setConfirming(next) {
    if (next) {
      serversRemoveConfirming.add(confirmKey);
      removeBtn.classList.add("hidden");
      confirmBtn.classList.remove("hidden");
    } else {
      serversRemoveConfirming.delete(confirmKey);
      removeBtn.classList.remove("hidden");
      confirmBtn.classList.add("hidden");
    }
  }

  removeBtn.addEventListener("click", () => {
    setConfirming(true);
  });

  confirmBtn.addEventListener("click", async () => {
    const path = item.type === "local" ? "/api/locals/" : "/api/servers/";
    setConfirming(false);
    await fetchJSON(`${path}${encodeURIComponent(item.server.name)}`, { method: "DELETE" });
    await refreshServers();
    await refreshStatus();
  });

  actions.append(editBtn, maintenanceBtn, removeBtn, confirmBtn);
  if (serversRemoveConfirming.has(confirmKey)) {
    setConfirming(true);
  }
  return actions;
}

function buildServersNameRow(item, statusLabel) {
  const infoKey = `${item.type}:${item.server.name}`;
  const info = cachedServerInfo[infoKey] || {};
  const row = document.createElement("div");
  row.className = "servers-name-row";

  const nameIcon = document.createElementNS("http://www.w3.org/2000/svg", "svg");
  nameIcon.setAttribute(
    "class",
    `status-card-icon status-icon status-icon-${statusLabel} icon icon-tabler icons-tabler-outline icon-tabler-server-bolt`
  );
  nameIcon.setAttribute("width", "24");
  nameIcon.setAttribute("height", "24");
  nameIcon.setAttribute("viewBox", "0 0 24 24");
  nameIcon.setAttribute("fill", "none");
  nameIcon.setAttribute("stroke", "currentColor");
  nameIcon.setAttribute("stroke-width", "2");
  nameIcon.setAttribute("stroke-linecap", "round");
  nameIcon.setAttribute("stroke-linejoin", "round");
  nameIcon.setAttribute("aria-hidden", "true");
  nameIcon.innerHTML = '<path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M3 7a3 3 0 0 1 3 -3h12a3 3 0 0 1 3 3v2a3 3 0 0 1 -3 3h-12a3 3 0 0 1 -3 -3" /><path d="M15 20h-9a3 3 0 0 1 -3 -3v-2a3 3 0 0 1 3 -3h12" /><path d="M7 8v.01" /><path d="M7 16v.01" /><path d="M20 15l-2 3h3l-2 3" />';

  const nameIconWrap = document.createElement("span");
  nameIconWrap.className = "status-card-icon-wrap has-tooltip";
  nameIconWrap.dataset.tooltipServerStatus = "true";
  nameIconWrap.dataset.statusLabel = statusLabel;
  nameIconWrap.dataset.checkedAt = info.checked_at || "";
  nameIconWrap.setAttribute("data-tooltip", `Status: ${statusLabel}`);
  nameIconWrap.setAttribute("aria-label", `Status: ${statusLabel}`);
  nameIconWrap.appendChild(nameIcon);

  const nameEl = document.createElement("strong");
  nameEl.textContent = item.server.name;

  const typeIcon = document.createElementNS("http://www.w3.org/2000/svg", "svg");
  typeIcon.setAttribute(
    "class",
    `status-card-type-icon icon icon-tabler icons-tabler-outline icon-tabler-${item.type === "local" ? "plug-connected" : "network"}`
  );
  typeIcon.setAttribute("width", "24");
  typeIcon.setAttribute("height", "24");
  typeIcon.setAttribute("viewBox", "0 0 24 24");
  typeIcon.setAttribute("fill", "none");
  typeIcon.setAttribute("stroke", "currentColor");
  typeIcon.setAttribute("stroke-width", "2");
  typeIcon.setAttribute("stroke-linecap", "round");
  typeIcon.setAttribute("stroke-linejoin", "round");
  typeIcon.setAttribute("aria-hidden", "true");
  const typeWrap = document.createElement("span");
  typeWrap.className = "status-card-type has-tooltip";
  if (item.type === "local") {
    typeWrap.setAttribute("data-tooltip", "Local socket");
    typeWrap.setAttribute("aria-label", "Local socket");
    typeIcon.innerHTML = '<path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M7 12l5 5l-1.5 1.5a3.536 3.536 0 1 1 -5 -5l1.5 -1.5" /><path d="M17 12l-5 -5l1.5 -1.5a3.536 3.536 0 1 1 5 5l-1.5 1.5" /><path d="M3 21l2.5 -2.5" /><path d="M18.5 5.5l2.5 -2.5" /><path d="M10 11l-2 2" /><path d="M13 14l-2 2" />';
  } else {
    typeWrap.setAttribute("data-tooltip", "Remote agent");
    typeWrap.setAttribute("aria-label", "Remote agent");
    typeIcon.innerHTML = '<path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M6 9a6 6 0 1 0 12 0a6 6 0 0 0 -12 0" /><path d="M12 3c1.333 .333 2 2.333 2 6s-.667 5.667 -2 6" /><path d="M12 3c-1.333 .333 -2 2.333 -2 6s.667 5.667 2 6" /><path d="M6 9h12" /><path d="M3 20h7" /><path d="M14 20h7" /><path d="M10 20a2 2 0 1 0 4 0a2 2 0 0 0 -4 0" /><path d="M12 15v3" />';
  }
  typeWrap.appendChild(typeIcon);

  row.append(nameIconWrap, nameEl, typeWrap);
  return row;
}

function renderServers(localServers, remoteServers) {
  if (!serversTableBody || !serversCardsEl || !serversTableWrap) return;
  const items = [
    ...localServers.map((server) => ({ type: "local", server })),
    ...remoteServers.map((server) => ({ type: "remote", server })),
  ];
  const query = normalizeQuery(serversSearchQuery);
  const filtered = items.filter((item) => {
    const isMaintenance = Boolean(item.server && item.server.maintenance);
    const infoKey = `${item.type}:${item.server.name}`;
    const statusValue = String(cachedServerInfo[infoKey]?.status || "").toLowerCase();
    if (serversFilterMode === "maintenance") return isMaintenance;
    if (serversFilterMode === "active") return !isMaintenance;
    if (serversFilterMode === "offline") return !isMaintenance && statusValue === "offline";
    if (!query) return true;
    const address = item.type === "local" ? item.server.socket : item.server.url;
    const haystack = normalizeQuery(`${item.server.name} ${address} ${item.type} ${statusValue}`);
    return haystack.includes(query);
  });

  const showTable = serversViewMode === "table";
  serversTableWrap.classList.toggle("hidden", !showTable);
  serversCardsEl.classList.toggle("hidden", showTable);

  serversTableBody.innerHTML = "";
  serversCardsEl.innerHTML = "";

  if (items.length === 0 || filtered.length === 0) {
    const message = items.length === 0 ? "No servers added." : "No matching servers.";
    if (showTable) {
      const row = document.createElement("tr");
      const cell = document.createElement("td");
      cell.colSpan = 5;
      cell.className = "servers-empty";
      cell.textContent = message;
      row.appendChild(cell);
      serversTableBody.appendChild(row);
    } else {
      const empty = document.createElement("div");
      empty.className = "servers-empty";
      empty.textContent = message;
      serversCardsEl.appendChild(empty);
    }
    return;
  }

  filtered.forEach((item) => {
    const infoKey = `${item.type}:${item.server.name}`;
    const info = cachedServerInfo[infoKey] || {};
    const isMaintenance = Boolean(item.server.maintenance);
    const infoChecking = debouncedCheckingVisible(infoKey, info);
    let statusValue = String(info.status || "").toLowerCase();
    if (infoChecking) statusValue = "checking";
    if (isMaintenance) statusValue = "maintenance";
    if (!statusValue) statusValue = "unknown";
    const statusClass = ["online", "offline", "maintenance", "restarting", "checking"].includes(statusValue)
      ? statusValue
      : "";
    const statusLabel = statusValue;

    const addressValue = item.type === "local" ? item.server.socket : item.server.url;
    const actions = buildServerActions(item);
    const versionValue = info && info.version ? formatVersion(info.version) : "unknown";

    if (showTable) {
      const row = document.createElement("tr");
      const nameCell = document.createElement("td");
      const nameWrap = document.createElement("div");
      nameWrap.className = "servers-name";
      const nameRow = buildServersNameRow(item, statusLabel);
      nameWrap.appendChild(nameRow);
      nameCell.appendChild(nameWrap);

      const addressCell = document.createElement("td");
      addressCell.className = "servers-address";
      addressCell.textContent = addressValue || "unknown";

      const versionCell = document.createElement("td");
      versionCell.className = "servers-version";
      versionCell.textContent = versionValue;

      const statusCell = document.createElement("td");
      const statusPill = document.createElement("span");
      statusPill.className = `server-status-pill${statusClass ? ` server-status-${statusClass}` : ""}`;
      statusPill.textContent = statusLabel;
      statusPill.classList.add("has-tooltip");
      statusPill.dataset.tooltipLastChecked = "true";
      statusPill.dataset.checkedAt = info.checked_at || "";
      statusPill.setAttribute("data-tooltip", "Last checked: unavailable");
      statusPill.setAttribute("aria-label", "Last checked: unavailable");
      statusCell.appendChild(statusPill);

      const actionsCell = document.createElement("td");
      actionsCell.appendChild(actions);

      row.append(nameCell, addressCell, versionCell, statusCell, actionsCell);
      serversTableBody.appendChild(row);
    } else {
      const card = document.createElement("div");
      card.className = "servers-card";
      const head = document.createElement("div");
      head.className = "servers-card-head";
      const meta = document.createElement("div");
      meta.className = "servers-card-meta";
      const title = buildServersNameRow(item, statusLabel);
      title.classList.add("servers-card-name");
      meta.appendChild(title);

      const statusPill = document.createElement("span");
      statusPill.className = `server-status-pill${statusClass ? ` server-status-${statusClass}` : ""}`;
      statusPill.textContent = statusLabel;
      statusPill.classList.add("has-tooltip");
      statusPill.dataset.tooltipLastChecked = "true";
      statusPill.dataset.checkedAt = info.checked_at || "";
      statusPill.setAttribute("data-tooltip", "Last checked: unavailable");
      statusPill.setAttribute("aria-label", "Last checked: unavailable");
      head.append(meta, statusPill);

      const address = document.createElement("div");
      address.className = "servers-card-address";
      address.textContent = addressValue || "unknown";

      const version = document.createElement("div");
      version.className = "servers-card-version";
      version.textContent = versionValue;

      const actionsWrap = document.createElement("div");
      actionsWrap.className = "servers-card-actions";
      actionsWrap.appendChild(actions);

      card.append(head, address, version, actionsWrap);
      serversCardsEl.appendChild(card);
    }
  });
}

function formatVersion(raw) {
  const version = String(raw || "").trim();
  if (!version) return "unknown";
  if (version.startsWith("v")) return version;
  return `v${version}`;
}

function formatLastChecked(value) {
  if (!value) return "";
  const checkedAt = new Date(value);
  if (Number.isNaN(checkedAt.getTime())) return "";
  if (checkedAt.getUTCFullYear() < 2000) return "";
  const diffMs = Date.now() - checkedAt.getTime();
  if (diffMs < 0) return "Last checked: just now";
  const seconds = Math.floor(diffMs / 1000);
  if (seconds < 10) return "Last checked: just now";
  if (seconds < 60) return `Last checked: ${seconds}s ago`;
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `Last checked: ${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `Last checked: ${hours}h ago`;
  const days = Math.floor(hours / 24);
  return `Last checked: ${days}d ago`;
}

function debouncedCheckingVisible(serverKey, info) {
  if (!info || !info.checking) return false;
  const startedAt = checkingStartMsByServerKey[serverKey];
  if (!startedAt) return false;
  return Date.now() - startedAt >= 300;
}

function scheduleCheckingDebounceRender(serverKey) {
  if (checkingDebounceTimersByServerKey[serverKey]) return;
  checkingDebounceTimersByServerKey[serverKey] = window.setTimeout(() => {
    delete checkingDebounceTimersByServerKey[serverKey];
    if (cachedServerInfo[serverKey] && cachedServerInfo[serverKey].checking) {
      renderServers(cachedLocals, cachedRemotes);
      if (currentView === "status") {
        renderStatus(lastStatusResults.length > 0 ? lastStatusResults : buildStatusPlaceholders());
      }
    }
  }, 320);
}

function applyCheckingTracking(serverKey, info) {
  if (!info || !info.checking) {
    delete checkingStartMsByServerKey[serverKey];
    if (checkingDebounceTimersByServerKey[serverKey]) {
      window.clearTimeout(checkingDebounceTimersByServerKey[serverKey]);
      delete checkingDebounceTimersByServerKey[serverKey];
    }
    return;
  }
  if (!checkingStartMsByServerKey[serverKey]) {
    checkingStartMsByServerKey[serverKey] = Date.now();
    scheduleCheckingDebounceRender(serverKey);
  }
}

function serverInfoCheckedAtMs(info) {
  if (!info || !info.checked_at) return 0;
  const checkedAt = new Date(info.checked_at);
  if (Number.isNaN(checkedAt.getTime())) return 0;
  return checkedAt.getTime();
}

function shouldReplaceServerInfo(existing, incoming) {
  if (!existing) return true;
  const existingAddress = String(existing.address || "");
  const incomingAddress = String(incoming.address || "");
  if (incomingAddress && existingAddress && incomingAddress !== existingAddress) {
    return true;
  }
  const incomingCheckedAt = serverInfoCheckedAtMs(incoming);
  const existingCheckedAt = serverInfoCheckedAtMs(existing);
  if (incomingCheckedAt > existingCheckedAt) return true;
  if (incomingCheckedAt < existingCheckedAt) return false;
  const existingChecking = Boolean(existing.checking);
  const incomingChecking = Boolean(incoming.checking);
  if (existingChecking && !incomingChecking) return true;
  if (!existingChecking && incomingChecking) return false;
  const existingStatus = String(existing.status || "");
  const incomingStatus = String(incoming.status || "");
  if (!existingStatus && incomingStatus) return true;
  if (existingStatus && !incomingStatus) return false;
  if (existingStatus.toLowerCase() === "unknown" && incomingStatus.toLowerCase() !== "unknown") return true;
  return false;
}

function mergeServerInfoSnapshot(list) {
  if (!Array.isArray(list)) return;
  list.forEach((item) => {
    if (!item || !item.name || !item.type) return;
    const key = `${item.type}:${item.name}`;
    applyCheckingTracking(key, item);
    const existing = cachedServerInfo[key];
    if (shouldReplaceServerInfo(existing, item)) {
      cachedServerInfo[key] = item;
    }
  });
}

function normalizeQuery(value) {
  return String(value || "").trim().toLowerCase();
}

function applySidebarFilter(queryOverride) {
  if (!sidebarSearch) return;
  const query = normalizeQuery(queryOverride ?? sidebarSearch.value);
  if (currentView === "servers") {
    serversSearchQuery = query;
    renderServers(cachedLocals, cachedRemotes);
    return;
  }
  if (currentView !== "status") return;
  const cards = Array.from(statusEl ? statusEl.querySelectorAll(".server-card") : []);
  cards.forEach((el) => {
    const source = el.dataset && el.dataset.search ? el.dataset.search : el.textContent;
    const haystack = normalizeQuery(source);
    const match = !query || haystack.includes(query);
    el.classList.toggle("filtered-out", !match);
  });
}

function updateServersFilterButtons() {
  const mode = serversFilterMode || "all";
  if (serversFilterSelect) {
    serversFilterSelect.value = mode;
  }
  if (serversFilterResetBtn) {
    const isAll = mode === "all";
    serversFilterResetBtn.disabled = isAll;
    serversFilterResetBtn.setAttribute("aria-label", "Clear filter");
    serversFilterResetBtn.setAttribute("data-tooltip", "Clear filter");
    const icon = serversFilterResetBtn.querySelector(".icon-action");
    if (icon) {
      icon.classList.toggle("icon-filter", isAll);
      icon.classList.toggle("icon-filter-off", !isAll);
    }
  }
}

function updateServersViewButtons() {
  if (!serversViewToggleBtn) return;
  const current = serversViewMode === "cards" ? "cards" : "table";
  const next = current === "cards" ? "table" : "cards";
  serversViewToggleBtn.dataset.mode = next;
  const label = next === "cards" ? "Switch to cards view" : "Switch to table view";
  serversViewToggleBtn.setAttribute("aria-label", label);
  serversViewToggleBtn.setAttribute("data-tooltip", label);
}

function setServersFilterMode(mode) {
  const next = ["all", "active", "maintenance", "offline"].includes(mode) ? mode : "all";
  serversFilterMode = next;
  updateServersFilterButtons();
  renderServers(cachedLocals, cachedRemotes);
}

function setServersViewMode(mode, persist = true) {
  serversViewMode = mode || "table";
  if (persist) {
    localStorage.setItem(serversViewStorageKey, serversViewMode);
  }
  updateServersViewButtons();
  renderServers(cachedLocals, cachedRemotes);
}

function isSelectiveScanEnabled() {
  return selectedScanScope === "selective";
}

function updateSelectiveScanToggle() {
  if (!statusSelectiveToggleBtn) return;
  const enabled = isSelectiveScanEnabled();
  const icon = statusSelectiveToggleBtn.querySelector(".icon-action");
  if (icon) {
    icon.classList.toggle("icon-filter", !enabled);
    icon.classList.toggle("icon-filter-off", enabled);
  }
  const label = enabled ? "Clear selective scan" : "Selective scan";
  statusSelectiveToggleBtn.setAttribute("aria-label", label);
  statusSelectiveToggleBtn.setAttribute("data-tooltip", label);
}

function setSelectiveScanEnabled(enabled) {
  if (enabled) {
    selectedScanScope = "selective";
  } else {
    selectedScanScope = "all";
    selectedScanServers.clear();
  }
  updateSelectiveScanToggle();
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
    themeToggleBtn.setAttribute("data-tooltip", `Switch to ${label} theme`);
  }
  if (themeLabel) {
    themeLabel.textContent = mode[0].toUpperCase() + mode.slice(1);
  }
}

function loadSchedulerAnchor() {
  const raw = localStorage.getItem(schedulerAnchorKey);
  if (!raw) return null;
  try {
    const parsed = JSON.parse(raw);
    if (!parsed || typeof parsed.anchorMs !== "number" || typeof parsed.intervalSec !== "number") {
      return null;
    }
    return parsed;
  } catch {
    return null;
  }
}

function saveSchedulerAnchor(anchorMs, intervalSec) {
  localStorage.setItem(schedulerAnchorKey, JSON.stringify({ anchorMs, intervalSec }));
}

function clearSchedulerAnchor() {
  localStorage.removeItem(schedulerAnchorKey);
}

function getSchedulerNextRun(cfg) {
  const enabled = Boolean(cfg && cfg.scheduler_enabled);
  if (!enabled) {
    return null;
  }
  const intervalSec = Number(cfg.scan_interval_sec);
  if (!Number.isFinite(intervalSec) || intervalSec <= 0) {
    return null;
  }
  const stored = loadSchedulerAnchor();
  if (!stored || stored.intervalSec !== intervalSec) {
    const anchorMs = Date.now();
    saveSchedulerAnchor(anchorMs, intervalSec);
    return new Date(anchorMs + intervalSec * 1000);
  }
  const intervalMs = intervalSec * 1000;
  const now = Date.now();
  const anchorMs = stored.anchorMs;
  if (now <= anchorMs) {
    return new Date(anchorMs + intervalMs);
  }
  const elapsed = now - anchorMs;
  const cycles = Math.floor(elapsed / intervalMs) + 1;
  return new Date(anchorMs + cycles * intervalMs);
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

// Status server scope dropdown removed; selective scan is toggled via icon button.

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
  const intervalSec = Number(cfg.scan_interval_sec);
  const enableChanged = lastSchedulerEnabled === null ? false : enabled !== lastSchedulerEnabled;
  const intervalChanged = lastSchedulerIntervalSec === null ? false : intervalSec !== lastSchedulerIntervalSec;
  if (!enabled) {
    clearSchedulerAnchor();
  } else if (enableChanged || intervalChanged) {
    saveSchedulerAnchor(Date.now(), intervalSec);
  } else {
    getSchedulerNextRun(cfg);
  }
  lastSchedulerEnabled = enabled;
  lastSchedulerIntervalSec = intervalSec;
  updateSchedulerNextRun(cfg);
    if (schedulerStateEl) {
      schedulerStateEl.dataset.enabled = enabled ? "true" : "false";
      const tooltip = enabled ? formatSchedulerNextRun(cfg) : "Scheduler disabled";
      schedulerStateEl.setAttribute("data-tooltip", tooltip);
      schedulerStateEl.setAttribute("aria-label", tooltip);
      schedulerStateEl.dataset.tooltipScheduler = "true";
    }
  }

function formatSchedulerNextRun(cfg) {
  const enabled = Boolean(cfg && cfg.scheduler_enabled);
  if (!enabled) {
    return "Next run: enable scheduler to check.";
  }
  const interval = Number(cfg.scan_interval_sec);
  if (!Number.isFinite(interval) || interval <= 0) {
    return "Next run: unavailable.";
  }
  const nextRun = getSchedulerNextRun(cfg);
  if (!nextRun) {
    return "Next run: unavailable.";
  }
  const timeLabel = nextRun.toLocaleString([], {
    year: "numeric",
    month: "short",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
  return `Next run: ${timeLabel}`;
}

function updateSchedulerNextRun(cfg) {
  if (!schedulerNextRunEl) return;
  schedulerNextRunEl.textContent = formatSchedulerNextRun(cfg);
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
  updateSchedulerNextRun(updated);
  return updated;
}

async function refreshServers() {
  const [locals, remotes] = await Promise.all([
    fetchJSON("/api/locals"),
    fetchJSON("/api/servers"),
  ]);
  cachedLocals = locals || [];
  cachedRemotes = remotes || [];
  if (serversRemoveConfirming.size > 0) {
    const valid = new Set([
      ...cachedLocals.map((server) => `local:${server.name}`),
      ...cachedRemotes.map((server) => `remote:${server.name}`),
    ]);
    serversRemoveConfirming = new Set(Array.from(serversRemoveConfirming).filter((key) => valid.has(key)));
  }
  renderServers(cachedLocals, cachedRemotes);
  updateServersFilterButtons();
  updateServersViewButtons();
  statusHintEl.classList.add("hidden");
}

async function triggerServersRefresh() {
  const info = await fetchJSON("/api/servers/refresh", { method: "POST" });
  mergeServerInfoSnapshot(info);
  renderServers(cachedLocals, cachedRemotes);
  if (currentView === "status") {
    renderStatus(lastStatusResults.length > 0 ? lastStatusResults : buildStatusPlaceholders());
  }
}

async function triggerStatusRefresh() {
  await fetchJSON("/api/status/refresh", { method: "POST" });
}

function applyStatusResults(results) {
  if (Array.isArray(results)) {
    results.sort((a, b) => {
      if (a.local === b.local) return 0;
      return a.local ? -1 : 1;
    });
  }
  lastStatusResults = Array.isArray(results) ? results : [];
  Object.entries(pendingSelfUpdates).forEach(([key, expiresAt]) => {
    if (!expiresAt || expiresAt <= Date.now()) {
      delete pendingSelfUpdates[key];
      return;
    }
    const sep = key.indexOf("::");
    if (sep < 0) return;
    const serverKey = key.slice(0, sep);
    const containerName = key.slice(sep + 2);
    if (!serverKey || !containerName) return;
    if (isSelfUpdateResolved(serverKey, containerName)) {
      clearPendingSelfUpdate(serverKey, containerName);
      delete restartingServers[serverKey];
    }
  });
  if (detailsModal && !detailsModal.classList.contains("hidden") && currentDetailsServerKey) {
    const target = lastStatusResults.find((entry) => {
      const key = serverScopeKey(entry);
      return key === currentDetailsServerKey;
    });
    if (target) {
      updateDetailsModal(target, Boolean(currentConfig && currentConfig.update_stopped_containers));
    }
  }
  if (lastStatusResults.length > 0) {
    const nextOverrides = { ...scanStateOverrides };
    lastStatusResults.forEach((result) => {
      const key = serverScopeKey(result);
      const state = String(result.scan_state || "").toLowerCase();
      if (state && state !== "idle") {
        delete nextOverrides[key];
      }
    });
    scanStateOverrides = nextOverrides;
  }
  updateScanPolling(results);
  renderStatus(results);
  return results;
}

async function refreshStatus() {
  const results = await fetchJSON("/api/aggregate");
  return applyStatusResults(results);
}

function upsertStatusResult(next) {
  if (!next) return;
  const key = serverScopeKey(next);
  const index = lastStatusResults.findIndex((entry) => serverScopeKey(entry) === key);
  if (index >= 0) {
    lastStatusResults[index] = next;
  } else {
    lastStatusResults.push(next);
  }
  applyStatusResults(lastStatusResults);
}

function initServerStream() {
  if (serverStream || typeof window.EventSource === "undefined") return;
  serverStream = new EventSource("/api/servers/stream");
  serverStream.addEventListener("server_info_snapshot", (event) => {
    try {
      const list = JSON.parse(event.data);
      mergeServerInfoSnapshot(list);
      renderServers(cachedLocals, cachedRemotes);
      if (currentView === "status") {
        renderStatus(lastStatusResults.length > 0 ? lastStatusResults : buildStatusPlaceholders());
      }
    } catch {
      return;
    }
  });
  serverStream.addEventListener("server_info", (event) => {
    try {
      const info = JSON.parse(event.data);
      if (!info || !info.name || !info.type) return;
      const key = `${info.type}:${info.name}`;
      applyCheckingTracking(key, info);
      const existing = cachedServerInfo[key];
      if (shouldReplaceServerInfo(existing, info)) {
        cachedServerInfo[key] = info;
      }
      renderServers(cachedLocals, cachedRemotes);
      if (currentView === "status") {
        renderStatus(lastStatusResults.length > 0 ? lastStatusResults : buildStatusPlaceholders());
      }
    } catch {
      return;
    }
  });
  serverStream.addEventListener("scan_snapshot", (event) => {
    try {
      const list = JSON.parse(event.data);
      if (!Array.isArray(list)) return;
      applyStatusResults(list);
    } catch {
      return;
    }
  });
  serverStream.addEventListener("scan_result", (event) => {
    try {
      const result = JSON.parse(event.data);
      upsertStatusResult(result);
    } catch {
      return;
    }
  });
}

async function refreshVersion() {
  if (!appVersionEl) return;
  try {
    const payload = await fetchJSON("/api/version");
    const raw = (payload && payload.version) ? String(payload.version) : "dev";
    const normalized = raw.startsWith("v") ? raw : `v${raw}`;
    const display = raw.startsWith("dev") ? `v${raw}` : normalized;
    appVersionEl.textContent = display;
  } catch (err) {
    appVersionEl.textContent = "vdev";
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
      const tsEl = document.createElement("span");
      tsEl.className = "log-ts";
      tsEl.textContent = ts;
      const levelEl = document.createElement("span");
      levelEl.className = levelClass;
      levelEl.textContent = level;
      const msgEl = document.createElement("span");
      msgEl.className = "log-msg";
      msgEl.textContent = entry.message;
      row.append(tsEl, levelEl, msgEl);
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
  logsAutoToggle.setAttribute("data-tooltip", `${state} logs auto refresh`);
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
  if (statusSelectiveToggleBtn) {
    statusSelectiveToggleBtn.disabled = isActive || updateInProgress;
  }
  scanBtn.classList.toggle("secondary", !isActive);
  scanBtn.classList.toggle("btn-warning", isActive);
  if (scanBtn) {
    const label = isActive ? "Stop scan" : "Check updates";
    scanBtn.setAttribute("aria-label", label);
    scanBtn.setAttribute("data-tooltip", label);
    const icon = scanBtn.querySelector(".icon-action");
    if (icon) {
      icon.classList.toggle("icon-stop", isActive);
      icon.classList.toggle("icon-refresh-dot", !isActive);
    }
  }
}

function updateScanPolling(results) {
  let active = Array.isArray(results)
    ? results.some((result) => ["pending", "scanning", "updating"].includes(String(result.scan_state || "")))
    : false;
  if (!active) {
    active = scanRequestActive || Boolean(currentScanController);
  }
  scanActive = active;
  setScanningUI(false);
  if (active && !scanPollingTimer) {
    scanPollingTimer = window.setInterval(async () => {
      await refreshStatus();
    }, 2000);
  } else if (!active && scanPollingTimer) {
    window.clearInterval(scanPollingTimer);
    scanPollingTimer = null;
    scanStateOverrides = {};
    currentScanStartedAtMs = null;
  }
}

function applyOptimisticScanState(scope) {
  const nextOverrides = {};
  currentScanStartedAtMs = Date.now();
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
  if (Array.isArray(scope)) {
    scope.forEach((key, index) => {
      nextOverrides[key] = index === 0 ? "scanning" : "pending";
    });
  } else if (scope && scope !== "all") {
    nextOverrides[scope] = "scanning";
  } else if (items.length > 0) {
    const onlineTargets = items
      .map((item) => serverScopeKey(item))
      .filter((key) => {
        const info = cachedServerInfo[key] || {};
        return String(info.status || "").toLowerCase() === "online";
      });
    onlineTargets.forEach((key, index) => {
      nextOverrides[key] = index === 0 ? "scanning" : "pending";
    });
  }
  scanStateOverrides = nextOverrides;
  if (lastStatusResults.length > 0) {
    renderStatus(lastStatusResults);
  }
}

function buildStatusPlaceholders() {
  return cachedLocals.map((server) => ({
    server_name: server.name,
    server_url: server.socket,
    local: true,
  })).concat(cachedRemotes.map((server) => ({
    server_name: server.name,
    server_url: server.url,
    local: false,
  })));
}

function renderStatusPlaceholders() {
  const placeholders = buildStatusPlaceholders();
  if (placeholders.length === 0) return;
  renderStatus(placeholders);
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

let tooltipTarget = null;

function showTooltip(target) {
  if (target && target.dataset && target.dataset.tooltipScheduler === "true") {
    const label = currentConfig && currentConfig.scheduler_enabled
      ? formatSchedulerNextRun(currentConfig)
      : "Scheduler disabled";
    target.setAttribute("data-tooltip", label);
    target.setAttribute("aria-label", label);
  }
  if (target && target.dataset && target.dataset.tooltipLastChecked === "true") {
    const label = formatLastChecked(target.dataset.checkedAt);
    if (label) {
      target.setAttribute("data-tooltip", label);
      target.setAttribute("aria-label", label);
    }
  }
  if (target && target.dataset && target.dataset.tooltipServerStatus === "true") {
    const statusLabel = String(target.dataset.statusLabel || "").trim();
    const lastChecked = formatLastChecked(target.dataset.checkedAt);
    const label = lastChecked ? `Status: ${statusLabel} · ${lastChecked}` : `Status: ${statusLabel}`;
    target.setAttribute("data-tooltip", label);
    target.setAttribute("aria-label", label);
  }
  const text = target?.dataset?.tooltip || target?.getAttribute("aria-label") || "";
  if (!text) return;
  tooltipEl.textContent = text;
  tooltipEl.classList.add("visible");

  const rect = target.getBoundingClientRect();
  const margin = 8;
  tooltipEl.style.left = "0px";
  tooltipEl.style.top = "0px";
  const tooltipRect = tooltipEl.getBoundingClientRect();

  let left = rect.left + rect.width / 2 - tooltipRect.width / 2;
  left = Math.max(margin, Math.min(left, window.innerWidth - tooltipRect.width - margin));

  let top = rect.top - tooltipRect.height - 10;
  let placement = "top";
  if (top < margin) {
    top = rect.bottom + 10;
    placement = "bottom";
  }
  if (top + tooltipRect.height > window.innerHeight - margin) {
    top = Math.max(margin, window.innerHeight - tooltipRect.height - margin);
  }

  tooltipEl.style.left = `${left}px`;
  tooltipEl.style.top = `${top}px`;
  tooltipEl.setAttribute("data-placement", placement);
}

function hideTooltip() {
  tooltipEl.classList.remove("visible");
  tooltipTarget = null;
}

document.addEventListener("mouseover", (event) => {
  const target = event.target.closest?.(".has-tooltip");
  if (!target) return;
  if (tooltipTarget === target) return;
  tooltipTarget = target;
  showTooltip(target);
});

document.addEventListener("mouseout", (event) => {
  if (!tooltipTarget) return;
  if (event.relatedTarget && tooltipTarget.contains(event.relatedTarget)) return;
  hideTooltip();
});

window.addEventListener("scroll", () => {
  if (tooltipTarget) {
    showTooltip(tooltipTarget);
  }
}, { passive: true });

window.addEventListener("resize", () => {
  if (tooltipTarget) {
    showTooltip(tooltipTarget);
  }
  updateTopbarHeight();
});

function refreshViewData(view) {
  const tasks = [];
  tasks.push(refreshConfig().catch(() => {}));
  if (view === "logs") {
    tasks.push(refreshLogs().catch(() => {}));
  }
  void Promise.all(tasks);
}

function updateTopbarHeight() {
  if (!topbarEl) return;
  const height = topbarEl.getBoundingClientRect().height;
  document.documentElement.style.setProperty("--topbar-h", `${Math.ceil(height)}px`);
}

function setView(nextView) {
  currentView = nextView;
  viewStatusEl.classList.toggle("hidden", nextView !== "status");
  viewSettingsEl.classList.toggle("hidden", nextView !== "settings");
  viewServersEl.classList.toggle("hidden", nextView !== "servers");
  viewLogsEl.classList.toggle("hidden", nextView !== "logs");
  if (topbarStatusEl) {
    topbarStatusEl.classList.toggle("hidden", nextView !== "status");
  }
  if (topbarSettingsEl) {
    topbarSettingsEl.classList.toggle("hidden", nextView !== "settings");
  }
  if (topbarServersEl) {
    topbarServersEl.classList.toggle("hidden", nextView !== "servers");
  }
  if (topbarLogsEl) {
    topbarLogsEl.classList.toggle("hidden", nextView !== "logs");
  }
  updateTopbarHeight();

  sidebar.querySelectorAll("[data-view]").forEach((btn) => {
    btn.classList.toggle("active", btn.getAttribute("data-view") === nextView);
  });

  if (nextView === "logs") {
    startLogsAutoRefresh();
  } else {
    stopLogsAutoRefresh();
  }
  if (nextView === "servers" && sidebarSearch) {
    serversSearchQuery = normalizeQuery(sidebarSearch.value);
  }
  if (nextView === "status" && lastStatusResults.length === 0) {
    renderStatusPlaceholders();
  }
  if (nextView === "servers") {
    setServersFilterMode("all");
    refreshServers()
      .then(() => triggerServersRefresh())
      .catch(() => {});
  }
  if (nextView === "status") {
    refreshServers()
      .then(() => triggerServersRefresh())
      .then(() => triggerStatusRefresh())
      .then(() => refreshStatus())
      .catch(() => {});
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
    currentScanStartedAtMs = null;
    await refreshStatus();
    return;
  }
  const wasSelectiveScan = isSelectiveScanEnabled();
  currentScanController = new AbortController();
  if (wasSelectiveScan && selectedScanServers.size === 0) {
    showToast("Select at least one server to scan.");
    currentScanController = null;
    scanActive = false;
    setScanningUI(false);
    return;
  }
  scanRequestActive = true;
  setScanningUI(true);
  startScanPolling();
  const scanTargets = wasSelectiveScan
    ? Array.from(selectedScanServers)
    : "all";
  applyOptimisticScanState(scanTargets);
  refreshStatus().catch(() => {});
  try {
    if (wasSelectiveScan) {
      for (const scope of scanTargets) {
        const url = `/api/scan?server=${encodeURIComponent(scope)}`;
        await fetchJSON(url, { method: "POST", signal: currentScanController.signal });
      }
    } else {
      await fetchJSON("/api/scan?server=all", { method: "POST", signal: currentScanController.signal });
    }
  } catch (err) {
    if (err.name === "AbortError" || /aborted|canceled|cancelled/i.test(err.message)) {
      alert("Scan cancelled.");
    } else {
      alert(err.message);
    }
  } finally {
    currentScanController = null;
    scanRequestActive = false;
    setScanningUI(false);
    if (wasSelectiveScan) {
      setSelectiveScanEnabled(false);
    }
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
      await refreshConfig();
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
    await refreshConfig();
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

if (localModalForm) {
  localModalForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    const payload = {
      name: localModalNameInput ? localModalNameInput.value.trim() : "",
      socket: localModalSocketInput ? localModalSocketInput.value.trim() : "/var/run/docker.sock",
    };
    if (editingLocalServer && editingLocalServer.maintenance) {
      payload.maintenance = true;
    }
    try {
      await fetchJSON("/api/locals", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      closeRemoteModal();
      await refreshServers();
      await refreshStatus();
    } catch (err) {
      alert(err.message);
    }
  });
}

if (localModalCancel) {
  localModalCancel.addEventListener("click", () => {
    closeRemoteModal();
  });
}

if (remoteModalSave) {
  remoteModalSave.addEventListener("click", async () => {
    const payload = {
      name: remoteModalNameInput ? remoteModalNameInput.value.trim() : "",
      url: normalizeRemoteUrl(
        remoteModalHostInput ? remoteModalHostInput.value.trim() : "",
        remoteModalPortInput ? remoteModalPortInput.value.trim() : "8080"
      ),
      token: remoteModalTokenInput ? remoteModalTokenInput.value.trim() : "",
    };
    if (editingRemoteServer && editingRemoteServer.maintenance) {
      payload.maintenance = true;
    }
    try {
      await fetchJSON("/api/servers", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      closeRemoteModal();
      await refreshServers();
      await refreshStatus();
    } catch (err) {
      alert(err.message);
    }
  });
}

if (remoteModalCancel) {
  remoteModalCancel.addEventListener("click", () => {
    closeRemoteModal();
  });
}

if (remoteModalClose) {
  remoteModalClose.addEventListener("click", () => {
    closeRemoteModal();
  });
}

if (remoteModalTokenCopy) {
  remoteModalTokenCopy.addEventListener("click", () => {
    copyToClipboard(remoteModalTokenInput ? remoteModalTokenInput.value : "", "Token");
  });
}

if (remoteModalComposeCopy) {
  remoteModalComposeCopy.addEventListener("click", () => {
    copyToClipboard(remoteModalCompose ? remoteModalCompose.textContent : "", "Compose");
  });
}

if (remoteModalPortInput) {
  remoteModalPortInput.addEventListener("input", () => {
    updateRemoteComposePreview();
  });
}

if (serverTabLocal) {
  serverTabLocal.addEventListener("click", () => {
    setServerModalTab("local");
    if (remoteModalTitle) {
      remoteModalTitle.textContent = editingLocalServer ? "Edit local server" : "Add local server";
    }
    if (!editingLocalServer && localModalSocketInput && !localModalSocketInput.value.trim()) {
      localModalSocketInput.value = "/var/run/docker.sock";
    }
    if (localModalNameInput) {
      localModalNameInput.focus();
    }
  });
}

if (serverTabRemote) {
  serverTabRemote.addEventListener("click", () => {
    setServerModalTab("remote");
    if (remoteModalTitle) {
      remoteModalTitle.textContent = editingRemoteServer ? "Edit remote server" : "Add remote server";
    }
    if (!editingRemoteServer && remoteModalTokenInput && !remoteModalTokenInput.value.trim()) {
      remoteModalTokenInput.value = generateToken(32);
      updateRemoteComposePreview();
    }
    if (remoteModalNameInput) {
      remoteModalNameInput.focus();
    }
  });
}

async function init() {
  sidebar.setAttribute("aria-hidden", "false");

  if (sidebarSearch) {
    sidebarSearch.addEventListener("input", (event) => {
      applySidebarFilter(event.target.value);
    });
  }

  if (statusSelectiveToggleBtn) {
    statusSelectiveToggleBtn.addEventListener("click", () => {
      setSelectiveScanEnabled(!isSelectiveScanEnabled());
      refreshStatus();
    });
    updateSelectiveScanToggle();
  }

  if (serversFilterSelect) {
    serversFilterSelect.addEventListener("change", () => {
      setServersFilterMode(serversFilterSelect.value);
    });
  }
  if (serversFilterResetBtn) {
    serversFilterResetBtn.addEventListener("click", () => {
      setServersFilterMode("all");
      if (serversFilterSelect) {
        serversFilterSelect.focus();
      }
    });
  }

  if (serversViewToggleBtn) {
    serversViewToggleBtn.addEventListener("click", () => {
      const next = serversViewMode === "cards" ? "table" : "cards";
      setServersViewMode(next);
    });
  }

  if (addServerBtn) {
    addServerBtn.addEventListener("click", () => {
      openRemoteModal("add");
    });
  }

  if (refreshStatusBtn) {
    refreshStatusBtn.addEventListener("click", async () => {
      const wasSelectiveScan = isSelectiveScanEnabled();
      showToast("Checking connection…", 2000);
      try {
        await refreshServers();
        await triggerServersRefresh();
        await triggerStatusRefresh();
        await refreshStatus();
        showToast("Connection checked", 2000);
      } finally {
        if (wasSelectiveScan) {
          setSelectiveScanEnabled(false);
        }
      }
    });
  }

  if (refreshServersBtn) {
    refreshServersBtn.addEventListener("click", async () => {
      showToast("Checking connection…", 2000);
      await refreshServers();
      await triggerServersRefresh();
      showToast("Connection checked", 2000);
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
  if (remoteModal) {
    remoteModal.addEventListener("click", (event) => {
      if (event.target && event.target.dataset && event.target.dataset.close) {
        closeRemoteModal();
      }
    });
  }
  if (globalPolicyInfoBtn) {
    globalPolicyInfoBtn.addEventListener("click", openPolicyModal);
  }
  if (policyCloseBtn) {
    policyCloseBtn.addEventListener("click", closePolicyModal);
  }
  if (policyModal) {
    policyModal.addEventListener("click", (event) => {
      if (event.target && event.target.dataset && event.target.dataset.close) {
        closePolicyModal();
      }
    });
  }
  document.addEventListener("keydown", (event) => {
    if (event.key === "Escape" && detailsModal && !detailsModal.classList.contains("hidden")) {
      closeDetailsModal();
    }
    if (event.key === "Escape" && policyModal && !policyModal.classList.contains("hidden")) {
      closePolicyModal();
    }
    if (event.key === "Escape" && remoteModal && !remoteModal.classList.contains("hidden")) {
      closeRemoteModal();
    }
  });
  document.addEventListener(
    "click",
    (event) => {
      if (serversRemoveConfirming.size === 0) return;
      const target = event.target;
      if (target && target.closest && target.closest(".servers-confirm-btn")) return;

      serversRemoveConfirming.clear();
      document.querySelectorAll(".servers-confirm-btn:not(.hidden)").forEach((btn) => {
        btn.classList.add("hidden");
        const actions = btn.closest(".servers-row-actions");
        const removeBtn = actions ? actions.querySelector(".servers-remove-btn") : null;
        if (removeBtn) removeBtn.classList.remove("hidden");
      });
    },
    true
  );

  initThemeToggle();
  updateTopbarHeight();
  const storedServersView = localStorage.getItem(serversViewStorageKey);
  if (storedServersView === "cards" || storedServersView === "table") {
    serversViewMode = storedServersView;
  }
  setView("status");
  setScanningUI(false);
  await refreshConfig();
  await refreshServers();
  initServerStream();
  await triggerServersRefresh();
  triggerStatusRefresh().catch(() => {});
  renderStatusPlaceholders();
  refreshStatus().catch(() => {});
  await refreshVersion();
  statusHintEl.classList.toggle("hidden", cachedLocals.length > 0);
}

init().catch((err) => {
  statusEl.textContent = `Error: ${err.message}`;
});
