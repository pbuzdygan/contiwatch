# Deployment security

This document describes the supported deployment model and current compatibility boundaries. It is not a claim that the entire application has passed a security audit. The accepted [authentication ADR](ADR/ADR-001-controller-authentication-boundaries.md) governs controller/agent compatibility.

## Required setup

- Set a unique, non-trivial `APP_PIN` containing 4–8 digits on controllers. A controller refuses to start without a valid PIN; avoid predictable values such as `1234`, `1111`, or `0000`.
- Mount persistent storage at `/data`, or configure an equivalent protected location with `CONTIWATCH_CONFIG`.
- Configure a unique `CONTIWATCH_AGENT_TOKEN` of at least 32 random characters on every agent. Do not configure controller PIN access on agents.
- Expose controllers and agents through HTTPS, or restrict plain HTTP to a trusted private network. The application HTTP listener does not terminate TLS itself.
- Restrict access to trusted administrators: Docker socket access permits management of host containers, images, networks, and volumes. Back up configuration and stack files before upgrades or destructive operations.

Deployment examples are in [installation](installation.md); defaults and aliases are in [configuration](configuration.md#environment-variables).

## Controller sessions

PIN verification establishes a server-side session; hiding the UI alone does not grant API access. Protected controller API routes require a valid session. HTTP Basic Auth has been removed from controller mode.

The public controller API exceptions are `/api/health`, `/api/version`, `/api/meta`, `/api/release`, `/api/pin/status`, `/api/pin/verify`, and `/api/pin/logout`. `/api/pin/ws-ticket` is protected. Static UI assets are served so the browser can display the lock screen.

The browser keeps its session per window/tab. Refresh in the same tab retains access; a new window/tab requires unlocking. The Lock action revokes the session. There is no short active-session timeout; the server expires sessions after 24 hours without authorized activity. Sessions are held in memory, so a process restart requires unlocking again.

Browser Shell/Logs WebSockets use session-bound, single-use tickets valid for 30 seconds. The legacy `pin_session` query authentication path remains for SSE and compatibility; it has not been removed by the ticket migration. Keep these URLs and other credentials out of logs. See [API authentication](api.md#authentication-and-agent-availability).

## Agents and compatibility

Agents expose a restricted API surface and authenticate requests using bearer tokens, with `/api/health` as the unauthenticated exception. Controller-to-agent HTTP and proxied WebSocket requests use bearer headers. The [API reference](api.md#authentication-and-agent-availability) records the discrepancy between the ADR's subprotocol wording and the current header-based implementation; this documentation update does not change either protocol or the accepted decision.

Existing short tokens and HTTP agent URLs remain accepted for upgrades but trigger security warnings. Rotate short tokens and use HTTPS or a trusted private network. These compatibility allowances are not recommended defaults for a new deployment.

## Reverse proxies and request controls

Leave `CONTIWATCH_TRUSTED_PROXIES` unset unless directly behind a trusted reverse proxy. Configure only that proxy's IP/CIDR. Forwarded client IPs are honored only when the direct peer is trusted; otherwise client-specific PIN lockouts use the peer address.

PIN verification uses per-client lockout/backoff, bounded attempt tracking, a global limiter, and a minimum response delay. Request bodies, headers, and HTTP connection lifetimes are bounded. The application sets CSP, framing restrictions, MIME sniffing protection, referrer policy, and browser permission restrictions. Keep proxy configuration compatible with same-origin HTTP, SSE, and WebSocket connections.

## Stored credentials and files

Controller read APIs hide webhook URLs and remote tokens, returning configured-state flags instead. Preserve or clear secrets using the [documented update semantics](configuration.md#secret-reads-and-updates).

Config, Compose, and stack environment files are written atomically with `0600` permissions; stack directories use `0700`. These files can contain credentials: protect their volumes and backups, and do not commit live files or put secret values in screenshots, logs, or issue reports.

Webhook configuration and test requests accept official Discord HTTPS webhook URLs. A webhook test sends a real notification; use an isolated destination only when intentionally testing delivery.

## Development references

- [Authentication ADR](ADR/ADR-001-controller-authentication-boundaries.md) — accepted protocol and proxy-trust decisions.
- [PIN Guard reference](PIN_GUARD.md) — historical Mopay material and explicitly separate Contiwatch references.
- [Testing](testing.md) — isolated verification resources and commands.
