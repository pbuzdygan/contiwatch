# Discord webhook — powiadomienia (treść i scenariusze)

Contiwatch wysyła powiadomienia na Discord jako **embed** (pole `embeds[]`). Stały kolor embeda to `0x3498DB`.

Logo:
- Contiwatch ustawia `username=<tytuł komunikatu>` (np. `Contiwatch started`, `Contiwatch updates`) — to jest “duży nagłówek” po lewej, obok avatara webhooka.
- Jeśli ustawisz `CONTIWATCH_PUBLIC_URL` (np. `https://contiwatch.example.com`), to do payload dodawane jest `avatar_url=<PUBLIC_URL>/icons/contiwatch_logo_small.png`.

## Warunki globalne (kiedy Contiwatch w ogóle wysyła)

Wysyłka powiadomień z serwera (runtime) zachodzi tylko gdy:
- `discord_webhook_url` jest ustawione (niepusty string),
- `discord_notifications_enabled=true`,
- oraz odpowiednia flaga per-scenariusz jest włączona.

Domyślnie (wg `internal/config/config.go`) wszystkie flagi Discord są ustawione na `false`, więc bez konfiguracji **nic nie jest wysyłane**.

## Typy komunikatów webhook (embed)

Poniżej: **co** jest wysyłane i **kiedy** (scenariusze).

### 1) Test webhooka (UI/API)

**Kiedy:** `POST /api/notifications/test` (nie zależy od zapisanego configu — URL webhooka jest przekazywany w body).

**Tytuł:** `Contiwatch test`  
**Opis:** `Webhook verified.`  
**Kolor:** `0x3498DB`

Body żądania:
```json
{ "webhook_url": "https://discord.com/api/webhooks/..." }
```

### 2) Start aplikacji (startup)

**Kiedy:** przy starcie procesu `contiwatch`, jeśli:
- `discord_webhook_url` jest ustawione,
- `discord_notifications_enabled=true`,
- `discord_notify_on_start=true`.

**Tytuł:** `Contiwatch started`  
**Kolor:** `0x3498DB`

**Opis (linie, przykładowy kształt):**
- `Scheduler: disabled` **lub** `Scheduler: enabled` / `Scheduler: enabled (every <duration>)`
- `Global policy: <...>`
- `Update stopped containers: <true|false>`
- `Discord notifications: <true|false>`
- `Remote servers: none configured` **lub** `Remote servers: <N>` + (opcjonalnie lista nazw w formie „- name”)
- `Local servers: none configured` **lub** `Local servers: <N>` + (opcjonalnie lista nazw w formie „- name”)

### 3) Podsumowanie skanu (wykryte aktualizacje i/lub zaktualizowane kontenery)

**Kiedy:** po zakończeniu procesu dla danego serwera (skan + ewentualne auto-update), jeśli:
- `discord_notify_on_update_detected=true` **i** wykryto przynajmniej 1 aktualizację (`updates > 0`), **lub**
- `discord_notify_on_container_updated=true` **i** w wyniku skanu przynajmniej 1 kontener ma `updated > 0`.

Uwaga: w trybie auto-update Contiwatch **nie wysyła** osobnych powiadomień per-kontener (pkt 4). Zamiast tego, wynik aktualizacji agreguje do tego jednego powiadomienia.

**Tytuł:** `Contiwatch updates`  
**Kolor:** `0x3498DB`

**Opis (linie):**
- `Server: <serverLabel> (local|remote)`
- `Scanned images: <total>`
- jeśli `discord_notify_on_update_detected=true`:
  - `Updates available: <updates>`
- jeśli `discord_notify_on_container_updated=true`:
  - `Updated: <updated>`

**Dodatkowe sekcje (doklejane, gdy > 0):**
- `Containers with updates:` + lista `- <containerName>`
- `Containers updated:` + lista `- <containerName>`

### 4) Wynik aktualizacji pojedynczego kontenera (manual lub auto-update)

**Kiedy:** po każdej próbie aktualizacji kontenera (manualnej lub automatycznej), jeśli:
- `discord_notify_on_container_updated=true`.

Dotyczy m.in.:
- manualnego update przez API/UI,
- (opcjonalnie) auto-update lokalnego po skanie,
- (opcjonalnie) auto-update zdalnego po skanie.

Uwaga: w obecnej implementacji Contiwatch **agreguje** wyniki auto-update do powiadomienia skanowego (pkt 3), więc powiadomienia per-kontener pojawiają się głównie przy manualnym update.

**Tytuł:** `Contiwatch updates`  
**Kolor:** `0x3498DB`

**Opis:**
```
Server: <serverName>
Container: <containerName>
Result: <status>
Previous state: <previous>
Current state: <current>
```

Gdzie `Result` przyjmuje:
- `updated` (gdy `Updated=true`),
- w przeciwnym razie `Message` z wyniku update (np. „update triggered; agent restarting”),
- a jeśli brak `Message`: `not updated`.

## Format payload (embed)

Każdy z powyższych komunikatów jest wysyłany jako embed, np.:
```json
{
  "username": "Contiwatch started",
  "avatar_url": "https://contiwatch.example.com/icons/contiwatch_logo_small.png",
  "embeds": [
    {
      "description": "...",
      "color": 3447003
    }
  ]
}
```
