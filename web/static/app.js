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
const expContainersInput = document.getElementById("exp-containers");
const expContainersSidebarInput = document.getElementById("exp-containers-sidebar");
const expContainerShellInput = document.getElementById("exp-container-shell");
const expContainerLogsInput = document.getElementById("exp-container-logs");
const expContainerResourcesInput = document.getElementById("exp-container-resources");
const expStacksInput = document.getElementById("exp-stacks");
const expImagesInput = document.getElementById("exp-images");
const addServerBtn = document.getElementById("add-server");
const refreshServersBtn = document.getElementById("refresh-servers");
const serversTableWrap = document.getElementById("servers-table-wrap");
const serversTableBody = document.getElementById("servers-table-body");
const serversCardsEl = document.getElementById("servers-cards");
const serversViewToggleBtn = document.getElementById("servers-view-toggle");
const serversFilterResetBtn = document.getElementById("servers-filter-reset");
const serversFilterBtn = document.getElementById("servers-filter-btn");
const serversFilterMenu = document.getElementById("servers-filter-menu");
const remoteModal = document.getElementById("remote-modal");
const remoteModalTitle = document.getElementById("remote-modal-title");
const remoteModalClose = document.getElementById("remote-modal-close");
const serverPanelRemote = document.getElementById("server-panel-remote");
const remoteModalNameInput = document.getElementById("remote-modal-name");
const remoteModalHostInput = document.getElementById("remote-modal-host");
const serverModalHostLabel = document.getElementById("server-modal-host-label");
const serverConnectionTypeSelect = document.getElementById("server-connection-type");
const serverConnectionTypeHelpBtn = document.getElementById("server-connection-type-help");
const serverRemotePortRow = document.getElementById("server-remote-port-row");
const serverTokenRow = document.getElementById("server-token-row");
const serverTokenHint = document.getElementById("server-token-hint");
const serverComposePanel = document.getElementById("server-compose-panel");
const remoteModalPublicIPInput = document.getElementById("remote-modal-public-ip");
const remoteModalPortInput = document.getElementById("remote-modal-port");
const remoteModalTokenInput = document.getElementById("remote-modal-token");
const remoteModalTokenCopy = document.getElementById("remote-modal-token-copy");
const remoteModalCompose = document.getElementById("remote-modal-compose");
const remoteModalComposeCopy = document.getElementById("remote-modal-compose-copy");
const remoteModalSave = document.getElementById("remote-modal-save");
const remoteModalCancel = document.getElementById("remote-modal-cancel");
const sidebar = document.getElementById("sidebar");
const sidebarSearch = document.getElementById("sidebar-search");
const sidebarSearchCountEl = document.getElementById("sidebar-search-count");
const sidebarCollapseToggleBtn = document.getElementById("sidebar-collapse-toggle");
const themeToggleBtn = document.getElementById("theme-toggle");
const themeLabel = document.getElementById("theme-label");
const topbarEl = document.querySelector(".topbar");
const topbarStatusEl = document.getElementById("topbar-status");
const topbarSettingsEl = document.getElementById("topbar-settings");
const topbarContainersEl = document.getElementById("topbar-containers");
const topbarServersEl = document.getElementById("topbar-servers");
const topbarLogsEl = document.getElementById("topbar-logs");
const viewStatusEl = document.getElementById("view-status");
const viewSettingsEl = document.getElementById("view-settings");
const viewContainersEl = document.getElementById("view-containers");
const topbarContainersStacksBtn = document.getElementById("topbar-containers-stacks");
const topbarContainersImagesBtn = document.getElementById("topbar-containers-images");
const viewServersEl = document.getElementById("view-servers");
const viewLogsEl = document.getElementById("view-logs");
const topbarContainersServerBtn = document.getElementById("topbar-containers-server-btn");
const topbarContainersServerMenu = document.getElementById("topbar-containers-server-menu");
const topbarContainersRefreshBtn = document.getElementById("topbar-containers-refresh");
const topbarContainersShellBtn = document.getElementById("topbar-containers-shell");
const topbarContainersLogsBtn = document.getElementById("topbar-containers-logs");
const topbarContainersResourcesBtn = document.getElementById("topbar-containers-resources");
const topbarContainersStatusEl = document.getElementById("topbar-containers-status");
const containersNameSortBtn = document.getElementById("containers-name-sort");
const containersStateSortBtn = document.getElementById("containers-state-sort");
const containersCpuSortBtn = document.getElementById("containers-cpu-sort");
const containersRamSortBtn = document.getElementById("containers-ram-sort");
const containersIpSortBtn = document.getElementById("containers-ip-sort");
const containersPortSortBtn = document.getElementById("containers-port-sort");
const containersTableBody = document.getElementById("containers-body");
const containersEmptyEl = document.getElementById("containers-empty");
const containersTableView = document.getElementById("containers-table-view");
const containersStacksView = document.getElementById("containers-stacks-view");
const containersImagesView = document.getElementById("containers-images-view");
const stacksTableBody = document.getElementById("stacks-body");
const stacksEmptyEl = document.getElementById("stacks-empty");
const stacksNewBtn = document.getElementById("stacks-new");
const stacksNameSortBtn = document.getElementById("stacks-name-sort");
const stacksStateSortBtn = document.getElementById("stacks-state-sort");
const imagesTableBody = document.getElementById("images-body");
const imagesEmptyEl = document.getElementById("images-empty");
const imagesRepoSortBtn = document.getElementById("images-repo-sort");
const imagesTagSortBtn = document.getElementById("images-tag-sort");
const imagesStateSortBtn = document.getElementById("images-state-sort");
const imagesSizeSortBtn = document.getElementById("images-size-sort");
const imagesPullBtn = document.getElementById("images-pull");
const imagesPruneUnusedBtn = document.getElementById("images-prune-unused");
const imagesPruneDanglingBtn = document.getElementById("images-prune-dangling");
const imagesRemoveBtn = document.getElementById("images-remove");
const imagesTotalCountEl = document.getElementById("images-total-count");
const imagesTotalSizeEl = document.getElementById("images-total-size");
const imagesPullModal = document.getElementById("images-pull-modal");
const imagesPullCloseBtn = document.getElementById("images-pull-close");
const imagesPullCancelBtn = document.getElementById("images-pull-cancel");
const imagesPullConfirmBtn = document.getElementById("images-pull-confirm");
const imagesPullRepoInput = document.getElementById("images-pull-repo");
const imagesPullTagInput = document.getElementById("images-pull-tag");
const imagesPullStatusEl = document.getElementById("images-pull-status");
const containersShellLayout = document.getElementById("containers-shell-layout");
const containersShellList = document.getElementById("containers-shell-list");
const containersShellPlaceholder = document.getElementById("containers-shell-placeholder");
const containersTerminalEl = document.getElementById("containers-terminal");
const containersShellTerminalEl = document.getElementById("containers-shell-terminal");
const containersLogsLayout = document.getElementById("containers-logs-layout");
const containersLogsList = document.getElementById("containers-logs-list");
const containersLogsPlaceholder = document.getElementById("containers-logs-placeholder");
const containersLogsView = document.getElementById("containers-logs-view");
const containersLogsOutput = document.getElementById("containers-logs-output");
const containersLogsStreamEl = document.getElementById("containers-logs-stream");
const containersLogsPauseBtn = document.getElementById("containers-logs-pause");
const containersLogsTimestampsBtn = document.getElementById("containers-logs-timestamps");
const containersLogsClearBtn = document.getElementById("containers-logs-clear");
const containersResourcesLayout = document.getElementById("containers-resources-layout");
const containersResourcesList = document.getElementById("containers-resources-list");
const containersResourcesPlaceholder = document.getElementById("containers-resources-placeholder");
const containersResourcesCards = document.getElementById("containers-resources-cards");
const containersResourcesTableWrap = document.getElementById("containers-resources-table-wrap");
const containersResourcesTableBody = document.getElementById("containers-resources-body");
const containersResourcesViewToggleBtn = document.getElementById("containers-resources-view-toggle");
const resourcesNameSortBtn = document.getElementById("resources-name-sort");
const resourcesCpuSortBtn = document.getElementById("resources-cpu-sort");
const resourcesRamSortBtn = document.getElementById("resources-ram-sort");
const resourcesStatusSortBtn = document.getElementById("resources-status-sort");
const resourcesUptimeSortBtn = document.getElementById("resources-uptime-sort");
const resourcesIpSortBtn = document.getElementById("resources-ip-sort");
const resourcesPortSortBtn = document.getElementById("resources-port-sort");
const refreshLogsBtn = document.getElementById("refresh-logs");
const clearLogsBtn = document.getElementById("clear-logs");
const logsLevelBtn = document.getElementById("logs-level-btn");
const logsLevelMenu = document.getElementById("logs-level-menu");
const logsAutoToggle = document.getElementById("logs-auto");
const logsListEl = document.getElementById("logs-list");
const appVersionEl = document.getElementById("app-version");
const appVersionUpdateEl = document.getElementById("app-version-update");
const aboutVersionLinkEl = document.getElementById("about-version-link");
const aboutUpdateStatusEl = document.getElementById("about-update-status");
const aboutUpdateLinkEl = document.getElementById("about-update-link");
const aboutChannelEl = document.getElementById("about-channel");
const aboutRepoLinkEl = document.getElementById("about-repo-link");
const aboutReleaseTagEl = document.getElementById("about-release-tag");
const testWebhookBtn = document.getElementById("test-webhook");
const saveIntervalBtn = document.getElementById("save-interval");
const detailsModal = document.getElementById("details-modal");
const detailsTitle = document.getElementById("details-title");
const detailsBody = document.getElementById("details-body");
const detailsCloseBtn = document.getElementById("details-close");
const policyModal = document.getElementById("policy-modal");
const policyCloseBtn = document.getElementById("policy-close");
const stackModal = document.getElementById("stack-modal");
const stackModalTitle = document.getElementById("stack-modal-title");
const stackModalClose = document.getElementById("stack-modal-close");
const stackModalSave = document.getElementById("stack-modal-save");
const stackModalSaveIcon = document.getElementById("stack-modal-save-icon");
const stackModalCancel = document.getElementById("stack-modal-cancel");
const stackModalComposeUp = document.getElementById("stack-modal-compose-up");
const stackModalComposeDown = document.getElementById("stack-modal-compose-down");
const stackNameInput = document.getElementById("stack-name-input");
const stackNameRow = document.getElementById("stack-name-row");
const stackComposeEditorEl = document.getElementById("stack-compose-editor");
const stackComposeInput = document.getElementById("stack-compose-input");
const stackEnvEditorEl = document.getElementById("stack-env-editor");
const stackEnvInput = document.getElementById("stack-env-input");
const stackEnvToggle = document.getElementById("stack-env-toggle");
const stackEnvToggleBtn = document.getElementById("stack-env-toggle-btn");
const stackModalErrorEl = document.getElementById("stack-modal-error");

let currentScanController = null;
let currentView = "status";
let updateInProgress = false;
let logsRefreshTimer = null;
let releaseCheckTimer = null;
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
let serverModalTab = "remote";
const schedulerAnchorKey = "contiwatch_scheduler_anchor";
const serversViewStorageKey = "contiwatch_servers_view";
let serversViewMode = "table";
const containersResourcesViewStorageKey = "contiwatch_container_resources_view";
let containersResourcesViewMode = "cards";
const sidebarCollapsedStorageKey = "contiwatch_sidebar_collapsed";
const releaseCheckPollMs = 1000 * 60 * 60 * 6;
const defaultReleaseRepo = "pbuzdygan/contiwatch";
let lastSchedulerEnabled = null;
let lastSchedulerIntervalSec = null;
let serverStream = null;
let checkingStartMsByServerKey = {};
let checkingDebounceTimersByServerKey = {};
let pendingSelfUpdates = {};
let pendingSelfUpdateTimers = {};
let logsLevelMode = "all";
let containersSelectedScope = "";
let containersSortMode = "name:asc";
let stacksSortMode = "name:asc";
let imagesSortMode = "repository:asc";
let containersRefreshTimer = null;
let containersUpdateInProgress = false;
let containersTableResourcesData = new Map();
let containersTableResourcesUpdateInProgress = false;
let containersTableResourcesRequestId = 0;
let containersTableResourcesScope = "";
let stacksRefreshTimer = null;
let stacksUpdateInProgress = false;
let imagesUpdateInProgress = false;
let containersCache = new Map();
let containersKillConfirming = new Set();
let stacksActionConfirming = new Set();
let stackStatusOverrides = new Map();
let imagesCache = [];
let imagesSelectedId = "";
let imagesExpandedRepos = new Set();
let imagesActionConfirming = new Set();
let imagesPullInProgress = false;
let imagesPullAutoCloseTimer = null;
let imagesPullAutoCloseRemainingSec = 0;
let imagesPullLastRef = "";
let containersViewMode = "table";
let pendingContainersMode = "";
let containersShellSelectedId = "";
let containersShellSocket = null;
let containersShellTerm = null;
let containersShellFitAddon = null;
let containersShellDataListener = null;
const containersShellEncoder = new TextEncoder();
let containersShellResizeTimer = null;
let containersShellResizeHandler = null;
let containersLogsSelectedId = "";
let containersLogsSocket = null;
let containersLogsBuffer = "";
let containersLogsPending = "";
let containersLogsLineBuffer = "";
let containersLogsBytes = 0;
let containersLogsAutoScroll = true;
let containersLogsPaused = false;
let containersResourcesSelectedIds = new Set();
let containersResourcesData = new Map();
let containersResourcesRefreshTimer = null;
let containersResourcesUpdateInProgress = false;
let containersResourcesSortMode = "name:asc";
const containersResourcesPinsKey = "contiwatch_container_resources_pins:v1";
let containersResourcesPins = null;
let containersResourcesLastInteractionAtMs = 0;
let containersLogsTimestamps = false;
const containersLogsMaxBytes = 300000;
let containersLogsSilentClose = false;
let containersLogsStreamActive = false;
let currentTimeZone = "";
let stacksCache = new Map();
let editingStackName = "";
let composeEditor = null;
let envEditor = null;

function showStackModalError(message) {
  if (!stackModalErrorEl) return;
  const text = String(message || "").trim();
  if (!text) {
    stackModalErrorEl.textContent = "";
    stackModalErrorEl.classList.add("hidden");
    return;
  }
  stackModalErrorEl.textContent = text;
  stackModalErrorEl.classList.remove("hidden");
}

function clearStackModalError() {
  showStackModalError("");
}

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

function applyServerModalTabState(tab) {
  const mode = tab === "local" ? "local" : "remote";
  serverModalTab = mode;
  const isLocal = mode === "local";
  if (serverConnectionTypeSelect) {
    serverConnectionTypeSelect.value = isLocal ? "local" : "remote";
  }
  if (serverModalHostLabel) {
    serverModalHostLabel.textContent = isLocal ? "Socket path" : "Host / IP";
  }
  if (remoteModalHostInput) {
    remoteModalHostInput.placeholder = isLocal ? "/var/run/docker.sock" : "10.0.0.10";
  }
  if (serverRemotePortRow) {
    serverRemotePortRow.classList.toggle("hidden", isLocal);
  }
  if (remoteModalPortInput) {
    remoteModalPortInput.disabled = isLocal;
  }
  if (serverTokenRow) {
    serverTokenRow.classList.toggle("hidden", isLocal);
  }
  if (serverTokenHint) {
    serverTokenHint.classList.toggle("hidden", isLocal);
  }
  if (remoteModalTokenInput) {
    remoteModalTokenInput.disabled = isLocal;
  }
  if (remoteModalTokenCopy) {
    remoteModalTokenCopy.disabled = isLocal;
  }
  if (remoteModalComposeCopy) {
    remoteModalComposeCopy.disabled = isLocal;
  }
  if (serverComposePanel) {
    serverComposePanel.classList.toggle("is-disabled", isLocal);
  }
  if (remoteModalCompose) {
    remoteModalCompose.classList.toggle("is-disabled", isLocal);
    remoteModalCompose.setAttribute("aria-disabled", isLocal ? "true" : "false");
  }
  if (!isLocal && remoteModalTokenInput && !remoteModalTokenInput.value.trim()) {
    remoteModalTokenInput.value = generateToken(32);
    updateRemoteComposePreview();
  }
  if (isLocal) {
    if (remoteModalTokenInput) remoteModalTokenInput.value = "";
    if (remoteModalCompose) remoteModalCompose.textContent = "";
  }
}

function setServerModalTab(tab) {
  const next = tab === "local" ? "local" : "remote";
  if (serverPanelRemote) {
    serverPanelRemote.classList.remove("hidden");
  }
  applyServerModalTabState(next);
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
  if (serverConnectionTypeSelect) {
    serverConnectionTypeSelect.disabled = Boolean(isEdit);
  }
  if (remoteModalNameInput) {
    remoteModalNameInput.value = server ? server.name || "" : "";
  }
  if (remoteModalHostInput) {
    remoteModalHostInput.value = server ? server.socket || "/var/run/docker.sock" : "/var/run/docker.sock";
  }
  if (remoteModalPublicIPInput) {
    remoteModalPublicIPInput.value = server ? server.public_ip || "" : "";
  }
  if (remoteModalPortInput) {
    remoteModalPortInput.value = "8080";
  }
  if (remoteModalTokenInput) {
    remoteModalTokenInput.value = "";
  }
  if (remoteModalCompose) {
    remoteModalCompose.textContent = "";
  }
  remoteModal.classList.remove("hidden");
  remoteModal.setAttribute("aria-hidden", "false");
  if (remoteModalNameInput) {
    remoteModalNameInput.focus();
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
  if (serverConnectionTypeSelect) {
    serverConnectionTypeSelect.disabled = Boolean(isEdit);
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
  if (remoteModalPublicIPInput) {
    remoteModalPublicIPInput.value = server ? server.public_ip || "" : "";
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

function isValidStackName(name) {
  return /^[A-Za-z0-9_-]+$/.test(String(name || "").trim());
}

function enableYamlIndent(textarea) {
  if (!textarea) return;
  textarea.addEventListener("keydown", (event) => {
    if (event.key === "Enter") {
      const start = textarea.selectionStart;
      const value = textarea.value;
      const lineStart = value.lastIndexOf("\n", start - 1) + 1;
      const line = value.slice(lineStart, start);
      const indentMatch = line.match(/^\s+/);
      const indent = indentMatch ? indentMatch[0] : "";
      event.preventDefault();
      const insert = `\n${indent}`;
      textarea.value = value.slice(0, start) + insert + value.slice(start);
      const nextPos = start + insert.length;
      textarea.selectionStart = nextPos;
      textarea.selectionEnd = nextPos;
      return;
    }
    if (event.key !== "Tab") return;
    event.preventDefault();
    const start = textarea.selectionStart;
    const end = textarea.selectionEnd;
    const value = textarea.value;
    const indent = "  ";
    if (event.shiftKey) {
      const before = value.slice(0, start);
      const lineStart = before.lastIndexOf("\n") + 1;
      const line = value.slice(lineStart, start);
      if (line.startsWith(indent)) {
        textarea.value = value.slice(0, lineStart) + line.slice(indent.length) + value.slice(start);
        textarea.selectionStart = start - indent.length;
        textarea.selectionEnd = end - indent.length;
      }
      return;
    }
    textarea.value = value.slice(0, start) + indent + value.slice(end);
    textarea.selectionStart = start + indent.length;
    textarea.selectionEnd = start + indent.length;
  });
}

function getComposeValue() {
  if (composeEditor && typeof composeEditor.getValue === "function") {
    return composeEditor.getValue();
  }
  return stackComposeInput ? stackComposeInput.value : "";
}

function setComposeValue(value) {
  if (composeEditor && typeof composeEditor.setValue === "function") {
    composeEditor.setValue(value || "");
    return;
  }
  if (stackComposeInput) {
    stackComposeInput.value = value || "";
  }
}

function getEnvValue() {
  return stackEnvInput ? stackEnvInput.value : "";
}

function setEnvValue(value) {
  if (envEditor && typeof envEditor.setValue === "function") {
    envEditor.setValue(value || "");
    return;
  }
  if (stackEnvInput) {
    stackEnvInput.value = value || "";
  }
}

function initComposeEditor() {
  if (!stackComposeEditorEl || !stackComposeInput || !window.CM6Compose || typeof window.CM6Compose.init !== "function") {
    if (stackComposeInput) stackComposeInput.classList.remove("hidden");
    if (stackComposeEditorEl) stackComposeEditorEl.classList.add("hidden");
    return;
  }
  stackComposeInput.classList.add("hidden");
  stackComposeEditorEl.classList.remove("hidden");
  composeEditor = window.CM6Compose.init({
    container: stackComposeEditorEl,
    textarea: stackComposeInput,
    getEnv: () => (stackEnvInput ? stackEnvInput.value : ""),
    getUseEnv: () => Boolean(stackEnvToggle && stackEnvToggle.checked),
    onChange: () => clearStackModalError(),
    mode: "yaml",
    lint: true,
  });
}

function initEnvEditor() {
  if (!stackEnvEditorEl || !stackEnvInput || !window.CM6Compose || typeof window.CM6Compose.init !== "function") {
    if (stackEnvInput) stackEnvInput.classList.remove("hidden");
    if (stackEnvEditorEl) stackEnvEditorEl.classList.add("hidden");
    return;
  }
  stackEnvInput.classList.add("hidden");
  stackEnvEditorEl.classList.remove("hidden");
  envEditor = window.CM6Compose.init({
    container: stackEnvEditorEl,
    textarea: stackEnvInput,
    onChange: () => {
      clearStackModalError();
      if (composeEditor && typeof composeEditor.forceLint === "function") {
        composeEditor.forceLint();
      }
    },
    mode: "env",
    lint: true,
  });
}

const envFileItemRegex = /^-\s*["']?\.env["']?\s*$/;

function countIndent(line) {
  const match = String(line || "").match(/^\s*/);
  return match ? match[0].length : 0;
}

function isCommentOrBlank(line) {
  const trimmed = String(line || "").trim();
  return trimmed === "" || trimmed.startsWith("#");
}

function isServiceHeader(line, indent, servicesIndent) {
  const trimmed = String(line || "").trim();
  if (!trimmed || trimmed.startsWith("#")) return false;
  if (!trimmed.endsWith(":")) return false;
  if (trimmed.startsWith("-")) return false;
  return indent > servicesIndent;
}

function findServicesBlock(lines) {
  for (let i = 0; i < lines.length; i += 1) {
    const line = String(lines[i] || "");
    if (line.trim() === "services:") {
      return { index: i, indent: countIndent(line) };
    }
  }
  return { index: -1, indent: 0 };
}

function ensureEnvFileForServices(text) {
  const lines = String(text || "").split(/\r?\n/);
  const servicesInfo = findServicesBlock(lines);
  if (servicesInfo.index < 0) {
    const trimmed = String(text || "").trim();
    if (!trimmed) {
      return "services:\n  app:\n    env_file:\n      - .env";
    }
    return `${trimmed}\n\nservices:\n  app:\n    env_file:\n      - .env`;
  }
  let i = servicesInfo.index + 1;
  let foundService = false;
  while (i < lines.length) {
    const line = lines[i];
    const indent = countIndent(line);
    if (indent <= servicesInfo.indent && !isCommentOrBlank(line)) {
      break;
    }
    if (isServiceHeader(line, indent, servicesInfo.indent)) {
      foundService = true;
      const serviceIndent = indent;
      let blockEnd = i + 1;
      while (blockEnd < lines.length) {
        const nextLine = lines[blockEnd];
        const nextIndent = countIndent(nextLine);
        if (nextIndent <= serviceIndent && !isCommentOrBlank(nextLine)) {
          break;
        }
        blockEnd += 1;
      }
      let envFileLine = -1;
      let envFileIndent = serviceIndent + 2;
      for (let j = i + 1; j < blockEnd; j += 1) {
        const current = lines[j];
        const currentIndent = countIndent(current);
        if (currentIndent === envFileIndent && String(current || "").trim().startsWith("env_file:")) {
          envFileLine = j;
          break;
        }
      }
      if (envFileLine >= 0) {
        let hasEnv = false;
        let insertAt = envFileLine + 1;
        for (let j = envFileLine + 1; j < blockEnd; j += 1) {
          const current = lines[j];
          const currentIndent = countIndent(current);
          if (currentIndent <= envFileIndent && !isCommentOrBlank(current)) {
            break;
          }
          if (currentIndent > envFileIndent && String(current || "").trim().startsWith("-")) {
            insertAt = j + 1;
            if (envFileItemRegex.test(String(current || "").trim())) {
              hasEnv = true;
            }
          }
        }
        if (!hasEnv) {
          lines.splice(insertAt, 0, `${" ".repeat(envFileIndent + 2)}- .env`);
          blockEnd += 1;
        }
      } else {
        let insertAt = i + 1;
        while (insertAt < blockEnd && isCommentOrBlank(lines[insertAt])) {
          insertAt += 1;
        }
        lines.splice(
          insertAt,
          0,
          `${" ".repeat(envFileIndent)}env_file:`,
          `${" ".repeat(envFileIndent + 2)}- .env`
        );
        blockEnd += 2;
      }
      i = blockEnd;
      continue;
    }
    i += 1;
  }
  if (!foundService) {
    const insertAt = servicesInfo.index + 1;
    lines.splice(
      insertAt,
      0,
      `${" ".repeat(servicesInfo.indent + 2)}app:`,
      `${" ".repeat(servicesInfo.indent + 4)}env_file:`,
      `${" ".repeat(servicesInfo.indent + 6)}- .env`
    );
  }
  return lines.join("\n");
}

function removeEnvFileFromServices(text) {
  const lines = String(text || "").split(/\r?\n/);
  const servicesInfo = findServicesBlock(lines);
  if (servicesInfo.index < 0) return text;
  const processed = [];
  processed.push(...lines.slice(0, servicesInfo.index + 1));
  let i = servicesInfo.index + 1;
  while (i < lines.length) {
    const currentLine = lines[i];
    const currentIndent = countIndent(currentLine);
    if (currentIndent <= servicesInfo.indent && !isCommentOrBlank(currentLine)) {
      processed.push(...lines.slice(i));
      return processed.join("\n");
    }
    if (!isServiceHeader(currentLine, currentIndent, servicesInfo.indent)) {
      processed.push(currentLine);
      i += 1;
      continue;
    }
    const serviceIndent = currentIndent;
    let blockEnd = i + 1;
    while (blockEnd < lines.length) {
      const nextLine = lines[blockEnd];
      const nextIndent = countIndent(nextLine);
      if (nextIndent <= serviceIndent && !isCommentOrBlank(nextLine)) {
        break;
      }
      blockEnd += 1;
    }
    const block = lines.slice(i, blockEnd);
    const newBlock = [block[0]];
    const envFileIndent = serviceIndent + 2;
    let k = 1;
    while (k < block.length) {
      const line = block[k];
      const indent = countIndent(line);
      const trimmed = String(line || "").trim();
      if (indent === envFileIndent && trimmed.startsWith("env_file:")) {
        let subEnd = k + 1;
        while (subEnd < block.length) {
          const subLine = block[subEnd];
          const subIndent = countIndent(subLine);
          if (subIndent <= envFileIndent && !isCommentOrBlank(subLine)) {
            break;
          }
          subEnd += 1;
        }
        const sub = block.slice(k + 1, subEnd);
        const kept = [];
        let hasItem = false;
        sub.forEach((subLine) => {
          const subIndent = countIndent(subLine);
          const subTrimmed = String(subLine || "").trim();
          if (subIndent > envFileIndent && subTrimmed.startsWith("-")) {
            if (envFileItemRegex.test(subTrimmed)) {
              return;
            }
            hasItem = true;
            kept.push(subLine);
            return;
          }
          kept.push(subLine);
        });
        if (hasItem) {
          newBlock.push(line, ...kept);
        } else {
          newBlock.push(...kept);
        }
        k = subEnd;
        continue;
      }
      newBlock.push(line);
      k += 1;
    }
    processed.push(...newBlock);
    i = blockEnd;
  }
  return processed.join("\n");
}

function updateStackEnvState() {
  if (!stackEnvToggle || !stackEnvInput) return;
  const enabled = Boolean(stackEnvToggle.checked);
  stackEnvInput.disabled = !enabled;
  if (envEditor && typeof envEditor.setReadOnly === "function") {
    envEditor.setReadOnly(!enabled);
  }
  if (stackEnvToggleBtn) {
    stackEnvToggleBtn.dataset.enabled = enabled ? "true" : "false";
    stackEnvToggleBtn.classList.toggle("is-active", enabled);
    const icon = stackEnvToggleBtn.querySelector(".icon-action");
    if (icon) {
      icon.className = `icon-action ${enabled ? "icon-key-off" : "icon-key"}`;
    }
    const label = enabled ? "Disable .env secrets" : "Use .env secrets";
    stackEnvToggleBtn.setAttribute("aria-label", label);
    stackEnvToggleBtn.setAttribute("data-tooltip", label);
    stackEnvToggleBtn.setAttribute("aria-pressed", enabled ? "true" : "false");
  }
  const current = getComposeValue();
  const next = enabled
    ? ensureEnvFileForServices(current)
    : removeEnvFileFromServices(current);
  setComposeValue(next);
  if (!enabled) {
    setEnvValue("");
  }
  if (composeEditor && typeof composeEditor.forceLint === "function") {
    composeEditor.forceLint();
  }
  if (envEditor && typeof envEditor.forceLint === "function") {
    envEditor.forceLint();
  }
}

async function openStackModal(name = "") {
  if (!stackModal) return;
  const scope = containersSelectedScope;
  if (!scope) {
    showToast("Select server first.");
    return;
  }
  editingStackName = name ? String(name) : "";
  if (stackModalTitle) {
    stackModalTitle.textContent = editingStackName ? `Edit stack: ${editingStackName}` : "New stack";
  }
  if (stackNameRow) {
    stackNameRow.classList.toggle("hidden", Boolean(editingStackName));
  }
  if (stackNameInput) {
    stackNameInput.value = editingStackName ? editingStackName : "";
    stackNameInput.disabled = Boolean(editingStackName);
  }
  setComposeValue("");
  setEnvValue("");
  clearStackModalError();
  if (stackEnvToggle) {
    stackEnvToggle.checked = false;
  }
  updateStackEnvState();
  stackModal.classList.remove("hidden");
  stackModal.setAttribute("aria-hidden", "false");
  if (editingStackName) {
    try {
      const payload = await fetchJSON(`/api/stacks/get?scope=${encodeURIComponent(scope)}&name=${encodeURIComponent(editingStackName)}`);
      if (payload) {
        setComposeValue(payload.compose_yaml || "");
        setEnvValue(payload.env || "");
        if (stackEnvToggle) stackEnvToggle.checked = Boolean(payload.has_env);
        updateStackEnvState();
        clearStackModalError();
      }
    } catch (err) {
      showToast(err.message || "Unable to load stack.");
    }
  } else if (stackNameInput) {
    stackNameInput.focus();
  }
}

function closeStackModal() {
  if (!stackModal) return;
  if (document.activeElement && stackModal.contains(document.activeElement)) {
    try {
      document.activeElement.blur();
    } catch {
      // ignore
    }
  }
  stackModal.classList.add("hidden");
  stackModal.setAttribute("aria-hidden", "true");
  editingStackName = "";
}

async function saveStackFromModal({ closeOnSuccess } = {}) {
  const scope = containersSelectedScope;
  if (!scope) {
    showToast("Select server first.");
    return false;
  }
  const name = editingStackName || (stackNameInput ? stackNameInput.value.trim() : "");
  if (!name || !isValidStackName(name)) {
    showToast("Stack name must match A-Z, a-z, 0-9, _ or -.");
    return false;
  }
  clearStackModalError();
  const composeYml = getComposeValue()
    .split(/\r?\n/)
    .map((line) => {
      const prefix = String(line || "").match(/^\s*/)?.[0] || "";
      if (!prefix.includes("\t")) return line;
      const fixed = prefix.replace(/\t/g, "  ");
      return fixed + String(line || "").slice(prefix.length);
    })
    .join("\n")
    .trim();
  setComposeValue(composeYml);
  if (!composeYml) {
    showStackModalError("docker-compose.yml is required.");
    return false;
  }
  const useEnv = Boolean(stackEnvToggle && stackEnvToggle.checked);
  const env = useEnv ? getEnvValue() : "";
  try {
    const validation = await fetchJSON("/api/stacks/validate", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ compose_yaml: composeYml, env, use_env: useEnv }),
    });
    if (!validation || validation.valid !== true) {
      const message = validation && validation.error ? validation.error : "Compose validation failed.";
      showStackModalError(message);
      return false;
    }
    await fetchJSON("/api/stacks/save", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ scope, name, compose_yaml: composeYml, env, use_env: useEnv }),
    });
    if (closeOnSuccess !== false) {
      closeStackModal();
    }
    await refreshStacks({ silent: true });
    return true;
  } catch (err) {
    showStackModalError(err.message || "Stack save failed.");
    return false;
  }
}

async function runStackActionFromModal(action) {
  const scope = containersSelectedScope;
  if (!scope) {
    showToast("Select server first.");
    return;
  }
  const name = editingStackName || (stackNameInput ? stackNameInput.value.trim() : "");
  if (!name || !isValidStackName(name)) {
    showToast("Stack name must match A-Z, a-z, 0-9, _ or -.");
    return;
  }
  const ok = await saveStackFromModal({ closeOnSuccess: false });
  if (!ok) return;
  await runStackAction(name, action);
}

async function copyToClipboard(value, label, options = {}) {
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
    if (!options.silent) {
      showToast(`${label || "Copied"} to clipboard.`);
    }
  } catch {
    if (!options.silent) {
      showToast("Copy failed.");
    }
    throw new Error("copy failed");
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
  let infoText = `Current image: ${currentImageID}\nNew image: ${newImageID}`;
  if (container.error && String(container.error).startsWith("skipped: digest unknown")) {
    const errText = String(container.error);
    let reason = "registry/local digest unavailable";
    if (errText.includes("local digest unavailable")) {
      reason = "local digest unavailable";
    } else if (errText.includes("registry digest unavailable")) {
      reason = "registry digest unavailable";
    } else if (errText.includes("registry error")) {
      reason = "registry error";
    }
    infoText += `\nSkipped: digest unknown (${reason}).`;
  }
  infoPanel.textContent = infoText;

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
    } else if (infoChecking && (!infoStatus || infoStatus === "unknown")) {
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
    if (!scanLabelText && infoChecking && (!infoStatus || infoStatus === "unknown")) {
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
    nameIcon.setAttribute("class", "status-card-icon icon icon-tabler icons-tabler-outline icon-tabler-server-bolt");
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
    nameIconWrap.className = "status-card-icon-wrap";
    nameIconWrap.appendChild(nameIcon);
    const name = document.createElement("strong");
    name.textContent = result.server_name || "unknown";
    nameWrap.appendChild(nameIconWrap);
    nameWrap.appendChild(name);
    headLeft.appendChild(nameWrap);

    const typeIcon = document.createElementNS("http://www.w3.org/2000/svg", "svg");
    typeIcon.setAttribute(
      "class",
      `status-card-type-icon status-icon status-icon-${statusLabel} icon icon-tabler icons-tabler-outline icon-tabler-${result.local ? "plug-connected" : "network"}`
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
    typeWrap.dataset.tooltipServerStatus = "true";
    typeWrap.dataset.statusLabel = statusLabel;
    typeWrap.dataset.checkedAt = infoCheckedAt || "";
    typeWrap.dataset.serverType = result.local ? "local" : "remote";
    if (result.local) {
      typeIcon.innerHTML = '<path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M7 12l5 5l-1.5 1.5a3.536 3.536 0 1 1 -5 -5l1.5 -1.5" /><path d="M17 12l-5 -5l1.5 -1.5a3.536 3.536 0 1 1 5 5l-1.5 1.5" /><path d="M3 21l2.5 -2.5" /><path d="M18.5 5.5l2.5 -2.5" /><path d="M10 11l-2 2" /><path d="M13 14l-2 2" />';
    } else {
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
  maintenanceBtn.className = isMaintenance
    ? "btn-warning icon-action-btn has-tooltip"
    : "secondary icon-action-btn has-tooltip";
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
    "status-card-icon icon icon-tabler icons-tabler-outline icon-tabler-server-bolt"
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
  nameIconWrap.className = "status-card-icon-wrap";
  nameIconWrap.appendChild(nameIcon);

  const nameEl = document.createElement("strong");
  nameEl.textContent = item.server.name;

  const typeIcon = document.createElementNS("http://www.w3.org/2000/svg", "svg");
  typeIcon.setAttribute(
    "class",
    `status-card-type-icon status-icon status-icon-${statusLabel} icon icon-tabler icons-tabler-outline icon-tabler-${item.type === "local" ? "plug-connected" : "network"}`
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
  typeWrap.dataset.tooltipServerStatus = "true";
  typeWrap.dataset.statusLabel = statusLabel;
  typeWrap.dataset.checkedAt = info.checked_at || "";
  typeWrap.dataset.serverType = item.type;
  if (item.type === "local") {
    typeIcon.innerHTML = '<path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M7 12l5 5l-1.5 1.5a3.536 3.536 0 1 1 -5 -5l1.5 -1.5" /><path d="M17 12l-5 -5l1.5 -1.5a3.536 3.536 0 1 1 5 5l-1.5 1.5" /><path d="M3 21l2.5 -2.5" /><path d="M18.5 5.5l2.5 -2.5" /><path d="M10 11l-2 2" /><path d="M13 14l-2 2" />';
  } else {
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
    if (infoChecking && (!statusValue || statusValue === "unknown")) statusValue = "checking";
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
  const existingChecking = Boolean(existing.checking);
  const incomingChecking = Boolean(incoming.checking);
  const incomingCheckedAt = serverInfoCheckedAtMs(incoming);
  const existingCheckedAt = serverInfoCheckedAtMs(existing);
  if (existingChecking && !incomingChecking) return true;
  if (!existingChecking && incomingChecking) {
    if (incomingCheckedAt >= existingCheckedAt) return true;
    return false;
  }
  if (incomingCheckedAt > existingCheckedAt) return true;
  if (incomingCheckedAt < existingCheckedAt) return false;
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

function updateSidebarSearchUI() {
  if (!sidebarSearch) return;
  if (currentView === "containers") {
    updateContainersSearchCount();
  } else if (sidebarSearchCountEl) {
    sidebarSearchCountEl.classList.add("hidden");
  }
}

function applySidebarFilter(queryOverride) {
  if (!sidebarSearch) return;
  updateSidebarSearchUI();
  const query = normalizeQuery(queryOverride ?? sidebarSearch.value);
  if (currentView === "containers") {
    if (containersViewMode === "shell" && containersShellList) {
      const items = Array.from(containersShellList.querySelectorAll(".containers-shell-item"));
      items.forEach((item) => {
        const source = item.dataset && item.dataset.search ? item.dataset.search : item.textContent;
        const haystack = normalizeQuery(source);
        const match = !query || haystack.includes(query);
        item.classList.toggle("filtered-out", !match);
      });
    } else if (containersViewMode === "logs" && containersLogsList) {
      const items = Array.from(containersLogsList.querySelectorAll(".containers-shell-item"));
      items.forEach((item) => {
        const source = item.dataset && item.dataset.search ? item.dataset.search : item.textContent;
        const haystack = normalizeQuery(source);
        const match = !query || haystack.includes(query);
        item.classList.toggle("filtered-out", !match);
      });
    } else if (containersViewMode === "resources" && containersResourcesList) {
      const items = Array.from(containersResourcesList.querySelectorAll(".containers-shell-item"));
      items.forEach((item) => {
        const source = item.dataset && item.dataset.search ? item.dataset.search : item.textContent;
        const haystack = normalizeQuery(source);
        const match = !query || haystack.includes(query);
        item.classList.toggle("filtered-out", !match);
      });
    } else if (containersViewMode === "stacks" && stacksTableBody) {
      const rows = Array.from(stacksTableBody.querySelectorAll("tr.stack-row"));
      rows.forEach((row) => {
        const source = row.dataset && row.dataset.search ? row.dataset.search : row.textContent;
        const haystack = normalizeQuery(source);
        const match = !query || haystack.includes(query);
        row.classList.toggle("filtered-out", !match);
      });
    } else if (containersViewMode === "images" && imagesTableBody) {
      const rows = Array.from(imagesTableBody.querySelectorAll("tr.images-row"));
      rows.forEach((row) => {
        const source = row.dataset && row.dataset.search ? row.dataset.search : row.textContent;
        const haystack = normalizeQuery(source);
        const match = !query || haystack.includes(query);
        row.classList.toggle("filtered-out", !match);
      });
    } else {
      const rows = Array.from(containersTableBody ? containersTableBody.querySelectorAll("tr") : []);
      rows.forEach((row) => {
        const source = row.dataset && row.dataset.search ? row.dataset.search : row.textContent;
        const haystack = normalizeQuery(source);
        const match = !query || haystack.includes(query);
        row.classList.toggle("filtered-out", !match);
      });
    }
    updateContainersSearchCount();
    return;
  }
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

function applyExperimentalFeatures(cfg) {
  const exp = cfg && cfg.experimental_features ? cfg.experimental_features : {};
  const flags = {
    containers: Boolean(exp.containers),
    container_shell: Boolean(exp.container_shell),
    container_logs: Boolean(exp.container_logs),
    container_resources: Boolean(exp.container_resources),
    stacks: Boolean(exp.stacks),
    images: Boolean(exp.images),
    containers_sidebar: Boolean(exp.containers_sidebar),
  };
  if (sidebar) {
    sidebar.querySelectorAll(".experimental-link").forEach((btn) => {
      const view = btn.getAttribute("data-view");
      const enabled = Boolean(flags[view]);
      btn.classList.toggle("hidden", !enabled);
    });
    const showContainerShortcuts = flags.containers && flags.containers_sidebar;
    sidebar.querySelectorAll(".containers-sidebar-link").forEach((btn) => {
      const requiredFlag = btn.getAttribute("data-feature");
      const enabled = showContainerShortcuts && Boolean(flags[requiredFlag]);
      btn.classList.toggle("hidden", !enabled);
    });
  }
  if (flags.containers) {
    updateContainersServerOptions();
  }
  if (topbarContainersShellBtn) {
    topbarContainersShellBtn.classList.toggle("hidden", !(flags.container_shell && flags.containers));
  }
  if (topbarContainersLogsBtn) {
    topbarContainersLogsBtn.classList.toggle("hidden", !(flags.container_logs && flags.containers));
  }
  if (topbarContainersResourcesBtn) {
    topbarContainersResourcesBtn.classList.toggle("hidden", !(flags.container_resources && flags.containers));
  }
  if (topbarContainersStacksBtn) {
    topbarContainersStacksBtn.classList.toggle("hidden", !(flags.stacks && flags.containers));
  }
  if (topbarContainersImagesBtn) {
    topbarContainersImagesBtn.classList.toggle("hidden", !(flags.images && flags.containers));
  }
  if (!flags.container_shell && containersViewMode === "shell") {
    setContainersViewMode("table");
  }
  if (!flags.container_logs && containersViewMode === "logs") {
    setContainersViewMode("table");
  }
  if (!flags.container_resources && containersViewMode === "resources") {
    setContainersViewMode("table");
  }
  if (!flags.stacks && containersViewMode === "stacks") {
    setContainersViewMode("table");
  }
  if (!flags.images && containersViewMode === "images") {
    setContainersViewMode("table");
  }
  if (viewContainersEl && !flags.containers) viewContainersEl.classList.add("hidden");
  if (currentView === "containers" && !flags.containers) {
    setView("status");
  }
  updateContainersExperimentalToggles();
  updateSidebarNavActive(currentView);
}

function isExperimentalEnabled(view) {
  const exp = currentConfig && currentConfig.experimental_features ? currentConfig.experimental_features : {};
  return Boolean(exp && exp[view]);
}

function updateServersFilterButtons() {
  const mode = serversFilterMode || "all";
  renderServersFilterButton(mode);
  updateServersFilterMenu(mode);
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

function getServersFilterLabel(mode) {
  switch (mode) {
    case "active":
      return "Active";
    case "maintenance":
      return "Maintenance";
    case "offline":
      return "Offline";
    default:
      return "All";
  }
}

function renderServersFilterButton(mode) {
  if (!serversFilterBtn) return;
  const content = serversFilterBtn.querySelector(".servers-filter-btn-content");
  if (!content) return;
  content.textContent = getServersFilterLabel(mode);
}

function closeServersFilterMenu() {
  if (!serversFilterMenu || !serversFilterBtn) return;
  serversFilterMenu.classList.add("hidden");
  serversFilterBtn.setAttribute("aria-expanded", "false");
}

function toggleServersFilterMenu() {
  if (!serversFilterMenu || !serversFilterBtn) return;
  const isOpen = !serversFilterMenu.classList.contains("hidden");
  if (isOpen) {
    closeServersFilterMenu();
  } else {
    serversFilterMenu.classList.remove("hidden");
    serversFilterBtn.setAttribute("aria-expanded", "true");
  }
}

function updateServersFilterMenu(mode = serversFilterMode) {
  if (!serversFilterMenu) return;
  const options = [
    { value: "all", label: "All" },
    { value: "active", label: "Active" },
    { value: "maintenance", label: "Maintenance" },
    { value: "offline", label: "Offline" },
  ];
  serversFilterMenu.innerHTML = "";
  options.forEach((option) => {
    const button = document.createElement("button");
    button.type = "button";
    button.className = "containers-server-option servers-filter-option";
    if (option.value === mode) {
      button.classList.add("active");
    }
    button.textContent = option.label;
    button.addEventListener("click", () => {
      setServersFilterMode(option.value);
      closeServersFilterMenu();
    });
    serversFilterMenu.appendChild(button);
  });
}

function updateContainersExperimentalToggles() {
  const enabled = expContainersInput ? expContainersInput.checked : false;
  const dependents = [expContainersSidebarInput, expContainerShellInput, expContainerLogsInput, expContainerResourcesInput, expStacksInput, expImagesInput];
  dependents.forEach((input) => {
    if (!input) return;
    input.disabled = !enabled;
    const row = input.closest(".toggle-row");
    if (row) {
      row.classList.toggle("toggle-disabled", !enabled);
    }
  });
}

function getLogsLevelLabel(mode) {
  switch (mode) {
    case "info":
      return "Info";
    case "warn":
      return "Warn";
    case "error":
      return "Error";
    default:
      return "All";
  }
}

function renderLogsLevelButton(mode) {
  if (!logsLevelBtn) return;
  const content = logsLevelBtn.querySelector(".logs-level-btn-content");
  if (!content) return;
  content.textContent = getLogsLevelLabel(mode);
}

function closeLogsLevelMenu() {
  if (!logsLevelMenu || !logsLevelBtn) return;
  logsLevelMenu.classList.add("hidden");
  logsLevelBtn.setAttribute("aria-expanded", "false");
}

function toggleLogsLevelMenu() {
  if (!logsLevelMenu || !logsLevelBtn) return;
  const isOpen = !logsLevelMenu.classList.contains("hidden");
  if (isOpen) {
    closeLogsLevelMenu();
  } else {
    logsLevelMenu.classList.remove("hidden");
    logsLevelBtn.setAttribute("aria-expanded", "true");
  }
}

function updateLogsLevelMenu(mode = logsLevelMode) {
  if (!logsLevelMenu) return;
  const options = [
    { value: "all", label: "All" },
    { value: "info", label: "Info" },
    { value: "warn", label: "Warn" },
    { value: "error", label: "Error" },
  ];
  logsLevelMenu.innerHTML = "";
  options.forEach((option) => {
    const button = document.createElement("button");
    button.type = "button";
    button.className = "containers-server-option logs-level-option";
    if (option.value === mode) {
      button.classList.add("active");
    }
    button.textContent = option.label;
    button.addEventListener("click", () => {
      logsLevelMode = option.value;
      renderLogsLevelButton(logsLevelMode);
      updateLogsLevelMenu(logsLevelMode);
      closeLogsLevelMenu();
      refreshLogs();
    });
    logsLevelMenu.appendChild(button);
  });
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

function updateContainersResourcesViewButtons() {
  if (!containersResourcesViewToggleBtn) return;
  const current = containersResourcesViewMode === "table" ? "table" : "cards";
  const next = current === "cards" ? "table" : "cards";
  containersResourcesViewToggleBtn.dataset.mode = next;
  const label = next === "cards" ? "Switch to cards view" : "Switch to table view";
  containersResourcesViewToggleBtn.setAttribute("aria-label", label);
  containersResourcesViewToggleBtn.setAttribute("data-tooltip", label);
}

function setContainersResourcesViewMode(mode, persist = true) {
  containersResourcesViewMode = mode === "table" ? "table" : "cards";
  if (persist) {
    try {
      localStorage.setItem(containersResourcesViewStorageKey, containersResourcesViewMode);
    } catch {
      // ignore
    }
  }
  updateContainersResourcesViewButtons();
  renderContainersResourcesView();
  updateContainersResourcesPlaceholder();
}

function updateContainersResourcesSortUI() {
  const buttons = [
    { key: "name", btn: resourcesNameSortBtn, label: "name" },
    { key: "cpu", btn: resourcesCpuSortBtn, label: "CPU" },
    { key: "ram", btn: resourcesRamSortBtn, label: "RAM" },
    { key: "status", btn: resourcesStatusSortBtn, label: "status" },
    { key: "uptime", btn: resourcesUptimeSortBtn, label: "uptime" },
    { key: "ip", btn: resourcesIpSortBtn, label: "IP" },
    { key: "port", btn: resourcesPortSortBtn, label: "port" },
  ];
  buttons.forEach(({ key, btn, label }) => {
    if (!btn) return;
    const icon = btn.querySelector(".icon-sort");
    if (!icon) return;
    icon.classList.remove("icon-sort-neutral", "icon-sort-asc", "icon-sort-desc");
    if (containersResourcesSortMode === `${key}:asc`) {
      icon.classList.add("icon-sort-asc");
      btn.setAttribute("aria-label", `Sort ${label} A → Z`);
      return;
    }
    if (containersResourcesSortMode === `${key}:desc`) {
      icon.classList.add("icon-sort-desc");
      btn.setAttribute("aria-label", `Sort ${label} Z → A`);
      return;
    }
    icon.classList.add("icon-sort-neutral");
    btn.setAttribute("aria-label", `Sort by ${label}`);
  });
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
  const exp = cfg.experimental_features || {};
  if (expContainersInput) expContainersInput.checked = Boolean(exp.containers);
  if (expContainersSidebarInput) expContainersSidebarInput.checked = Boolean(exp.containers_sidebar);
  if (expContainerShellInput) expContainerShellInput.checked = Boolean(exp.container_shell);
  if (expContainerLogsInput) expContainerLogsInput.checked = Boolean(exp.container_logs);
  if (expContainerResourcesInput) expContainerResourcesInput.checked = Boolean(exp.container_resources);
  if (expStacksInput) expStacksInput.checked = Boolean(exp.stacks);
  if (expImagesInput) expImagesInput.checked = Boolean(exp.images);
  currentTimeZone = cfg.time_zone ? String(cfg.time_zone).trim() : "";
  setContainersLogsTimestamps(containersLogsTimestamps);
  applyExperimentalFeatures(cfg);
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
    experimental_features: {
      containers: expContainersInput ? expContainersInput.checked : false,
      containers_sidebar: expContainersSidebarInput ? expContainersSidebarInput.checked : false,
      container_shell: expContainerShellInput ? expContainerShellInput.checked : false,
      container_logs: expContainerLogsInput ? expContainerLogsInput.checked : false,
      container_resources: expContainerResourcesInput ? expContainerResourcesInput.checked : false,
      stacks: expStacksInput ? expStacksInput.checked : false,
      images: expImagesInput ? expImagesInput.checked : false,
    },
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
  updateContainersServerOptions();
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

function formatUptime(seconds) {
  const total = Number(seconds);
  if (!Number.isFinite(total) || total <= 0) return "—";
  const days = Math.floor(total / 86400);
  const hours = Math.floor((total % 86400) / 3600);
  const minutes = Math.floor((total % 3600) / 60);
  const parts = [];
  if (days > 0) parts.push(`${days}d`);
  if (hours > 0 || days > 0) parts.push(`${hours}h`);
  parts.push(`${minutes}m`);
  return parts.join(" ");
}

function formatBytes(value) {
  const total = Number(value);
  if (!Number.isFinite(total) || total <= 0) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  let size = total;
  let unitIndex = 0;
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024;
    unitIndex += 1;
  }
  return `${size.toFixed(size >= 10 || unitIndex === 0 ? 0 : 1)} ${units[unitIndex]}`;
}

function formatBytesMBOrGB(value) {
  const total = Number(value);
  if (!Number.isFinite(total) || total <= 0) return "0 MB";
  const mb = total / (1024 * 1024);
  // UX: keep the RAM column narrow. Above ~999 MB we switch to GB so the value stays compact.
  if (mb < 1000) {
    return `${mb.toFixed(mb >= 10 ? 0 : 1)} MB`;
  }
  const gb = mb / 1024;
  return `${gb.toFixed(gb >= 10 ? 0 : 1)} GB`;
}

function formatPercent(value) {
  const total = Number(value);
  if (!Number.isFinite(total) || total <= 0) return "0%";
  if (total >= 100) return "100%";
  return `${total.toFixed(total >= 10 ? 0 : 1)}%`;
}

function formatImageCreated(value) {
  const date = new Date(value);
  if (!Number.isFinite(date.getTime())) return "—";
  return date.toLocaleString(undefined, {
    year: "numeric",
    month: "short",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function normalizeContainerState(state) {
  return String(state || "").toLowerCase();
}

function containersSort(list, mode) {
  const [field, order] = String(mode || "name:asc").split(":");
  const dir = order === "desc" ? "desc" : "asc";
  const getResourcesSortValue = (container) => {
    const resource = containersTableResourcesData.get(container.id);
    if (field === "cpu") return resource && Number.isFinite(Number(resource.cpu_percent)) ? Number(resource.cpu_percent) : null;
    if (field === "ram") return resource && Number.isFinite(Number(resource.mem_usage_bytes)) ? Number(resource.mem_usage_bytes) : null;
    if (field === "ip") {
      const ips = Array.isArray(resource && resource.ip_addresses)
        ? resource.ip_addresses
        : Array.isArray(container.ip_addresses)
            ? container.ip_addresses
            : [];
      return ips.length > 0 ? String(ips[0].ip || "") : null;
    }
    if (field === "port") {
      const ports = getResourcesHostPorts(resource ? resource.ports : container.ports);
      return ports.length > 0 ? Number(ports[0]) : null;
    }
    return null;
  };

  return list.slice().sort((a, b) => {
    const aVal = field === "state"
      ? normalizeContainerState(a.state)
      : field === "cpu" || field === "ram" || field === "ip" || field === "port"
          ? getResourcesSortValue(a)
          : String(a.name || "");
    const bVal = field === "state"
      ? normalizeContainerState(b.state)
      : field === "cpu" || field === "ram" || field === "ip" || field === "port"
          ? getResourcesSortValue(b)
          : String(b.name || "");

    return compareContainersResourcesValues(aVal, bVal, dir);
  });
}

function stacksSort(list, mode) {
  const [field, order] = String(mode || "name:asc").split(":");
  const dir = order === "desc" ? -1 : 1;
  return list.slice().sort((a, b) => {
    const aVal = field === "state" ? String(a.status || "") : String(a.name || "");
    const bVal = field === "state" ? String(b.status || "") : String(b.name || "");
    if (aVal === bVal) return 0;
    return aVal > bVal ? dir : -dir;
  });
}

function imagesSortKey(image, field) {
  if (!image) return "";
  switch (field) {
    case "tags":
      return String(image.tag || "");
    case "state":
      return normalizeQuery(imageStateLabel(image));
    case "size":
      return Number(image.size_bytes || 0);
    default:
      return String(image.repository || "");
  }
}

function updateStacksSortUI() {
  const buttons = [
    { key: "name", btn: stacksNameSortBtn, label: "name" },
    { key: "state", btn: stacksStateSortBtn, label: "state" },
  ];
  buttons.forEach(({ key, btn, label }) => {
    if (!btn) return;
    const icon = btn.querySelector(".icon-sort");
    if (!icon) return;
    icon.classList.remove("icon-sort-neutral", "icon-sort-asc", "icon-sort-desc");
    if (stacksSortMode === `${key}:asc`) {
      icon.classList.add("icon-sort-asc");
      btn.setAttribute("aria-label", `Sort ${label} A → Z`);
      return;
    }
    if (stacksSortMode === `${key}:desc`) {
      icon.classList.add("icon-sort-desc");
      btn.setAttribute("aria-label", `Sort ${label} Z → A`);
      return;
    }
    icon.classList.add("icon-sort-neutral");
    btn.setAttribute("aria-label", `Sort by ${label}`);
  });
}

function updateImagesSortUI() {
  const buttons = [
    { key: "repository", btn: imagesRepoSortBtn, label: "repository" },
    { key: "tags", btn: imagesTagSortBtn, label: "tags" },
    { key: "state", btn: imagesStateSortBtn, label: "state" },
    { key: "size", btn: imagesSizeSortBtn, label: "size" },
  ];
  buttons.forEach(({ key, btn, label }) => {
    if (!btn) return;
    const icon = btn.querySelector(".icon-sort");
    if (!icon) return;
    icon.classList.remove("icon-sort-neutral", "icon-sort-asc", "icon-sort-desc");
    if (imagesSortMode === `${key}:asc`) {
      icon.classList.add("icon-sort-asc");
      btn.setAttribute("aria-label", `Sort ${label} A → Z`);
      return;
    }
    if (imagesSortMode === `${key}:desc`) {
      icon.classList.add("icon-sort-desc");
      btn.setAttribute("aria-label", `Sort ${label} Z → A`);
      return;
    }
    icon.classList.add("icon-sort-neutral");
    btn.setAttribute("aria-label", `Sort by ${label}`);
  });
}

function updateContainersSortUI() {
  const buttons = [
    { key: "name", btn: containersNameSortBtn, label: "name" },
    { key: "state", btn: containersStateSortBtn, label: "state" },
    { key: "cpu", btn: containersCpuSortBtn, label: "CPU" },
    { key: "ram", btn: containersRamSortBtn, label: "RAM" },
    { key: "ip", btn: containersIpSortBtn, label: "IP" },
    { key: "port", btn: containersPortSortBtn, label: "port" },
  ];
  buttons.forEach(({ key, btn, label }) => {
    if (!btn) return;
    const icon = btn.querySelector(".icon-sort");
    if (!icon) return;
    icon.classList.remove("icon-sort-neutral", "icon-sort-asc", "icon-sort-desc");
    if (containersSortMode === `${key}:asc`) {
      icon.classList.add("icon-sort-asc");
      btn.setAttribute("aria-label", `Sort ${label} A → Z`);
      return;
    }
    if (containersSortMode === `${key}:desc`) {
      icon.classList.add("icon-sort-desc");
      btn.setAttribute("aria-label", `Sort ${label} Z → A`);
      return;
    }
    icon.classList.add("icon-sort-neutral");
    btn.setAttribute("aria-label", `Sort by ${label}`);
  });
}

function getScopeInfo(scope) {
  const [type, name] = String(scope || "").split(":");
  if (!type || !name) return null;
  const server = type === "local"
    ? cachedLocals.find((item) => item.name === name)
    : cachedRemotes.find((item) => item.name === name);
  const infoKey = `${type}:${name}`;
  const info = cachedServerInfo[infoKey] || {};
  const status = String(info.status || "").toLowerCase();
  const maintenance = Boolean(server && server.maintenance);
  const checking = debouncedCheckingVisible(infoKey, info);
  return { type, name, server, status, maintenance, checking };
}

function updateContainersStatus(scope, errorMessage) {
  if (!topbarContainersStatusEl) return;
  if (!scope) {
    topbarContainersStatusEl.innerHTML = "";
    return;
  }
  const info = getScopeInfo(scope);
  let message = "";
  let variant = "online";
  if (info && info.maintenance) {
    message = "Maintenance mode — operations may be limited.";
    variant = "maintenance";
  }
  if (info && info.status === "offline") {
    message = "Offline — data unavailable.";
    variant = "offline";
  }
  if (errorMessage) {
    message = `Error: ${errorMessage}`;
    variant = "offline";
  }
  if (!message) {
    topbarContainersStatusEl.innerHTML = "";
    return;
  }
  topbarContainersStatusEl.innerHTML = "";
  const pill = document.createElement("span");
  pill.className = "status-pill";
  pill.dataset.variant = variant;
  pill.textContent = message;
  topbarContainersStatusEl.appendChild(pill);
}

function updateContainersSearchCount() {
  if (!sidebarSearchCountEl || !sidebarSearch) return;
  if (currentView !== "containers") {
    sidebarSearchCountEl.classList.add("hidden");
    return;
  }
  let total = 0;
  let visible = 0;
  if (containersViewMode === "shell" && containersShellList) {
    const items = Array.from(containersShellList.querySelectorAll(".containers-shell-item"));
    total = items.length;
    visible = items.filter((item) => !item.classList.contains("filtered-out")).length;
  } else if (containersViewMode === "logs" && containersLogsList) {
    const items = Array.from(containersLogsList.querySelectorAll(".containers-shell-item"));
    total = items.length;
    visible = items.filter((item) => !item.classList.contains("filtered-out")).length;
  } else if (containersViewMode === "resources" && containersResourcesList) {
    const items = Array.from(containersResourcesList.querySelectorAll(".containers-shell-item"));
    total = items.length;
    visible = items.filter((item) => !item.classList.contains("filtered-out")).length;
  } else if (containersViewMode === "stacks" && stacksTableBody) {
    const rows = Array.from(stacksTableBody.querySelectorAll("tr.stack-row"));
    total = rows.length;
    visible = rows.filter((row) => !row.classList.contains("filtered-out")).length;
  } else if (containersViewMode === "images" && imagesTableBody) {
    const rows = Array.from(imagesTableBody.querySelectorAll("tr.images-image-row"));
    total = rows.length;
    visible = rows.filter((row) => !row.classList.contains("filtered-out") && !row.classList.contains("images-row-hidden")).length;
  } else if (containersTableBody) {
    const rows = Array.from(containersTableBody.querySelectorAll("tr"));
    total = rows.length;
    visible = rows.filter((row) => !row.classList.contains("filtered-out")).length;
  }
  const query = sidebarSearch.value.trim();
  const focused = document.activeElement === sidebarSearch;
  const shouldShow = focused || query.length > 0;
  if (!shouldShow) {
    sidebarSearchCountEl.classList.add("hidden");
    return;
  }
  sidebarSearchCountEl.classList.remove("hidden");
  sidebarSearchCountEl.textContent = query ? `${visible}/${total}` : `${total}`;
}

function setContainersViewMode(mode) {
  const next = mode === "shell"
    ? "shell"
    : mode === "logs"
        ? "logs"
        : mode === "resources"
            ? "resources"
            : mode === "stacks"
                ? "stacks"
                : mode === "images"
                    ? "images"
                    : "table";
  containersViewMode = next;
  if (topbarContainersEl) {
    const title = topbarContainersEl.querySelector("h2");
    if (title) {
      title.textContent = next === "shell"
        ? "Containers Shell"
        : next === "logs"
            ? "Containers Logs"
            : next === "resources"
                ? "Containers Resources"
                : next === "stacks"
                    ? "Containers Stacks"
                    : next === "images"
                        ? "Containers Images"
                        : "Containers";
    }
  }
  const isSplit = next === "shell" || next === "logs" || next === "resources";
  if (viewContainersEl) {
    viewContainersEl.classList.toggle("shell-mode", isSplit);
  }
  if (containersTableView) {
    containersTableView.classList.toggle("hidden", next !== "table");
  }
  if (containersStacksView) {
    containersStacksView.classList.toggle("hidden", next !== "stacks");
  }
  if (containersImagesView) {
    containersImagesView.classList.toggle("hidden", next !== "images");
  }
  if (containersShellLayout) {
    containersShellLayout.classList.toggle("hidden", next !== "shell");
  }
  if (containersLogsLayout) {
    containersLogsLayout.classList.toggle("hidden", next !== "logs");
  }
  if (containersResourcesLayout) {
    containersResourcesLayout.classList.toggle("hidden", next !== "resources");
  }
  if (topbarContainersShellBtn) {
    topbarContainersShellBtn.classList.toggle("is-active", next === "shell");
  }
  if (topbarContainersLogsBtn) {
    topbarContainersLogsBtn.classList.toggle("is-active", next === "logs");
  }
  if (topbarContainersResourcesBtn) {
    topbarContainersResourcesBtn.classList.toggle("is-active", next === "resources");
  }
  if (topbarContainersStacksBtn) {
    topbarContainersStacksBtn.classList.toggle("is-active", next === "stacks");
  }
  if (topbarContainersImagesBtn) {
    topbarContainersImagesBtn.classList.toggle("is-active", next === "images");
  }
  if (next === "shell") {
    closeContainersLogsSession("switch to shell");
    setContainersLogsPaused(false);
    const list = Array.from(containersCache.values()).map((entry) => entry.data);
    renderContainersShellList(list);
    stopContainersAutoRefresh();
    stopStacksAutoRefresh();
    refreshContainers({ silent: true }).catch(() => {});
  } else if (next === "logs") {
    closeContainersShellSession("switch to logs");
    closeContainersLogsSession("enter logs");
    setContainersLogsPaused(false);
    updateContainersLogsStreamIndicator();
    const list = Array.from(containersCache.values()).map((entry) => entry.data);
    renderContainersLogsList(list);
    stopContainersAutoRefresh();
    stopStacksAutoRefresh();
    refreshContainers({ silent: true }).catch(() => {});
  } else if (next === "resources") {
    closeContainersShellSession("switch to resources");
    closeContainersLogsSession("switch to resources");
    setContainersLogsPaused(false);
    const list = Array.from(containersCache.values()).map((entry) => entry.data);
    renderContainersResourcesList(list);
    stopContainersAutoRefresh();
    stopStacksAutoRefresh();
    startContainersResourcesAutoRefresh();
    refreshContainers({ silent: true }).catch(() => {});
    refreshContainersResources({ silent: true, force: true }).catch(() => {});
  } else if (next === "stacks") {
    closeContainersShellSession("switch to stacks");
    closeContainersLogsSession("switch to stacks");
    setContainersLogsPaused(false);
    stopContainersAutoRefresh();
    startStacksAutoRefresh();
    refreshStacks({ silent: true }).catch(() => {});
  } else if (next === "images") {
    closeContainersShellSession("switch to images");
    closeContainersLogsSession("switch to images");
    setContainersLogsPaused(false);
    stopContainersAutoRefresh();
    stopStacksAutoRefresh();
    refreshImages({ silent: true }).catch(() => {});
  } else {
    closeContainersShellSession("switch to table");
    closeContainersLogsSession("switch to table");
    stopStacksAutoRefresh();
    startContainersAutoRefresh();
    refreshContainers({ silent: true }).catch(() => {});
  }
  if (next !== "resources") {
    stopContainersResourcesAutoRefresh();
    containersResourcesSelectedIds.clear();
    updateContainersResourcesPlaceholder();
  }
  if (sidebarSearch && sidebarSearch.value.trim()) {
    applySidebarFilter(sidebarSearch.value);
  } else {
    updateContainersSearchCount();
  }
  updateSidebarNavActive(currentView);
}

function renderContainersShellList(list) {
  if (!containersShellList) return;
  containersShellList.innerHTML = "";
  if (!Array.isArray(list) || list.length === 0) {
    const empty = document.createElement("div");
    empty.className = "containers-shell-placeholder";
    empty.textContent = "No containers found.";
    containersShellList.appendChild(empty);
    return;
  }
  const sorted = containersSort(list, containersSortMode);
  let selectedFound = false;
  sorted.forEach((container) => {
    if (!container || !container.id) return;
    const button = document.createElement("button");
    button.type = "button";
    button.className = "containers-shell-item";
    if (container.id === containersShellSelectedId) {
      button.classList.add("active");
      selectedFound = true;
    }
    button.dataset.id = container.id;
    button.dataset.search = `${container.name || ""} ${container.state || ""}`;
    button.textContent = "";
    const icon = document.createElement("span");
    icon.className = "icon-action icon-package name-leading-icon";
    icon.setAttribute("aria-hidden", "true");
    button.appendChild(icon);
    button.appendChild(document.createTextNode(container.name || container.id.slice(0, 12)));
    button.addEventListener("click", () => {
      openContainersShell(container);
    });
    containersShellList.appendChild(button);
  });
  if (!selectedFound && containersShellSelectedId) {
    containersShellSelectedId = "";
    closeContainersShellSession("selection missing");
  }
  if (sidebarSearch && sidebarSearch.value.trim()) {
    applySidebarFilter(sidebarSearch.value);
  }
}

function renderContainersLogsList(list) {
  if (!containersLogsList) return;
  containersLogsList.innerHTML = "";
  if (!Array.isArray(list) || list.length === 0) {
    const empty = document.createElement("div");
    empty.className = "containers-shell-placeholder";
    empty.textContent = "No containers found.";
    containersLogsList.appendChild(empty);
    return;
  }
  const sorted = containersSort(list, containersSortMode);
  let selectedFound = false;
  sorted.forEach((container) => {
    if (!container || !container.id) return;
    const button = document.createElement("button");
    button.type = "button";
    button.className = "containers-shell-item";
    if (container.id === containersLogsSelectedId) {
      button.classList.add("active");
      selectedFound = true;
    }
    button.dataset.id = container.id;
    button.dataset.search = `${container.name || ""} ${container.state || ""}`;
    button.textContent = "";
    const icon = document.createElement("span");
    icon.className = "icon-action icon-package name-leading-icon";
    icon.setAttribute("aria-hidden", "true");
    button.appendChild(icon);
    button.appendChild(document.createTextNode(container.name || container.id.slice(0, 12)));
    button.addEventListener("click", () => {
      openContainersLogs(container);
    });
    containersLogsList.appendChild(button);
  });
  if (!selectedFound && containersLogsSelectedId) {
    containersLogsSelectedId = "";
    closeContainersLogsSession("selection missing");
  }
  if (sidebarSearch && sidebarSearch.value.trim()) {
    applySidebarFilter(sidebarSearch.value);
  }
}

function loadContainersResourcesPins() {
  if (containersResourcesPins) return containersResourcesPins;
  try {
    const raw = localStorage.getItem(containersResourcesPinsKey);
    const parsed = raw ? JSON.parse(raw) : {};
    if (!parsed || typeof parsed !== "object") {
      containersResourcesPins = {};
    } else {
      containersResourcesPins = parsed;
    }
  } catch {
    containersResourcesPins = {};
  }
  return containersResourcesPins;
}

function saveContainersResourcesPins() {
  if (!containersResourcesPins) return;
  try {
    localStorage.setItem(containersResourcesPinsKey, JSON.stringify(containersResourcesPins));
  } catch {
    // ignore
  }
}

function getContainersResourcesPins(scope) {
  if (!scope) return new Set();
  const store = loadContainersResourcesPins();
  const list = Array.isArray(store[scope]) ? store[scope] : [];
  return new Set(list);
}

function setContainersResourcesPins(scope, pins) {
  if (!scope) return;
  const store = loadContainersResourcesPins();
  store[scope] = Array.from(pins || []);
  containersResourcesPins = store;
  saveContainersResourcesPins();
}

function updateContainersResourcesPlaceholder(message) {
  if (!containersResourcesPlaceholder) return;
  const hasCards = Boolean(containersResourcesCards && containersResourcesCards.children.length > 0);
  const hasRows = Boolean(containersResourcesTableBody && containersResourcesTableBody.children.length > 0);
  const hasContent = containersResourcesViewMode === "table" ? hasRows : hasCards;
  if (hasContent) {
    containersResourcesPlaceholder.classList.add("hidden");
    if (containersResourcesCards) {
      containersResourcesCards.classList.toggle("hidden", containersResourcesViewMode !== "cards");
    }
    if (containersResourcesTableWrap) {
      containersResourcesTableWrap.classList.toggle("hidden", containersResourcesViewMode !== "table");
    }
    return;
  }
  containersResourcesPlaceholder.textContent = message || "Select containers to view resources.";
  containersResourcesPlaceholder.classList.remove("hidden");
  if (containersResourcesCards) {
    containersResourcesCards.classList.add("hidden");
  }
  if (containersResourcesTableWrap) {
    containersResourcesTableWrap.classList.add("hidden");
  }
}

function toggleContainersResourcesSelection(containerId) {
  if (!containerId) return;
  containersResourcesLastInteractionAtMs = Date.now();
  if (containersResourcesSelectedIds.has(containerId)) {
    containersResourcesSelectedIds.delete(containerId);
  } else {
    containersResourcesSelectedIds.add(containerId);
  }
  const list = Array.from(containersCache.values()).map((entry) => entry.data);
  renderContainersResourcesList(list);
  refreshContainersResources({ silent: true, force: true }).catch(() => {});
}

function toggleContainersResourcesPin(container) {
  const scope = containersSelectedScope;
  if (!scope || !container || !container.name) return;
  containersResourcesLastInteractionAtMs = Date.now();
  const pins = getContainersResourcesPins(scope);
  if (pins.has(container.name)) {
    pins.delete(container.name);
  } else {
    pins.add(container.name);
  }
  setContainersResourcesPins(scope, pins);
  renderContainersResourcesView();
  refreshContainersResources({ silent: true, force: true }).catch(() => {});
}

function renderContainersResourcesList(list) {
  if (!containersResourcesList) return;
  const prevScrollTop = containersResourcesList.scrollTop;
  containersResourcesList.innerHTML = "";
  if (!Array.isArray(list) || list.length === 0) {
    const empty = document.createElement("div");
    empty.className = "containers-shell-placeholder";
    empty.textContent = "No containers found.";
    containersResourcesList.appendChild(empty);
    containersResourcesSelectedIds.clear();
    renderContainersResourcesView();
    updateContainersResourcesPlaceholder();
    return;
  }
  const sorted = containersSort(list, containersSortMode);
  const existingIds = new Set();
  sorted.forEach((container) => {
    if (!container || !container.id) return;
    existingIds.add(container.id);
    const button = document.createElement("button");
    button.type = "button";
    button.className = "containers-shell-item";
    if (containersResourcesSelectedIds.has(container.id)) {
      button.classList.add("active");
    }
    button.dataset.id = container.id;
    button.dataset.search = `${container.name || ""} ${container.state || ""}`;
    button.textContent = "";
    const icon = document.createElement("span");
    icon.className = "icon-action icon-package name-leading-icon";
    icon.setAttribute("aria-hidden", "true");
    button.appendChild(icon);
    button.appendChild(document.createTextNode(container.name || container.id.slice(0, 12)));
    button.addEventListener("click", () => {
      toggleContainersResourcesSelection(container.id);
    });
    containersResourcesList.appendChild(button);
  });
  containersResourcesSelectedIds.forEach((id) => {
    if (!existingIds.has(id)) {
      containersResourcesSelectedIds.delete(id);
    }
  });
  renderContainersResourcesView();
  updateContainersResourcesPlaceholder();
  requestAnimationFrame(() => {
    if (containersResourcesList) {
      containersResourcesList.scrollTop = prevScrollTop;
    }
  });
  if (sidebarSearch && sidebarSearch.value.trim()) {
    applySidebarFilter(sidebarSearch.value);
  }
}

function getResourcesHostPorts(ports) {
  if (!Array.isArray(ports) || ports.length === 0) return [];
  const unique = new Set();
  ports.forEach((port) => {
    if (!port) return;
    const hostPort = Number(port.host_port || 0);
    if (hostPort > 0) unique.add(hostPort);
  });
  return Array.from(unique).sort((a, b) => a - b);
}

function formatResourcesIPValue(ips) {
  if (!Array.isArray(ips) || ips.length === 0) return "—";
  return ips.map((ip) => ip.ip).filter(Boolean).join(", ");
}

function formatResourcesIPTooltip(ips) {
  if (!Array.isArray(ips) || ips.length === 0) return "";
  return ips.map((ip) => `${ip.network}: ${ip.ip}`).filter(Boolean).join(", ");
}

function parseContainersResourcesSortMode(mode) {
  const [key, dir] = String(mode || "").split(":");
  const cleanKey = ["name", "cpu", "ram", "status", "uptime", "ip", "port"].includes(key) ? key : "name";
  const cleanDir = dir === "desc" ? "desc" : "asc";
  return { key: cleanKey, dir: cleanDir };
}

function compareContainersResourcesValues(a, b, dir) {
  const missingA = a === null || a === undefined || a === "";
  const missingB = b === null || b === undefined || b === "";
  if (missingA && missingB) return 0;
  if (missingA) return 1;
  if (missingB) return -1;
  if (typeof a === "number" && typeof b === "number") {
    return dir === "desc" ? b - a : a - b;
  }
  return dir === "desc"
    ? String(b).localeCompare(String(a), undefined, { numeric: true, sensitivity: "base" })
    : String(a).localeCompare(String(b), undefined, { numeric: true, sensitivity: "base" });
}

function renderContainersResourcesTable() {
  if (!containersResourcesTableBody) return;
  const scope = containersSelectedScope;
  const list = Array.from(containersCache.values()).map((entry) => entry.data);
  const byId = new Map(list.map((container) => [container.id, container]));
  const byName = new Map(list.map((container) => [container.name, container]));
  const pins = getContainersResourcesPins(scope);
  const pinned = [];
  pins.forEach((name) => {
    const container = byName.get(name);
    if (container) pinned.push(container);
  });
  const selected = [];
  containersResourcesSelectedIds.forEach((id) => {
    const container = byId.get(id);
    if (container && !pins.has(container.name)) {
      selected.push(container);
    }
  });

  const { key, dir } = parseContainersResourcesSortMode(containersResourcesSortMode);
  const getSortValue = (container) => {
    const resource = containersResourcesData.get(container.id);
    if (key === "name") return String(container.name || container.id || "");
    if (key === "cpu") return resource && Number.isFinite(Number(resource.cpu_percent)) ? Number(resource.cpu_percent) : null;
    if (key === "ram") return resource && Number.isFinite(Number(resource.mem_usage_bytes)) ? Number(resource.mem_usage_bytes) : null;
    if (key === "status") return String((resource && resource.state) ? resource.state : (container.state || ""));
    if (key === "uptime") {
      const val = resource && Number.isFinite(Number(resource.uptime_sec)) ? Number(resource.uptime_sec) : Number(container.uptime_sec);
      return Number.isFinite(val) ? val : null;
    }
    if (key === "ip") {
      const ips = resource && Array.isArray(resource.ip_addresses)
        ? resource.ip_addresses
        : Array.isArray(container.ip_addresses)
            ? container.ip_addresses
            : [];
      return ips.length > 0 ? String(ips[0].ip || "") : null;
    }
    if (key === "port") {
      const ports = getResourcesHostPorts(resource ? resource.ports : container.ports);
      return ports.length > 0 ? Number(ports[0]) : null;
    }
    return String(container.name || container.id || "");
  };
  const sorter = (a, b) => compareContainersResourcesValues(getSortValue(a), getSortValue(b), dir);
  pinned.sort(sorter);
  selected.sort(sorter);

  containersResourcesTableBody.innerHTML = "";
  const fragment = document.createDocumentFragment();
  const renderRow = (container) => {
    const resource = containersResourcesData.get(container.id);
    const row = document.createElement("tr");
    row.dataset.id = container.id;
    row.dataset.search = `${container.name || ""} ${container.state || ""}`;

    const pinCell = document.createElement("td");
    pinCell.className = "containers-resources-pin-cell";
    const pinBtn = document.createElement("button");
    pinBtn.type = "button";
    pinBtn.className = "secondary btn-small icon-action-btn has-tooltip containers-resources-pin-btn";
    const isPinned = pins.has(container.name);
    if (isPinned) {
      pinBtn.classList.add("is-active");
    }
    pinBtn.setAttribute("aria-label", isPinned ? "Unpin row" : "Pin row");
    pinBtn.setAttribute("data-tooltip", isPinned ? "Unpin row" : "Pin row");
    const pinIcon = document.createElement("span");
    pinIcon.className = `icon-action ${isPinned ? "icon-pinned-off" : "icon-pin"}`;
    pinIcon.setAttribute("aria-hidden", "true");
    pinBtn.appendChild(pinIcon);
    pinBtn.addEventListener("click", () => {
      toggleContainersResourcesPin(container);
    });
    pinCell.appendChild(pinBtn);

    const nameCell = document.createElement("td");
    nameCell.className = "containers-resources-name-cell";
    const nameIcon = document.createElement("span");
    nameIcon.className = "icon-action icon-package name-leading-icon";
    nameIcon.setAttribute("aria-hidden", "true");
    nameCell.appendChild(nameIcon);
    nameCell.appendChild(document.createTextNode(container.name || container.id.slice(0, 12)));

    const cpuCell = document.createElement("td");
    cpuCell.textContent = resource ? formatPercent(resource.cpu_percent) : "—";

    const ramCell = document.createElement("td");
    ramCell.textContent = resource ? formatBytesMBOrGB(resource.mem_usage_bytes) : "—";

    const statusCell = document.createElement("td");
    const stateValue = normalizeContainerState(resource && resource.state ? resource.state : container.state);
    const stateText = (resource && resource.state) ? resource.state : (container.state || "unknown");
    statusCell.className = "containers-state-cell";
    const badge = document.createElement("span");
    badge.className = `containers-state-badge ${stateValue ? `containers-state-${stateValue}` : "containers-state-unknown"}`;
    badge.textContent = stateText;
    statusCell.appendChild(badge);

    const uptimeCell = document.createElement("td");
    uptimeCell.textContent = formatUptime(resource ? resource.uptime_sec : container.uptime_sec);

    const ipCell = document.createElement("td");
    const ipAddresses = resource && Array.isArray(resource.ip_addresses)
      ? resource.ip_addresses
      : Array.isArray(container.ip_addresses)
          ? container.ip_addresses
          : [];
    ipCell.textContent = ipAddresses.length > 0 ? formatResourcesIPValue(ipAddresses) : "—";

    const portCell = document.createElement("td");
    const hostPorts = getResourcesHostPorts(resource ? resource.ports : container.ports);
    const scopeInfo = getScopeInfo(containersSelectedScope);
    const publicHost = scopeInfo && scopeInfo.server ? String(scopeInfo.server.public_ip || "").trim() : "";
    if (hostPorts.length === 0) {
      portCell.textContent = "—";
    } else {
      hostPorts.forEach((port, index) => {
        if (index > 0) {
          portCell.appendChild(document.createTextNode(", "));
        }
        const text = String(port);
        if (publicHost) {
          const link = document.createElement("a");
          link.href = `http://${publicHost}:${port}`;
          link.target = "_blank";
          link.rel = "noreferrer";
          link.className = "containers-resource-link";
          link.textContent = text;
          portCell.appendChild(link);
        } else {
          portCell.appendChild(document.createTextNode(text));
        }
      });
    }

    row.append(pinCell, nameCell, statusCell, uptimeCell, cpuCell, ramCell, ipCell, portCell);
    return row;
  };

  pinned.forEach((container) => fragment.appendChild(renderRow(container)));
  if (pinned.length > 0 && selected.length > 0) {
    const dividerRow = document.createElement("tr");
    dividerRow.className = "containers-resources-pinned-divider";
    const cell = document.createElement("td");
    cell.colSpan = 8;
    const line = document.createElement("div");
    line.className = "containers-resources-pinned-divider-line";
    cell.appendChild(line);
    dividerRow.appendChild(cell);
    fragment.appendChild(dividerRow);
  }
  selected.forEach((container) => fragment.appendChild(renderRow(container)));
  containersResourcesTableBody.appendChild(fragment);
}

function renderContainersResourcesCards() {
  if (!containersResourcesCards) return;
  const scope = containersSelectedScope;
  const list = Array.from(containersCache.values()).map((entry) => entry.data);
  const byId = new Map(list.map((container) => [container.id, container]));
  const byName = new Map(list.map((container) => [container.name, container]));
  const pins = getContainersResourcesPins(scope);
  const pinned = [];
  pins.forEach((name) => {
    const container = byName.get(name);
    if (container) pinned.push(container);
  });
  pinned.sort((a, b) => String(a.name || "").localeCompare(String(b.name || "")));
  const selected = [];
  containersResourcesSelectedIds.forEach((id) => {
    const container = byId.get(id);
    if (container && !pins.has(container.name)) {
      selected.push(container);
    }
  });
  selected.sort((a, b) => String(a.name || "").localeCompare(String(b.name || "")));
  const cards = [...pinned, ...selected];
  containersResourcesCards.innerHTML = "";
  if (cards.length === 0) {
    updateContainersResourcesPlaceholder();
    return;
  }
  cards.forEach((container) => {
    const resource = containersResourcesData.get(container.id);
    const stateValue = normalizeContainerState(resource && resource.state ? resource.state : container.state);
    const stateText = (resource && resource.state) ? resource.state : (container.state || "unknown");
    const card = document.createElement("div");
    card.className = "containers-resource-card";
    card.dataset.id = container.id;
    const header = document.createElement("div");
    header.className = "containers-resource-card-header";
    const title = document.createElement("div");
    title.className = "containers-resource-card-title";
    const icon = document.createElement("span");
    icon.className = "icon-action icon-package";
    icon.setAttribute("aria-hidden", "true");
    const name = document.createElement("span");
    name.textContent = container.name || container.id.slice(0, 12);
    title.append(icon, name);
    const actions = document.createElement("div");
    actions.className = "containers-resource-card-actions";
    const badge = document.createElement("span");
    badge.className = `containers-state-badge ${stateValue ? `containers-state-${stateValue}` : "containers-state-unknown"}`;
    badge.textContent = stateText;
    const pinBtn = document.createElement("button");
    pinBtn.type = "button";
    pinBtn.className = "secondary btn-small icon-action-btn has-tooltip";
    const isPinned = pins.has(container.name);
    if (isPinned) {
      pinBtn.classList.add("is-active");
    }
    pinBtn.setAttribute("aria-label", isPinned ? "Unpin card" : "Pin card");
    pinBtn.setAttribute("data-tooltip", isPinned ? "Unpin card" : "Pin card");
    const pinIcon = document.createElement("span");
    pinIcon.className = `icon-action ${isPinned ? "icon-pinned-off" : "icon-pin"}`;
    pinIcon.setAttribute("aria-hidden", "true");
    pinBtn.appendChild(pinIcon);
    pinBtn.addEventListener("click", () => {
      toggleContainersResourcesPin(container);
    });
    actions.append(badge, pinBtn);
    header.append(title, actions);
    const metrics = document.createElement("div");
    metrics.className = "containers-resource-metrics-line";
    const cpuIcon = document.createElement("span");
    cpuIcon.className = "icon-action icon-cpu has-tooltip";
    cpuIcon.setAttribute("aria-hidden", "true");
    cpuIcon.setAttribute("data-tooltip", "CPU | RAM values");
    cpuIcon.setAttribute("aria-label", "CPU | RAM values");
    const cpuValue = document.createElement("span");
    cpuValue.className = "containers-resource-cpu-value";
    cpuValue.textContent = resource ? formatPercent(resource.cpu_percent) : "—";
    const ramValue = document.createElement("span");
    ramValue.className = "containers-resource-ram-value";
    ramValue.textContent = resource ? formatBytesMBOrGB(resource.mem_usage_bytes) : "—";
    const sep1 = document.createElement("span");
    sep1.className = "containers-resource-separator";
    sep1.textContent = "/";
    const sep2 = document.createElement("span");
    sep2.className = "containers-resource-separator";
    sep2.textContent = "/";
    const uptimeIcon = document.createElement("span");
    uptimeIcon.className = "icon-action icon-stopwatch has-tooltip";
    uptimeIcon.setAttribute("aria-hidden", "true");
    uptimeIcon.setAttribute("data-tooltip", "Uptime");
    uptimeIcon.setAttribute("aria-label", "Uptime");
    const uptimeValue = document.createElement("span");
    uptimeValue.className = "containers-resource-uptime-value";
    uptimeValue.textContent = formatUptime(resource ? resource.uptime_sec : container.uptime_sec);
    metrics.append(cpuIcon, cpuValue, sep1, ramValue, sep2, uptimeIcon, uptimeValue);
    const details = document.createElement("div");
    details.className = "containers-resource-details";
    const ip = document.createElement("div");
    ip.className = "containers-resource-detail";
    const ipIcon = document.createElement("span");
    ipIcon.className = "icon-action icon-world-bolt has-tooltip";
    ipIcon.setAttribute("aria-hidden", "true");
    const ipValue = document.createElement("span");
    ipValue.className = "containers-resource-value";
    const ipAddresses = resource && Array.isArray(resource.ip_addresses)
      ? resource.ip_addresses
      : Array.isArray(container.ip_addresses)
          ? container.ip_addresses
          : [];
    ipValue.textContent = ipAddresses.length > 0 ? formatResourcesIPValue(ipAddresses) : "—";
    const ipTooltip = ipAddresses.length > 0 ? formatResourcesIPTooltip(ipAddresses) : "";
    if (ipTooltip) {
      ipIcon.setAttribute("data-tooltip", ipTooltip);
      ipIcon.setAttribute("aria-label", ipTooltip);
    }
    ip.append(ipIcon, ipValue);
    const ports = document.createElement("div");
    ports.className = "containers-resource-detail";
    const portsIcon = document.createElement("span");
    portsIcon.className = "icon-action icon-plug has-tooltip";
    portsIcon.setAttribute("aria-hidden", "true");
    portsIcon.setAttribute("data-tooltip", "Ports");
    portsIcon.setAttribute("aria-label", "Ports");
    const portsValue = document.createElement("span");
    portsValue.className = "containers-resource-value";
    const hostPorts = getResourcesHostPorts(resource ? resource.ports : container.ports);
    const scopeInfo = getScopeInfo(containersSelectedScope);
    const publicHost = scopeInfo && scopeInfo.server ? String(scopeInfo.server.public_ip || "").trim() : "";
    if (hostPorts.length === 0) {
      portsValue.textContent = "—";
    } else {
      hostPorts.forEach((port, index) => {
        if (index > 0) {
          portsValue.append(document.createTextNode(", "));
        }
        const text = String(port);
        if (publicHost) {
          const link = document.createElement("a");
          link.href = `http://${publicHost}:${port}`;
          link.target = "_blank";
          link.rel = "noreferrer";
          link.className = "containers-resource-link";
          link.textContent = text;
          portsValue.appendChild(link);
        } else {
          portsValue.appendChild(document.createTextNode(text));
        }
      });
    }
    ports.append(portsIcon, portsValue);
    details.append(ip, ports);
    card.append(header, metrics, details);
    containersResourcesCards.appendChild(card);
  });
}

function renderContainersResourcesView() {
  if (containersResourcesViewMode === "table") {
    if (containersResourcesCards) containersResourcesCards.innerHTML = "";
    renderContainersResourcesTable();
  } else {
    if (containersResourcesTableBody) containersResourcesTableBody.innerHTML = "";
    renderContainersResourcesCards();
  }
  updateContainersResourcesPlaceholder();
}

function showContainersShellPlaceholder(message) {
  if (containersShellPlaceholder) {
    containersShellPlaceholder.textContent = message;
    containersShellPlaceholder.classList.remove("hidden");
  }
  if (containersTerminalEl) {
    containersTerminalEl.classList.add("hidden");
  }
  if (containersShellTerminalEl) {
    containersShellTerminalEl.classList.remove("shell-active");
  }
}

function hideContainersShellPlaceholder() {
  if (containersShellPlaceholder) {
    containersShellPlaceholder.classList.add("hidden");
  }
  if (containersTerminalEl) {
    containersTerminalEl.classList.remove("hidden");
  }
  if (containersShellTerminalEl) {
    containersShellTerminalEl.classList.add("shell-active");
  }
}

function showContainersLogsPlaceholder(message) {
  if (containersLogsPlaceholder) {
    containersLogsPlaceholder.textContent = message;
    containersLogsPlaceholder.classList.remove("hidden");
  }
  if (containersLogsView) {
    containersLogsView.classList.add("hidden");
  }
}

function hideContainersLogsPlaceholder() {
  if (containersLogsPlaceholder) {
    containersLogsPlaceholder.classList.add("hidden");
  }
  if (containersLogsView) {
    containersLogsView.classList.remove("hidden");
  }
}

function updateContainersLogsStreamIndicator() {
  if (!containersLogsStreamEl) return;
  containersLogsStreamEl.innerHTML = "";
  containersLogsStreamEl.classList.remove("status-icon-online", "status-icon-maintenance");
  containersLogsStreamEl.classList.add(containersLogsStreamActive ? "status-icon-online" : "status-icon-maintenance");
  const icon = document.createElement("span");
  icon.className = `icon-action ${containersLogsStreamActive ? "icon-wifi" : "icon-wifi-off"}`;
  icon.setAttribute("aria-hidden", "true");
  containersLogsStreamEl.appendChild(icon);
  const label = containersLogsStreamActive ? "Log Live Stream Active" : "Log Live Stream Inactive";
  containersLogsStreamEl.setAttribute("data-tooltip", label);
  containersLogsStreamEl.setAttribute("aria-label", label);
}

function setContainersLogsStreamActive(active) {
  containersLogsStreamActive = Boolean(active);
  updateContainersLogsStreamIndicator();
}

function resetContainersLogsBuffer() {
  containersLogsBuffer = "";
  containersLogsPending = "";
  containersLogsLineBuffer = "";
  containersLogsBytes = 0;
  containersLogsAutoScroll = true;
  if (containersLogsOutput) {
    containersLogsOutput.textContent = "";
    containersLogsOutput.scrollTop = 0;
  }
}

function formatLogTimestamp(date) {
  if (!(date instanceof Date) || Number.isNaN(date.getTime())) return "";
  const options = {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    hour12: false,
  };
  if (currentTimeZone) {
    try {
      return new Intl.DateTimeFormat("sv-SE", { ...options, timeZone: currentTimeZone }).format(date);
    } catch {
      // fallback to browser timezone
    }
  }
  return new Intl.DateTimeFormat("sv-SE", options).format(date);
}

function normalizeLogLineTimestamp(line) {
  const match = line.match(/^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})?)\s*/);
  if (!match) return line;
  const raw = match[1];
  const date = new Date(raw);
  const formatted = formatLogTimestamp(date);
  if (!formatted) return line;
  const rest = line.slice(match[0].length);
  return rest ? `${formatted} ${rest}` : formatted;
}

function processContainersLogsChunk(chunk) {
  if (!chunk) return;
  const text = typeof chunk === "string" ? chunk : String(chunk);
  if (!containersLogsTimestamps) {
    appendContainersLogs(text);
    return;
  }
  containersLogsLineBuffer += text;
  const lines = containersLogsLineBuffer.split("\n");
  containersLogsLineBuffer = lines.pop() || "";
  lines.forEach((line) => {
    appendContainersLogs(`${normalizeLogLineTimestamp(line)}\n`);
  });
}

function appendContainersLogs(data) {
  if (!data) return;
  const text = typeof data === "string" ? data : String(data);
  if (containersLogsPaused) {
    containersLogsPending += text;
    if (containersLogsPending.length > Math.floor(containersLogsMaxBytes / 2)) {
      const excess = containersLogsPending.length - Math.floor(containersLogsMaxBytes / 2);
      containersLogsPending = containersLogsPending.slice(excess);
    }
    return;
  }
  containersLogsBuffer += text;
  containersLogsBytes = containersLogsBuffer.length;
  if (containersLogsBytes > containersLogsMaxBytes) {
    const excess = containersLogsBytes - containersLogsMaxBytes;
    containersLogsBuffer = containersLogsBuffer.slice(excess);
    containersLogsBytes = containersLogsBuffer.length;
  }
  if (containersLogsOutput) {
    containersLogsOutput.textContent = containersLogsBuffer;
    if (containersLogsAutoScroll) {
      containersLogsOutput.scrollTop = containersLogsOutput.scrollHeight;
    }
  }
}

function setContainersLogsPaused(paused) {
  containersLogsPaused = Boolean(paused);
  if (containersLogsPauseBtn) {
    const icon = containersLogsPauseBtn.querySelector(".icon-action");
    if (icon) {
      icon.classList.toggle("icon-pause", !containersLogsPaused);
      icon.classList.toggle("icon-play", containersLogsPaused);
    }
    containersLogsPauseBtn.classList.toggle("is-active", containersLogsPaused);
    const label = containersLogsPaused ? "Resume logs" : "Pause logs";
    containersLogsPauseBtn.setAttribute("aria-label", label);
    containersLogsPauseBtn.setAttribute("data-tooltip", label);
  }
  if (!containersLogsPaused && containersLogsPending) {
    const pending = containersLogsPending;
    containersLogsPending = "";
    appendContainersLogs(pending);
  }
}

function setContainersLogsTimestamps(enabled) {
  containersLogsTimestamps = Boolean(enabled);
  if (containersLogsTimestampsBtn) {
    const icon = containersLogsTimestampsBtn.querySelector(".icon-action");
    if (icon) {
      icon.classList.toggle("icon-clock", !containersLogsTimestamps);
      icon.classList.toggle("icon-clock-off", containersLogsTimestamps);
    }
    containersLogsTimestampsBtn.classList.toggle("is-active", containersLogsTimestamps);
    const tzInfo = currentTimeZone ? ` (TZ: ${currentTimeZone})` : "";
    const label = containersLogsTimestamps ? "Timestamp off" : "Timestamp on";
    containersLogsTimestampsBtn.setAttribute("aria-label", `${label}${tzInfo}`);
    containersLogsTimestampsBtn.setAttribute("data-tooltip", `${label}${tzInfo}`);
  }
}

function closeContainersLogsSession(reason, options = {}) {
  const preserve = Boolean(options && options.preserve);
  const silent = Boolean(options && options.silent);
  if (reason) {
    console.debug(`[containers-logs] closing: ${reason}`);
  }
  setContainersLogsStreamActive(false);
  if (containersLogsSocket) {
    containersLogsSilentClose = silent;
    try {
      containersLogsSocket.close();
    } catch (err) {
      // ignore
    }
    containersLogsSocket = null;
  }
  if (!preserve) {
    resetContainersLogsBuffer();
  }
  showContainersLogsPlaceholder("Select a container to view logs.");
}

function closeContainersShellSession(reason) {
  if (reason) {
    console.debug(`[containers-shell] closing: ${reason}`);
  }
  if (containersShellSocket) {
    try {
      containersShellSocket.close();
    } catch (err) {
      // ignore
    }
    containersShellSocket = null;
  }
  if (containersShellDataListener) {
    containersShellDataListener.dispose();
    containersShellDataListener = null;
  }
  if (containersShellTerm) {
    containersShellTerm.dispose();
    containersShellTerm = null;
    containersShellFitAddon = null;
  }
  showContainersShellPlaceholder("Select a container to open shell.");
}

function scheduleContainersShellResize() {
  if (containersShellResizeTimer) {
    window.clearTimeout(containersShellResizeTimer);
  }
  containersShellResizeTimer = window.setTimeout(() => {
    if (!containersShellSocket || containersShellSocket.readyState !== WebSocket.OPEN) return;
    if (containersShellFitAddon && typeof containersShellFitAddon.fit === "function") {
      containersShellFitAddon.fit();
    }
    if (!containersShellTerm) return;
    const cols = containersShellTerm.cols || 80;
    const rows = containersShellTerm.rows || 24;
    containersShellSocket.send(JSON.stringify({ type: "resize", cols, rows }));
  }, 120);
}

function openContainersShell(container) {
  if (!container || !container.id) return;
  containersShellSelectedId = container.id;
  renderContainersShellList(Array.from(containersCache.values()).map((entry) => entry.data));
  const state = normalizeContainerState(container.state);
  if (state !== "running") {
    closeContainersShellSession("container not running");
    showContainersShellPlaceholder("Container is not running.");
    return;
  }
  if (!containersTerminalEl) return;
  if (typeof window.Terminal === "undefined") {
    showContainersShellPlaceholder("Terminal library not loaded.");
    return;
  }
  closeContainersShellSession("open new shell");
  if (containersShellTerm) {
    containersShellTerm.dispose();
    containersShellTerm = null;
    containersShellFitAddon = null;
  }
  const shellBin = container.shell || "/bin/sh";
  const shellUser = container.user || "root";
  containersShellTerm = new window.Terminal({
    convertEol: true,
    scrollback: 2000,
    theme: {
      background: "transparent",
      foreground: "#e2e8f0",
      cursor: "#94a3b8",
      selection: "#334155",
      black: "#0b0f1a",
      red: "#ef4444",
      green: "#22c55e",
      yellow: "#eab308",
      blue: "#3b82f6",
      magenta: "#a855f7",
      cyan: "#06b6d4",
      white: "#e2e8f0",
    },
  });
  if (typeof window.FitAddon !== "undefined") {
    containersShellFitAddon = new window.FitAddon.FitAddon();
    containersShellTerm.loadAddon(containersShellFitAddon);
  }
  containersTerminalEl.innerHTML = "";
  containersShellTerm.open(containersTerminalEl);
  containersShellTerm.focus();
  containersShellTerm.reset();
  containersShellTerm.writeln(`Connecting to "${container.name || container.id.slice(0, 12)}"...`);
  containersShellTerm.writeln(`Shell: ${shellBin}, User: ${shellUser}`);
  containersShellTerm.writeln("");
  hideContainersShellPlaceholder();

  const wsUrl = `/api/containers/shell?scope=${encodeURIComponent(containersSelectedScope)}&container_id=${encodeURIComponent(container.id)}`;
  const ws = new WebSocket(wsUrl);
  ws.binaryType = "arraybuffer";
  containersShellSocket = ws;

  ws.addEventListener("open", () => {
    if (containersShellSocket !== ws) return;
    console.debug("[containers-shell] ws open");
    scheduleContainersShellResize();
  });

  ws.addEventListener("message", (event) => {
    if (containersShellSocket !== ws) return;
    if (!containersShellTerm) return;
    if (typeof event.data === "string") {
      try {
        const payload = JSON.parse(event.data);
        if (payload && payload.type === "error" && payload.message) {
          showToast(payload.message);
          return;
        }
      } catch (err) {
        containersShellTerm.write(event.data);
        return;
      }
      return;
    }
    const data = event.data instanceof ArrayBuffer ? new Uint8Array(event.data) : new Uint8Array();
    if (data.length > 0) {
      containersShellTerm.write(data);
    }
  });

  ws.addEventListener("close", (event) => {
    if (containersShellSocket !== ws) return;
    console.debug(`[containers-shell] ws closed code=${event.code} reason=${event.reason || "n/a"}`);
    showContainersShellPlaceholder("Shell disconnected.");
  });

  ws.addEventListener("error", () => {
    if (containersShellSocket !== ws) return;
    console.debug("[containers-shell] ws error");
    showContainersShellPlaceholder("Shell connection failed.");
  });

  containersShellDataListener = containersShellTerm.onData((data) => {
    if (containersShellSocket !== ws) return;
    if (!containersShellSocket || containersShellSocket.readyState !== WebSocket.OPEN) return;
    containersShellSocket.send(containersShellEncoder.encode(data));
  });
  if (!containersShellResizeHandler) {
    containersShellResizeHandler = () => scheduleContainersShellResize();
    window.addEventListener("resize", containersShellResizeHandler);
  }
}

function openContainersLogs(container, options = {}) {
  if (!container || !container.id) return;
  containersLogsSelectedId = container.id;
  renderContainersLogsList(Array.from(containersCache.values()).map((entry) => entry.data));
  if (!containersLogsOutput) {
    showContainersLogsPlaceholder("Logs panel unavailable.");
    return;
  }
  setContainersLogsPaused(false);
  setContainersLogsTimestamps(containersLogsTimestamps);
  setContainersLogsStreamActive(false);
  const preserve = Boolean(options && options.preserve);
  const silentClose = Boolean(options && options.silentClose);
  closeContainersLogsSession("open new logs", { preserve, silent: silentClose });
  if (!preserve) {
    resetContainersLogsBuffer();
  }
  containersLogsAutoScroll = true;
  hideContainersLogsPlaceholder();
  appendContainersLogs(`Connecting to "${container.name || container.id.slice(0, 12)}"...\n`);

  updateContainersLogsStreamIndicator();
  const timestamps = containersLogsTimestamps ? "1" : "0";
  const wsUrl = `/api/containers/logs?scope=${encodeURIComponent(containersSelectedScope)}&container_id=${encodeURIComponent(container.id)}&follow=1&tail=100&timestamps=${timestamps}`;
  const ws = new WebSocket(wsUrl);
  containersLogsSocket = ws;

  ws.addEventListener("open", () => {
    if (containersLogsSocket !== ws) return;
    console.debug("[containers-logs] ws open");
    setContainersLogsStreamActive(true);
  });

  ws.addEventListener("message", (event) => {
    if (containersLogsSocket !== ws) return;
    if (typeof event.data === "string") {
      try {
        const payload = JSON.parse(event.data);
        if (payload && payload.type === "error" && payload.message) {
          showToast(payload.message);
          return;
        }
      } catch (err) {
        processContainersLogsChunk(event.data);
        return;
      }
      return;
    }
    if (event.data instanceof ArrayBuffer) {
      const decoder = new TextDecoder("utf-8");
      processContainersLogsChunk(decoder.decode(event.data));
    }
  });

  ws.addEventListener("close", (event) => {
    if (containersLogsSocket !== ws) return;
    setContainersLogsStreamActive(false);
    if (containersLogsSilentClose) {
      containersLogsSilentClose = false;
      return;
    }
    console.debug(`[containers-logs] ws closed code=${event.code} reason=${event.reason || "n/a"}`);
    appendContainersLogs("\n[logs disconnected]\n");
  });

  ws.addEventListener("error", () => {
    if (containersLogsSocket !== ws) return;
    setContainersLogsStreamActive(false);
    console.debug("[containers-logs] ws error");
    appendContainersLogs("\n[logs connection failed]\n");
  });
}

function resolveServerStatusLabel(info) {
  if (!info) return "unknown";
  if (info.maintenance) return "maintenance";
  if (info.status === "offline") return "offline";
  if (info.status === "restarting") return "restarting";
  if (info.checking && (!info.status || info.status === "unknown")) return "checking";
  if (!info.status) return "unknown";
  return "online";
}

function buildServerStatusIcon(statusLabel) {
  const icon = document.createElementNS("http://www.w3.org/2000/svg", "svg");
  icon.setAttribute(
    "class",
    "status-card-icon icon icon-tabler icons-tabler-outline icon-tabler-server-bolt"
  );
  icon.setAttribute("width", "24");
  icon.setAttribute("height", "24");
  icon.setAttribute("viewBox", "0 0 24 24");
  icon.setAttribute("fill", "none");
  icon.setAttribute("stroke", "currentColor");
  icon.setAttribute("stroke-width", "2");
  icon.setAttribute("stroke-linecap", "round");
  icon.setAttribute("stroke-linejoin", "round");
  icon.setAttribute("aria-hidden", "true");
  icon.innerHTML = '<path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M3 7a3 3 0 0 1 3 -3h12a3 3 0 0 1 3 3v2a3 3 0 0 1 -3 3h-12a3 3 0 0 1 -3 -3" /><path d="M15 20h-9a3 3 0 0 1 -3 -3v-2a3 3 0 0 1 3 -3h12" /><path d="M7 8v.01" /><path d="M7 16v.01" /><path d="M20 15l-2 3h3l-2 3" />';
  return icon;
}

function buildServerTypeIcon(serverType, statusLabel) {
  const icon = document.createElementNS("http://www.w3.org/2000/svg", "svg");
  icon.setAttribute(
    "class",
    `status-card-type-icon status-icon status-icon-${statusLabel || "unknown"} icon icon-tabler icons-tabler-outline icon-tabler-${serverType === "local" ? "plug-connected" : "network"}`
  );
  icon.setAttribute("width", "24");
  icon.setAttribute("height", "24");
  icon.setAttribute("viewBox", "0 0 24 24");
  icon.setAttribute("fill", "none");
  icon.setAttribute("stroke", "currentColor");
  icon.setAttribute("stroke-width", "2");
  icon.setAttribute("stroke-linecap", "round");
  icon.setAttribute("stroke-linejoin", "round");
  icon.setAttribute("aria-hidden", "true");
  if (serverType === "local") {
    icon.innerHTML = '<path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M7 12l5 5l-1.5 1.5a3.536 3.536 0 1 1 -5 -5l1.5 -1.5" /><path d="M17 12l-5 -5l1.5 -1.5a3.536 3.536 0 1 1 5 5l-1.5 1.5" /><path d="M3 21l2.5 -2.5" /><path d="M18.5 5.5l2.5 -2.5" /><path d="M10 11l-2 2" /><path d="M13 14l-2 2" />';
  } else {
    icon.innerHTML = '<path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M6 9a6 6 0 1 0 12 0a6 6 0 0 0 -12 0" /><path d="M12 3c1.333 .333 2 2.333 2 6s-.667 5.667 -2 6" /><path d="M12 3c-1.333 .333 -2 2.333 -2 6s.667 5.667 2 6" /><path d="M6 9h12" /><path d="M3 20h7" /><path d="M14 20h7" /><path d="M10 20a2 2 0 1 0 4 0a2 2 0 0 0 -4 0" /><path d="M12 15v3" />';
  }
  return icon;
}

function renderContainersServerButton(info) {
  if (!topbarContainersServerBtn) return;
  const content = topbarContainersServerBtn.querySelector(".containers-server-btn-content");
  if (!content) return;
  content.innerHTML = "";
  if (!info || !info.server) {
    content.textContent = "Select server";
    return;
  }
  const statusLabel = resolveServerStatusLabel(info);
  const statusIcon = buildServerStatusIcon(statusLabel);
  const nameEl = document.createElement("span");
  nameEl.className = "containers-server-name";
  nameEl.textContent = info.server.name;
  const typeIcon = buildServerTypeIcon(info.type, statusLabel);
  content.append(statusIcon, nameEl, typeIcon);
}

function closeContainersServerMenu() {
  if (!topbarContainersServerMenu || !topbarContainersServerBtn) return;
  topbarContainersServerMenu.classList.add("hidden");
  topbarContainersServerBtn.setAttribute("aria-expanded", "false");
}

function toggleContainersServerMenu() {
  if (!topbarContainersServerMenu || !topbarContainersServerBtn) return;
  const isOpen = !topbarContainersServerMenu.classList.contains("hidden");
  if (isOpen) {
    closeContainersServerMenu();
  } else {
    topbarContainersServerMenu.classList.remove("hidden");
    topbarContainersServerBtn.setAttribute("aria-expanded", "true");
  }
}

function updateContainersServerOptions() {
  if (!topbarContainersServerMenu) return;
  const options = [
    ...cachedLocals.map((server) => ({
      scope: `local:${server.name}`,
      label: server.name,
      type: "local",
      server,
    })),
    ...cachedRemotes.map((server) => ({
      scope: `remote:${server.name}`,
      label: server.name,
      type: "remote",
      server,
    })),
  ];
  const previous = containersSelectedScope;
  topbarContainersServerMenu.innerHTML = "";
  options.forEach((item) => {
    const info = getScopeInfo(item.scope) || { type: item.type, server: item.server, status: "", maintenance: false, checking: false };
    const statusLabel = resolveServerStatusLabel(info);
    const button = document.createElement("button");
    button.type = "button";
    button.className = "containers-server-option";
    if (item.scope === previous) {
      button.classList.add("active");
    }
    const statusIcon = buildServerStatusIcon(statusLabel);
    const nameEl = document.createElement("span");
    nameEl.className = "containers-server-name";
    nameEl.textContent = item.label;
    const typeIcon = buildServerTypeIcon(item.type, statusLabel);
	    button.append(statusIcon, nameEl, typeIcon);
      button.addEventListener("click", async () => {
        containersSelectedScope = item.scope;
        topbarContainersServerMenu.querySelectorAll(".containers-server-option").forEach((el) => {
          el.classList.remove("active");
        });
        button.classList.add("active");
      renderContainersServerButton(info);
      closeContainersServerMenu();
      updateContainersLogsStreamIndicator();
      if (containersViewMode === "resources") {
        containersResourcesSelectedIds.clear();
        containersResourcesData = new Map();
        updateContainersResourcesPlaceholder("Select containers to view resources.");
      }
      if (containersViewMode === "stacks") {
        await refreshStacks({ silent: true });
      } else if (containersViewMode === "images") {
        await refreshImages({ silent: true });
      } else {
        await refreshContainers({ silent: true });
        if (containersViewMode === "resources") {
          await refreshContainersResources({ silent: true });
        }
      }
    });
    topbarContainersServerMenu.appendChild(button);
  });
  const fallback = options.length > 0 ? options[0].scope : "";
  const next = options.find((item) => item.scope === previous) ? previous : fallback;
  containersSelectedScope = next;
  if (next) {
    const info = getScopeInfo(next);
    renderContainersServerButton(info);
  } else {
    renderContainersServerButton(null);
  }
  if (currentView === "containers") {
    if (containersViewMode === "stacks") {
      refreshStacks({ silent: true }).catch(() => {});
    } else if (containersViewMode === "images") {
      refreshImages({ silent: true }).catch(() => {});
    } else {
      refreshContainers({ silent: true }).catch(() => {});
      if (containersViewMode === "resources") {
        refreshContainersResources({ silent: true }).catch(() => {});
      }
    }
  }
}

function buildContainerActionButton(iconClass, label, onClick) {
  const button = document.createElement("button");
  button.type = "button";
  button.className = "secondary icon-action-btn has-tooltip";
  button.setAttribute("aria-label", label);
  button.setAttribute("data-tooltip", label);
  const icon = document.createElement("span");
  icon.className = `icon-action ${iconClass}`;
  icon.setAttribute("aria-hidden", "true");
  button.appendChild(icon);
  button.addEventListener("click", onClick);
  return button;
}

function setContainerOptimisticState(containerId, state, scope) {
  if (!containerId) return;
  const entry = containersCache.get(containerId);
  if (!entry || !entry.data) return;
  entry.data = { ...entry.data, state };
  updateContainerRow(entry.row, entry.data, scope);
  if (containersSortMode.startsWith("state:")) {
    const list = Array.from(containersCache.values()).map((item) => item.data);
    if (list.length > 0 && containersSelectedScope) {
      renderContainers(list, containersSelectedScope);
    }
  }
}

function clearContainersKillConfirmations() {
  if (containersKillConfirming.size === 0) return;
  containersKillConfirming.clear();
  if (!containersSelectedScope) return;
  const list = Array.from(containersCache.values()).map((entry) => entry.data);
  if (list.length > 0) {
    renderContainers(list, containersSelectedScope);
  }
}

function stackActionKey(name, action) {
  return `${name || ""}::${action || ""}`;
}

function clearStacksActionConfirmations() {
  if (stacksActionConfirming.size === 0) return;
  stacksActionConfirming.clear();
  const list = Array.from(stacksCache.values()).map((entry) => entry && entry.data).filter(Boolean);
  if (list.length > 0) {
    renderStacks(list);
  }
}

function updateContainerRow(row, container, scope) {
  row.dataset.id = container.id;
  const resource = containersTableResourcesData.get(container.id);
  const hostPorts = getResourcesHostPorts(resource ? resource.ports : container.ports);
  const ipAddresses = resource && Array.isArray(resource.ip_addresses)
    ? resource.ip_addresses
    : Array.isArray(container.ip_addresses)
        ? container.ip_addresses
        : [];
  const ipSearchValue = ipAddresses.length > 0 ? ipAddresses.map((ip) => ip.ip).filter(Boolean).join(" ") : "";
  row.dataset.search = `${container.name || ""} ${container.image || ""} ${container.stack || ""} ${container.state || ""} ${ipSearchValue} ${hostPorts.join(" ")}`;
  const cells = row.querySelectorAll("td");
  if (cells.length < 10) return;
  // 0 Name, 1 Image, 2 State, 3 Uptime, 4 CPU, 5 RAM, 6 IP, 7 Port, 8 Stack, 9 Actions
  cells[0].textContent = "";
  const nameIcon = document.createElement("span");
  nameIcon.className = "icon-action icon-package name-leading-icon";
  nameIcon.setAttribute("aria-hidden", "true");
  cells[0].appendChild(nameIcon);
  cells[0].appendChild(document.createTextNode(container.name || container.id.slice(0, 12)));
  cells[1].textContent = container.image || "";

  const stateValue = normalizeContainerState(container.state);
  const stateText = container.state || "unknown";
  cells[2].textContent = "";
  const stateBadge = document.createElement("span");
  stateBadge.className = `containers-state-badge ${stateValue ? `containers-state-${stateValue}` : "containers-state-unknown"}`;
  stateBadge.textContent = stateText;
  cells[2].className = "containers-state-cell";
  cells[2].appendChild(stateBadge);

  cells[3].textContent = formatUptime(container.uptime_sec);
  cells[4].textContent = resource ? formatPercent(resource.cpu_percent) : "—";
  cells[5].textContent = resource ? formatBytesMBOrGB(resource.mem_usage_bytes) : "—";
  cells[6].textContent = ipAddresses.length > 0 ? formatResourcesIPValue(ipAddresses) : "—";

  cells[7].textContent = "";
  if (hostPorts.length === 0) {
    cells[7].textContent = "—";
  } else {
    const scopeInfo = getScopeInfo(scope);
    const publicHost = scopeInfo && scopeInfo.server ? String(scopeInfo.server.public_ip || "").trim() : "";
    hostPorts.forEach((port, index) => {
      if (index > 0) {
        cells[7].appendChild(document.createTextNode(", "));
      }
      if (publicHost) {
        const link = document.createElement("a");
        link.href = `http://${publicHost}:${port}`;
        link.target = "_blank";
        link.rel = "noreferrer";
        link.className = "containers-resource-link";
        link.textContent = String(port);
        cells[7].appendChild(link);
      } else {
        cells[7].appendChild(document.createTextNode(String(port)));
      }
    });
  }

  cells[8].textContent = container.stack || "-";
  row.dataset.state = stateValue;
  const actionsCell = cells[9];
  let actionsWrap = actionsCell.querySelector(".containers-actions");
  if (!actionsWrap) {
    actionsWrap = document.createElement("div");
    actionsWrap.className = "containers-actions";
    actionsCell.innerHTML = "";
    actionsCell.appendChild(actionsWrap);
  } else {
    actionsWrap.innerHTML = "";
  }
  const state = stateValue;
  const isRunning = state === "running";
  const isPaused = state === "paused";
  const isRestarting = state === "restarting";
  const canStart = !(isRunning || isPaused || isRestarting);
  const canStop = isRunning || isPaused || isRestarting;
  const canRestart = isRunning || isPaused || isRestarting;
  const canPause = isRunning || isPaused;
  const canKill = isRunning || isPaused || isRestarting;
  if (!canKill) {
    containersKillConfirming.delete(container.id);
  }

  const startBtn = buildContainerActionButton("icon-play", "Start", async () => {
    containersKillConfirming.delete(container.id);
    await runContainerAction(scope, container.id, "start");
  });
  startBtn.classList.add("btn-success");
  startBtn.disabled = !canStart;

  const stopBtn = buildContainerActionButton("icon-stop", "Stop", async () => {
    containersKillConfirming.delete(container.id);
    await runContainerAction(scope, container.id, "stop");
  });
  stopBtn.classList.add("btn-danger");
  stopBtn.disabled = !canStop;

  const restartBtn = buildContainerActionButton("icon-reload", "Restart", async () => {
    containersKillConfirming.delete(container.id);
    setContainerOptimisticState(container.id, "restarting", scope);
    await runContainerAction(scope, container.id, "restart");
  });
  restartBtn.classList.add("btn-info");
  restartBtn.disabled = !canRestart;

  const pauseLabel = isPaused ? "Unpause" : "Pause";
  const pauseIcon = isPaused ? "icon-play" : "icon-pause";
  const pauseAction = isPaused ? "unpause" : "pause";
  const pauseBtn = buildContainerActionButton(pauseIcon, pauseLabel, async () => {
    containersKillConfirming.delete(container.id);
    await runContainerAction(scope, container.id, pauseAction);
  });
  pauseBtn.classList.add("btn-warning");
  pauseBtn.disabled = !canPause;

  const isConfirmingKill = containersKillConfirming.has(container.id);
  const killBtn = buildContainerActionButton(
    isConfirmingKill ? "icon-check" : "icon-skull",
    isConfirmingKill ? "Confirm kill" : "Kill",
    async () => {
      if (!canKill) return;
      if (!containersKillConfirming.has(container.id)) {
        containersKillConfirming.add(container.id);
        updateContainerRow(row, container, scope);
        return;
      }
      containersKillConfirming.delete(container.id);
      await runContainerAction(scope, container.id, "kill");
    }
  );
  killBtn.classList.add("btn-danger");
  killBtn.classList.add("containers-kill-btn");
  if (isConfirmingKill) {
    killBtn.classList.add("is-confirming");
  }
  killBtn.disabled = !canKill;

  actionsWrap.append(startBtn, stopBtn, restartBtn, pauseBtn, killBtn);
}

function clearContainersTable(message) {
  if (containersTableBody) {
    containersTableBody.innerHTML = "";
  }
  containersCache = new Map();
  containersKillConfirming.clear();
  containersTableResourcesData = new Map();
  containersTableResourcesScope = "";
  if (containersShellList) {
    containersShellList.innerHTML = "";
    if (containersViewMode === "shell") {
      const empty = document.createElement("div");
      empty.className = "containers-shell-placeholder";
      empty.textContent = message || "No containers found.";
      containersShellList.appendChild(empty);
    }
  }
  if (containersLogsList) {
    containersLogsList.innerHTML = "";
    if (containersViewMode === "logs") {
      const empty = document.createElement("div");
      empty.className = "containers-shell-placeholder";
      empty.textContent = message || "No containers found.";
      containersLogsList.appendChild(empty);
    }
  }
  if (containersEmptyEl) {
    if (message) {
      containersEmptyEl.textContent = message;
      containersEmptyEl.classList.remove("hidden");
    } else {
      containersEmptyEl.classList.add("hidden");
    }
  }
  updateContainersSearchCount();
}

function renderContainers(list, scope) {
  if (!containersTableBody) return;
  if (!Array.isArray(list) || list.length === 0) {
    clearContainersTable("No containers found.");
    return;
  }
  if (containersEmptyEl) {
    containersEmptyEl.classList.add("hidden");
  }
  const previous = containersCache;
  const next = new Map();
  list.forEach((container) => {
    if (!container || !container.id) return;
    let entry = previous.get(container.id);
    let row = entry ? entry.row : null;
    if (!row) {
      row = document.createElement("tr");
      for (let i = 0; i < 10; i += 1) {
        row.appendChild(document.createElement("td"));
      }
    }
    updateContainerRow(row, container, scope);
    next.set(container.id, { row, data: container });
  });
  previous.forEach((entry, id) => {
    if (!next.has(id)) {
      entry.row.remove();
    }
  });
  containersCache = next;
  const sorted = containersSort(Array.from(next.values()).map((entry) => entry.data), containersSortMode);
  const fragment = document.createDocumentFragment();
  sorted.forEach((container) => {
    const entry = containersCache.get(container.id);
    if (entry) {
      fragment.appendChild(entry.row);
    }
  });
  containersTableBody.appendChild(fragment);
  if (containersViewMode === "shell") {
    renderContainersShellList(sorted);
  } else if (containersViewMode === "logs") {
    renderContainersLogsList(sorted);
  } else if (containersViewMode === "resources") {
    renderContainersResourcesList(sorted);
  }
  updateContainersSearchCount();
}

async function refreshContainers(options = {}) {
  if (!containersTableBody) return;
  if (containersUpdateInProgress) return;
  containersUpdateInProgress = true;
  const scope = containersSelectedScope;
  if (!scope) {
    clearContainersTable("No servers configured. Add one in Servers.");
    updateContainersStatus("", "");
    containersUpdateInProgress = false;
    return;
  }
  containersSelectedScope = scope;
  updateContainersStatus(scope, "");
  try {
    const payload = await fetchJSON(`/api/containers?scope=${encodeURIComponent(scope)}`);
    if (!payload || payload.error) {
      const errorMessage = payload && payload.error ? payload.error : "Unable to load containers.";
      clearContainersTable("No containers data available.");
      updateContainersStatus(scope, errorMessage);
      return;
    }
    const list = Array.isArray(payload.containers) ? payload.containers : [];
    if (containersTableResourcesScope !== scope) {
      containersTableResourcesData = new Map();
      containersTableResourcesScope = scope;
    }
    renderContainers(list, scope);
    if (currentView === "containers" && containersViewMode === "table") {
      refreshContainersTableResources(scope, list, { silent: true }).catch(() => {});
    }
    updateContainersStatus(scope, "");
  } catch (err) {
    clearContainersTable("No containers data available.");
    updateContainersStatus(scope, err.message);
  } finally {
    containersUpdateInProgress = false;
  }
}

async function refreshContainersTableResources(scope, list, options = {}) {
  if (!scope) return;
  if (!Array.isArray(list) || list.length === 0) {
    containersTableResourcesData = new Map();
    containersTableResourcesScope = scope;
    return;
  }
  if (containersTableResourcesUpdateInProgress && containersTableResourcesScope === scope) return;
  containersTableResourcesUpdateInProgress = true;
  containersTableResourcesScope = scope;
  const requestId = (containersTableResourcesRequestId += 1);
  try {
    const containerIds = list.map((c) => c && c.id).filter(Boolean);
    const payload = await fetchJSON("/api/containers/resources", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ scope, container_ids: containerIds }),
    });
    if (requestId !== containersTableResourcesRequestId) return;
    if (!payload || payload.error) {
      if (!options.silent) {
        showToast(payload && payload.error ? payload.error : "Unable to load container resources.");
      }
      return;
    }
    const resources = Array.isArray(payload.resources) ? payload.resources : [];
    containersTableResourcesData = new Map(resources.map((item) => [item.id, item]));
    if (currentView === "containers" && containersViewMode === "table" && containersSelectedScope === scope) {
      const currentList = Array.from(containersCache.values()).map((entry) => entry.data);
      renderContainers(currentList, scope);
    }
  } catch (err) {
    if (requestId !== containersTableResourcesRequestId) return;
    if (!options.silent) {
      showToast(err.message || "Unable to load container resources.");
    }
  } finally {
    if (requestId === containersTableResourcesRequestId) {
      containersTableResourcesUpdateInProgress = false;
    }
  }
}

async function refreshContainersResources(options = {}) {
  if (containersResourcesUpdateInProgress) return;
  if (currentView !== "containers" || containersViewMode !== "resources") return;
  if (!options.force && Date.now() - containersResourcesLastInteractionAtMs < 450) return;
  const scope = containersSelectedScope;
  if (!scope) {
    updateContainersResourcesPlaceholder("Select a server to view resources.");
    return;
  }
  const list = Array.from(containersCache.values()).map((entry) => entry.data);
  const byId = new Map(list.map((container) => [container.id, container]));
  const byName = new Map(list.map((container) => [container.name, container]));
  const pins = getContainersResourcesPins(scope);
  const targetIds = [];
  pins.forEach((name) => {
    const container = byName.get(name);
    if (container && container.id) {
      targetIds.push(container.id);
    }
  });
  containersResourcesSelectedIds.forEach((id) => {
    if (byId.has(id) && !targetIds.includes(id)) {
      targetIds.push(id);
    }
  });
  if (targetIds.length === 0) {
    containersResourcesData = new Map();
    renderContainersResourcesView();
    updateContainersResourcesPlaceholder();
    return;
  }
  containersResourcesUpdateInProgress = true;
  try {
    const payload = await fetchJSON("/api/containers/resources", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ scope, container_ids: targetIds }),
    });
    if (!payload || payload.error) {
      if (!options.silent) {
        showToast(payload && payload.error ? payload.error : "Unable to load container resources.");
      }
      containersResourcesData = new Map();
      renderContainersResourcesView();
      updateContainersResourcesPlaceholder("Unable to load resources.");
      return;
    }
    const resources = Array.isArray(payload.resources) ? payload.resources : [];
    containersResourcesData = new Map(resources.map((item) => [item.id, item]));
    renderContainersResourcesView();
  } catch (err) {
    if (!options.silent) {
      showToast(err.message || "Unable to load container resources.");
    }
    containersResourcesData = new Map();
    renderContainersResourcesView();
    updateContainersResourcesPlaceholder("Unable to load resources.");
  } finally {
    containersResourcesUpdateInProgress = false;
  }
}

function clearStacksTable(message) {
  if (stacksTableBody) {
    stacksTableBody.innerHTML = "";
  }
  stacksCache = new Map();
  if (stacksEmptyEl) {
    if (message) {
      stacksEmptyEl.textContent = message;
      stacksEmptyEl.classList.remove("hidden");
    } else {
      stacksEmptyEl.classList.add("hidden");
    }
  }
  updateContainersSearchCount();
}

function formatStackContainers(stack) {
  const total = Number(stack.containers_total || 0);
  const running = Number(stack.containers_running || 0);
  return total > 0 ? `${running}/${total}` : "0";
}

function buildStackStatusBadge(status) {
  const badge = document.createElement("span");
  const normalized = String(status || "not_deployed").toLowerCase();
  badge.className = `stack-status-badge stack-status-${normalized}`;
  badge.textContent = normalized.replace(/_/g, " ");
  return badge;
}

function setStackStatusOverride(name, action, status, baseline) {
  if (!name || !status || !action) return;
  const baseStatus = baseline && baseline.status ? String(baseline.status) : "";
  const baseTotal = Number(baseline && baseline.containers_total ? baseline.containers_total : 0);
  const baseRunning = Number(baseline && baseline.containers_running ? baseline.containers_running : 0);
  stackStatusOverrides.set(name, {
    action,
    status,
    baseStatus,
    baseTotal,
    baseRunning,
    startedAtMs: Date.now(),
    completedAtMs: 0,
  });
}

function markStackStatusOverrideCompleted(name) {
  const entry = stackStatusOverrides.get(name);
  if (!entry) return;
  entry.completedAtMs = Date.now();
  stackStatusOverrides.set(name, entry);
}

function maybeClearStackStatusOverride(stack) {
  if (!stack || !stack.name) return;
  const entry = stackStatusOverrides.get(stack.name);
  if (!entry) return;

  const now = Date.now();
  const maxAgeMs = 5 * 60 * 1000;
  if (now - entry.startedAtMs > maxAgeMs) {
    stackStatusOverrides.delete(stack.name);
    return;
  }

  const currentStatus = String(stack.status || "");
  const currentTotal = Number(stack.containers_total || 0);
  const currentRunning = Number(stack.containers_running || 0);

  if (entry.baseStatus && currentStatus && currentStatus.toLowerCase() !== String(entry.baseStatus).toLowerCase()) {
    stackStatusOverrides.delete(stack.name);
    return;
  }
  if (currentTotal !== Number(entry.baseTotal || 0) || currentRunning !== Number(entry.baseRunning || 0)) {
    stackStatusOverrides.delete(stack.name);
    return;
  }

  if (entry.completedAtMs) {
    const fallbackByActionMs = {
      up: 30000,
      down: 30000,
      restart: 5000,
      start: 5000,
      stop: 5000,
      kill: 5000,
      rm: 5000,
    };
    const ttlMs = fallbackByActionMs[String(entry.action)] || 12000;
    if (now - entry.completedAtMs > ttlMs) {
      stackStatusOverrides.delete(stack.name);
      return;
    }
  }
}

function getStackStatusOverride(name, stack) {
  if (stack) {
    maybeClearStackStatusOverride(stack);
  }
  const entry = stackStatusOverrides.get(name);
  return entry ? entry.status : "";
}

function buildStackActionButton(iconClass, label, onClick, variant) {
  const button = document.createElement("button");
  button.type = "button";
  button.className = "secondary icon-action-btn has-tooltip";
  if (variant) {
    button.classList.add(variant);
  }
  button.setAttribute("aria-label", label);
  button.setAttribute("data-tooltip", label);
  const icon = document.createElement("span");
  icon.className = `icon-action ${iconClass}`;
  icon.setAttribute("aria-hidden", "true");
  button.appendChild(icon);
  button.addEventListener("click", onClick);
  return button;
}

function updateStackRow(entry, stack) {
  entry.data = stack;
  const row = entry.row;
  row.className = "stack-row";
  row.dataset.search = `${stack.name || ""} ${stack.status || ""}`;
  const cells = row.querySelectorAll("td");
  if (cells.length < 4) return;
  cells[0].textContent = "";
  const nameIcon = document.createElement("span");
  nameIcon.className = "icon-action icon-stack name-leading-icon";
  nameIcon.setAttribute("aria-hidden", "true");
  cells[0].appendChild(nameIcon);
  cells[0].appendChild(document.createTextNode(stack.name));
  cells[1].textContent = "";
  const optimistic = getStackStatusOverride(stack.name, stack);
  cells[1].appendChild(buildStackStatusBadge(optimistic || stack.status));
  cells[2].textContent = formatStackContainers(stack);
  const actionsCell = cells[3];
  let actionsWrap = actionsCell.querySelector(".containers-actions");
  if (!actionsWrap) {
    actionsWrap = document.createElement("div");
    actionsWrap.className = "containers-actions";
    actionsCell.innerHTML = "";
    actionsCell.appendChild(actionsWrap);
  } else {
    actionsWrap.innerHTML = "";
  }
  const editBtn = buildStackActionButton("icon-edit", "Edit stack", () => openStackModal(stack.name));
  const upBtn = buildStackActionButton("icon-chevrons-up", "Compose up", () => runStackAction(stack.name, "up"), "btn-success");
  const downBtn = buildStackActionButton("icon-chevrons-down", "Compose down", () => runStackAction(stack.name, "down"), "btn-warning");
  const restartBtn = buildStackActionButton("icon-reload", "Restart stack", () => runStackAction(stack.name, "restart"), "btn-info");
  const startBtn = buildStackActionButton("icon-play", "Start stack", () => runStackAction(stack.name, "start"), "btn-success");
  const stopBtn = buildStackActionButton("icon-stop", "Stop stack", () => runStackAction(stack.name, "stop"), "btn-warning");
  const killKey = stackActionKey(stack.name, "kill");
  const isConfirmingKill = stacksActionConfirming.has(killKey);
  const killBtn = buildStackActionButton(
    isConfirmingKill ? "icon-check" : "icon-skull",
    isConfirmingKill ? "Confirm kill" : "Kill stack",
    async () => {
      if (!stacksActionConfirming.has(killKey)) {
        stacksActionConfirming.add(killKey);
        updateStackRow(entry, stack);
        return;
      }
      stacksActionConfirming.delete(killKey);
      await runStackAction(stack.name, "kill");
    },
    "btn-danger"
  );
  killBtn.classList.add("containers-kill-btn");
  if (isConfirmingKill) {
    killBtn.classList.add("is-confirming");
  }

  const rmKey = stackActionKey(stack.name, "rm");
  const isConfirmingRm = stacksActionConfirming.has(rmKey);
  const rmBtn = buildStackActionButton(
    isConfirmingRm ? "icon-check" : "icon-trash",
    isConfirmingRm ? "Confirm remove" : "Remove stack",
    async () => {
      if (!stacksActionConfirming.has(rmKey)) {
        stacksActionConfirming.add(rmKey);
        updateStackRow(entry, stack);
        return;
      }
      stacksActionConfirming.delete(rmKey);
      await runStackAction(stack.name, "rm");
    },
    "btn-danger"
  );
  rmBtn.classList.add("containers-kill-btn");
  if (isConfirmingRm) {
    rmBtn.classList.add("is-confirming");
  }
  actionsWrap.append(editBtn, upBtn, downBtn, restartBtn, startBtn, stopBtn, killBtn, rmBtn);
}

function renderStacks(list) {
  if (!stacksTableBody) return;
  if (!Array.isArray(list) || list.length === 0) {
    clearStacksTable("No stacks found.");
    return;
  }
  if (stacksEmptyEl) {
    stacksEmptyEl.classList.add("hidden");
  }
  const sorted = stacksSort(list, stacksSortMode);
  const previous = stacksCache;
  const next = new Map();
  sorted.forEach((stack) => {
    if (!stack || !stack.name) return;
    let entry = previous.get(stack.name);
    let row = entry ? entry.row : null;
    if (!row) {
      row = document.createElement("tr");
      for (let i = 0; i < 4; i += 1) {
        row.appendChild(document.createElement("td"));
      }
    }
    entry = { ...(entry || {}), row, name: stack.name };
    updateStackRow(entry, stack);
    next.set(stack.name, entry);
  });
  previous.forEach((entry, name) => {
    if (!next.has(name)) {
      entry.row.remove();
    }
  });
  stacksCache = next;
  const fragment = document.createDocumentFragment();
  sorted.forEach((stack) => {
    const entry = stacksCache.get(stack.name);
    if (entry) {
      fragment.appendChild(entry.row);
    }
  });
  stacksTableBody.appendChild(fragment);
  if (sidebarSearch && sidebarSearch.value.trim()) {
    applySidebarFilter(sidebarSearch.value);
  } else {
    updateContainersSearchCount();
  }
}

async function refreshStacks(options = {}) {
  if (!stacksTableBody) return;
  if (stacksUpdateInProgress) return;
  stacksUpdateInProgress = true;
  const scope = containersSelectedScope;
  if (!scope) {
    clearStacksTable("No servers configured. Add one in Servers.");
    updateContainersStatus("", "");
    stacksUpdateInProgress = false;
    return;
  }
  containersSelectedScope = scope;
  updateContainersStatus(scope, "");
  try {
    const payload = await fetchJSON(`/api/stacks?scope=${encodeURIComponent(scope)}`);
    const list = payload && Array.isArray(payload.stacks) ? payload.stacks : [];
    if (payload && payload.error) {
      renderStacks(list);
      updateContainersStatus(scope, payload.error);
      return;
    }
    if (!payload) {
      clearStacksTable("No stacks data available.");
      updateContainersStatus(scope, "Unable to load stacks.");
      return;
    }
    renderStacks(list);
    updateContainersStatus(scope, "");
  } catch (err) {
    clearStacksTable("No stacks data available.");
    updateContainersStatus(scope, err.message);
  } finally {
    stacksUpdateInProgress = false;
  }
}

function clearImagesTable(message) {
  if (imagesTableBody) {
    imagesTableBody.innerHTML = "";
  }
  imagesCache = [];
  imagesSelectedId = "";
  imagesActionConfirming.clear();
  updateImagesConfirmButtons();
  if (imagesRemoveBtn) {
    imagesRemoveBtn.disabled = true;
  }
  if (imagesEmptyEl) {
    if (message) {
      imagesEmptyEl.textContent = message;
      imagesEmptyEl.classList.remove("hidden");
    } else {
      imagesEmptyEl.classList.add("hidden");
    }
  }
  updateImagesSummary([]);
  updateContainersSearchCount();
}

function updateImagesSummary(list) {
  if (imagesTotalCountEl) {
    const valueEl = imagesTotalCountEl.querySelector(".summary-value");
    if (valueEl) {
      valueEl.textContent = String(list.length);
    } else {
      imagesTotalCountEl.textContent = String(list.length);
    }
  }
  if (!imagesTotalSizeEl) return;
  const sizes = new Map();
  list.forEach((item) => {
    if (!item || !item.id) return;
    if (!sizes.has(item.id)) {
      sizes.set(item.id, Number(item.size_bytes || 0));
    }
  });
  const total = Array.from(sizes.values()).reduce((acc, val) => acc + val, 0);
  const sizeEl = imagesTotalSizeEl.querySelector(".summary-value");
  if (sizeEl) {
    sizeEl.textContent = formatBytes(total);
  } else {
    imagesTotalSizeEl.textContent = formatBytes(total);
  }
}

function imageStateLabel(image) {
  if (image && image.dangling) return "Dangling";
  const count = Number(image && image.containers_count ? image.containers_count : 0);
  return count > 0 ? "In use" : "Unused";
}

function imageStateClass(state) {
  const normalized = String(state || "").toLowerCase();
  if (normalized === "in use") return "image-state-in-use";
  if (normalized === "dangling") return "image-state-dangling";
  if (normalized === "multiple") return "image-state-unused";
  return "image-state-unused";
}

function buildImageStateBadge(state) {
  const badge = document.createElement("span");
  badge.className = `image-state-badge ${imageStateClass(state)}`;
  badge.textContent = String(state || "Unused");
  return badge;
}

function buildImageGroups(list) {
  const groups = new Map();
  list.forEach((image) => {
    if (!image) return;
    const repo = image.repository || "<none>";
    if (!groups.has(repo)) {
      groups.set(repo, []);
    }
    groups.get(repo).push(image);
  });
  return groups;
}

function toggleImagesGroup(repo) {
  if (!repo) return;
  if (imagesExpandedRepos.has(repo)) {
    imagesExpandedRepos.delete(repo);
  } else {
    imagesExpandedRepos.add(repo);
  }
  renderImages(imagesCache, containersSelectedScope);
}

function selectImageRow(imageId) {
  imagesSelectedId = imageId || "";
  if (imagesRemoveBtn) {
    imagesRemoveBtn.disabled = !imagesSelectedId;
  }
  if (imagesActionConfirming.size > 0) {
    imagesActionConfirming.clear();
    updateImagesConfirmButtons();
  }
  if (!imagesTableBody) return;
  imagesTableBody.querySelectorAll("tr.images-image-row").forEach((row) => {
    row.classList.toggle("is-selected", row.dataset && row.dataset.id === imagesSelectedId);
  });
}

function updateImagesConfirmButtons() {
  const actions = [
    { key: "prune-unused", btn: imagesPruneUnusedBtn, defaultIcon: "icon-square-minus" },
    { key: "prune-dangling", btn: imagesPruneDanglingBtn, defaultIcon: "icon-copy-minus" },
    { key: "remove", btn: imagesRemoveBtn, defaultIcon: "icon-trash" },
  ];
  actions.forEach(({ key, btn, defaultIcon }) => {
    if (!btn) return;
    const icon = btn.querySelector(".icon-action");
    if (!icon) return;
    if (imagesActionConfirming.has(key)) {
      icon.classList.remove(defaultIcon);
      icon.classList.add("icon-check");
      btn.classList.add("is-confirming");
      if (key === "remove") {
        btn.classList.add("btn-danger");
      } else {
        btn.classList.add("btn-warning");
      }
    } else {
      icon.classList.remove("icon-check");
      icon.classList.add(defaultIcon);
      btn.classList.remove("is-confirming");
      btn.classList.remove("btn-warning");
      btn.classList.remove("btn-danger");
    }
  });
}

function toggleImagesConfirm(key) {
  if (!key) return;
  if (imagesActionConfirming.has(key)) {
    imagesActionConfirming.delete(key);
  } else {
    imagesActionConfirming.clear();
    imagesActionConfirming.add(key);
  }
  updateImagesConfirmButtons();
}

function clearImagesConfirmations() {
  if (imagesActionConfirming.size === 0) return;
  imagesActionConfirming.clear();
  updateImagesConfirmButtons();
}

function clearImagesSelection() {
  if (!imagesSelectedId) return;
  imagesSelectedId = "";
  if (imagesRemoveBtn) {
    imagesRemoveBtn.disabled = true;
  }
  if (imagesTableBody) {
    imagesTableBody.querySelectorAll("tr.images-image-row").forEach((row) => {
      row.classList.remove("is-selected");
    });
  }
  clearImagesConfirmations();
}

function renderImages(list) {
  if (!imagesTableBody) return;
  if (!Array.isArray(list) || list.length === 0) {
    clearImagesTable("No images found.");
    return;
  }
  if (imagesEmptyEl) {
    imagesEmptyEl.classList.add("hidden");
  }
  imagesCache = list;
  updateImagesSummary(list);
  const groups = buildImageGroups(list);
  const [field, order] = String(imagesSortMode || "repository:asc").split(":");
  const dir = order === "desc" ? -1 : 1;
  const groupEntries = Array.from(groups.entries()).map(([repo, items]) => {
    const groupSize = items.reduce((acc, image) => {
      if (image && image.id && !acc.ids.has(image.id)) {
        acc.ids.add(image.id);
        acc.total += Number(image.size_bytes || 0);
      }
      return acc;
    }, { total: 0, ids: new Set() }).total;
    const isSingle = items.length === 1;
    const firstImage = isSingle ? items[0] : null;
    const groupState = isSingle && firstImage ? imageStateLabel(firstImage) : "Multiple";
    const groupTag = isSingle && firstImage ? String(firstImage.tag || "") : "Multiple";
    return {
      repo,
      items,
      size: groupSize,
      state: groupState,
      tag: groupTag,
    };
  });
  groupEntries.sort((a, b) => {
    let aVal = "";
    let bVal = "";
    switch (field) {
      case "tags":
        aVal = a.tag;
        bVal = b.tag;
        break;
      case "state":
        aVal = normalizeQuery(a.state);
        bVal = normalizeQuery(b.state);
        break;
      case "size":
        aVal = a.size;
        bVal = b.size;
        break;
      default:
        aVal = a.repo;
        bVal = b.repo;
        break;
    }
    if (aVal === bVal) return 0;
    return aVal > bVal ? dir : -dir;
  });
  const fragment = document.createDocumentFragment();
  groupEntries.forEach((entry) => {
    const repo = entry.repo;
    const group = entry.items.slice();
    const expanded = imagesExpandedRepos.has(repo);
    const isSingle = group.length === 1;
    const firstImage = isSingle ? group[0] : null;
    const fieldForItems = field === "repository" ? "tags" : field;
    group.sort((a, b) => {
      const aVal = imagesSortKey(a, fieldForItems);
      const bVal = imagesSortKey(b, fieldForItems);
      if (aVal === bVal) return 0;
      return aVal > bVal ? dir : -dir;
    });
    if (isSingle && firstImage) {
      const image = firstImage;
      const imageRow = document.createElement("tr");
      imageRow.className = "images-image-row images-row";
      imageRow.dataset.id = image.id;
      imageRow.dataset.repo = repo;
      imageRow.dataset.search = `${repo} ${image.tag || ""} ${image.id || ""} ${imageStateLabel(image)}`;
      for (let i = 0; i < 6; i += 1) {
        imageRow.appendChild(document.createElement("td"));
      }
      const imageCells = imageRow.querySelectorAll("td");
      imageCells[0].textContent = "";
      const imageIcon = document.createElement("span");
      imageIcon.className = "icon-action icon-cube-3d-sphere name-leading-icon";
      imageIcon.setAttribute("aria-hidden", "true");
      imageCells[0].appendChild(imageIcon);
      imageCells[0].appendChild(document.createTextNode(image.repository || repo));
      imageCells[1].textContent = image.tag || "—";
      imageCells[2].textContent = "";
      imageCells[2].appendChild(buildImageStateBadge(imageStateLabel(image)));
      imageCells[3].textContent = formatBytes(image.size_bytes || 0);
      imageCells[4].textContent = "1";
      imageCells[5].textContent = formatImageCreated(image.created_at);
      if (imagesSelectedId && image.id === imagesSelectedId) {
        imageRow.classList.add("is-selected");
      }
      imageRow.addEventListener("click", () => selectImageRow(image.id));
      fragment.appendChild(imageRow);
      return;
    }
    const row = document.createElement("tr");
    row.className = "images-group-row images-row";
    row.dataset.repo = repo;
    const groupTags = group.map((image) => image && image.tag ? image.tag : "").filter(Boolean).join(" ");
    row.dataset.search = `${repo} ${groupTags} Multiple`;
    for (let i = 0; i < 6; i += 1) {
      row.appendChild(document.createElement("td"));
    }
    const cells = row.querySelectorAll("td");
    const toggle = document.createElement("span");
    toggle.className = `images-group-toggle ${expanded ? "is-expanded" : ""}`;
    const icon = document.createElement("span");
    icon.className = `icon-action ${expanded ? "icon-chevron-down" : "icon-chevron-right"}`;
    icon.setAttribute("aria-hidden", "true");
    toggle.appendChild(icon);
    const label = document.createElement("span");
    label.textContent = repo;
    toggle.appendChild(label);
    toggle.addEventListener("click", () => toggleImagesGroup(repo));
    cells[0].textContent = "";
    cells[0].appendChild(toggle);
    cells[1].textContent = "Multiple";
    cells[2].textContent = "";
    cells[2].appendChild(buildImageStateBadge("Multiple"));
    const uniqueSizes = new Map();
    let latestCreated = 0;
    group.forEach((image) => {
      if (image && image.id && !uniqueSizes.has(image.id)) {
        uniqueSizes.set(image.id, Number(image.size_bytes || 0));
      }
      const createdAt = new Date(image && image.created_at ? image.created_at : 0).getTime();
      if (Number.isFinite(createdAt)) {
        latestCreated = Math.max(latestCreated, createdAt);
      }
    });
    const groupSize = Array.from(uniqueSizes.values()).reduce((acc, val) => acc + val, 0);
    cells[3].textContent = formatBytes(groupSize);
    cells[4].textContent = String(group.length);
    cells[5].textContent = latestCreated ? formatImageCreated(new Date(latestCreated).toISOString()) : "—";
    fragment.appendChild(row);
    group.forEach((image) => {
      if (!image) return;
      const imageRow = document.createElement("tr");
      imageRow.className = `images-image-row images-row${expanded ? "" : " images-row-hidden"}`;
      imageRow.dataset.id = image.id;
      imageRow.dataset.repo = repo;
      imageRow.dataset.search = `${repo} ${image.tag || ""} ${image.id || ""} ${imageStateLabel(image)}`;
      for (let i = 0; i < 6; i += 1) {
        imageRow.appendChild(document.createElement("td"));
      }
      const imageCells = imageRow.querySelectorAll("td");
      imageCells[0].textContent = "";
      const imageIcon = document.createElement("span");
      imageIcon.className = "icon-action icon-cube-3d-sphere name-leading-icon";
      imageIcon.setAttribute("aria-hidden", "true");
      imageCells[0].appendChild(imageIcon);
      imageCells[0].appendChild(document.createTextNode(image.repository || repo));
      imageCells[0].classList.add("images-child-cell");
      imageCells[1].textContent = image.tag || "—";
      imageCells[2].textContent = "";
      imageCells[2].appendChild(buildImageStateBadge(imageStateLabel(image)));
      imageCells[3].textContent = formatBytes(image.size_bytes || 0);
      imageCells[4].textContent = "1";
      imageCells[5].textContent = formatImageCreated(image.created_at);
      if (imagesSelectedId && image.id === imagesSelectedId) {
        imageRow.classList.add("is-selected");
      }
      imageRow.addEventListener("click", () => selectImageRow(image.id));
      fragment.appendChild(imageRow);
    });
  });
  imagesTableBody.innerHTML = "";
  imagesTableBody.appendChild(fragment);
  if (imagesSelectedId) {
    const exists = list.some((image) => image && image.id === imagesSelectedId);
    if (!exists) {
      imagesSelectedId = "";
      if (imagesRemoveBtn) imagesRemoveBtn.disabled = true;
    }
  }
  if (sidebarSearch && sidebarSearch.value.trim()) {
    applySidebarFilter(sidebarSearch.value);
  } else {
    updateContainersSearchCount();
  }
}

async function refreshImages(options = {}) {
  if (!imagesTableBody) return;
  if (imagesUpdateInProgress) return;
  imagesUpdateInProgress = true;
  const scope = containersSelectedScope;
  if (!scope) {
    clearImagesTable("No servers configured. Add one in Servers.");
    updateContainersStatus("", "");
    imagesUpdateInProgress = false;
    return;
  }
  containersSelectedScope = scope;
  updateContainersStatus(scope, "");
  try {
    const payload = await fetchJSON(`/api/images?scope=${encodeURIComponent(scope)}`);
    if (!payload || payload.error) {
      const errorMessage = payload && payload.error ? payload.error : "Unable to load images.";
      clearImagesTable("No images data available.");
      updateContainersStatus(scope, errorMessage);
      return;
    }
    const list = Array.isArray(payload.images) ? payload.images : [];
    renderImages(list, scope);
    updateContainersStatus(scope, "");
  } catch (err) {
    clearImagesTable("No images data available.");
    updateContainersStatus(scope, err.message);
  } finally {
    imagesUpdateInProgress = false;
  }
}

function openImagesPullModal() {
  if (!imagesPullModal) return;
  if (imagesPullRepoInput) imagesPullRepoInput.value = "";
  if (imagesPullTagInput) imagesPullTagInput.value = "latest";
  setImagesPullInProgress(false, "");
  imagesPullModal.classList.remove("hidden");
}

function closeImagesPullModal() {
  if (!imagesPullModal) return;
  if (imagesPullInProgress) return;
  stopImagesPullAutoClose();
  imagesPullModal.classList.add("hidden");
}

function setImagesPullInProgress(inProgress, message, variant = "") {
  imagesPullInProgress = Boolean(inProgress);
  if (imagesPullRepoInput) imagesPullRepoInput.disabled = imagesPullInProgress;
  if (imagesPullTagInput) imagesPullTagInput.disabled = imagesPullInProgress;
  if (imagesPullCloseBtn) imagesPullCloseBtn.disabled = imagesPullInProgress;
  if (imagesPullCancelBtn) imagesPullCancelBtn.disabled = imagesPullInProgress;
  if (imagesPullBtn) imagesPullBtn.disabled = imagesPullInProgress;
  if (imagesPruneUnusedBtn) imagesPruneUnusedBtn.disabled = imagesPullInProgress;
  if (imagesPruneDanglingBtn) imagesPruneDanglingBtn.disabled = imagesPullInProgress;
  if (imagesRemoveBtn) imagesRemoveBtn.disabled = imagesPullInProgress ? true : !imagesSelectedId;

  if (imagesPullConfirmBtn) {
    imagesPullConfirmBtn.disabled = imagesPullInProgress;
    if (imagesPullInProgress) {
      imagesPullConfirmBtn.innerHTML = '<span class="icon-action icon-reload icon-spin" aria-hidden="true"></span><span>Pulling…</span>';
    } else {
      imagesPullConfirmBtn.textContent = "Pull";
    }
  }

  if (!imagesPullStatusEl) return;
  const text = String(message || "").trim();
  imagesPullStatusEl.classList.toggle("is-empty", !text);
  imagesPullStatusEl.classList.toggle("is-error", variant === "error");
  imagesPullStatusEl.classList.toggle("is-success", variant === "success");
  if (text) {
    imagesPullStatusEl.textContent = text;
  } else {
    imagesPullStatusEl.textContent = "";
  }
}

function stopImagesPullAutoClose() {
  if (!imagesPullAutoCloseTimer) return;
  window.clearInterval(imagesPullAutoCloseTimer);
  imagesPullAutoCloseTimer = null;
  imagesPullAutoCloseRemainingSec = 0;
  imagesPullLastRef = "";
}

function startImagesPullAutoClose(repository, tag, seconds = 3) {
  stopImagesPullAutoClose();
  imagesPullLastRef = `${repository}:${tag}`;
  imagesPullAutoCloseRemainingSec = Math.max(1, Number(seconds) || 3);
  const update = () => {
    const suffix = imagesPullAutoCloseRemainingSec > 0
      ? `Auto close in ${imagesPullAutoCloseRemainingSec}s (click to keep open).`
      : "";
    setImagesPullInProgress(false, `Pulled ${imagesPullLastRef}.\n${suffix}`.trim(), "success");
  };
  update();
  imagesPullAutoCloseTimer = window.setInterval(() => {
    imagesPullAutoCloseRemainingSec -= 1;
    if (imagesPullAutoCloseRemainingSec <= 0) {
      stopImagesPullAutoClose();
      closeImagesPullModal();
      return;
    }
    update();
  }, 1000);
}

function cancelImagesPullAutoClose() {
  if (!imagesPullAutoCloseTimer) return;
  stopImagesPullAutoClose();
  setImagesPullInProgress(false, "", "");
  if (imagesPullRepoInput) imagesPullRepoInput.value = "";
  if (imagesPullTagInput) imagesPullTagInput.value = "latest";
  if (imagesPullRepoInput) imagesPullRepoInput.focus();
}

async function runImagePull() {
  const scope = containersSelectedScope;
  if (!scope) {
    showToast("Select server first.");
    return;
  }
  const repository = imagesPullRepoInput ? imagesPullRepoInput.value.trim() : "";
  let tag = imagesPullTagInput ? imagesPullTagInput.value.trim() : "";
  if (!repository) {
    showToast("Repository is required.");
    return;
  }
  if (!tag) tag = "latest";
  try {
    stopImagesPullAutoClose();
    setImagesPullInProgress(true, `Pulling ${repository}:${tag}…`, "");
    await fetchJSON("/api/images/pull", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ scope, repository, tag }),
    });
    startImagesPullAutoClose(repository, tag, 6);
    await refreshImages({ silent: true });
  } catch (err) {
    stopImagesPullAutoClose();
    setImagesPullInProgress(false, err.message || "Pull image failed.", "error");
  }
}

async function runImagePrune(mode) {
  const scope = containersSelectedScope;
  if (!scope) {
    showToast("Select server first.");
    return;
  }
  try {
    imagesActionConfirming.clear();
    updateImagesConfirmButtons();
    await fetchJSON("/api/images/prune", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ scope, mode }),
    });
    await refreshImages({ silent: true });
  } catch (err) {
    showToast(err.message || "Prune images failed.");
  }
}

async function runImageRemove() {
  const scope = containersSelectedScope;
  if (!scope) {
    showToast("Select server first.");
    return;
  }
  if (!imagesSelectedId) {
    showToast("Select an image first.");
    return;
  }
  try {
    imagesActionConfirming.clear();
    updateImagesConfirmButtons();
    await fetchJSON("/api/images/remove", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ scope, image_id: imagesSelectedId }),
    });
    imagesSelectedId = "";
    if (imagesRemoveBtn) imagesRemoveBtn.disabled = true;
    await refreshImages({ silent: true });
  } catch (err) {
    showToast(err.message || "Remove image failed.");
  }
}

async function runStackAction(name, action) {
  if (!name || !action) return;
  const scope = containersSelectedScope;
  if (!scope) {
    showToast("Select server first.");
    return;
  }
  const optimisticByAction = {
    up: "deploying",
    down: "tearing_down",
    restart: "restarting",
    start: "starting",
    stop: "stopping",
    kill: "killing",
    rm: "removing",
  };
  if (optimisticByAction[action]) {
    const cached = stacksCache.get(name);
    setStackStatusOverride(name, action, optimisticByAction[action], cached ? cached.data : null);
    if (cached && cached.data) {
      updateStackRow(cached, cached.data);
    }
  }
  try {
    clearStacksActionConfirmations();
    await fetchJSON("/api/stacks/action", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ scope, name, action }),
    });
    markStackStatusOverrideCompleted(name);
    await refreshStacks({ silent: true });
  } catch (err) {
    stackStatusOverrides.delete(name);
    showToast(err.message || "Stack action failed.");
  }
}

function startStacksAutoRefresh() {
  if (stacksRefreshTimer) return;
  stacksRefreshTimer = window.setInterval(() => {
    refreshStacks({ silent: true });
  }, 6000);
}

function stopStacksAutoRefresh() {
  if (!stacksRefreshTimer) return;
  window.clearInterval(stacksRefreshTimer);
  stacksRefreshTimer = null;
}

async function runContainerAction(scope, containerID, action) {
  if (!scope || !containerID || !action) return;
  try {
    const payload = await fetchJSON("/api/containers/action", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ scope, container_id: containerID, action }),
    });
    if (payload && payload.container) {
      const entry = containersCache.get(payload.container.id);
      const row = entry ? entry.row : null;
      if (row) {
        updateContainerRow(row, payload.container, scope);
      }
    }
    await refreshContainers({ silent: true });
  } catch (err) {
    showToast(err.message || "Container action failed.");
  }
}

function startContainersAutoRefresh() {
  if (containersRefreshTimer) return;
  containersRefreshTimer = window.setInterval(() => {
    refreshContainers({ silent: true });
  }, 5000);
}

function stopContainersAutoRefresh() {
  if (!containersRefreshTimer) return;
  window.clearInterval(containersRefreshTimer);
  containersRefreshTimer = null;
}

function startContainersResourcesAutoRefresh() {
  if (containersResourcesRefreshTimer) return;
  containersResourcesRefreshTimer = window.setInterval(() => {
    refreshContainersResources({ silent: true });
  }, 5000);
}

function stopContainersResourcesAutoRefresh() {
  if (!containersResourcesRefreshTimer) return;
  window.clearInterval(containersResourcesRefreshTimer);
  containersResourcesRefreshTimer = null;
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

function formatAppVersionLabel(version) {
  const raw = version ? String(version) : "dev";
  if (raw.startsWith("v")) return raw;
  return `v${raw}`;
}

function buildReleaseTag(version) {
  if (!version) return "";
  const raw = String(version);
  const lower = raw.toLowerCase();
  if (lower.startsWith("vdev")) return raw.slice(1);
  if (lower.startsWith("dev")) return raw;
  if (lower.startsWith("v")) return raw;
  return `v${raw}`;
}

function buildReleaseUrl(meta) {
  const repo = meta && meta.repo ? meta.repo : defaultReleaseRepo;
  if (!repo) return "";
  const tag = meta && meta.release_tag ? meta.release_tag : buildReleaseTag(meta && meta.version ? meta.version : "");
  if (!tag) return `https://github.com/${repo}/releases`;
  return `https://github.com/${repo}/releases/tag/${tag}`;
}

function applyAppVersion(meta) {
  const versionLabel = formatAppVersionLabel(meta && meta.version ? meta.version : "dev");
  const url = buildReleaseUrl(meta);
  if (appVersionEl) {
    appVersionEl.textContent = versionLabel;
    if (url) {
      appVersionEl.href = url;
      appVersionEl.classList.add("app-version-link");
    } else {
      appVersionEl.removeAttribute("href");
      appVersionEl.classList.remove("app-version-link");
    }
  }
  if (aboutVersionLinkEl) {
    aboutVersionLinkEl.textContent = versionLabel;
    if (url) {
      aboutVersionLinkEl.href = url;
      aboutVersionLinkEl.classList.add("about-value");
    } else {
      aboutVersionLinkEl.removeAttribute("href");
      aboutVersionLinkEl.classList.remove("about-value");
    }
  }
  if (aboutChannelEl) {
    aboutChannelEl.textContent = meta && meta.channel ? meta.channel : "—";
  }
  if (aboutRepoLinkEl) {
    const repo = meta && meta.repo ? meta.repo : defaultReleaseRepo;
    if (repo) {
      aboutRepoLinkEl.textContent = repo;
      aboutRepoLinkEl.href = `https://github.com/${repo}`;
      aboutRepoLinkEl.classList.add("about-value");
    } else {
      aboutRepoLinkEl.textContent = "—";
      aboutRepoLinkEl.removeAttribute("href");
      aboutRepoLinkEl.classList.remove("about-value");
    }
  }
  if (aboutReleaseTagEl) {
    const tag = meta && meta.release_tag ? meta.release_tag : buildReleaseTag(meta && meta.version ? meta.version : "");
    aboutReleaseTagEl.textContent = tag || "—";
  }
}

function applyReleaseBadge(release) {
  if (!appVersionUpdateEl) return;
  appVersionUpdateEl.classList.add("hidden");
  appVersionUpdateEl.classList.remove("has-tooltip");
  appVersionUpdateEl.removeAttribute("data-tooltip");
  appVersionUpdateEl.removeAttribute("aria-label");
  if (aboutUpdateLinkEl && aboutUpdateStatusEl) {
    aboutUpdateLinkEl.classList.add("hidden");
    aboutUpdateLinkEl.classList.remove("has-tooltip");
    aboutUpdateLinkEl.removeAttribute("data-tooltip");
    aboutUpdateLinkEl.removeAttribute("aria-label");
    aboutUpdateStatusEl.classList.remove("hidden");
  }
  if (!release || !release.update_available || !release.latest || !release.latest.url) {
    return;
  }
  const label = release.latest.tag || release.latest.version || "";
  const tooltip = label ? `Update available: ${label}` : "Update available";
  appVersionUpdateEl.href = release.latest.url;
  appVersionUpdateEl.classList.remove("hidden");
  appVersionUpdateEl.classList.add("has-tooltip");
  appVersionUpdateEl.setAttribute("data-tooltip", tooltip);
  appVersionUpdateEl.setAttribute("aria-label", tooltip);
  if (aboutUpdateLinkEl && aboutUpdateStatusEl) {
    aboutUpdateStatusEl.classList.add("hidden");
    aboutUpdateLinkEl.href = release.latest.url;
    aboutUpdateLinkEl.classList.remove("hidden");
    aboutUpdateLinkEl.classList.add("has-tooltip");
    aboutUpdateLinkEl.setAttribute("data-tooltip", tooltip);
    aboutUpdateLinkEl.setAttribute("aria-label", tooltip);
  }
}

async function fetchMeta() {
  try {
    return await fetchJSON("/api/meta");
  } catch {
    const payload = await fetchJSON("/api/version");
    return { version: payload && payload.version ? String(payload.version) : "dev" };
  }
}

async function refreshReleaseStatus() {
  if (!appVersionEl) return;
  let release = null;
  let meta = null;
  try {
    release = await fetchJSON("/api/release");
    if (release && release.meta) {
      meta = release.meta;
    }
  } catch {
    release = null;
  }
  if (!meta) {
    try {
      meta = await fetchMeta();
    } catch {
      meta = { version: "dev" };
    }
  }
  applyAppVersion(meta);
  applyReleaseBadge(release);
}

async function refreshLogs() {
  if (!logsListEl) return;
  if (viewLogsEl.classList.contains("hidden")) return;
  try {
    const logs = await fetchJSON("/api/logs");
    const level = logsLevelMode || "all";
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

const pulseStates = new WeakMap();

function pulseButton(btn, durationMs = 2000) {
  if (!btn) return;
  const prevState = pulseStates.get(btn);
  if (prevState && prevState.timer) {
    window.clearTimeout(prevState.timer);
  }
  const icon = btn.querySelector(".icon-action");
  let originalIconClass = prevState ? prevState.originalIconClass : "";
  if (icon) {
    if (!originalIconClass) {
      originalIconClass = Array.from(icon.classList).find((name) => name.startsWith("icon-") && name !== "icon-action") || "";
    }
    if (originalIconClass) {
      icon.classList.remove(originalIconClass);
    }
    icon.classList.add("icon-rotate-clockwise-2");
  }
  btn.classList.add("is-pulsing");
  const timer = window.setTimeout(() => {
    btn.classList.remove("is-pulsing");
    const current = pulseStates.get(btn);
    const currentIcon = btn.querySelector(".icon-action");
    if (currentIcon) {
      currentIcon.classList.remove("icon-rotate-clockwise-2");
      if (current && current.originalIconClass) {
        currentIcon.classList.add(current.originalIconClass);
      }
    }
    pulseStates.delete(btn);
  }, durationMs);
  pulseStates.set(btn, { timer, originalIconClass });
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
const tooltipHoverQuery = window.matchMedia("(hover: hover) and (pointer: fine)");
const coarsePointerQuery = window.matchMedia("(pointer: coarse)");

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
    const statusLabel = String(target.dataset.statusLabel || "").trim() || "unknown";
    const lastChecked = formatLastChecked(target.dataset.checkedAt);
    const serverType = String(target.dataset.serverType || "").trim();
    const typeLabel = serverType === "local" ? "Local socket" : "Remote agent";
    const base = `${typeLabel} ${statusLabel}`;
    const label = lastChecked ? `${base} - ${lastChecked}` : base;
    target.setAttribute("data-tooltip", label);
    target.setAttribute("aria-label", label);
  }
  const text = target?.dataset?.tooltip || target?.getAttribute("aria-label") || "";
  if (!text) return;
  tooltipEl.classList.toggle("tooltip-compact", target?.dataset?.tooltipCompact === "true");
  tooltipEl.textContent = text;
  tooltipEl.classList.add("visible");

  const rect = target.getBoundingClientRect();
  const margin = 8;
  tooltipEl.style.left = "0px";
  tooltipEl.style.top = "0px";
  const tooltipRect = tooltipEl.getBoundingClientRect();

  let left = rect.left + rect.width / 2 - tooltipRect.width / 2;
  left = Math.max(margin, Math.min(left, window.innerWidth - tooltipRect.width - margin));

  const spaceAbove = rect.top - margin;
  const spaceBelow = window.innerHeight - rect.bottom - margin;
  const offset = 10;
  let placement = "top";
  let top = rect.top - tooltipRect.height - offset;
  if (spaceAbove < tooltipRect.height + offset && spaceBelow >= tooltipRect.height + offset) {
    placement = "bottom";
    top = rect.bottom + offset;
  } else if (spaceAbove < tooltipRect.height + offset && spaceBelow < tooltipRect.height + offset) {
    // Not enough space either side: pick the larger space and clamp.
    if (spaceBelow > spaceAbove) {
      placement = "bottom";
      top = rect.bottom + offset;
    } else {
      placement = "top";
      top = rect.top - tooltipRect.height - offset;
    }
  }
  if (top < margin) {
    top = margin;
  }
  if (top + tooltipRect.height > window.innerHeight - margin) {
    top = window.innerHeight - tooltipRect.height - margin;
  }

  tooltipEl.style.left = `${left}px`;
  tooltipEl.style.top = `${top}px`;
  tooltipEl.setAttribute("data-placement", placement);
}

function hideTooltip() {
  tooltipEl.classList.remove("visible");
  tooltipTarget = null;
}

function showCopiedFeedback(button) {
  if (!button) return;
  const previousTooltip = button.dataset ? button.dataset.tooltip : "";
  const previousAria = button.getAttribute("aria-label") || "";
  const icon = button.querySelector(".icon-action");
  if (icon) {
    icon.classList.remove("icon-copy");
    icon.classList.add("icon-copy-check");
  }
  if (button.dataset) {
    button.dataset.tooltip = "Copied!";
  } else {
    button.setAttribute("data-tooltip", "Copied!");
  }
  button.setAttribute("aria-label", "Copied!");
  tooltipTarget = button;
  showTooltip(button);
  window.setTimeout(() => {
    if (tooltipTarget === button) hideTooltip();
    if (button.dataset) {
      button.dataset.tooltip = previousTooltip;
    } else {
      button.setAttribute("data-tooltip", previousTooltip);
    }
    button.setAttribute("aria-label", previousAria);
    if (icon) {
      icon.classList.remove("icon-copy-check");
      icon.classList.add("icon-copy");
    }
  }, 950);
}

document.addEventListener("mouseover", (event) => {
  if (!tooltipHoverQuery.matches) return;
  const target = event.target.closest?.(".has-tooltip");
  if (!target) return;
  if (tooltipTarget === target) return;
  tooltipTarget = target;
  showTooltip(target);
});

document.addEventListener("mouseout", (event) => {
  if (!tooltipHoverQuery.matches) return;
  if (!tooltipTarget) return;
  if (event.relatedTarget && tooltipTarget.contains(event.relatedTarget)) return;
  hideTooltip();
});

window.addEventListener(
  "scroll",
  () => {
    if (tooltipTarget && tooltipHoverQuery.matches) {
      showTooltip(tooltipTarget);
    }
  },
  { passive: true }
);

window.addEventListener("resize", () => {
  if (tooltipTarget && tooltipHoverQuery.matches) {
    showTooltip(tooltipTarget);
  }
  updateTopbarHeight();
  updateMobileNavHeight();
});

function refreshViewData(view) {
  const tasks = [];
  tasks.push(refreshConfig().catch(() => {}));
  if (view === "containers") {
    if (containersViewMode === "stacks") {
      tasks.push(refreshStacks({ silent: true }).catch(() => {}));
    } else {
      tasks.push(refreshContainers({ silent: true }).catch(() => {}));
    }
  }
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

const mobileNavQuery = window.matchMedia("(max-width: 720px)");
let mobileNavHeight = 0;

function updateMobileNavHeight() {
  if (!topbarEl || !sidebar) return;
  if (!mobileNavQuery.matches) {
    document.documentElement.style.removeProperty("--mobile-nav-h");
    return;
  }
  const topbarHeight = topbarEl.getBoundingClientRect().height;
  const sidebarHeight = sidebar.getBoundingClientRect().height;
  mobileNavHeight = Math.ceil(topbarHeight + sidebarHeight);
  document.documentElement.style.setProperty("--mobile-nav-h", `${mobileNavHeight}px`);
  updateTopbarHeight();
}

mobileNavQuery.addEventListener("change", () => {
  updateMobileNavHeight();
});

function updateSidebarNavActive(nextView) {
  if (!sidebar) return;
  const hasVisibleContainersShortcuts = Array.from(sidebar.querySelectorAll(".containers-sidebar-link"))
    .some((btn) => !btn.classList.contains("hidden"));
  sidebar.querySelectorAll("[data-view]").forEach((btn) => {
    const view = btn.getAttribute("data-view");
    const mode = btn.getAttribute("data-containers-mode");
    let active = false;
    if (view !== nextView) {
      active = false;
    } else if (view === "containers") {
      if (mode) {
        active = containersViewMode === mode;
      } else {
        active = !hasVisibleContainersShortcuts || containersViewMode === "table";
      }
    } else {
      active = true;
    }
    btn.classList.toggle("active", active);
  });
}

function isSidebarCollapsed() {
  return document.body && document.body.dataset && document.body.dataset.sidebar === "collapsed";
}

function updateSidebarCollapseUI() {
  if (!sidebarCollapseToggleBtn) return;
  const collapsed = isSidebarCollapsed();
  const icon = sidebarCollapseToggleBtn.querySelector(".icon-action");
  if (icon) {
    icon.classList.toggle("icon-layout-sidebar-left-collapse", !collapsed);
    icon.classList.toggle("icon-layout-sidebar-left-expand", collapsed);
  }
  const label = collapsed ? "Expand sidebar" : "Collapse sidebar";
  sidebarCollapseToggleBtn.setAttribute("aria-label", label);
  sidebarCollapseToggleBtn.setAttribute("data-tooltip", label);
}

function setSidebarCollapsed(collapsed, persist = true) {
  if (!document.body) return;
  if (collapsed) {
    document.body.dataset.sidebar = "collapsed";
  } else if (document.body.dataset && document.body.dataset.sidebar) {
    delete document.body.dataset.sidebar;
  }
  if (persist) {
    try {
      localStorage.setItem(sidebarCollapsedStorageKey, collapsed ? "1" : "0");
    } catch {
      // ignore
    }
  }
  updateSidebarCollapseUI();
  updateTopbarHeight();
  updateMobileNavHeight();
}

function setView(nextView) {
  const prevView = currentView;
  if (["containers", "stacks", "images"].includes(nextView) && !isExperimentalEnabled(nextView)) {
    nextView = "status";
  }
  currentView = nextView;
  viewStatusEl.classList.toggle("hidden", nextView !== "status");
  viewSettingsEl.classList.toggle("hidden", nextView !== "settings");
  if (viewContainersEl) viewContainersEl.classList.toggle("hidden", nextView !== "containers");
  viewServersEl.classList.toggle("hidden", nextView !== "servers");
  viewLogsEl.classList.toggle("hidden", nextView !== "logs");
  if (topbarStatusEl) {
    topbarStatusEl.classList.toggle("hidden", nextView !== "status");
  }
  if (topbarSettingsEl) {
    topbarSettingsEl.classList.toggle("hidden", nextView !== "settings");
  }
  if (topbarContainersEl) {
    topbarContainersEl.classList.toggle("hidden", nextView !== "containers");
  }
  if (topbarServersEl) {
    topbarServersEl.classList.toggle("hidden", nextView !== "servers");
  }
  if (topbarLogsEl) {
    topbarLogsEl.classList.toggle("hidden", nextView !== "logs");
  }
  if (sidebarSearch && prevView !== nextView) {
    sidebarSearch.value = "";
    applySidebarFilter("");
    updateSidebarSearchUI();
  }
  updateTopbarHeight();
  updateMobileNavHeight();

  // Mobile UX: each view should start at the top instead of carrying over the
  // scroll position from the previous view.
  if (mobileNavQuery.matches) {
    hideTooltip();
    // Depending on the browser, the scroll container can be the window or `main`,
    // and scroll anchoring can nudge the viewport after content visibility changes.
    // Reset both to ensure we always start from the top.
    const resetScroll = () => {
      const mainEl = document.querySelector("main");
      if (mainEl) {
        mainEl.scrollTop = 0;
      } else {
        window.scrollTo({ top: 0, left: 0, behavior: "auto" });
        document.documentElement.scrollTop = 0;
        document.body.scrollTop = 0;
      }
    };
    resetScroll();
    requestAnimationFrame(resetScroll);
  }

  if (nextView === "logs") {
    startLogsAutoRefresh();
  } else {
    stopLogsAutoRefresh();
  }
  if (nextView === "containers") {
    const requestedMode = pendingContainersMode;
    pendingContainersMode = "";
    if (prevView !== "containers") {
      containersShellSelectedId = "";
      containersLogsSelectedId = "";
    }
    if (requestedMode) {
      setContainersViewMode(requestedMode);
    } else if (prevView !== "containers") {
      setContainersViewMode("table");
    }
    if (containersViewMode === "table") {
      refreshContainers({ silent: true }).catch(() => {});
    }
  } else {
    stopContainersAutoRefresh();
    stopStacksAutoRefresh();
    stopContainersResourcesAutoRefresh();
    if (prevView === "containers") {
      closeContainersShellSession("leave containers view");
      closeContainersLogsSession("leave containers view");
      setContainersLogsPaused(false);
      if (containersViewMode === "resources") {
        containersResourcesSelectedIds.clear();
        updateContainersResourcesPlaceholder();
      }
    }
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
  updateSidebarNavActive(nextView);
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
attachImmediateSave(expContainersInput);
attachImmediateSave(expContainersSidebarInput);
attachImmediateSave(expContainerShellInput);
attachImmediateSave(expContainerLogsInput);
attachImmediateSave(expContainerResourcesInput);
attachImmediateSave(expStacksInput);
attachImmediateSave(expImagesInput);
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

if (remoteModalSave) {
  remoteModalSave.addEventListener("click", async () => {
    const isLocal = serverModalTab === "local";
    if (isLocal) {
      const payload = {
        name: remoteModalNameInput ? remoteModalNameInput.value.trim() : "",
        socket: remoteModalHostInput ? remoteModalHostInput.value.trim() : "/var/run/docker.sock",
        public_ip: remoteModalPublicIPInput ? remoteModalPublicIPInput.value.trim() : "",
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
      return;
    }
    const payload = {
      name: remoteModalNameInput ? remoteModalNameInput.value.trim() : "",
      url: normalizeRemoteUrl(
        remoteModalHostInput ? remoteModalHostInput.value.trim() : "",
        remoteModalPortInput ? remoteModalPortInput.value.trim() : "8080"
      ),
      token: remoteModalTokenInput ? remoteModalTokenInput.value.trim() : "",
      public_ip: remoteModalPublicIPInput ? remoteModalPublicIPInput.value.trim() : "",
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

if (stacksNewBtn) {
  stacksNewBtn.addEventListener("click", () => {
    openStackModal("");
  });
}

if (stackModalSave) {
  stackModalSave.addEventListener("click", () => {
    saveStackFromModal();
  });
}

if (stackModalSaveIcon) {
  stackModalSaveIcon.addEventListener("click", () => {
    saveStackFromModal({ closeOnSuccess: false });
  });
}

if (stackModalComposeUp) {
  stackModalComposeUp.addEventListener("click", () => {
    runStackActionFromModal("up");
  });
}

if (stackModalComposeDown) {
  stackModalComposeDown.addEventListener("click", () => {
    runStackActionFromModal("down");
  });
}

if (stackModalCancel) {
  stackModalCancel.addEventListener("click", () => {
    closeStackModal();
  });
}

if (stackModalClose) {
  stackModalClose.addEventListener("click", () => {
    closeStackModal();
  });
}

if (stackEnvToggle) {
  stackEnvToggle.addEventListener("change", () => {
    updateStackEnvState();
  });
}
if (stackEnvToggleBtn && stackEnvToggle) {
  stackEnvToggleBtn.addEventListener("click", () => {
    stackEnvToggle.checked = !stackEnvToggle.checked;
    updateStackEnvState();
  });
}
if (stackComposeInput) {
  stackComposeInput.addEventListener("input", () => {
    clearStackModalError();
  });
}
if (stackEnvInput) {
  stackEnvInput.addEventListener("input", () => {
    clearStackModalError();
  });
}

if (remoteModalTokenCopy) {
  remoteModalTokenCopy.addEventListener("click", () => {
    copyToClipboard(remoteModalTokenInput ? remoteModalTokenInput.value : "", "Token", { silent: true })
      .then(() => showCopiedFeedback(remoteModalTokenCopy))
      .catch(() => showToast("Copy failed."));
  });
}

if (remoteModalComposeCopy) {
  remoteModalComposeCopy.addEventListener("click", () => {
    if (serverModalTab === "local") return;
    copyToClipboard(remoteModalCompose ? remoteModalCompose.textContent : "", "Compose", { silent: true })
      .then(() => showCopiedFeedback(remoteModalComposeCopy))
      .catch(() => showToast("Copy failed."));
  });
}

if (remoteModalPortInput) {
  remoteModalPortInput.addEventListener("input", () => {
    updateRemoteComposePreview();
  });
}

if (serverConnectionTypeSelect) {
  serverConnectionTypeSelect.addEventListener("change", () => {
    const value = serverConnectionTypeSelect.value === "local" ? "local" : "remote";
    setServerModalTab(value);
    if (remoteModalTitle) {
      remoteModalTitle.textContent = value === "local"
        ? (editingLocalServer ? "Edit local server" : "Add local server")
        : (editingRemoteServer ? "Edit remote server" : "Add remote server");
    }
    if (value === "local") {
      if (!editingLocalServer && remoteModalHostInput && !remoteModalHostInput.value.trim()) {
        remoteModalHostInput.value = "/var/run/docker.sock";
      }
    } else if (!editingRemoteServer && remoteModalTokenInput && !remoteModalTokenInput.value.trim()) {
      remoteModalTokenInput.value = generateToken(32);
      updateRemoteComposePreview();
    }
  });
}

// Connection type help uses tooltip (no modal)

async function init() {
  sidebar.setAttribute("aria-hidden", "false");

  if (sidebarSearch) {
    sidebarSearch.addEventListener("input", (event) => {
      applySidebarFilter(event.target.value);
    });
    sidebarSearch.addEventListener("focus", () => {
      updateSidebarSearchUI();
    });
    sidebarSearch.addEventListener("blur", () => {
      updateSidebarSearchUI();
    });
  }
  if (sidebarCollapseToggleBtn) {
    sidebarCollapseToggleBtn.addEventListener("click", () => {
      setSidebarCollapsed(!isSidebarCollapsed());
    });
  }
  initComposeEditor();
  initEnvEditor();
  enableYamlIndent(stackComposeInput);
  enableYamlIndent(stackEnvInput);
  updateContainersSortUI();
  updateStacksSortUI();
  updateImagesSortUI();
  updateContainersResourcesSortUI();
  updateContainersResourcesViewButtons();

  if (statusSelectiveToggleBtn) {
    statusSelectiveToggleBtn.addEventListener("click", () => {
      setSelectiveScanEnabled(!isSelectiveScanEnabled());
      refreshStatus();
    });
    updateSelectiveScanToggle();
  }

  if (serversFilterResetBtn) {
    serversFilterResetBtn.addEventListener("click", () => {
      setServersFilterMode("all");
      if (serversFilterBtn) {
        serversFilterBtn.focus();
      }
    });
  }

  if (serversViewToggleBtn) {
    serversViewToggleBtn.addEventListener("click", () => {
      const next = serversViewMode === "cards" ? "table" : "cards";
      setServersViewMode(next);
    });
  }

  if (containersResourcesViewToggleBtn) {
    containersResourcesViewToggleBtn.addEventListener("click", () => {
      const next = containersResourcesViewMode === "table" ? "cards" : "table";
      setContainersResourcesViewMode(next);
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

  if (topbarContainersServerBtn) {
    topbarContainersServerBtn.addEventListener("click", () => {
      toggleContainersServerMenu();
    });
  }
  if (serversFilterBtn) {
    serversFilterBtn.addEventListener("click", () => {
      toggleServersFilterMenu();
    });
  }
  if (expContainersInput) {
    expContainersInput.addEventListener("change", () => {
      updateContainersExperimentalToggles();
    });
  }
  if (logsLevelBtn) {
    logsLevelBtn.addEventListener("click", () => {
      toggleLogsLevelMenu();
    });
    renderLogsLevelButton(logsLevelMode);
    updateLogsLevelMenu(logsLevelMode);
  }
  if (containersNameSortBtn) {
    containersNameSortBtn.addEventListener("click", () => {
      if (containersSortMode !== "name:asc" && containersSortMode !== "name:desc") {
        containersSortMode = "name:asc";
      } else {
        containersSortMode = containersSortMode === "name:asc" ? "name:desc" : "name:asc";
      }
      updateContainersSortUI();
      const list = Array.from(containersCache.values()).map((entry) => entry.data);
      if (list.length > 0 && containersSelectedScope) {
        renderContainers(list, containersSelectedScope);
      }
    });
  }
  const toggleContainersSortMode = (key) => {
    if (!key) return;
    const asc = `${key}:asc`;
    const desc = `${key}:desc`;
    if (containersSortMode !== asc && containersSortMode !== desc) {
      containersSortMode = asc;
    } else {
      containersSortMode = containersSortMode === asc ? desc : asc;
    }
    updateContainersSortUI();
    const list = Array.from(containersCache.values()).map((entry) => entry.data);
    if (list.length > 0 && containersSelectedScope) {
      renderContainers(list, containersSelectedScope);
    }
  };
  if (containersCpuSortBtn) {
    containersCpuSortBtn.addEventListener("click", () => toggleContainersSortMode("cpu"));
  }
  if (containersRamSortBtn) {
    containersRamSortBtn.addEventListener("click", () => toggleContainersSortMode("ram"));
  }
  if (containersIpSortBtn) {
    containersIpSortBtn.addEventListener("click", () => toggleContainersSortMode("ip"));
  }
  if (containersPortSortBtn) {
    containersPortSortBtn.addEventListener("click", () => toggleContainersSortMode("port"));
  }
  const toggleResourcesSort = (key) => {
    if (!key) return;
    const asc = `${key}:asc`;
    const desc = `${key}:desc`;
    if (containersResourcesSortMode !== asc && containersResourcesSortMode !== desc) {
      containersResourcesSortMode = asc;
    } else {
      containersResourcesSortMode = containersResourcesSortMode === asc ? desc : asc;
    }
    updateContainersResourcesSortUI();
    if (currentView === "containers" && containersViewMode === "resources") {
      renderContainersResourcesView();
    }
  };
  if (resourcesNameSortBtn) {
    resourcesNameSortBtn.addEventListener("click", () => toggleResourcesSort("name"));
  }
  if (resourcesCpuSortBtn) {
    resourcesCpuSortBtn.addEventListener("click", () => toggleResourcesSort("cpu"));
  }
  if (resourcesRamSortBtn) {
    resourcesRamSortBtn.addEventListener("click", () => toggleResourcesSort("ram"));
  }
  if (resourcesStatusSortBtn) {
    resourcesStatusSortBtn.addEventListener("click", () => toggleResourcesSort("status"));
  }
  if (resourcesUptimeSortBtn) {
    resourcesUptimeSortBtn.addEventListener("click", () => toggleResourcesSort("uptime"));
  }
  if (resourcesIpSortBtn) {
    resourcesIpSortBtn.addEventListener("click", () => toggleResourcesSort("ip"));
  }
  if (resourcesPortSortBtn) {
    resourcesPortSortBtn.addEventListener("click", () => toggleResourcesSort("port"));
  }
  if (containersStateSortBtn) {
    containersStateSortBtn.addEventListener("click", () => {
      if (containersSortMode !== "state:asc" && containersSortMode !== "state:desc") {
        containersSortMode = "state:asc";
      } else {
        containersSortMode = containersSortMode === "state:asc" ? "state:desc" : "state:asc";
      }
      updateContainersSortUI();
      const list = Array.from(containersCache.values()).map((entry) => entry.data);
      if (list.length > 0 && containersSelectedScope) {
        renderContainers(list, containersSelectedScope);
      }
    });
  }
  if (stacksNameSortBtn) {
    stacksNameSortBtn.addEventListener("click", () => {
      if (stacksSortMode !== "name:asc" && stacksSortMode !== "name:desc") {
        stacksSortMode = "name:asc";
      } else {
        stacksSortMode = stacksSortMode === "name:asc" ? "name:desc" : "name:asc";
      }
      updateStacksSortUI();
      const list = Array.from(stacksCache.values()).map((entry) => entry.data);
      if (list.length > 0) {
        renderStacks(list);
      }
    });
  }
  if (stacksStateSortBtn) {
    stacksStateSortBtn.addEventListener("click", () => {
      if (stacksSortMode !== "state:asc" && stacksSortMode !== "state:desc") {
        stacksSortMode = "state:asc";
      } else {
        stacksSortMode = stacksSortMode === "state:asc" ? "state:desc" : "state:asc";
      }
      updateStacksSortUI();
      const list = Array.from(stacksCache.values()).map((entry) => entry.data);
      if (list.length > 0) {
        renderStacks(list);
      }
    });
  }
  if (imagesRepoSortBtn) {
    imagesRepoSortBtn.addEventListener("click", () => {
      if (imagesSortMode !== "repository:asc" && imagesSortMode !== "repository:desc") {
        imagesSortMode = "repository:asc";
      } else {
        imagesSortMode = imagesSortMode === "repository:asc" ? "repository:desc" : "repository:asc";
      }
      updateImagesSortUI();
      if (imagesCache.length > 0) {
        renderImages(imagesCache, containersSelectedScope);
      }
    });
  }
  if (imagesTagSortBtn) {
    imagesTagSortBtn.addEventListener("click", () => {
      if (imagesSortMode !== "tags:asc" && imagesSortMode !== "tags:desc") {
        imagesSortMode = "tags:asc";
      } else {
        imagesSortMode = imagesSortMode === "tags:asc" ? "tags:desc" : "tags:asc";
      }
      updateImagesSortUI();
      if (imagesCache.length > 0) {
        renderImages(imagesCache, containersSelectedScope);
      }
    });
  }
  if (imagesStateSortBtn) {
    imagesStateSortBtn.addEventListener("click", () => {
      if (imagesSortMode !== "state:asc" && imagesSortMode !== "state:desc") {
        imagesSortMode = "state:asc";
      } else {
        imagesSortMode = imagesSortMode === "state:asc" ? "state:desc" : "state:asc";
      }
      updateImagesSortUI();
      if (imagesCache.length > 0) {
        renderImages(imagesCache, containersSelectedScope);
      }
    });
  }
  if (imagesSizeSortBtn) {
    imagesSizeSortBtn.addEventListener("click", () => {
      if (imagesSortMode !== "size:asc" && imagesSortMode !== "size:desc") {
        imagesSortMode = "size:asc";
      } else {
        imagesSortMode = imagesSortMode === "size:asc" ? "size:desc" : "size:asc";
      }
      updateImagesSortUI();
      if (imagesCache.length > 0) {
        renderImages(imagesCache, containersSelectedScope);
      }
    });
  }
  if (topbarContainersRefreshBtn) {
    topbarContainersRefreshBtn.addEventListener("click", async () => {
      pulseButton(topbarContainersRefreshBtn, 2000);
      if (containersViewMode === "stacks") {
        await refreshStacks({ silent: true });
      } else if (containersViewMode === "images") {
        await refreshImages({ silent: true });
      } else if (containersViewMode === "resources") {
        await refreshContainers({ silent: true });
        await refreshContainersResources({ silent: true });
      } else {
        await refreshContainers({ silent: true });
      }
    });
  }
  if (topbarContainersShellBtn) {
    topbarContainersShellBtn.addEventListener("click", () => {
      const enabled = currentConfig && currentConfig.experimental_features
        ? Boolean(currentConfig.experimental_features.container_shell)
        : false;
      if (!enabled) {
        showToast("Container shell is disabled in Experimental features.");
        return;
      }
      setContainersViewMode(containersViewMode === "shell" ? "table" : "shell");
    });
  }
  if (topbarContainersLogsBtn) {
    topbarContainersLogsBtn.addEventListener("click", () => {
      const enabled = currentConfig && currentConfig.experimental_features
        ? Boolean(currentConfig.experimental_features.container_logs)
        : false;
      if (!enabled) {
        showToast("Container logs is disabled in Experimental features.");
        return;
      }
      setContainersViewMode(containersViewMode === "logs" ? "table" : "logs");
    });
  }
  if (topbarContainersResourcesBtn) {
    topbarContainersResourcesBtn.addEventListener("click", () => {
      const enabled = currentConfig && currentConfig.experimental_features
        ? Boolean(currentConfig.experimental_features.container_resources)
        : false;
      if (!enabled) {
        showToast("Container resources is disabled in Experimental features.");
        return;
      }
      setContainersViewMode(containersViewMode === "resources" ? "table" : "resources");
    });
  }
  if (topbarContainersStacksBtn) {
    topbarContainersStacksBtn.addEventListener("click", () => {
      const enabled = currentConfig && currentConfig.experimental_features
        ? Boolean(currentConfig.experimental_features.stacks)
        : false;
      if (!enabled) {
        showToast("Container stacks is disabled in Experimental features.");
        return;
      }
      setContainersViewMode(containersViewMode === "stacks" ? "table" : "stacks");
    });
  }
  if (topbarContainersImagesBtn) {
    topbarContainersImagesBtn.addEventListener("click", () => {
      const enabled = currentConfig && currentConfig.experimental_features
        ? Boolean(currentConfig.experimental_features.images)
        : false;
      if (!enabled) {
        showToast("Container images is disabled in Experimental features.");
        return;
      }
      setContainersViewMode(containersViewMode === "images" ? "table" : "images");
    });
  }

  if (imagesPullBtn) {
    imagesPullBtn.addEventListener("click", openImagesPullModal);
  }
  if (imagesPruneUnusedBtn) {
    imagesPruneUnusedBtn.addEventListener("click", () => {
      if (!imagesActionConfirming.has("prune-unused")) {
        toggleImagesConfirm("prune-unused");
        return;
      }
      runImagePrune("unused");
    });
  }
  if (imagesPruneDanglingBtn) {
    imagesPruneDanglingBtn.addEventListener("click", () => {
      if (!imagesActionConfirming.has("prune-dangling")) {
        toggleImagesConfirm("prune-dangling");
        return;
      }
      runImagePrune("dangling");
    });
  }
  if (imagesRemoveBtn) {
    imagesRemoveBtn.addEventListener("click", () => {
      if (!imagesActionConfirming.has("remove")) {
        toggleImagesConfirm("remove");
        return;
      }
      runImageRemove();
    });
  }
  if (imagesPullConfirmBtn) {
    imagesPullConfirmBtn.addEventListener("click", runImagePull);
  }
  if (imagesPullCloseBtn) {
    imagesPullCloseBtn.addEventListener("click", closeImagesPullModal);
  }
  if (imagesPullCancelBtn) {
    imagesPullCancelBtn.addEventListener("click", closeImagesPullModal);
  }
  if (imagesPullModal) {
    imagesPullModal.addEventListener("click", (event) => {
      if (event.target && event.target.dataset && event.target.dataset.close) {
        closeImagesPullModal();
      }
    });
    imagesPullModal.addEventListener("click", () => {
      cancelImagesPullAutoClose();
    });
    imagesPullModal.addEventListener("keydown", () => {
      cancelImagesPullAutoClose();
    });
  }

  document.addEventListener("click", (event) => {
    if (imagesActionConfirming.size === 0) return;
    const target = event.target;
    if (!target || !target.closest) return;
    const inside = target.closest("#images-prune-unused, #images-prune-dangling, #images-remove");
    if (!inside) {
      clearImagesConfirmations();
    }
  });

  document.addEventListener("click", (event) => {
    if (currentView !== "containers" || containersViewMode !== "images") return;
    if (!imagesSelectedId) return;
    const target = event.target;
    if (!target || !target.closest) return;
    if (target.closest("#images-pull-modal")) return;
    if (target.closest("#images-remove, #images-prune-unused, #images-prune-dangling, #images-pull")) return;
    if (target.closest("#containers-images-view .containers-table-wrap")) return;
    clearImagesSelection();
  });

  if (containersLogsClearBtn) {
    containersLogsClearBtn.addEventListener("click", () => {
      resetContainersLogsBuffer();
    });
  }
  if (containersLogsPauseBtn) {
    setContainersLogsPaused(false);
    containersLogsPauseBtn.addEventListener("click", () => {
      setContainersLogsPaused(!containersLogsPaused);
    });
  }
  if (containersLogsTimestampsBtn) {
    setContainersLogsTimestamps(false);
    containersLogsTimestampsBtn.addEventListener("click", () => {
      setContainersLogsTimestamps(!containersLogsTimestamps);
      if (!containersLogsSelectedId) return;
      const entry = containersCache.get(containersLogsSelectedId);
      if (entry && entry.data) {
        setContainersLogsPaused(false);
        openContainersLogs(entry.data, { preserve: false, silentClose: true });
      }
    });
  }
  if (containersLogsOutput) {
    containersLogsOutput.addEventListener("scroll", () => {
      const threshold = 24;
      const distance = containersLogsOutput.scrollHeight - containersLogsOutput.scrollTop - containersLogsOutput.clientHeight;
      containersLogsAutoScroll = distance <= threshold;
    });
  }

  sidebar.querySelectorAll("[data-view]").forEach((btn) => {
    btn.addEventListener("click", () => {
      const nextView = btn.getAttribute("data-view");
      if (nextView) {
        if (nextView === "containers") {
          pendingContainersMode = btn.getAttribute("data-containers-mode") || "table";
        } else {
          pendingContainersMode = "";
        }
        setView(nextView);
      }
      // On touch devices, keep taps from leaving a "stuck" focused nav button,
      // which can look like a hover/active highlight.
      if (coarsePointerQuery.matches && btn && typeof btn.blur === "function") {
        btn.blur();
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
  if (stackModal) {
    stackModal.addEventListener("click", (event) => {
      if (event.target && event.target.dataset && event.target.dataset.close) {
        closeStackModal();
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

  document.addEventListener("click", (event) => {
    if (!topbarContainersServerMenu || !topbarContainersServerBtn) return;
    if (topbarContainersServerBtn.contains(event.target)) return;
    if (topbarContainersServerMenu.contains(event.target)) return;
    closeContainersServerMenu();
  });

  document.addEventListener("click", (event) => {
    if (!serversFilterMenu || !serversFilterBtn) return;
    if (serversFilterBtn.contains(event.target)) return;
    if (serversFilterMenu.contains(event.target)) return;
    closeServersFilterMenu();
  });

  document.addEventListener("click", (event) => {
    if (!logsLevelMenu || !logsLevelBtn) return;
    if (logsLevelBtn.contains(event.target)) return;
    if (logsLevelMenu.contains(event.target)) return;
    closeLogsLevelMenu();
  });

	  document.addEventListener(
	    "click",
	    (event) => {
	      if (containersKillConfirming.size === 0) return;
	      const target = event.target;
	      if (target && target.closest && target.closest(".containers-kill-btn")) return;
	      clearContainersKillConfirmations();
	    },
	    true
	  );

	  document.addEventListener(
	    "click",
	    (event) => {
	      if (stacksActionConfirming.size === 0) return;
	      const target = event.target;
	      if (target && target.closest && target.closest(".containers-kill-btn")) return;
	      clearStacksActionConfirmations();
	    },
	    true
	  );

  initThemeToggle();
  updateTopbarHeight();
  updateMobileNavHeight();
  try {
    const stored = localStorage.getItem(sidebarCollapsedStorageKey);
    if (stored === "1") {
      setSidebarCollapsed(true, false);
    } else if (stored === "0") {
      setSidebarCollapsed(false, false);
    }
  } catch {
    // ignore
  }
  updateSidebarCollapseUI();
  const storedServersView = localStorage.getItem(serversViewStorageKey);
  if (storedServersView === "cards" || storedServersView === "table") {
    serversViewMode = storedServersView;
  }
  const storedResourcesView = localStorage.getItem(containersResourcesViewStorageKey);
  if (storedResourcesView === "cards" || storedResourcesView === "table") {
    containersResourcesViewMode = storedResourcesView;
  }
  updateContainersResourcesViewButtons();
  setView("status");
  setScanningUI(false);
  await refreshConfig();
  await refreshServers();
  initServerStream();
  await triggerServersRefresh();
  triggerStatusRefresh().catch(() => {});
  renderStatusPlaceholders();
  refreshStatus().catch(() => {});
  await refreshReleaseStatus();
  if (!releaseCheckTimer) {
    releaseCheckTimer = window.setInterval(() => {
      refreshReleaseStatus().catch(() => {});
    }, releaseCheckPollMs);
  }
  statusHintEl.classList.toggle("hidden", cachedLocals.length > 0);
}

init().catch((err) => {
  statusEl.textContent = `Error: ${err.message}`;
});
