# Events Debug filter — implementation proposal

Status: **Proposed; configuration decisions remain open.** The current Events menu supports All, Info, Warn, and Error. None of the debug environment variables or config fields suggested below is an implemented setting.

## Goal and current evidence

Add a Debug level filter to Events so users can inspect diagnostic entries without flooding normal logs.

- [app.js](../web/static/app.js): `logsLevelMode`, `getLogsLevelLabel()`, and `updateLogsLevelMenu()` filter client-side by `entry.level`; there is no Debug menu option.
- [server.go](../internal/server/server.go): `/api/logs` reads/appends entries with fields including `level` and `message`.
- [style.css](../web/static/style.css): `.log-level-debug` already exists. Its presence does not mean a debug-mode switch is implemented.

## Proposed behavior

The Debug filter shows normalized `level === "debug"` entries. Debug generation is off by default and requires an explicit debug-mode switch. When no matching entries exist, show a useful hint such as `No debug logs (enable debug mode)` using the configuration mechanism eventually selected.

## Open decisions

1. Configuration location: an environment setting (recommended), such as `CONTIWATCH_LOG_LEVEL=debug|info` or `CONTIWATCH_DEBUG=true`, versus a config field such as `debug_logs` or `log_level`.
2. Unknown log levels: normalize to `info` or reject them. Choose deliberately rather than silently changing the API contract.
3. Whether to persist the selected Events filter in `localStorage`.

These are alternatives, not settings to add to a deployment today.

## Implementation steps

### Normalize levels at the backend boundary

Allow `debug`, `info`, `warn`, and `error`; trim and lowercase levels. Apply the chosen unknown-level policy consistently in `addLog()` and `POST /api/logs`.

### Gate debug generation

Implement the selected switch and helpers such as `s.isDebugEnabled()` and `s.debugf()` or `s.addDebug()`. When debug mode is off, `addLog("debug", ...)` should not store diagnostic entries.

### Add useful diagnostics without secrets

Log remote operation durations/timeouts, server/container/record counts, and error context such as endpoint and scope. Do not log tokens, authorization headers, complete credential-bearing URLs, sensitive payloads, or full configuration dumps.

### Add the UI option

Add `{ value: "debug", label: "Debug" }` in `updateLogsLevelMenu()` and the corresponding label in `getLogsLevelLabel()`. Keep filtering consistent with the other levels. If persistence is chosen, restore only a supported filter value.

### Explain empty results

When Debug is selected and there are no entries, explain how to enable it using the final chosen setting. Do not display instructions for an unimplemented alternative.

### Update documentation and examples

Document the implemented setting in [configuration](configuration.md), link it from the README when useful, and update the relevant safe deployment example. Mark the decisions resolved only after implementation and verification.

## Acceptance criteria

- Events offers Debug alongside the existing filters.
- Explicitly enabled debug mode produces backend diagnostic entries.
- Debug mode off suppresses diagnostic generation.
- Normalization and malformed/unknown levels follow the selected policy consistently.
- Debug messages contain no credentials or sensitive payloads.
- Empty results explain the actual enablement mechanism.
- Tests cover switch behavior, filtering, normalization, and secret-safe diagnostic content with isolated fixtures.
