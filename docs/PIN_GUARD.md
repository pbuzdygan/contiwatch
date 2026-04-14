# PIN Guard – analiza i wytyczne implementacyjne

Ten dokument opisuje jak w Mopay zbudowany jest mechanizm PIN Guard, jak dziala i jakie ma zabezpieczenia. Druga czesc to uniwersalne wytyczne wdrozenia podobnego mechanizmu w dowolnej aplikacji.

Powiazane informacje znajduja sie tez w `docs/ARCHITECTURE.md` w sekcjach o szyfrowaniu i w opisie komponentu PinGuard.

## 1. Jak PIN Guard jest zbudowany w Mopay

### Frontend (React)

Pliki kluczowe:
- `frontend/src/components/PinGuard.tsx`
- `frontend/src/App.tsx`
- `frontend/src/store.ts`
- `frontend/src/components/MainBar.tsx`
- `frontend/src/components/modals/SettingsModal.tsx`

Mechanizm UI:
- `PinGuard` to pelnoekranowa nakladka blokujaca UI, renderowana zawsze w `App.tsx`.
- Odblokowanie odbywa sie tylko po poprawnej weryfikacji PIN (API `/api/pin/verify`).
- Stan sesji jest przechowywany w Zustand (`pinSession`) oraz w `sessionStorage` pod kluczem `pin-ok`.
- Po odswiezeniu strony w tej samej sesji przegladarki PinGuard sprawdza `sessionStorage` i odblokowuje UI bez ponownego pytania o PIN.
- Reczne zablokowanie sesji:
  - ikona klodki w `MainBar` usuwa `sessionStorage` i zeruje `pinSession`.
  - przelacznik „Screen Lock” w `SettingsModal` robi to samo po krotkim opoznieniu.
- Komponent reaguje na klawiature mobilna i wylicza przesuniecie karty PIN przy pomocy `visualViewport`, zeby nie byla zaslonieta.

Wejscie PIN:
- `inputMode=numeric`, filtr tylko cyfry, do 8 znakow.
- Minimalna dlugosc 4, maksymalna 8.
- Bledy sa komunikowane lokalnie i blokowany jest szybki re-try przez 1.8s.

### Backend (Node/Express)

Pliki kluczowe:
- `backend/pin.js`
- `backend/server.js`
- `backend/encryption.js`

Weryfikacja PIN:
- Endpoint `POST /api/pin/verify` przyjmuje `{ pin }` i zwraca `{ ok: true }` albo HTTP 401.
- `verifyPinValue` porownuje PIN z rekordem z tabeli `meta`.

Przechowywanie PIN:
- PIN nie jest przechowywany w plain text.
- Podczas startu aplikacji `initializePin`:
  - waliduje PIN z `APP_PIN` (4–8 cyfr),
  - tworzy rekord `{ salt, hash }`, gdzie `hash = scrypt(pin, salt)`,
  - rekord jest szyfrowany (`AES-256-GCM`) i zapisywany w `meta.key = 'pin_hash'`.
- Przy kazdym starcie, gdy `APP_PIN` sie zmieni, rekord jest nadpisywany nowa wartoscia.
- Porownanie hasha odbywa sie w stalej dlugosci z `crypto.timingSafeEqual`.

Szyfrowanie rekordu PIN:
- Do szyfrowania uzywany jest klucz `APP_ENC_KEY`.
- Ten sam mechanizm szyfruje dane finansowe w bazie.

## 2. Jak to dziala z perspektywy uzytkownika

- Po otwarciu aplikacji pojawia sie overlay PIN.
- Po poprawnym PIN, UI odblokowuje sie do konca sesji przegladarki (do zamkniecia karty lub recznego „Lock”).
- PIN nie jest wymagany przy kazdym odswiezeniu, o ile sesja trwa.

## 2.1 Logiczne zachowanie i granice sesji (Mopay)

Opisuje kiedy PIN Guard sie aktywuje i co realnie blokuje:

- Nowa karta / nowe okno przegladarki: zawsze zobaczysz overlay PIN, bo `sessionStorage` jest per-karta i nie dziedziczy sie na nowe okna.
- Odswiezenie strony w tej samej karcie: nie pyta ponownie o PIN, bo `sessionStorage["pin-ok"]` zostaje.
- Reczne zablokowanie (ikona klodki / Screen Lock): natychmiast wymusza ponowne podanie PIN.
- Overlay blokuje caly UI aplikacji: nie da sie kliknac nic pod spodem, dopoki PIN nie zostanie zaakceptowany.
- Nie ma sposobu „obejscia” modalu/overlaya z poziomu UI, bo jest renderowany na wierzchu i sterowany przez stan sesji.

Uwaga: to logika UI. Backendowe API nie jest blokowane przez PIN Guard (patrz sekcja 4).

## 3. Co realnie chroni PIN Guard w Mopay

- Chroni dostep do UI w przegladarce przed przypadkowym lub prostym wejsciem.
- PIN jest bezpiecznie przechowywany po stronie backendu (hash + salt + szyfrowanie), wiec nie wycieka z bazy w postaci jawnej.
- PIN nie jest przechowywany w frontendzie ani w localStorage.

## 4. Ograniczenia i ryzyka w Mopay

- PIN Guard nie blokuje backendu jako takiego. API poza `/api/pin/verify` nie wymaga tokenu ani sesji. Oznacza to, ze atakujacy majacy dostep do hosta lub sieci moze wywolac API bez PIN.
- Brak rate limitu lub lockout po wielu nieudanych probach.
- `APP_PIN` musi byc ustawione, inaczej backend przerywa start (to wymaga sensownego ustawienia w srodowisku).

Wniosek: to jest mechanizm UI-lock, a nie pelna autoryzacja API.

## 5. Uniwersalne wytyczne implementacji PIN Guard (dla innych aplikacji)

Ponizej masz schemat, ktory da sie przeniesc do praktycznie kazdej aplikacji web, desktop lub mobile.

### 5.1 Backend: model bezpiecznego PIN

Rekomendowane zasady:
- Nigdy nie przechowuj PIN w postaci jawnej.
- Uzywaj silnego hashu z sola (scrypt, Argon2 lub bcrypt). Scrypt w Mopay jest wystarczajacy i prosty do wdrozenia.
- Rekord przechowuj jako `{ salt, hash }`.
- Jesli mozesz, dodatkowo zaszyfruj rekord (np. AES-256-GCM) kluczem aplikacyjnym lub kluczem z HSM.

Minimalna implementacja serwerowa:
- Endpoint `POST /pin/verify`:
  - waliduje format PIN,
  - liczy hash i porownuje w czasie stalym,
  - zwraca `ok: true/false`.

Dodatkowe zabezpieczenia, ktore warto dodac:
- rate limiting (np. 5-10 prob / minute),
- czasowa blokada po X bledach,
- logowanie i alarmowanie nieudanych prob,
- opcjonalny TTL sesji po udanym PIN.

### 5.2 Frontend: blokada UI

Wzorzec UI:
- Pelnoekranowy overlay blokujacy interfejs.
- Stan odblokowania przechowywany w pamieci aplikacji oraz w storage per-sesja.
- Po odblokowaniu sesji, mozna utrzymac stan w `sessionStorage` lub analogicznym mechanizmie (np. secure storage w mobile).

Wymagane elementy UX:
- Czyszczenie pola PIN po bledzie.
- Maly cooldown (np. 1-2 sekundy) po bledzie.
- Przyciski „Clear” i „Unlock”.
- Akcja „Lock” dostepna zawsze (ikona klodki).

### 5.3 Sesja i TTL

Decyzje projektowe:
- Czy PIN ma dzialac tylko w danej karcie (sessionStorage), czy pomiedzy sesjami (localStorage lub token)?
- Czy po okresie bezczynnosci automatycznie blokowac UI?

W Mopay:
- stan jest tylko w `sessionStorage`, brak TTL.

### 5.4 Bezpieczenstwo realne vs UX

- Jezeli aplikacja jest lokalna lub self-hosted (jak Mopay), PIN Guard jest glownie bariera UX.
- Jezeli aplikacja jest zdalna i dostepna publicznie, konieczna jest realna warstwa autoryzacji API (token, session, OAuth, cookie + CSRF itp.).
- PIN Guard moze byc dobrym „screen lock”, ale nie powinien byc jedynym mechanizmem ochrony danych.

### 5.5 Wzorzec API + pseudo-kod (do przeniesienia)

Minimalny kontrakt API:
- `POST /pin/verify` → `{ pin: "1234" }` → `{ ok: true }` lub HTTP 401 `{ ok: false }`

Opcjonalnie (jesli chcesz sesje serwerowa):
- `POST /pin/session` → tworzy sesje i zwraca token/cookie
- `DELETE /pin/session` → uniewaznia sesje
- Middleware `requirePinSession` chroni wszystkie API poza `/pin/verify`

#### Backend (pseudo-kod)

```pseudo
// Inicjalizacja PIN (np. przy starcie aplikacji)
function initializePin(appPin, encKey):
  assert appPin matches /^[0-9]{4,8}$/
  if meta.pin_hash missing OR hash not match appPin:
     salt = randomBytes(16)
     hash = scrypt(appPin, salt)
     record = { salt, hash }
     encrypted = aes256gcm_encrypt(JSON.stringify(record), encKey)
     save meta.pin_hash = encrypted

function verifyPin(candidatePin, encKey):
  if candidatePin not string: return false
  encrypted = read meta.pin_hash
  if missing: return false
  record = JSON.parse(aes256gcm_decrypt(encrypted, encKey))
  computed = scrypt(candidatePin, record.salt)
  return timingSafeEqual(computed, record.hash)

POST /pin/verify:
  if verifyPin(body.pin): return { ok: true }
  else return 401 { ok: false }
```

#### Frontend (pseudo-kod)

```pseudo
state.pinOk = false

onAppStart():
  if sessionStorage.get("pin-ok") == "1":
     state.pinOk = true

render():
  if !state.pinOk:
     show PinOverlay()
  else:
     show AppUI()

PinOverlay.submit(pin):
  if pin.length < 4 or pin.length > 8: return
  res = POST /pin/verify { pin }
  if res.ok:
     sessionStorage.set("pin-ok", "1")
     state.pinOk = true
  else:
     showError("Wrong PIN")
     clearInput()
     cooldown(1-2s)

LockButton.onClick():
  sessionStorage.remove("pin-ok")
  state.pinOk = false
```

Uwagi do rozszerzenia:
- Dodaj rate limit w `/pin/verify` (np. 5 prob / minute / IP).
- Dodaj TTL w `sessionStorage` (np. znacznik czasu i auto-lock po X minutach).
- Jesli chcesz twardej ochrony API, dodaj server-side session + middleware.

## 6. Minimalny przepis wdrozenia (checklista)

1. Wygeneruj PIN i przechowuj go tylko w backendzie jako hash + salt.
2. Dodaj endpoint weryfikacji PIN.
3. Dodaj frontendowy overlay blokujacy UI.
4. Po poprawnej weryfikacji zapisz stan sesji.
5. Dodaj przycisk „Lock”, ktory usuwa stan sesji.
6. (Opcjonalnie) Dodaj rate limit i blokady przy bledach.
7. (Opcjonalnie) Dodaj TTL i auto-lock po bezczynnosci.

## 7. Powiazane pliki w Mopay

- `frontend/src/components/PinGuard.tsx`
- `frontend/src/components/MainBar.tsx`
- `frontend/src/components/modals/SettingsModal.tsx`
- `frontend/src/store.ts`
- `backend/pin.js`
- `backend/server.js`
- `backend/encryption.js`
- `docs/ARCHITECTURE.md`
