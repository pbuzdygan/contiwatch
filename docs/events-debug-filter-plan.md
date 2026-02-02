# Events: filtr logów `Debug` — plan wdrożenia

Cel: dodać w menu **Events** (UI) opcję filtra logów `Debug`, która pozwoli wyświetlić wpisy logów o poziomie debugowania.

## Stan obecny (referencje w kodzie)
- UI w **Events** filtruje logi client-side po `entry.level`, ale menu ma tylko `All/Info/Warn/Error` (zob. `web/static/app.js`, funkcje związane z `logsLevelMode`).
- Backend `/api/logs` przyjmuje i zwraca listę logów z polami m.in. `level` i `message` (zob. `internal/server/server.go`, handler `handleLogs`).
- CSS ma już styl `.log-level-debug` (zob. `web/static/style.css`), ale obecnie w praktyce rzadko generujemy logi o poziomie `debug`.

## Proponowane zachowanie (kontrakt)
- Filtr `Debug` w UI:
  - Pokazuje tylko wpisy z `level === "debug"` (case-insensitive; po normalizacji).
  - Domyślnie aplikacja nie “spamuje” debugami — debug logi pojawiają się dopiero po włączeniu trybu debug (env/config).
- Jeśli user wybierze `Debug`, a wpisów brak:
  - UI pokazuje czytelny komunikat, np. `No debug logs (enable debug mode)`.

## Decyzja do podjęcia (konfiguracja)
Wybierz jeden z wariantów:
1) **ENV (rekomendowane)**: `CONTIWATCH_LOG_LEVEL=debug|info` albo uproszczone `CONTIWATCH_DEBUG=true`.
2) **config.json**: np. pole `debug_logs: true` / `log_level: "debug"`.

## Plan wdrożenia

### 1) Backend: normalizacja poziomów logów
Cel: spójne poziomy i brak “śmieciowych” wartości.
- Dodać whitelistę poziomów: `debug`, `info`, `warn`, `error`.
- Normalizować `level` (trim + `strings.ToLower`).
- Wszystko spoza listy mapować na `info` (lub odrzucać — do decyzji).
- Zastosować to zarówno w:
  - `addLog(level, message)`
  - `POST /api/logs` (w `handleLogs`)

### 2) Backend: tryb debug (gating)
Cel: debug logi pojawiają się tylko, gdy debug jest aktywny.
- Wprowadzić przełącznik debug (zgodnie z decyzją: ENV vs config).
- Dodać pomocnicze API w serwerze, np.:
  - `s.isDebugEnabled()` oraz `s.debugf(...)`/`s.addDebug(...)`
- Reguła: jeśli debug jest OFF, to `addLog("debug", ...)` nie zapisuje wpisu.

### 3) Backend: gdzie i co logować w debug (bez sekretów)
Cel: debug ma pomagać diagnozować, nie wyciekać danych.
- Logować informacje diagnostyczne typu:
  - czasy odpowiedzi / timeouty dla wywołań do agentów (remote),
  - liczby: ile serwerów, ile kontenerów, ile rekordów,
  - kontekst błędów (endpoint, scope), bez tokenów i bez wrażliwych payloadów.
- Unikać:
  - tokenów, nagłówków auth, pełnych URL zawierających sekrety,
  - dumpów całych konfiguracji.

### 4) UI: dodać opcję `Debug` w filtrze Events
Cel: user ma wybór `Debug` obok `Info/Warn/Error`.
- W `updateLogsLevelMenu()` dodać opcję `{ value: "debug", label: "Debug" }`.
- W `getLogsLevelLabel()` obsłużyć `"debug" -> "Debug"`.
- Upewnić się, że filtr działa tak samo jak pozostałe (client-side po `entry.level`).
- Opcjonalnie: zapisać wybór w `localStorage`, żeby filtr nie resetował się po odświeżeniu (jeśli to pożądane).

### 5) UX: komunikat przy braku debug logów
Cel: brak logów debug nie wygląda jak błąd.
- Jeśli filtr `Debug` jest aktywny i wyników 0:
  - wyświetlić hint: jak włączyć debug (konkretna instrukcja zależna od wariantu konfiguracji).

### 6) Dokumentacja
Cel: użytkownik wie jak włączyć i do czego służy debug.
- `README.md`: dodać opis `CONTIWATCH_LOG_LEVEL` / `CONTIWATCH_DEBUG` (albo config.json) i przykłady.
- Jeśli wybierzemy `config.json`: dopisać pole do sekcji “Config file”.

## Kryteria akceptacji
- W Events w filtrze poziomu logów jest opcja `Debug`.
- Przy aktywnym debug mode pojawiają się wpisy `debug` (w tym z backendu, nie tylko z UI).
- Przy braku debug mode `debug` nie zalewa logów (opcjonalnie: w ogóle nie jest zapisywany).
- Debug logi nie zawierają sekretów (tokenów / auth / wrażliwych payloadów).

