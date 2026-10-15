# System Architecture

Three parts: static site, Go API, Turso database. The site must work without the API.

```
Browser ──► GitHub Pages (site/, branch gh-pages)
   │
   │ fetch (CORS, JSON)            API_BASE in site/js/config.js
   ▼
Render free web service (api/, Docker, distroless)
   │ libSQL over HTTPS
   ▼
Turso (wishes, scores)
```

Asset tool (`tools/fetch-assets/`) runs offline on a dev machine. It writes `site/assets/` and is not in the request path.

## Frontend (`site/`)

### Boot sequence (`site/js/main.js`)
1. `initMute`, `initDock`, `initTilt`, `renderWearing`, `initLetter`.
2. `mountAll()` dynamically imports each game/guestbook module and calls `mount(el)`.
3. `await runLoader({ onPlay: unlockAudio })`. Resolves when Play is clicked.
4. `playBirthdaySong()`, `burst()`, `runLevelUp()`.

`main.js` adds `html.booted` as soon as its static imports load. This cancels the loader failsafe (below).

### Module contracts

| Contract | Owner | Consumer | Notes |
|---|---|---|---|
| `mount(el)` | `games/steal-an-egg.js`, `games/fruit-gacha.js`, `games/cake.js`, `guestbook.js` | `main.js` `mountAll` | Sets `el.innerHTML` to a static `TEMPLATE`, then wires events. If import or mount throws, `main.js` writes the fallback text from `MODULES` into the element. |
| `egg:gameover` (window CustomEvent) | `steal-an-egg.js` (dispatched near end of game) | `leaderboard.js` | `detail: { score, eggs }`. Leaderboard reads `detail.score`, defaults to 0. |
| `playSfx(name)` | `audio-8bit.js` | games, hero, letter | Named sound effects. |
| `unlockAudio()`, `setMuted()`, `isMuted()` | `audio-8bit.js` | `main.js`, loader | Mute state in `localStorage` key `bd-muted`. |
| `badge(title, subtitle, iconUrl)` | `toast.js` | games, hero | Queued toasts. |
| `burst({ x, y, count })` | `confetti.js` | games, main | Canvas confetti. |

Loader (`site/js/loader.js`) takes `{ onPlay }` and returns a promise. `onPlay` runs inside the click so the audio unlock counts as a user gesture.

### Loader failsafe
`site/css/base.css`: `.loader` runs `loader-failsafe` for 10 s, then sets opacity 0 and `visibility: hidden`. `html.booted .loader` cancels it. If JS fails before `main.js` boots, the join screen goes away on its own.

### Offline fallback
- `API_BASE = ''` (value in `site/js/config.js`): `api-client.js` returns `{ ok: false, error: 'offline' }` with no request sent.
- Guestbook renders `assets/data/seed-wishes.json` immediately, then merges live wishes if the API answers.
- Leaderboard shows local best (`bd-best`).
- Games stay playable. Only score submission is skipped.
- Wake-up: guestbook and leaderboard share one `ping()` (`GET /healthz`, 60 s timeout). Status chip shows waking / offline.

### Request policy (`site/js/api-client.js`)
- Never throws. Every call resolves to `{ ok, status, data?, error? }`.
- Timeouts: reads 15 s, writes 8 s, wake 60 s.
- GET retries once on network error (covers cold-start responses without CORS headers).

## HTTP API (`api/`)

Base: `API_BASE` (empty = offline). Errors are JSON `{"error": "..."}`.

| Method | Path | Body / query | Success | Errors |
|---|---|---|---|---|
| GET | `/healthz` | — | 200 `{"ok":true}` | 503 store down |
| HEAD | `/healthz` | — | 200 / 503 | — |
| GET | `/api/wishes` | `?limit=` (default 50, max 100) | 200 `{"wishes":[Wish]}` | 500 |
| POST | `/api/wishes` | `{name, message, website}` | 201 `Wish` | 400 validation, 413 body > 4 KB, 415 not JSON, 429 rate limit, 500 |
| GET | `/api/scores` | `?limit=` (default 10, max 50) | 200 `{"scores":[Score]}` | 500 |
| POST | `/api/scores` | `{name, score}` (score 0..1000, required) | 201 `Score` | 400, 413, 415, 429, 500 |

- `Wish` = `{id, name, message, createdAt}`. `Score` = `{id, name, score, createdAt}`.
- `website` is a honeypot. If non-empty, the API returns 201 and stores nothing.
- Name ≤ 30 runes (wish) / ≤ 20 (score). Message ≤ 280 runes, max 8 newlines.
- Rate limit: 5 requests/min per IP, burst 3, one bucket per POST route (`server.go`).
- 429 and 413 (with CORS headers) are covered by `api/server_test.go`. 415 and the other rows are from code reading only.

Route source: `api/server.go` (`newServer`). Handlers: `wish_handler.go`, `score_handler.go`.

### Middleware order
Recover → RequestLogger → CORS → BodyLimit(4K) → RateLimiter (POST only). CORS runs before BodyLimit so 413 responses still carry `Access-Control-Allow-Origin`.

### Client IP
- Default: `X-Forwarded-For` via Echo extractor.
- `RENDER` env set: reads `True-Client-IP` / `CF-Connecting-IP` first (Cloudflare in front of Render LB). See `api/client_ip.go`.

## Storage (`api/store*.go`)

`Store` interface (`store.go`): `ListWishes`, `AddWish`, `TopScores`, `AddScore`, `Ping`.

| Implementation | When | Notes |
|---|---|---|
| `tursoStore` (`store_turso.go`) | `TURSO_DATABASE_URL` set | Pure-Go libSQL client, no CGO. Creates tables on start. |
| `memoryStore` (`store_memory.go`) | URL empty, not on Render | Data lost on restart. Dev and tests only. |

Startup refuses the memory store when `RENDER` is set (`main.go`).

Schema (`store_turso.go`):
- `wishes(id TEXT PK, name, message, created_at TEXT)`
- `scores(id TEXT PK, name, score INTEGER, created_at TEXT)` + index `idx_scores_rank (score DESC, created_at)`

Timestamps: fixed-width UTC `2006-01-02T15:04:05.000Z`. Every query bounded by 5 s.

## Asset pipeline (`tools/fetch-assets/`)
Downloads avatar, headshot, game icons and thumbs from Roblox public APIs, plus fruit images from the fruityblox CDN. Optimizes PNG/JPEG. Writes `site/assets/data/fruits.json` and `wearing.json`. Skips existing files unless `-force`.

Avatar or headshot failure makes the process exit 1. Wearing items are best-effort.

## Related
- [Deployment guide](./deployment-guide.md)
- [Codebase summary](./codebase-summary.md)
