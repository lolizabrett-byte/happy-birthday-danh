# Deployment Guide

Two hosts: frontend on GitHub Pages, API on Render with Turso as the database. The frontend works without the API, so the API can be deployed last or rolled back alone.

Live site: https://lolizabrett-byte.github.io/happy-birthday-danh/
Repo: `https://github.com/lolizabrett-byte/happy-birthday-danh` (from `.git/config`).

## 1. Frontend (GitHub Pages)

Pages serves the `gh-pages` branch, which holds only the contents of `site/`. No GitHub Actions workflow is used (the token lacked `workflow` scope, see `plans/261015-1941-roblox-birthday-site/phase-06-deploy.md`).

```bash
git push origin main
git subtree push --prefix site origin gh-pages
```

Check: `curl -sI https://lolizabrett-byte.github.io/happy-birthday-danh/` returns 200. Reported at cook time, not re-run for this doc.

Pages source must be set to the `gh-pages` branch in repo Settings → Pages.

## 2. Database (Turso)

Commands from the README. Not re-run for this doc.

```bash
turso db create danh-bd
turso db show danh-bd --url          # → TURSO_DATABASE_URL
turso db tokens create danh-bd       # → TURSO_AUTH_TOKEN
```

Tables `wishes` and `scores` are created by the API on start (`CREATE TABLE IF NOT EXISTS`). No manual migration.

Caveat (from the plan, unverified): Turso may archive idle databases after ~10 days. `/healthz` pings the DB on each visit, which keeps it active while the site is in use.

## 3. API (Render)

`render.yaml` is a Blueprint: one Docker web service `danh-birthday-api`, plan free, region singapore, health check `/healthz`, `autoDeploy: true`, build filter `api/**`.

1. Render dashboard → New → Blueprint → select this repo.
2. Fill the three secret env vars below (`sync: false`, never in git).
3. Wait for the first deploy, then check `https://<service>.onrender.com/healthz` returns `{"ok":true}`.

| Env var | Set where | Value | Required |
|---|---|---|---|
| `TURSO_DATABASE_URL` | Render dashboard | `libsql://danh-bd-<org>.turso.io` | Yes on Render. Startup fails without it (`main.go`) |
| `TURSO_AUTH_TOKEN` | Render dashboard | token from step 2 | Yes. Code sends it only when non-empty (`store_turso.go`); whether Turso rejects an empty token is unverified |
| `ALLOWED_ORIGINS` | Render dashboard | `https://lolizabrett-byte.github.io` | Yes. If unset, code falls back to `http://localhost:8000,http://127.0.0.1:8000` and the Pages origin is blocked by CORS |
| `PORT` | Render (injected) | — | No. Code defaults to 8080 and the Dockerfile sets `PORT=8080` |
| `RENDER` | Render (platform-set) | — | No. Code reads it to enable Render IP handling and to refuse the memory store. Not set in `render.yaml`; Render is expected to set it, unverified here |

`ALLOWED_ORIGINS` is a comma-separated list. Entries must have no path and no trailing `/` (trailing slashes are trimmed by code, but keep them clean).

Free tier sleeps after ~15 min without requests. First request can take ~1 min. The site calls `/healthz` on load to wake it.

## 4. Connect the frontend to the API

Edit `site/js/config.js`:

```js
export const API_BASE = 'https://<service>.onrender.com';   // no trailing slash
```

Then push both branches (step 1 commands). Until this is set, the site runs in offline mode (`API_BASE = ''`).

Verify on the live site: guestbook status chip goes from waking to online, a test wish posts, and the leaderboard submits a score.

## 5. Local run

From the README:

```bash
python -m http.server 8000 -d site        # frontend, http://localhost:8000
cd api && go run .                         # API on :8080, memory store
cd api && go test -race ./...
cd tools/fetch-assets && go run . -out ../../site/assets      # add -force to re-download
```

To point local frontend at local API, set `API_BASE = 'http://localhost:8080'` in `site/js/config.js`. Do not commit that change.

## Rollback

| Layer | Steps | Notes |
|---|---|---|
| Frontend | `git revert <sha>` on `main`, `git push origin main`, `git subtree push --prefix site origin gh-pages` | The plan's "workflow redeploys" wording is stale: no workflow exists |
| API | Render dashboard → service → Deploys → roll back to previous deploy | Dashboard step, not verified from repo |
| Kill switch | Set `API_BASE = ''` in `site/js/config.js`, push both branches | Site runs offline. No data is touched |
| Data | Nothing destructive. Tables are additive. Delete rows in Turso only if needed | Use `turso db shell danh-bd` (not verified here) |

## Checklist after a change
- [ ] `go test -race ./...` passes in `api/`
- [ ] Public URL returns 200
- [ ] `/healthz` returns `{"ok":true}`
- [ ] Wish and score POST work from the Pages origin (CORS)

## Related
- [System architecture](./system-architecture.md)
- [Code standards](./code-standards.md)
