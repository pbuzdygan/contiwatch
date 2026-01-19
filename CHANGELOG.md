# Changelog

## v1.0.2

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
- Status, Settings, and Logs headers/actions moved into the top bar.
- Servers header and actions moved into the top bar for consistency.
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
