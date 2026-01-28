# Contiwatch

<p align="center">
  <img src="branding/contiwatch_banner.png" alt="CONTIWATCH Banner" width="50%">
</p>

**Contiwatch** is minimal Docker image watcher inspired by Watchtower. Scans local containers, checks for new images, optionally recreates containers, and sends Discord webhook notifications.

## Features
- ✅ Scan local Docker daemon for running containers
- ✅ Remote agent support (token-authenticated)
- ✅ Pull image tags and detect updates
- ✅ Global and per-container policy (`contiwatch.policy` label)
- ✅ Optional update (recreate container) or notify-only
- ✅ Simple HTML UI for updates, servers, events, and settings
- ✅ Server maintenance mode to pause scans and updates per server
- ✅ Experimental containers management UI (opt-in)
- ✅ Experimental container shell (opt-in)
- ✅ Experimental container logs (opt-in)
- ✅ Discord webhook notifications

---
## Demo / Screenshots

### Main UI

## Policies
Set on containers via label:
- `contiwatch.policy=update`
- `contiwatch.policy=notify_only`
- `contiwatch.policy=skip`

Global default is `notify_only` unless changed in the UI or config.

Policy behavior:
- `notify_only`: pulls image metadata, detects updates, sends notification only (no container changes).
- `update`: pulls the image and recreates the container with the same config; it only starts the new container if it was running before. By default, non-running containers are skipped for updates (reported as `Skipped`); enable `update_stopped_containers` to update stopped containers but keep them stopped.
- `skip`: ignores the container entirely.

When using remote agents, the controller syncs `global_policy` to agents; per-container labels still override the global setting.
Status summary includes a Skipped metric for containers that were intentionally skipped.

## Run (container)
```bash
docker build -t contiwatch .

docker run -d \
  --name contiwatch \
  -p 8080:8080 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v contiwatch-data:/data \
  contiwatch
```

Open `http://localhost:8080`.

## Agent mode (remote)
Run on a remote host with Docker socket access and a token:
```bash
docker run -d \
  --name contiwatch-agent \
  -p 8080:8080 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v contiwatch-agent-data:/data \
  -e CONTIWATCH_AGENT=true \
  -e CONTIWATCH_AGENT_TOKEN="<TOKEN>" \
  contiwatch
```

## Container images (GHCR)
Main releases (multi-arch):
```bash
docker pull ghcr.io/<owner>/<repo>:latest
docker pull ghcr.io/<owner>/<repo>:<version>
```

Dev releases (multi-arch):
```bash
docker pull ghcr.io/<owner>/<repo>:dev_latest
docker pull ghcr.io/<owner>/<repo>:dev_<version>
```

## Environment
- `CONTIWATCH_ADDR` (default `:8080`)
- `CONTIWATCH_CONFIG` (default `/data/config.json`)
- `TZ` (optional; e.g. `Europe/Warsaw` for local timestamps)
- `CONTIWATCH_AGENT` (optional; set to `true` to run in agent mode)
- `CONTIWATCH_AGENT_TOKEN` (required in agent mode; bearer token for API access)

## Config file
`/data/config.json` fields include:
- `scan_interval_sec` (in seconds; UI shows minutes)
- `scheduler_enabled` (if `true`, periodic scans run every `scan_interval_sec`)
- `global_policy`
- `discord_webhook_url`
- `discord_notifications_enabled` (if `false`, no Discord notifications are sent)
- `discord_notify_on_start`
- `discord_notify_on_update_detected`
- `discord_notify_on_container_updated`
- `update_stopped_containers` (if `true`, `update` policy also updates stopped containers but keeps them stopped)
- `prune_dangling_images` (if `true`, prune dangling images after updates)
- `experimental_features` (object of feature flags: `containers`, `containers_sidebar`, `stacks`, `images`, `container_shell`, `container_logs`)
- `experimental_features.container_shell` (enables container shell UI)
- `experimental_features.container_logs` (enables container logs UI)
- `experimental_features.stacks` / `experimental_features.images` (enables Container stacks/images buttons in the Containers top bar)
- `experimental_features.containers_sidebar` (shows enabled container subfeatures in the sidebar as shortcuts)
- `local_servers` (list of local Docker daemons with `name`, `socket`, and optional `maintenance`)
- `remote_servers` (list of remote servers with `name`, `url`, optional `token`, and optional `maintenance`)

## API
- `GET /api/version`
- `POST /api/scan` run scan
- `POST /api/scan/stop` cancel scan
- `GET /api/scan/state` scan running status
- `GET /api/status` last scan (local or agent)
- `GET /api/aggregate` local + remote status
- `GET/PUT /api/config`
- `GET/POST /api/servers` (remote servers)
- `DELETE /api/servers/{name}`
- `GET/POST /api/locals` (local servers)
- `GET /api/servers/info` versions + reachability
- `GET /api/servers/stream` live server info + scan updates (SSE)
- `POST /api/servers/refresh` trigger on-demand reachability checks (updates stream + returns snapshot)
- `POST /api/status/refresh` pull last scan snapshots from online agents (updates stream)
- `GET /api/containers?scope=local:{name}|remote:{name}` list containers for a selected server
- `POST /api/containers/action` run container action (`start`, `stop`, `restart`, `pause`, `unpause`, `kill`)
- `GET /api/containers/shell` (WebSocket) interactive shell for a container
- `GET /api/containers/logs` (WebSocket) stream logs for a container
- `GET /api/images?scope=local:{name}|remote:{name}` list images for a selected server
- `POST /api/images/pull` pull image (`repository`, optional `tag`)
- `POST /api/images/prune` prune images (`mode=unused|dangling`)
- `POST /api/images/remove` remove image by `image_id`
- `GET /api/stacks?scope=local:{name}|remote:{name}` list compose stacks stored on the controller
- `GET /api/stacks/get?scope=local:{name}|remote:{name}&name={stack}` fetch compose + env content
- `PUT /api/stacks/save` save compose + env without deploy
- `POST /api/stacks/validate` validate compose yaml (Docker Compose config)
- `POST /api/stacks/action` run stack action (`up`, `down`, `start`, `stop`, `restart`, `kill`, `rm`)
- `POST /api/update/{container_id}` update container
- `POST /api/self-update?container={container_id}` update agent container via helper (agent mode only)
- `GET/POST/DELETE /api/logs`
- `POST /api/notifications/test` test Discord webhook

Notes:
- `POST /api/scan` is a one-off trigger; if a scan is already running it returns `409`.
- `POST /api/update/{container_id}` returns `old_image_id`, `new_image_id`, and `applied_image_id` to help debug tag/image mismatches.
- Periodic scans are disabled by default; enable via `scheduler_enabled` in the config (UI).
- Agent mode exposes a limited API surface (token required).

UI timing:
- `Last scan` refers to the last update-check scan (`/api/scan` / agent `/api/status`).
- `Last checked` in server tooltips refers to the last reachability check (`POST /api/servers/refresh`), not the scan time.

## Run (docker compose)
```bash
docker compose up -d --build
```

Stop:
```bash
docker compose down
```

## Permissions
If `/data/config.json` shows permission errors, set user IDs in compose:
```yaml
environment:
  PUID: "${PUID}"
  PGID: "${PGID}"
```
Use your host user IDs (from `id`).

If you see `permission denied while trying to connect to the Docker daemon socket at unix:///var/run/docker.sock`, ensure the socket is mounted and the container user can access it. Contiwatch tries to detect the socket group ID automatically; if your setup needs it explicitly, set `DOCKER_GID` to the group id of `/var/run/docker.sock` on the host (e.g. from `stat -c '%g' /var/run/docker.sock`).

## Buy Me a Coffee
If You like results of my efforts, feel free to show that by supporting me.

[!["Buy Me A Coffee"](https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png)](https://www.buymeacoffee.com/pbuzdygan)
<p align="left">
  <img src="branding/bmc_qr.png" width="25%" alt="BMC QR code">
</p>
