# Changelog

## v1.2.3

## Bug fixes
- Images: fixed remote agent mode requests (no longer incorrectly passing `server=<remote name>` which could cause `local server not found`).

## New features
- UI: Update notification in UI - version footer now links to the GitHub release and can show an update badge when a newer release is available.

## Improvements
- Containers: IP/Port columns now populate immediately on first load (no need to wait for `/api/containers/resources`).
- Containers: CPU/RAM now load in a 2-step way (server-side cache returned immediately when available, then refreshed in background).
- Containers: adjusted table column order + narrower RAM column (999 MB max) to free space for wider Uptime.


## v1.2.2

## Bug fixes

## New features

## Improvements
- Container Stacks: UI adjustment, space using optimisation.
- Container Stacks: New/Edit stack modal now includes quick actions (Save, Compose up/down) in the header row.
- UI: Improving user experience by icon-action increase (without changing button size).
- Containers Resources: new experimental Resources view with multi-select cards, pins, and 5s resource refresh.
- Containers Resources: faster initial loading (parallel fetch), list scroll no longer jumps on refresh, IPv4-only port bindings, and a more compact card layout.
- Servers: Remote/Local server modal now includes optional `public_ip` field for opening container services.
- Containers Resources: UI adjustment, space using optimisation, introducing icons
- Containers Resources: hyperlink on ports - link to open container services.
- Containers Resources: added Cards/Table view toggle with a Containers-style table (sortable columns, pinned rows section, default Name asc).
- Containers: table view now includes CPU/RAM/IP/Port columns (sortable).
- Images: repository grouping now only shows a group header when a repository has 2+ images.
- UI: moved the global search field from the sidebar to the topbar (centered).
- UI: added collapsible sidebar (icon-only mode with banner swapped to logo).
- Servers: Add/Edit server modal redesigned with Connection type selector (Remote agent vs Local Unix socket) plus inline help and consistent layout.

## v1.2.1

## Bug fixes
- Container stacks: `docker compose` actions no longer fail when stack name contains uppercase letters (project name is normalized to lowercase).
- Container stacks: validation markers now point to the offending line more reliably (includes `Line N` in the tooltip; best-effort mapping for Compose config errors).

## New features
- Container stacks: Compose/.env editor migrated to CodeMirror 6 (bundled in the image; no CDN).
- Container stacks: inline validation while editing (Compose via `/api/stacks/validate`, `.env` via local lint rules).
- Container stacks: destructive actions (Remove/Kill) now use an inline confirm click (no browser `confirm()` popup).
- Container stacks: optimistic per-action status labels while an operation is in progress (Deploying/Teardown/Restarting/Killing/Removing).
- Container images: new images view with repository grouping, manual refresh, pull, prune (unused/dangling), and remove-by-image-id.

## Improvements
- Container stacks: removed row expand/collapse (container list under a stack) to simplify the table UX (API `/api/stacks/containers` removed as well).
- Editor UX: YAML-friendly defaults (2-space indent, word wrap, folding, indent guides) plus basic autocomplete for common Compose keys/values.
- Dark mode: improved editor theming (gutter/cursor/tooltips/autocomplete are theme-aware).
- Containers UI: Shell/Logs/Stacks topbar title now reflects the current view.
- Containers UI: consistent typography for container lists (table vs shell/logs list), including header styling.
- Containers UI: name-leading icons added plus spacing tweaks.
- Logs UI: sidebar label updated to “Events” (container logs remain unchanged).
- Settings: optional sidebar shortcuts for container subviews (Shell/Logs/Stacks/Images) via “Container features in side bar”.

## v1.2.0

## New features
- UI ergonomy improvements
- Containers feature (experimental):
    - settings: Experimental features toggle section (Containers + Container shell/logs/stacks/images)
    - new view with server picker and container table (Name, Image, State, Uptime, Stack) plus actions (start/stop/restart/pause/kill).
    - per-column sort icons for Name and State; inline confirmation for Kill.
    - shell view (top bar button) with selectable container list and embedded terminal.
    - logs view (top bar button) with selectable container list, live stream, pause/resume, and optional timestamps.
- Container stacks (experimental):
    - stacks table with status, container counts, and per-stack actions.
    - create/edit stack compose files without deploying.
    - stack containers expansion with downed containers highlighted.

## Improvements
- Containers feature:
    - auto-refresh every 5s with in-place row updates (no list reset), manual refresh, and sort by name/state
    - action buttons use colored icons with hover backgrounds; state column uses badge-style labels
    - logs: default tail reduced to 100 lines for faster initial load.
- Container stacks:
    - stored in `/data/stacks/<server>/<stack>/` with `docker-compose.yml` and optional `.env`.
    - compose actions available: up/down/start/stop/restart/kill/rm.
- Server reachability: checking state no longer overrides online/offline once known.
- Remote shell/logs proxy: improved WebSocket close handling when switching sessions.

## Bug fixes
- Settings: Containers-dependent experimental toggles are disabled when Containers is off.
- Servers/Updates: fixed cases where remote servers could remain stuck in Checking in the UI.

## v1.1.2

## New features

## Improvements
- Mobile and responsive view: improvements and adjustments. Now mobile view is usable.
- UI: further icons implementations and UI ergonomy fixes.

## Bug fixes

## v1.1.1

## New features
- Update API now returns `old_image_id`, `new_image_id`, and `applied_image_id` for troubleshooting.

## Improvements
- Status cards: Details action now uses an icon button.
- Settings: Renamed “Global policy” to “Global update policy”.
- Update logs now include the old/new/applied image IDs.
- Servers: Server actions now use icon buttons (Edit, Maintenance/End Maintenance, Remove/Confirm remove).
- Servers: Status filter reset icon is disabled when the filter is set to All and shows “Clear filter” tooltip.
- Servers (table view): Column sizing improved (Address narrower, Status wider) and row content vertically centered.
- Status: Update checks no longer pull images; they compare registry vs local digest and keep pulls for updates only.
- Logs: Scan now reports “digest unknown”/registry digest issues in API logs for easier troubleshooting.

## Bug fixes
- Update action now verifies the recreated container uses the pulled image and retries once using the resolved image ID if needed.
- Servers: Remove flow no longer changes Edit into Cancel; clicking outside Confirm now cancels the remove confirmation.
- Status: Selective scan mode now auto-clears after a scan completes or after “Check connection”.
- Status: When updating a stopped container with “Update stopped containers (keep them stopped)” disabled, UI now shows a clear message explaining how to enable updates.
- Status: Unknown digest checks are marked as skipped with a short explanation in Details → Info.

## v1.1.0

## New features
- Live server status updates via server stream (SSE) with on-demand reachability checks.
- New CHECKING state for servers awaiting reachability confirmation.
- Maintenance mode for local/remote servers, reflected in status and scan targeting.
- New add/edit modals for local and remote servers, including token + compose copy for agents.

## Improvements
- Servers view rebuilt with table/cards layout, active/maintenance filters, and view toggle.
- Sidebar search now filters the Servers list.
- Server list renders immediately; health/status updates arrive asynchronously.
- Server name rows now use consistent status/type icons.
- Status, Settings, Servers and Logs headers/actions moved into the top bar.
- Servers filter controls consolidated into a dropdown (All/Active/Maintenance/Offline) with a one-click reset icon.
- Add server flow unified into a single modal with Local/Remote tabs.
- Servers view toggle consolidated into a single icon switch (table/cards).
- Manual refresh buttons for Servers and Status (reachability checks are on-demand).
- Reachability checks now use `/api/version`; `Last checked` reflects connection checks, not scan time.
- Added `POST /api/status/refresh` to repopulate remote scan snapshots after controller restarts.
- Added `POST /api/self-update` for safe agent self-updates.

## Bug fixes
- Offline servers no longer block the initial Servers view render.
- Newly added offline servers appear immediately while status checks run.
- Remove-confirm state no longer resets on background refresh.
- Fixed status races where stale reachability snapshots could keep a server stuck in CHECKING.
- Fixed agent self-update flow: controller now refreshes status after agent restarts and outdated counts clear correctly.

## v1.0.1

- Global policy is now synced from the controller to remote agents.
- Scheduled scans now run scan-only first, then perform updates with an explicit updating phase (local + remote).
- Status cards show an Updating badge and an Updated metric for recent updates.
- Status summary now includes a Skipped metric.

## v1.0.0

## :fire: First main release after initial development phase

## New features
- Status view redesigned with server cards, inline actions, and richer scan state badges.
- Details modal rebuilt with collapsible groups and container cards for updates and scanned items.
- Persistent local scan state across restarts.
- New notification toggles in Settings for update events.
- Remote agent support improvements: clearer server list indicators and diagnostics.

## Improvements
- Unified icon system across sidebar, status, servers list, and settings.
- Custom tooltips with better positioning and theme-aware styling.
- Theme toggle simplified to light/dark with icon indicators.
- Status summary metrics updated with clearer labels and icons.
- Global scheduler indicators added to status and server cards.

## Bug fixes
- Scan state handling corrected (pending/scanning visibility and cleanup).
- Local/remote status handling refined during scan cancel and restart.
- Self-update flow stabilized and helper logs surfaced for troubleshooting.
