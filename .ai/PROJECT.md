# Project Profile

## Scope and verification

- Goal and scope: maintain Contiwatch's Docker image monitoring, optional container updates, controller/agent integration, and management UI. See [README](../README.md) for supported behavior and contracts.
- Required verification: preserve the previous requirement to check compilation, lint/static analysis, and tests before a commit. For Go changes, run targeted tests during iteration, then `go test ./...` and `go vet ./...` from the repository root. Test business logic, public API behavior, authentication/authorization, and bug regressions.
- Use the [development and verification guide](../docs/testing.md) for build prerequisites, available checks, and isolation requirements. There is no dedicated test/lint CI. Do not claim checks ran when unavailable. For frontend build changes, use the scripts in [package.json](../package.json) and build steps in [Dockerfile](../Dockerfile); inspect commands before executing them. For documentation/instruction-only changes, check Markdown structure/rendering, whitespace, links, scope, and examples against repository evidence instead of rebuilding the application.
- Protected resources: real Docker daemons and their containers, images, networks, volumes, stack files, `/data/config.json`, credentials, and Discord webhook destinations. Tests must use isolated fixtures; do not run Compose actions, prune/remove, container recreation, self-update, or notification tests against live resources without explicit authorization.

## Contracts and source precedence

- Before changing authentication, proxy trust, sessions, or controller/agent compatibility, read the accepted [authentication ADR](../docs/ADR/ADR-001-controller-authentication-boundaries.md) and relevant [deployment security sections](../docs/security.md). Keep their compatibility constraints; do not replace existing protocols with generic template recommendations.
- Before changing API responses, secret updates, configuration, or stack environment handling, read the relevant [API](../docs/api.md), [configuration](../docs/configuration.md), and [stack editor](../docs/user-guide.md#stack-editor-and-env) sections. Preserve established contracts instead of imposing a new response envelope, route prefix, or validation dependency.
- [PIN_GUARD.md](../docs/PIN_GUARD.md) retains historical Mopay and generic implementation guidance, with separate links to the Contiwatch contract. Its historical UI-only protection model must not override the accepted ADR or deployment security guide.
- For release/channel work, inspect the actual [release workflow](../.github/workflows/release-image.yml), which uses `main` and `dev`. The old generic Git Flow requirement for a `develop` integration branch does not describe this pipeline; do not create or rename branches to satisfy it.
- If accepted specifications and implementation conflict, report the discrepancy before changing a security or compatibility boundary. Generic examples and proposals do not supersede accepted decisions.

## Retained development rules

- Use Conventional Commits and atomic changes when commits are authorized. Merge branches through a PR; describe the behavior, verification, documentation impact, and any configuration changes. Do not commit or publish merely because checks pass.
- Keep public functions documented using the language's established documentation convention. Document new environment variables in the [configuration reference](../docs/configuration.md#environment-variables) and existing relevant safe configuration examples. Keep essential startup variables and a link to the full reference in the README; create an example file only when needed, never with real secrets.
- For UI changes, preserve semantic controls, accessible names, visible keyboard focus, and responsive behavior. Cover loading, error, empty, and success states when fetching data.
- Keep durable architecture decisions under existing `docs/ADR/`; do not start a parallel decision archive. Read related ADRs only when the affected design requires them.

## Existing plans and supporting documents

Read these only when the task concerns their topic; preserve their unresolved choices and compare them with current code before treating work as pending or complete:

- [Events debug filter plan](../docs/events-debug-filter-plan.md): proposed filtering and debug configuration; configuration choices remain open in the document.
- [Toast overlay record](../docs/toast-overlay-plan.md): core behavior is implemented; original design guidance and browser validation criteria remain. Verify the current implementation before treating any work as pending.
- [Discord notification scenarios](../docs/discord-webhook.md): consult when changing notification content or triggers.
- [Release check guide](../docs/release-check.md): Contiwatch image tagging, version resolution, backend caching, and channel-selection behavior; verify against the actual workflow and current implementation.

Keep existing plans at their current paths. Save a new plan under `.ai/plans/` only when it needs to survive a session, and link existing evidence rather than duplicating it. No handoff note is needed for completed work.
