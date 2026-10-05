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

Logout closes WebSockets and SSE streams associated with the revoked session, including WebSockets opened using its single-use tickets. Shell and Logs share a limit of 64 active connections per process. Client messages are limited to 256 KiB and incoming remote-agent messages to 16 MiB; proxy payloads are streamed without allocating an entire message. Standard ping/pong runs every 30 seconds with a 90-second pong deadline, and writes have a 30-second stall limit. Keep reverse proxies compatible with these control frames. This does not shorten container-update or Compose execution budgets.

## Agents and compatibility

Agents expose a restricted API surface and authenticate requests using bearer tokens, with `/api/health` as the unauthenticated exception. Controller-to-agent HTTP and proxied WebSocket requests use bearer headers. The [API reference](api.md#authentication-and-agent-availability) records the discrepancy between the ADR's subprotocol wording and the current header-based implementation; this documentation update does not change either protocol or the accepted decision.

Existing short tokens and HTTP agent URLs remain accepted for upgrades but trigger security warnings. Rotate short tokens and use HTTPS or a trusted private network. These compatibility allowances are not recommended defaults for a new deployment.

Agent HTTP responses have a 32 MiB decoded-body limit. HTTPS redirects cannot downgrade to HTTP; configure the final agent URL when a deployment previously depended on such redirects. Private/loopback HTTP destinations remain supported, TLS certificate checks remain enabled, and Go's standard authorization-header redirect policy is retained.

## Reverse proxies and request controls

Leave `CONTIWATCH_TRUSTED_PROXIES` unset unless directly behind a trusted reverse proxy. Configure only that proxy's IP/CIDR. Forwarded client IPs are honored only when the direct peer is trusted; otherwise client-specific PIN lockouts use the peer address.

PIN verification uses per-client lockout/backoff, bounded attempt tracking, a global limiter, and a minimum response delay. Request bodies, headers, and HTTP connection lifetimes are bounded. The application sets CSP, framing restrictions, MIME sniffing protection, referrer policy, and browser permission restrictions. Keep proxy configuration compatible with same-origin HTTP, SSE, and WebSocket connections.

## Stored credentials and files

Controller read APIs hide webhook URLs and remote tokens, returning configured-state flags instead. Preserve or clear secrets using the [documented update semantics](configuration.md#secret-reads-and-updates).

Config, Compose, and stack environment files are written atomically with `0600` permissions; stack directories use `0700`. These files can contain credentials: protect their volumes and backups, and do not commit live files or put secret values in screenshots, logs, or issue reports.

Stored stack-file reads/writes and deletion use filesystem roots. Compose and `.env` reads require regular files and reject symlinks/devices; each file is limited to 4 MiB. Use ordinary files inside the stack directory. Compose execution retains the stored working directory so relative bind mounts and project files keep their existing meaning. Protect the entire stack volume from untrusted writers: filesystem-root confinement is not a sandbox for administrator-authored Compose definitions or the Docker daemon.

Compose subprocesses do not inherit `APP_PIN`, `CONTIWATCH_APP_PIN`, `CONTIWATCH_AGENT_TOKEN`, `CONTIWATCH_GITHUB_TOKEN`, or `GITHUB_TOKEN` from Contiwatch. Docker host, TLS, credential-helper, proxy, and other custom environment settings remain available. Stack `.env` entries still take precedence and may explicitly define any valid variable, including the excluded names when intentionally deploying another Contiwatch instance. Failed Compose commands retain a diagnostic tail of at most 64 KiB, with a truncation marker; output volume alone does not stop an image pull.

Webhook configuration and test requests accept official Discord HTTPS webhook URLs. A webhook test sends a real notification; use an isolated destination only when intentionally testing delivery.

## Dependency and release maintenance

The [checks workflow](../.github/workflows/checks.yml) runs tests with the race detector, static checks, frontend builds, npm audit, vendored xterm hash/OSV checks, and govulncheck. A separate isolated build scans the amd64 runtime image's OS packages with Trivy and fails on HIGH/CRITICAL findings, including unfixed vulnerabilities. Go libraries use the reachability-aware govulncheck gate; npm and manually vendored assets have their own gates. The [release workflow](../.github/workflows/release-image.yml) requires these checks before publication, keeps main/dev tagging, and emits SBOM/provenance. ARM runtime-package equivalence still needs verification for the actual multi-platform release artifacts.

Base images and workflow actions are pinned to verified digests/commits. [Dependabot](../.github/dependabot.yml) checks Go/npm dependencies, Docker bases, and Actions weekly and proposes reviewable updates; it does not merge or deploy them automatically. Keep the CI Go version aligned with the Docker builder when accepting toolchain updates. The vendored terminal inventory and MIT licenses are in [manifest.json](../web/static/vendor/xterm/manifest.json), separate from the npm lockfile.

The legacy Docker SDK is retained to preserve existing hosts. The maintained `github.com/moby/moby/client@v0.6.1` enforces minimum API 1.40; regression tests retain successful negotiation with API 1.39. A future SDK migration needs an explicit supported-host policy and end-to-end update/self-update/Compose compatibility verification.

Two unreviewed, module-wide Go vulnerability records currently flag ordinary SDK calls for daemon-only issues: [plugin privilege validation](https://github.com/moby/moby/security/advisories/GHSA-pxq6-2prw-chj9) and [AuthZ bypass](https://github.com/moby/moby/security/advisories/GHSA-x744-4wpc-v9h2). The binary imports client/types/stdcopy, not the daemon or plugin implementation. Narrow [review records](../.github/scripts/go-vulnerability-reviews.json) apply only to the exact SDK version and those client packages, and expire on **2026-11-05**. New reachable findings, refined database records, changed versions, daemon packages, and expired reviews fail CI. Module/package-only findings remain visible for review. These exceptions do not assert that the deployed Docker Engine is patched.

Administrators must verify and patch the actual Docker Engine on controllers and remote hosts. The plugin/AuthZ fixes are in Engine 29.3.1; the [archive/copy mount race](https://github.com/moby/moby/security/advisories/GHSA-rg2x-37c3-w2rh), [symlink race](https://github.com/moby/moby/security/advisories/GHSA-vp62-88p7-qqf5), and [archive extraction execution](https://github.com/moby/moby/security/advisories/GHSA-x86f-5xw2-fm2r) fixes are in 29.5.1. Use a maintained release containing the applicable fixes. Updating Contiwatch's SDK or Docker CLI does not update the host daemon.

## Development references

- [Authentication ADR](ADR/ADR-001-controller-authentication-boundaries.md) — accepted protocol and proxy-trust decisions.
- [PIN Guard reference](PIN_GUARD.md) — historical Mopay material and explicitly separate Contiwatch references.
- [Testing](testing.md) — isolated verification resources and commands.
