# Release checks and image tagging

This is the Contiwatch implementation guide. It replaces the earlier generic examples; the [release workflow](../.github/workflows/release-image.yml), [Dockerfile](../Dockerfile), [backend](../internal/server/release_check.go), and [frontend](../web/static/app.js) define the behavior.

## Build versions and channels

Publishing a GitHub release triggers the image workflow. `target_commitish` must be exactly `main` or `dev`; other targets fail the tag-resolution step. Images are published under `ghcr.io/${github.repository}` for `linux/amd64`, `linux/arm/v7`, and `linux/arm64`.

| Release target | Image tags | Embedded `VERSION` |
| --- | --- | --- |
| `main` | `latest`, the exact release tag | Exact release tag. |
| `dev` | `dev_latest`, `dev_<release-tag>` | Exact tag when it starts with `dev`; otherwise `dev` followed by the tag. |

For example, a `main` release tagged `1.3.4` produces `latest` and `1.3.4`; a `dev` release tagged `12` produces `dev_latest` and `dev_12`, with embedded version `dev12`.

The Docker build argument `VERSION` is passed to Go as `-ldflags="-X main.Version=${VERSION}"`. The workflow does not inject runtime `APP_VERSION`, `APP_REPO`, or `APP_CHANNEL` variables. Use the [configuration reference](configuration.md#environment-variables) for runtime overrides and compatibility aliases.

Keep stable tags suitable for semantic version comparison and development tags consistent. Metadata infers `dev` from versions starting with `dev` or `vdev`; otherwise it infers `main`. `CONTIWATCH_CHANNEL` explicitly overrides this inference. The runtime strips a leading `v` from embedded versions. `release_tag` is derived from the resolved version, not a separately stored original workflow tag.

## Backend metadata and release API

`GET /api/meta` returns `version`, `repo`, `channel`, and optional `release_tag`. `GET /api/release` returns:

```json
{
  "meta": {
    "version": "1.3.4",
    "repo": "pbuzdygan/contiwatch",
    "channel": "main",
    "release_tag": "1.3.4"
  },
  "latest": {
    "version": "1.3.4",
    "tag": "1.3.4",
    "url": "https://github.com/pbuzdygan/contiwatch/releases/tag/1.3.4"
  },
  "update_available": false,
  "checked_at": "2026-10-05T10:00:00Z"
}
```

This is an illustrative response, not a claim about the latest published release. Responses may include `error`. Metadata is public on controllers; agents require their bearer token for these endpoints.

The backend fetches up to 30 releases from `https://api.github.com/repos/<owner>/<name>/releases?per_page=30`, using a 10-second HTTP timeout and `If-None-Match`/ETag caching. The controller monitor checks at startup and every six hours. Agents do not start the background monitor, but their release endpoint can refresh data on demand.

An endpoint request refreshes absent or older-than-six-hours state. A `304` keeps the previous result and advances its check time. A fetch failure preserves the previous release/update state and records an error and check time, so consumers must not interpret stale data as a successful fresh check.

Setting either `CONTIWATCH_RELEASE_CHECK` or the legacy `APP_RELEASE_CHECK` to false disables checks. The endpoint then returns HTTP `200`, `update_available=false`, and `error="release check disabled"`. GitHub credentials stay in the backend, with `CONTIWATCH_GITHUB_TOKEN` taking precedence over `GITHUB_TOKEN`.

## Release selection and comparison

A release is classified as development when its tag or name starts with `dev`, or its `target_commitish` is `dev`. The first filtering pass skips drafts and selects the requested channel; releases are sorted by the backend version comparator. The `prerelease` flag alone does not define the development channel.

If filtering produces no candidates, the current implementation falls back to the entire fetched list and sorts it. This fallback can cross channels and also bypass the first-pass draft exclusion. Do not describe channel isolation as guaranteed or silently change this compatibility behavior during documentation work.

Version comparison normalizes a leading `v`, compares semantic versions numerically, and handles development versions through the existing comparator, including numeric development indices. Stable and development versions have distinct ordering. The [release tests](../internal/server/release_check_test.go) cover stable, development semantic, numeric development, and channel-selection cases.

## Frontend behavior

The browser loads metadata and obtains release status through `/api/release`; it does not implement a second GitHub release-fetching/selection pipeline. The sidebar update indication and Settings About view use the backend result. If modifying the presentation, use text nodes for external version text and validate the protocol/destination before assigning a release link. Do not interpolate external release data into raw `innerHTML`.

## Verification and maintenance

When changing release behavior, compare the workflow tag mapping, embedded version, metadata, selected release, and UI result together. Test no releases, no matching channel, disabled checks, transport failures, version prefixes, and older cached data with isolated fixtures. See [testing](testing.md) for commands and [installation](installation.md#published-images-and-upgrades) for image usage.
