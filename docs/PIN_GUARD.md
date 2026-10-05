# PIN Guard — historical reference and implementation guidance

Status: **Reference material from Mopay, not the Contiwatch authentication specification.** The original notes below describe another application's design and have not been verified against a Mopay checkout in this repository.

For Contiwatch, use [deployment security](security.md), [PIN/session API contracts](api.md#metadata-and-pin), and the accepted [controller authentication ADR](ADR/ADR-001-controller-authentication-boundaries.md). Its protected APIs require server-side sessions; the historical UI-only design must not replace that boundary.

## Contiwatch implementation references

- [pin_guard.go](../internal/server/pin_guard.go): startup PIN validation, verification, sessions, lockouts, trusted proxies, and single-use WebSocket tickets.
- [server.go](../internal/server/server.go): controller session enforcement and agent API allowlist/bearer authentication.
- [app.js](../web/static/app.js): browser lock screen, session handling, and authenticated requests.
- [PIN tests](../internal/server/pin_guard_test.go): activity refresh, expiration, lockout/proxy behavior, oversized bodies, and ticket reuse.

Contiwatch validates a startup PIN of 4–8 digits from `APP_PIN`, with `CONTIWATCH_APP_PIN` as a legacy fallback. Its current implementation derives an in-memory salted SHA-256 value and compares candidates in constant time; it does not use Mopay's database, encryption key, or scrypt storage scheme. This describes the implementation, not an endorsement of SHA-256 as a password-storage algorithm. Changing credential derivation needs a separate security review and regression tests.

Successful verification creates a cryptographically random server-side session token, rather than just a browser `pin-ok` flag. The [security guide](security.md#controller-sessions) owns the session lifecycle and compatibility details; do not copy the Mopay pseudocode into Contiwatch.

## Historical Mopay frontend

The original notes identify a React `PinGuard` overlay rendered in `App.tsx`, Zustand state named `pinSession`, and `sessionStorage["pin-ok"]`. A successful `/api/pin/verify` request unlocks the UI. Refreshing the same tab keeps the browser flag; the intended lifecycle ends when the tab closes or the user locks it.

The lock icon in `MainBar` clears state/storage; the Screen Lock control in `SettingsModal` does the same after a short delay. `visualViewport` adjusts the PIN card for mobile keyboards. Input uses numeric mode, allows 4–8 digits, reports local errors, clears failed input, and delays retry by about 1.8 seconds.

This overlay blocks interaction with the rendered UI. A client-side flag is not proof of server authorization, even when the overlay covers the entire page.

## Historical Mopay backend and storage

The original Node/Express notes describe `POST /api/pin/verify` accepting `{ pin }`, returning `{ ok: true }` or HTTP `401`, and comparing the candidate against a `meta` table record.

On startup, `initializePin` validates `APP_PIN`, creates a salt and `scrypt(pin, salt)` hash, encrypts the record with AES-256-GCM, and stores it as `meta.key = 'pin_hash'`. A changed startup PIN replaces the record. Verification uses `crypto.timingSafeEqual`. `APP_ENC_KEY` supplies the encryption key, also used for financial data in that application.

These database, encryption, and framework details belong to Mopay. There is no `APP_ENC_KEY` setting or equivalent `meta` table contract in Contiwatch.

## Historical limitations

The original Mopay design protects casual access to the browser UI but does not require a PIN session on other backend APIs. A caller with network access can therefore bypass the overlay. The notes also describe no backend rate limit/lockout and no session TTL, despite mandatory startup PIN configuration.

Do not treat local or self-hosted deployment as sufficient reason to rely on a UI lock for sensitive operations. Contiwatch's accepted design protects the underlying controller API as well.

## General design guidance

For a new PIN/session implementation, define the server trust boundary before building the lock screen:

1. Validate PIN format and use an established adaptive password-hashing mechanism with a salt, such as scrypt, Argon2, or bcrypt. Never persist plaintext PINs.
2. Protect stored credentials and any encryption keys separately. Additional record encryption may be appropriate, but does not replace password hashing or key management.
3. Verify candidates on the server and establish an authenticated session with cryptographically random credentials.
4. Enforce that session on every protected operation; a frontend flag alone is insufficient.
5. Apply backend rate limiting and lockouts, then add a short UI retry cooldown for feedback.
6. Define idle expiration, revocation, logout, browser/tab lifecycle, and restart behavior explicitly.
7. Make the overlay accessible and mobile-friendly, clear failed input, and provide Clear, Unlock, and Lock actions.
8. Keep PINs and session credentials out of logs and public URLs; choose storage/transport appropriate to the architecture.

The original alternatives included per-tab `sessionStorage`, longer-lived storage, optional server sessions, and optional TTL. Those were generic design choices, not accepted changes for Contiwatch. Do not add a short active-session timeout or change the agent protocol to match a template.

## Historical pseudocode

The following is a sketch of the original Mopay mechanism, not executable code or a supported Contiwatch API:

```text
initializePin(appPin, encryptionKey):
  require appPin to contain 4–8 digits
  if stored record is missing or no longer matches:
    salt = randomBytes(16)
    hash = scrypt(appPin, salt)
    save encrypted {salt, hash} using AES-256-GCM

verifyPin(candidatePin, encryptionKey):
  reject invalid input or a missing record
  decrypt the stored record
  computed = scrypt(candidatePin, record.salt)
  return timingSafeEqual(computed, record.hash)

POST /pin/verify:
  if verifyPin(body.pin): return {ok: true}
  otherwise: return HTTP 401 {ok: false}

onBrowserStartup:
  restore the local pin-ok flag

onSuccessfulVerification:
  save the local pin-ok flag and show the UI

onLock:
  remove the local flag and display the overlay
```

A secure API design additionally issues/revokes a server session and requires it in middleware. The historical generic examples suggested session creation/deletion endpoints and a `requirePinSession` middleware; these are not Contiwatch route names. Its actual routes are listed in the [API reference](api.md).

## Original source paths outside this repository

The Mopay notes refer to `frontend/src/components/PinGuard.tsx`, `frontend/src/App.tsx`, `frontend/src/store.ts`, `frontend/src/components/MainBar.tsx`, `frontend/src/components/modals/SettingsModal.tsx`, `backend/pin.js`, `backend/server.js`, and `backend/encryption.js`.

They also mention a Mopay `docs/ARCHITECTURE.md` covering encryption and PinGuard. These are external historical references, not links to files in this repository.
