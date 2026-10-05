# Development and verification

Run commands from the repository root. Use the toolchain required by [go.mod](../go.mod) (Go 1.26.0 or newer) and Node.js/npm for the frontend bundle. The [Dockerfile](../Dockerfile) is the reference for the container build; [package.json](../package.json) defines the frontend build script.

## Go checks

Run targeted tests during implementation, for example:

```bash
go test ./internal/server -run 'TestPin|TestClientIP|TestSelectRelease|TestCompareVersions'
```

For Go changes, complete the repository-wide checks:

```bash
go test ./...
go vet ./...
go build -mod=readonly -o /tmp/contiwatch-doc-check ./cmd/contiwatch
```

The existing suites cover config permissions, PIN sessions and lockouts, trusted proxies, WebSocket tickets, release selection/version comparison, update summaries and self-detection, stack environments and background jobs, Docker storage summaries, and resource caching. HTTP integration tests use local `httptest` servers; Compose tests use temporary executable fixtures.

## Frontend and image build

```bash
npm ci --no-fund
npm run build:cm6
```

The generated bundle at `web/static/vendor/cm6/compose.bundle.js` is ignored by Git. Do not commit it or dependency caches. There are no `npm test` or `npm run lint` scripts in the current package.

The stack editor regression tests run with Node.js and mocked DOM/API fixtures, without dependencies or Docker access:

```bash
node --test web/tests/stack-modal.test.cjs
node --check web/static/app.js
```

These checks cover dirty fields, cancellation, saves, loading and save failures, and saving before stack operations. They do not replace browser checks.

For changes to the image build, verify the Dockerfile with:

```bash
docker build -t contiwatch-check .
```

Building an image does not verify runtime Docker management. UI changes also need browser checks for responsive behavior, keyboard focus, accessible controls, and loading/error/empty/success states. Do not claim browser behavior was tested when only a bundle build ran.

## Documentation checks

For documentation-only changes, check Markdown structure/rendering, relative links, target headings, whitespace, and examples against current code and deployment files. Do not rebuild the application solely for prose changes.

```bash
git diff --check
```

Check new untracked files separately because `git diff --check` does not include them. Keep one source of truth per topic and update inbound links when moving headings. New variables belong in [configuration](configuration.md#environment-variables); retain essential startup settings and a reference link in the README.

## Isolation and limitations

Never use production Docker daemons, containers, stack files, `/data/config.json`, credentials, or Discord destinations as test fixtures. Use temporary directories, synthetic data, mock transports, and isolated HTTP servers.

Compose actions, container recreation, prune/remove, self-update, and real notification delivery require explicit authorization for the selected isolated environment. An environment without Go, dependencies, Docker build access, or a browser may limit verification; report the unavailable check and do not substitute a claim of success. The release image workflow is not a dedicated test/lint pipeline.
