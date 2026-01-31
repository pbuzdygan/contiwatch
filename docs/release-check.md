# Release Check - uniwersalny schemat

Ten dokument opisuje uniwersalny mechanizm sprawdzania nowej wersji aplikacji, z rozdzieleniem kanałów wydań (np. `main` / `dev`, `stable` / `beta`, `prod` / `staging`).

Zawiera przykładową implementację dla:
- Backend: Go (HTTP API, scheduler, logi)
- Frontend: statyczne HTML/CSS + vanilla JS

## 1) Konwencje wersji i kanałów

Ustal zasady nazewnictwa wydań tak, aby dało się je rozdzielić między kanały:
- **kanał stabilny**: tagi semver, np. `v1.2.3`
- **kanał niestabilny**: tagi dev/beta/rc, np. `dev12`, `beta-3`, `rc.1`

Wymagania:
- Kanał musi być rozpoznawalny po tagu/nazwie release albo po `target_commitish`.
- Jeśli kanałów jest więcej niż dwa, zdefiniuj reguły priorytetu i fallback.

## 2) Pipeline (build i tagowanie artefaktów)

W pipeline ustawiasz:
- `APP_VERSION` (np. tag release)
- `APP_REPO` (np. `org/app`)
- `APP_CHANNEL` (np. `main`, `dev`, `beta`)
- Tag artefaktu/obrazu: `latest` dla kanału stabilnego, osobny tag dla innych kanałów (np. `dev_latest`)

Minimalny pseudokod (GitHub Actions):

```yaml
on:
  release:
    types: [ published ]

jobs:
  build:
    if: github.event.release.target_commitish == 'main' || github.event.release.target_commitish == 'dev'
    steps:
      - checkout
      - set VERSION = tag
      - set CHANNEL = (target_commitish == dev ? dev : main)
      - set PRIMARY_TAG = (CHANNEL == dev ? dev_latest : latest)
      - build artifact with APP_VERSION/APP_REPO/APP_CHANNEL
      - publish artifact tagged with PRIMARY_TAG
```

## 3) Backend - API meta

Backend jest źródłem prawdy o wersji i kanale. Wystaw endpoint:

`GET /api/meta` -> `{ version, repo, channel }`

Opcjonalnie dodaj:
- `release_tag` (dokładny tag release do linków)
- `release_check_enabled` (flaga do wyłączenia w instalacjach offline)

W praktyce warto **nie odpytywać GitHuba bezpośrednio z przeglądarki** (CORS, rate limit, prywatne repo). Zalecane jest proxy w backendzie z cachem/ETag.

### Minimalny kod (Go)

```go
package main

import (
  "encoding/json"
  "log"
  "net/http"
  "os"
)

type MetaResponse struct {
  Version string `json:"version"`
  Repo    string `json:"repo"`
  Channel string `json:"channel"`
}

func metaHandler(w http.ResponseWriter, r *http.Request) {
  resp := MetaResponse{
    Version: getenv("APP_VERSION", "dev"),
    Repo:    getenv("APP_REPO", "org/app"),
    Channel: getenv("APP_CHANNEL", "main"),
  }
  w.Header().Set("Content-Type", "application/json")
  _ = json.NewEncoder(w).Encode(resp)
}

func getenv(k, fallback string) string {
  v := os.Getenv(k)
  if v == "" {
    return fallback
  }
  return v
}

func main() {
  mux := http.NewServeMux()
  mux.HandleFunc("/api/meta", metaHandler)

  log.Println("HTTP on :8080")
  _ = http.ListenAndServe(":8080", mux)
}
```

## 4) Frontend (vanilla JS) - pobieranie i filtrowanie release

Frontend robi dwa kroki:
1) `GET /api/meta` -> ustawia `appVersion` i `releaseChannel`
2) pobiera release (najlepiej z backendu) i wybiera odpowiedni wg kanału

### Struktura danych release (GitHub API)

`GET https://api.github.com/repos/<repo>/releases?per_page=30`

Ważne pola:
- `tag_name`
- `name`
- `target_commitish`
- `html_url`

### Logika wyboru release (kanały)

```js
function isDevRelease(rel) {
  const tag = (rel.tag_name || "").toLowerCase();
  const name = (rel.name || "").toLowerCase();
  const branch = (rel.target_commitish || "").toLowerCase();
  return tag.startsWith("dev") || name.startsWith("dev") || branch === "dev";
}

function selectReleaseForChannel(releases, channel) {
  const normalized = (channel || "main").toLowerCase();
  const filtered = releases.filter((r) => (normalized === "dev" ? isDevRelease(r) : !isDevRelease(r)));
  if (!filtered.length) return releases[0] || null;
  filtered.sort((a, b) => compareVersions(releaseVersion(b), releaseVersion(a)));
  return filtered[0];
}

function releaseVersion(rel) {
  return rel.tag_name || rel.name || null;
}
```

### Porównywanie wersji (semver + niestabilne tagi)

```js
function normalizeVersion(v) {
  if (!v) return null;
  return v.trim().replace(/^v/i, "");
}

function isDevVersion(v) {
  return !!v && /^dev/i.test(v);
}

function extractDevIndex(v) {
  if (!v) return 0;
  const m = v.match(/^dev[-_]?(\d+)/i);
  return m ? Number(m[1]) || 0 : 0;
}

function compareVersions(a, b) {
  const left = normalizeVersion(a);
  const right = normalizeVersion(b);
  const leftIsDev = isDevVersion(left);
  const rightIsDev = isDevVersion(right);

  if (leftIsDev || rightIsDev) {
    if (leftIsDev && !rightIsDev) return -1;
    if (!leftIsDev && rightIsDev) return 1;
    const diff = extractDevIndex(left) - extractDevIndex(right);
    if (diff > 0) return 1;
    if (diff < 0) return -1;
    return (left || "").localeCompare(right || "");
  }

  if (!left && !right) return 0;
  if (!left) return -1;
  if (!right) return 1;
  const l = left.split(".").map((p) => Number(p) || 0);
  const r = right.split(".").map((p) => Number(p) || 0);
  const len = Math.max(l.length, r.length);
  for (let i = 0; i < len; i++) {
    const la = l[i] || 0;
    const rb = r[i] || 0;
    if (la > rb) return 1;
    if (la < rb) return -1;
  }
  return 0;
}
```

### Przykładowy skrypt (frontend)

```html
<div id="release-status">Loading...</div>
<script>
  const statusEl = document.getElementById("release-status");
  const POLL_MS = 1000 * 60 * 60 * 6;

  async function fetchMeta() {
    const res = await fetch("/api/meta");
    if (!res.ok) throw new Error("meta failed");
    return await res.json();
  }

  async function fetchReleases(repo) {
    // Wersja bezpośrednia (GitHub API).
    // W produkcji preferuj backend-proxy (CORS, rate limit, prywatne repo).
    const res = await fetch(`https://api.github.com/repos/${repo}/releases?per_page=30`, {
      headers: { Accept: "application/vnd.github+json" },
    });
    if (!res.ok) throw new Error("releases failed");
    const data = await res.json();
    return Array.isArray(data) ? data : [];
  }

  function render(appVersion, latestVersion, latestUrl) {
    if (!appVersion) {
      statusEl.textContent = "Dev build";
      return;
    }
    if (latestVersion && compareVersions(appVersion, latestVersion) < 0) {
      statusEl.innerHTML = `Update available: ${latestVersion} <a href="${latestUrl}" target="_blank" rel="noreferrer">release</a>`;
      return;
    }
    statusEl.textContent = `Version: ${appVersion}`;
  }

  async function refresh() {
    try {
      const meta = await fetchMeta();
      const releases = await fetchReleases(meta.repo);
      const release = selectReleaseForChannel(releases, meta.channel);
      const latestVersion = release ? (release.tag_name || release.name) : null;
      const latestUrl = release ? release.html_url : null;
      render(meta.version, latestVersion, latestUrl);
    } catch (e) {
      statusEl.textContent = "Release check unavailable";
    }
  }

  refresh();
  setInterval(refresh, POLL_MS);
</script>
```

## 5) Podsumowanie logiki (flow)

1) Pipeline buduje artefakt z `APP_CHANNEL=<kanał>` i publikuje go z odpowiednim tagiem.
2) Backend `/api/meta` zwraca aktualną wersję, repo i kanał.
3) Backend lub frontend pobiera release, filtruje wg kanału.
4) `compareVersions` ustala, czy jest nowsza wersja.
5) UI pokazuje "Update available" tylko w odpowiednim kanale.

## 6) Wskazówki wdrożeniowe

- Nie mieszaj tagów semver i tagów niestabilnych w tym samym kanale.
- Dla kanału niestabilnego utrzymuj spójną numerację (`dev1`, `dev2`, `rc.1`, itp.).
- Jeśli brak release dla kanału, UI może pokazać brak aktualizacji lub fallback do najnowszego release.
- Przy prywatnym repo użyj tokena w backendzie i proxy zamiast bezpośredniego GitHub API w przeglądarce.
