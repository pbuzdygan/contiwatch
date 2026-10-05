# User guide

## Container policies

Set a container label to override the global policy:

- `contiwatch.policy=notify_only`: compares registry metadata with the local image digest and detects updates without recreating containers. Notifications depend on the Discord settings.
- `contiwatch.policy=update`: pulls the image and recreates the container with its existing configuration when an update is needed. A previously running container is started again.
- `contiwatch.policy=skip`: skips update checking for the container and records it as intentionally skipped.

The default global policy is `notify_only`. Controllers synchronize it to remote agents; labels still take precedence. Stopped containers are skipped for updates unless `update_stopped_containers` is enabled, in which case they are updated but remain stopped. Paused containers are skipped. The status summary includes a Skipped metric.

## Scans, servers, and events

Run a manual scan from Updates or enable the scheduler in settings. Automatic scans are disabled by default. Basic, Cron, and legacy interval schedules are described in [configuration](configuration.md#scheduler).

A full manual or scheduled batch has no shared wall-clock deadline. Individual remote scan, update, and restart-verification operations remain bounded. Stop a manual batch through the UI or `POST /api/scan/stop`.

Server maintenance mode pauses scans and updates per server. `Last scan` refers to the update-check scan; `Last checked` in server tooltips refers to a reachability check, not a scan. Events supports All, Info, Warn, and Error level filters. A Debug filter and debug configuration are still [proposed](events-debug-filter-plan.md), not available settings.

Settings → Menu visibility controls access to container management views and optional sidebar shortcuts. Views cover containers, stacks, images, networks, volumes, shell, logs, and resource metrics. Notification scenarios and their switches are described in [Discord notifications](discord-webhook.md).

## Stack editor and env

The stack editor stores `docker-compose.yml` and an optional `.env` independently. The `.env` file is passed to Docker Compose for `${VARIABLE}` interpolation and placed next to the Compose file for local and remote operations.

The editor does not add or remove service-level `env_file` entries. Add `env_file: .env` to a service only when all values should also be injected into its container environment. It is not required for Compose interpolation. Removing `.env` is an explicit, confirmed editor action.

Validation evaluates the interpolated Compose model. Stack `.env` variables take precedence over same-named variables inherited from the Contiwatch process, preventing controller configuration from changing a managed stack accidentally.

In New stack and Edit stack, clicking outside the editor, Close, or Escape keeps the editor open when Name, Compose, or `.env` has unsaved changes. The warning lists the changed fields. Save validates and persists the configuration; Cancel discards unsaved edits and leaves the previously saved files intact. Saving with the toolbar icon keeps the editor open. Closing and editing are disabled while loading, saving, or running an editor action.

The editor's Compose up, Compose down, Redeploy, and Restart stack buttons all validate and save the current configuration first. If validation or saving fails, the operation does not start. Once saving succeeds, the configuration remains saved even if the subsequent operation fails; Cancel cannot undo that save. The corresponding buttons on the stacks list operate on the saved files directly.

- **Compose up** runs `docker compose up -d`, creating, starting, or recreating services as needed to apply the saved configuration.
- **Compose down** runs `docker compose down`, stopping and removing the stack's containers and networks without requesting volume deletion.
- **Redeploy** runs `docker compose pull` followed by `docker compose up -d`; it does not force recreation of unchanged containers.
- **Restart stack** runs `docker compose restart` for existing containers. Restart does not apply changed service settings or environment variables to those containers; use Compose up or Redeploy to apply such changes.

### Background actions and timeouts

Stack actions run as background jobs polled by the UI. A second action for the same stack is rejected while one runs. Limits are 10 minutes for `up`/`pull`, 20 minutes for `redeploy`, and 3 minutes for other actions.

About 90 seconds before the limit, a warning offers an Extend button that adds 10 minutes. Without extension, the action stops when its timeout expires. Remote extension requires agent version `1.3.4` or newer; older agents fall back to synchronous execution without extension. See the [stack API](api.md#stacks).

## Server health check

The Servers view provides an on-demand health check for local Docker daemons and remote Contiwatch agents. It reports Docker Engine/API versions, OS/architecture, CPU and memory capacity, container state and health counts, and Docker disk usage for images, containers, volumes, and build cache. Storage results include potentially reclaimable amounts reported by Docker.

It does not report live host CPU/RAM percentages or host filesystem free space, which require host-level metrics beyond Docker socket access. Health data is collected only on request because disk usage calculation can be expensive.

## Updates and agent self-update

Update results include old, new, and applied image IDs for troubleshooting. Scan entries use `self=true` only for the container running the responding Contiwatch process. Controllers update that agent last and verify its post-restart status. Detection also uses Docker runtime metadata so a stale generated hostname does not route self-update through regular in-process recreation.

Legacy agents without the self marker are recognized conservatively by their Contiwatch image and agent-style container name. See [API update contracts](api.md#scan-and-update-contracts) for the helper endpoint and [installation](installation.md#published-images-and-upgrades) for deployment upgrades.
