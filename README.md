# Contiwatch

Minimal Docker image watcher inspired by Watchtower. Scans local containers, checks for new images, optionally recreates containers, and sends Discord webhook notifications.

## Features
- Scan local Docker daemon for running containers
- Pull image tags and detect updates
- Global and per-container policy (`contiwatch.policy` label)
- Optional update (recreate container) or notify-only
- Simple HTML UI for status + remote instance list
- Discord webhook notifications

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
- `scan_interval_sec`
- `scheduler_enabled` (if `true`, periodic scans run every `scan_interval_sec`)
- `global_policy`
- `discord_webhook_url`
- `discord_notifications_enabled` (if `false`, no Discord notifications are sent)
- `update_stopped_containers` (if `true`, `update` policy also updates stopped containers but keeps them stopped)
- `local_servers` (list of local Docker daemons with `name` and `socket`)
- `remote_servers` (list of remote servers with `name`, `url`, and optional `token`)

## API
- `POST /api/scan` run scan
- `GET /api/status` last scan
- `GET /api/aggregate` local + remote status
- `GET/PUT /api/config`
- `GET/POST /api/servers`
- `DELETE /api/servers/{name}`

Notes:
- `POST /api/scan` is a one-off trigger; if a scan is already running it returns `409`.
- Periodic scans are disabled by default; enable via `scheduler_enabled` in the config (UI).

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
  PUID: "1026"
  PGID: "100"
```
Use your host user IDs (from `id`).

If you see `permission denied while trying to connect to the Docker daemon socket at unix:///var/run/docker.sock`, ensure the socket is mounted and the container user can access it. Contiwatch tries to detect the socket group ID automatically; if your setup needs it explicitly, set `DOCKER_GID` to the group id of `/var/run/docker.sock` on the host (e.g. from `stat -c '%g' /var/run/docker.sock`).
