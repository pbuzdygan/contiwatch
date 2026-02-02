# Toast overlay notifications (bottom-right) — implementation plan

This document describes a **universal**, app-agnostic plan for implementing modern, non-blocking notifications as an **overlay stack** in the bottom-right corner of the UI.

Naming in examples (`toast-stack`, `toast`, `notify`) is illustrative — keep whatever naming fits your app conventions.

## Goal
- Notifications do **not** take layout space (no toolbar/topbar displacement).
- Notifications appear **above all UI** (overlay) and work like “real app” notifications.
- Support typed variants: `success`, `info`, `warning`, `error`.
- Allow manual dismiss via **X**.
- Auto-dismiss:
  - `error` auto-dismisses too, but with **longer** timeout.
- Show up to **3** notifications at once; the rest waits in a **queue**.
- No “Undo” actions (by design).
- Optional: show a **progress bar** indicating remaining time until auto-dismiss.

## UX principles
- Short, scannable text (avoid paragraphs).
- Errors stay long enough to be noticed, but don’t block user interaction.
- Consistent placement, spacing, and animation timing.
- Accessible by default (screen readers, reduced motion).

## Data model
Define a normalized payload for notifications:

```ts
type ToastType = "success" | "info" | "warning" | "error";

type ToastPayload = {
  id?: string;             // optional stable id (for dedupe); otherwise auto-generated
  type: ToastType;
  message: string;         // required, trimmed
  timeoutMs?: number;      // optional override; otherwise defaults per type
};
```

## DOM structure
Add a single overlay container close to the end of `<body>`:

```html
<div id="toast-stack" class="toast-stack" aria-live="polite" aria-relevant="additions"></div>
```

Each toast is rendered as a card:

```html
<div class="toast toast--info" role="status">
  <span class="toast__icon" aria-hidden="true"></span>
  <div class="toast__content">
    <div class="toast__message">…</div>
  </div>
  <button class="toast__close" type="button" aria-label="Dismiss notification">×</button>
</div>
```

Optional progress bar (recommended for “auto-dismiss” UX clarity):

```html
<div class="toast__progress" aria-hidden="true">
  <div class="toast__progress-bar"></div>
</div>
```

Accessibility:
- `info/success/warning`: `role="status"` and `aria-live="polite"`.
- `error`: `role="alert"` (announced immediately); keep everything else `polite`.
- Respect `prefers-reduced-motion`.
- Do not steal focus on show; allow keyboard focus to remain where user is working.

## Styling (CSS)
Key requirements:
- `position: fixed; right: 16px; bottom: 16px; z-index: very high;`
- Consider safe areas on mobile (e.g. `env(safe-area-inset-bottom)` / `env(safe-area-inset-right)`).
- Column layout with spacing; max width and responsive behavior:
  - `max-width: min(420px, calc(100vw - 32px));`
- Allow clicks on toasts but avoid blocking the app behind the stack:
  - `pointer-events: none` on the container
  - `pointer-events: auto` on each toast card
- Per-type color accent (left border, top bar, or icon tint):
  - `success` -> green, `warning` -> amber, `error` -> red, `info` -> blue/primary.
- Animation:
  - enter: slide-up + fade-in
  - exit: fade-out + slight translate
- Hover behavior:
  - optional pause of auto-dismiss timer (recommended).

### Recommended visual (Option B + progress)
If you want a more “premium app” look:
- **Glass card**: translucent background + blur (`backdrop-filter`)
- **Icon bubble** on the left (type-colored)
- **Subtle glow** (very light outline) for type separation
- **Progress bar** at the bottom (type-colored) that shrinks over time

Implementation notes:
- Keep glass effects subtle; ensure readability on both light and dark backgrounds.
- Prefer a small border/outline + shadow over heavy glows.
- The progress bar should be purely decorative (`aria-hidden="true"`).

## Runtime behavior (JS)

### Public API
Expose a single function:

```js
notify({ type, message, timeoutMs })
```

Optionally keep backward compatibility:
- If the app previously had `showToast(message, timeoutMs)`:
  - implement `showToast()` as a thin wrapper that calls `notify({ type: "info", ... })`

### Defaults
Define default timeouts per type (example values):
- `success`: 2500–3500ms
- `info`: 3500–5000ms
- `warning`: 5000–6500ms
- `error`: 9000–12000ms (still auto-dismisses)

### Queue + concurrency
- Maintain:
  - `visibleToasts: ToastInstance[]`
  - `queuedToasts: ToastPayload[]`
- Constraint:
  - max visible = **3**
- Algorithm:
  1) `notify(payload)` enqueues.
  2) `drainQueue()` renders until visible=3 or queue empty.
  3) When a toast is dismissed or auto-expires:
     - remove it from DOM/visible
     - call `drainQueue()` to show next
- Ordering recommendation:
  - FIFO for the queue (older notifications first)
  - newest toast can be inserted at the bottom of the stack to preserve chronological reading

### Dismissal
- Manual:
  - clicking X removes immediately.
- Auto:
  - start timer on render using effective timeout (per-type default or override).
  - on hover: pause timer (optional but recommended).
- Ensure no memory leaks:
  - clear timers on removal.

### Progress bar behavior (if enabled)
Goal: progress bar communicates **remaining time** and stays in sync with pause/resume.

Recommended approach (works with pause/resume and “remainingMs” timers):
1) On render, set progress bar to **100%** width (scaleX(1)).
2) In the next animation frame, transition it to **0%** (scaleX(0)) with a duration equal to `remainingMs`.
3) On hover pause:
   - stop the auto-dismiss timer
   - compute `remainingMs` (`remainingMs -= elapsed`)
   - **freeze** the progress bar at its current scale (read computed transform or track % using elapsed/timeoutMs)
   - remove/disable the transition so it doesn’t keep shrinking
4) On hover resume:
   - restart the auto-dismiss timer for the new `remainingMs`
   - re-apply a transition of `remainingMs` and continue shrinking to 0%

Implementation options:
- **CSS transition** (recommended): easiest to “restart” with a new duration when resuming.
- **CSS animation**: possible, but pause/resume with accurate remaining time often becomes trickier across browsers unless you manage animation state carefully.

### Content safety
- Render notification text via `textContent` (or equivalent) and avoid injecting raw HTML in `message`.
- If rich content is required, define a safe, explicit schema and sanitize inputs.

### Deduplication (optional)
If needed in future:
- If `payload.id` is provided:
  - if a toast with same id exists (visible or queued), update it instead of adding a new one.

## Integration checklist
- Replace “system” dialogs (`alert/confirm/prompt`) with `notify(...)`.
- Replace any inline topbar “toast” placeholders that occupy layout space.
- Map error sources to `type="error"`:
  - network failures, validation failures, server errors.
- Map non-critical confirmations to `success/info`.
- Internationalization: keep messages short and translatable; avoid hard-coded punctuation-heavy strings.

## Validation checklist
- Keyboard: close button focusable; doesn’t trap focus.
- Screen readers: message announced (polite vs assertive).
- Reduced motion: animations disabled.
- Responsive: toasts stay within viewport on small screens.
- Flooding: more than 3 toasts results in queue behavior.
