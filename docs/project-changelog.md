# Project Changelog

Dated entries, newest first. Only significant changes; see git log for the rest.

## 2026-10-15 — Initial release (frontend + API code, frontend live)

### Added
- Static birthday site in `site/`: join screen, 8-bit birthday song, avatar level-up, "Badge Awarded" toasts, confetti, profile card with real avatar and wearing items, letter from kho as Roblox chat.
- Mini-games: Steal An Egg (`egg:gameover` leaderboard hook), Blox Fruit gacha, birthday cake (6 candles, tap or mic).
- Guestbook and egg leaderboard. Seed wishes show instantly; live data merges when the API answers.
- Go API in `api/`: `GET/POST /api/wishes`, `GET/POST /api/scores`, `GET /healthz`. Turso store, in-memory store for dev. Rate limit 5/min/IP, 4 KB body limit, honeypot.
- Asset tool `tools/fetch-assets/`: downloads and optimizes avatar, game art, fruit images, and wearing items into `site/assets/`.
- `render.yaml` Blueprint and `api/Dockerfile` for Render.

### Deployed
- Frontend published to GitHub Pages via `git subtree push --prefix site origin gh-pages`. Repo commit on `main`: `6acaed9` (per cook report).
- API: not deployed. Waiting on Turso and Render setup. `API_BASE` is still `''`.

### Verification (reported at cook time)
- `go vet` clean; `go test -race ./...` in `api/`: 67 PASS, 0 FAIL.
- Asset tool rerun is idempotent (20 skipped). Images reduced from 1.5 MB to 318 KB.
- Headless browser e2e: shot, games, guestbook offline and online against local API, regression. Live URL games e2e green; `og:image` returns 200.
- Code reviews for phases 04 and 02–03 recorded in `plans/reports/`. Findings at or above 80 fixed or ruled out.

### Deviations from plan
- 6 candles instead of 7.
- Leaderboard split out of `guestbook.js` into `leaderboard.js`.
- GitHub Actions Pages workflow dropped. Token lacked `workflow` scope; Pages serves `gh-pages` via subtree push.

### Known gaps
- Not verified: iOS Safari device test, live Turso store, Docker image build, Lighthouse score.
- `api/server_test.go` is 393 LOC (above the ~200 guideline).
