# API reference

The API uses `/api/` routes with endpoint-specific JSON responses, SSE streams, and WebSockets. It has no universal `data`/`error` envelope or `/api/v1` prefix. [Route registration and access checks](../internal/server/server.go) are the implementation reference.

## Authentication and agent availability

Controller requests to protected routes require `X-Contiwatch-Pin-Session: <session-token>`. `POST /api/pin/verify` returns the token after successful PIN verification. Unauthorized protected requests return `401` with `error: "pin_required"`.

The public controller exceptions are `/api/health`, `/api/version`, `/api/meta`, `/api/release`, `/api/pin/status`, `/api/pin/verify`, and `/api/pin/logout`. Browser Shell/Logs WebSockets use `ws_ticket` query credentials obtained from `/api/pin/ws-ticket`: tickets expire after 30 seconds, are session-bound, and can be consumed once. The legacy `pin_session` query path remains for SSE and compatibility. See the [security model](security.md#controller-sessions) and [accepted ADR](ADR/ADR-001-controller-authentication-boundaries.md).

Agents require `Authorization: Bearer <agent-token>` for all allowed routes except `/api/health`, including WebSocket upgrade requests from the controller proxy. Agents do not serve the UI and do not expose the controller PIN, server-directory, aggregate, reachability-stream, or `/api/scan/state` routes. They support metadata, config, scan/stop/status, scheduler, events, notification tests, server health, container, image, network, volume, stack, update, and self-update routes.

The accepted ADR mentions preserving a WebSocket subprotocol, but the current `authorize()` check and Shell/Logs proxy dialers use the bearer header and do not implement bearer-subprotocol authentication. That wording discrepancy does not establish an additional supported credential transport; keep the accepted compatibility boundary and review the ADR separately before changing it.

Controller resource requests select a server with `scope=local:{name}` or `scope=remote:{name}`, in the query or JSON body as indicated below. Direct agent resource requests generally use the `server` query parameter for one of the agent's local daemons; several handlers default it when exactly one local server exists. Do not pass a controller's remote-server name as an agent-local server name.

## Metadata and PIN

| Method | Route | Contract |
| --- | --- | --- |
| GET | `/api/health` | Process health: `{"status":"ok"}`; distinct from Docker server health. |
| GET | `/api/version` | Current application version. |
| GET | `/api/meta` | `version`, `channel`, `repo`, and optional `release_tag`. |
| GET | `/api/release` | Cached release metadata, `latest`, `update_available`, `checked_at`, and optional `error`; see [release checks](release-check.md). |
| GET | `/api/pin/status` | `enabled` and `unlocked`; supply a session header to check that session. |
| POST | `/api/pin/verify` | Body `{"pin":"<4–8 digits>"}`; success returns `ok` and `session_token`. Incorrect PIN returns `401`; lockout returns `429`, `Retry-After`, and `retry_after_seconds`. |
| POST | `/api/pin/logout` | Revokes the supplied session and returns `{"ok":true}`. |
| POST | `/api/pin/ws-ticket` | Protected; returns `ticket` and `expires_in` (30 seconds). |

## Settings and servers

| Method | Route | Contract |
| --- | --- | --- |
| GET, PUT | `/api/config` | Read/update settings; hidden webhook and field-specific update semantics are documented in [configuration](configuration.md#secret-reads-and-updates). |
| GET, POST | `/api/servers` | List or create/update a named remote server; tokens are hidden in responses. |
| DELETE | `/api/servers/{name}` | Remove a remote server definition. |
| GET, POST | `/api/locals` | List or create/update a named local Docker server. |
| DELETE | `/api/locals/{name}` | Remove a local server definition. |
| GET | `/api/servers/info` | Server versions and reachability snapshot. |
| GET | `/api/servers/stream` | Live server info and scan updates via SSE. |
| POST | `/api/servers/refresh` | Trigger reachability checks; update the stream and return a snapshot. |
| GET | `/api/servers/health?scope=local:{name}` | On-demand Docker health/storage summary; controller also supports `remote:{name}`. |
| POST | `/api/status/refresh` | Pull last scan snapshots from online agents and update the stream. |

## Scheduler

| Method | Route | Contract |
| --- | --- | --- |
| GET | `/api/scheduler/status` | `enabled`, `mode`, `tz`, and `next_runs`; currently returns at most the stored next run. |
| GET | `/api/scheduler/preview/basic` | Query `days=1,2,3,4,5`, `time=03:00`, optional `count`; preview Basic runs. |
| GET | `/api/scheduler/preview/cron` | Query `expr` (URL-encoded five-field Cron expression), optional `count`; preview Cron runs. |

Previews return `tz`, `now`, `next_runs`, and optional `error`. Invalid preview inputs can return HTTP `200` with `error`; callers must inspect the body. Enabled invalid schedule updates return `400`. See [scheduler configuration](configuration.md#scheduler).

## Scans and updates

| Method | Route | Contract |
| --- | --- | --- |
| POST | `/api/scan` | Trigger a one-off scan; returns `409` if a scan already runs. |
| POST | `/api/scan/stop` | Cancel a scan/batch. |
| GET | `/api/scan/state` | Controller scan running state. |
| GET | `/api/status` | Last local/agent scan snapshot. |
| GET | `/api/aggregate` | Controller local and remote status. |
| POST | `/api/update/{container_id}` | Update a selected container. |
| POST | `/api/self-update?container={container_id}` | Agent-only helper update for the container running that agent. |

### Scan and update contracts

Update results include `old_image_id`, `new_image_id`, and `applied_image_id` to diagnose tag/image mismatches. Scan entries mark the responding Contiwatch container with `self=true`; the controller updates that agent last and verifies its post-restart status. Runtime metadata supplements hostname detection. Legacy agents without the marker are recognized conservatively by their Contiwatch image and agent-style name.

A full manual or scheduled batch has no shared deadline; individual remote operations remain bounded. Periodic scans are disabled by default and use the configured schedule when enabled. See [policies and scan behavior](user-guide.md#container-policies).

## Containers and images

| Method | Route | Contract |
| --- | --- | --- |
| GET | `/api/containers?scope=local:{name}` | List containers; controller also supports remote scope. |
| POST | `/api/containers/action` | Body includes `scope`, `container_id`, `action`: `start`, `stop`, `restart`, `pause`, `unpause`, `kill`, or `rm`. |
| GET / WebSocket | `/api/containers/shell` | Interactive container shell. |
| GET / WebSocket | `/api/containers/logs` | Container log stream. |
| POST | `/api/containers/resources` | Fetch resource metrics for `scope` and `container_ids`. |
| GET | `/api/images?scope=local:{name}` | List images; controller also supports remote scope. |
| POST | `/api/images/pull` | Body includes `scope`, `repository`, optional `tag`. |
| POST | `/api/images/prune` | Body includes `scope`, `mode=unused` or `dangling`. |
| POST | `/api/images/remove` | Body includes `scope`, `image_id`. |

## Networks and volumes

| Method | Route | Contract |
| --- | --- | --- |
| GET | `/api/networks` | List networks for query `scope`. |
| GET | `/api/networks/details` | Query `scope` and network `id`. |
| POST | `/api/networks/remove` | Body `scope`, `network_id`; an in-use network returns `409` with `blocked_by`. |
| POST | `/api/networks/prune` | Body `scope`; prune unused networks. |
| POST | `/api/networks/connect` | Body `scope`, `network_id`, `container_id`. |
| POST | `/api/networks/disconnect` | Body `scope`, `network_id`, `container_id`. |
| GET | `/api/volumes` | List volumes for query `scope`. |
| POST | `/api/volumes/remove` | Body `scope`, volume `name`; an in-use volume returns `409` with `blocked_by`. |
| POST | `/api/volumes/prune` | Body `scope`; prune unused volumes. |

## Stacks

| Method | Route | Contract |
| --- | --- | --- |
| GET | `/api/stacks?scope=local:{name}` | List stored Compose stacks; controller also supports remote scope. |
| GET | `/api/stacks/get?scope=local:{name}&name={stack}` | Fetch Compose and environment content. |
| PUT | `/api/stacks/save` | Save Compose/environment content without deployment. |
| POST | `/api/stacks/validate` | Validate the interpolated Docker Compose model. |
| POST | `/api/stacks/action` | Synchronous `up`, `down`, `pull`, `redeploy`, `start`, `stop`, `restart`, `kill`, or `rm`; compatibility path. |
| POST | `/api/stacks/jobs` | Same request body as stack action; returns `202` with a job or `409` if one already runs for that stack. |
| GET | `/api/stacks/jobs?id={job}` | Job status: `running`, `succeeded`, or `failed`, with error details where applicable. |
| POST | `/api/stacks/jobs/extend` | Body `{"id":"{job}"}`; extend the timeout by 10 minutes, forwarded to remote agents. |

Jobs include `id`, `status`, `remaining_ms`, and `extendable`. Action limits are 10 minutes for `up`/`pull`, 20 minutes for `redeploy`, and 3 minutes for other actions. Remote timeout extension requires agent `1.3.4` or newer; older agents use synchronous fallback without extension. See [stack editor and environment handling](user-guide.md#stack-editor-and-env).

## Events and notifications

| Method | Route | Contract |
| --- | --- | --- |
| GET, POST, DELETE | `/api/logs` | Read, append, or clear application event logs. |
| POST | `/api/notifications/test` | Body `webhook_url`; validates an official Discord HTTPS webhook URL and sends a test independently of saved notification settings. |

Application event logs are separate from container log WebSockets. See [Discord notification scenarios](discord-webhook.md).
