# ADR-001: Controller Authentication Boundaries

## Status

Accepted

## Context

Contiwatch has two distinct authentication paths: browser users unlock the controller with a PIN session, while controllers call remote agents with bearer tokens. Browser WebSockets previously placed the long-lived PIN session token in the request URL, and PIN lockout identity accepted forwarded client addresses without proving that the direct peer was a trusted proxy. Hardening these paths must not change the remote agent protocol during an in-place upgrade.

## Decision

- Keep the controller-to-agent bearer header and WebSocket subprotocol unchanged.
- Authenticate browser WebSockets with a random ticket that expires after 30 seconds, is bound to the active PIN session, and can be consumed once.
- Ignore forwarded client IP headers by default. Honor them only when the direct peer matches an IP or CIDR in `CONTIWATCH_TRUSTED_PROXIES`, then walk the forwarding chain from the trusted edge toward the client.
- Retain the legacy PIN-session query path temporarily for SSE and compatibility, while new browser WebSockets always use tickets.
- Accept existing short agent tokens and plain HTTP remote URLs during upgrades, but emit security warnings recommending token rotation and HTTPS or a trusted private network.

## Consequences

- Browser session credentials no longer appear in Shell or Logs WebSocket URLs.
- Deployments behind a reverse proxy must explicitly configure its address for client-specific PIN lockouts; misconfigured or spoofed forwarding headers safely fall back to the direct peer.
- Existing agents continue to connect without coordinated rollout or token migration.
- The legacy query authentication path remains a documented compatibility surface that can be removed only in a future breaking release after SSE migration.
