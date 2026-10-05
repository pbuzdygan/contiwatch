# Configuration reference

Contiwatch reads startup settings from the environment and stores application settings in `/data/config.json`, unless `CONTIWATCH_CONFIG` overrides the path. Use the UI or documented API to update settings. Keep credentials out of committed configuration and examples.

Sources: [application startup](../cmd/contiwatch/main.go), [config model and defaults](../internal/config/config.go), [config API](../internal/server/server.go), and [image entrypoint](../entrypoint.sh).

## Environment variables

| Variable | Default or requirement | Purpose |
| --- | --- | --- |
| `CONTIWATCH_ADDR` | `:8080` | HTTP listen address. |
| `CONTIWATCH_CONFIG` | `/data/config.json` | Persistent configuration path. |
| `TZ` | Runtime local timezone when unset | Scheduler and log timezone; for example `Europe/Warsaw`. The image includes timezone data. |
| `APP_PIN` | Required for controllers; 4–8 digits | Unlocks a server-validated controller session. Use a unique, non-trivial PIN. |
| `CONTIWATCH_APP_PIN` | Unset | Legacy fallback when `APP_PIN` is empty; prefer `APP_PIN`. |
| `CONTIWATCH_AGENT` | Disabled | Set to `true` to enable remote agent mode. |
| `CONTIWATCH_AGENT_TOKEN` | Required in agent mode | Bearer credential. Use at least 32 random characters; shorter existing tokens remain accepted with a warning. |
| `CONTIWATCH_TRUSTED_PROXIES` | Unset | Comma-separated proxy IPs/CIDRs trusted for `X-Forwarded-For`. Leave unset unless needed. |
| `CONTIWATCH_PUBLIC_URL` | Unset | Public base URL for the Discord avatar at `/icons/contiwatch_logo_small.png`. Does not configure the HTTP listener or HTTPS. |
| `CONTIWATCH_REPO` | `pbuzdygan/contiwatch` | GitHub `owner/name` for release metadata and checks. |
| `CONTIWATCH_CHANNEL` | Inferred from version | Set `main` or `dev` to override the release channel. Versions starting with `dev` or `vdev` imply `dev`; other versions imply `main`. |
| `CONTIWATCH_RELEASE_CHECK` | Enabled | `0` or `false` disables release checks. |
| `CONTIWATCH_GITHUB_TOKEN` | Unset | Optional GitHub API credential for private repositories or higher request limits; used by the backend only. |
| `CONTIWATCH_VERSION` | Unset | Version fallback for builds without an embedded release version. An embedded non-`dev` version takes precedence. |
| `PUID` | `1000` in the image entrypoint | Application user ID and `/data` ownership. |
| `PGID` | `1000` in the image entrypoint | Application primary group ID. |
| `DOCKER_GID` | Detected from mounted socket | Additional Docker socket group ID; optional explicit override. |

Release metadata also supports `APP_REPO` and `APP_CHANNEL` when the respective `CONTIWATCH_*` value is empty, and `GITHUB_TOKEN` when `CONTIWATCH_GITHUB_TOKEN` is empty. `APP_RELEASE_CHECK` is an additional compatibility switch: setting either it or `CONTIWATCH_RELEASE_CHECK` to false disables checks. `APP_VERSION` is not read by Contiwatch.

Version resolution is: embedded non-`dev` `main.Version`, `CONTIWATCH_VERSION`, Go build metadata, then `dev`. The release workflow embeds the version using the Docker build argument `VERSION`; see [release checks](release-check.md).

Use [docker-compose.yml](../docker-compose.yml) and [compose_agent.yml](../compose_agent.yml) as safe deployment examples. Replace credential placeholders before starting; do not add real credentials to the repository.

## Config file

Defaults below are from `DefaultConfig()`, before startup adds any host-specific server entries.

| Field | Default | Behavior |
| --- | --- | --- |
| `scan_interval_sec` | `86400` | Interval in seconds for the legacy scheduler mode. |
| `scheduler_enabled` | `false` | Enables automatic scans using `scheduler_plan`. |
| `scheduler_plan` | `{"mode":"interval_legacy"}` | Legacy interval, Basic days/time, or Cron expression; see below. |
| `global_policy` | `notify_only` | Default container policy; labels override it. |
| `discord_webhook_url` | Empty | Secret notification destination. Read APIs hide the value. |
| `discord_notifications_enabled` | `false` | Master notification switch. |
| `discord_notify_on_start` | `false` | Startup notification. |
| `discord_notify_on_update_detected` | `false` | Scan notification when updates are detected. |
| `discord_notify_on_container_updated` | `false` | Scan/update result notifications. |
| `update_stopped_containers` | `false` | Allows updates of stopped containers while keeping them stopped. |
| `prune_dangling_images` | `false` | Prunes dangling images after updates. |
| `time_zone` | Omitted when empty | Config model field; `GET /api/config` reports the current `TZ` environment value. The config update handler does not set it. Configure the scheduler timezone through `TZ`. |
| `experimental_features` | All flags `false` | Menu visibility settings; the historical JSON name is retained. |
| `local_servers` | Empty list | Local Docker daemon definitions. |
| `remote_servers` | Empty list | Remote agent definitions. |

The visibility flags are `containers`, `containers_sidebar`, `stacks`, `images`, `networks`, `volumes`, `container_shell`, `container_logs`, and `container_resources`. Settings → Menu visibility controls them. `containers_sidebar` adds enabled container subfeatures as sidebar shortcuts; the other flags expose their corresponding views and toolbar controls.

### Scheduler

The scheduler uses the process local timezone, configured with `TZ`. Supported plans:

```json
{"scheduler_enabled": true, "scan_interval_sec": 86400, "scheduler_plan": {"mode": "interval_legacy"}}
```

```json
{"scheduler_enabled": true, "scheduler_plan": {"mode": "basic", "basic": {"days": [1, 2, 3, 4, 5], "time": "03:00"}}}
```

```json
{"scheduler_enabled": true, "scheduler_plan": {"mode": "cron", "cron": {"expr": "0 3 * * 1-5"}}}
```

These are field examples, not complete `PUT /api/config` request bodies. Basic days use `1` for Monday through `7` for Sunday, require at least one day, and use `HH:MM`. Cron uses five fields: minute, hour, day of month, month, day of week. `scan_interval_sec` controls only `interval_legacy`, not Basic or Cron. The UI provides schedule previews; see [scheduler API](api.md#scheduler).

### Servers

Local definitions have `name`, `socket`, optional `public_ip`, and `maintenance`. Remote definitions have `name`, `url`, `token`, optional `public_ip`, and `maintenance`.

`public_ip` supplies an IP/host for opening container services from the UI. Maintenance mode pauses scans and updates for that server. Per-container policy labels still override the global policy synchronized from the controller to agents.

Create or update servers through `/api/locals` and `/api/servers`; `PUT /api/config` does not update these lists. Names must not overlap between local and remote server definitions.

## Secret reads and updates

Config, Compose, and stack environment files use atomic writes with `0600` permissions; stack directories use `0700`.

`GET /api/config` and successful `PUT /api/config` responses hide `discord_webhook_url` and report `discord_webhook_configured`. Both responses also omit every nested remote-server token and report `token_configured` for each remote. `GET /api/servers` uses the same remote-server representation. Redaction does not modify saved credentials or controller-to-agent bearer headers.

For `PUT /api/config`, `discord_webhook_url="__keep__"` explicitly preserves the saved webhook, and an empty/omitted value also preserves an existing one. `"__clear__"` removes it; a new valid URL replaces it. Sending an empty `token` while updating an existing named remote server preserves its token.

Config updates have field-specific semantics rather than generic JSON merge behavior. Send the complete settings intended for the update: omitted non-pointer boolean settings can reset to false, and omitted policy/interval values normalize to defaults. Preserve hidden secrets with the documented values rather than reconstructing them from read responses.

See [Discord scenarios](discord-webhook.md), [security](security.md), and [stack environment handling](user-guide.md#stack-editor-and-env).
