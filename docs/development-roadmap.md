# Development Roadmap

Source: `plans/261015-1941-roblox-birthday-site/`. Status as of 2026-10-15.

## Phases

| # | Phase | Status | Notes |
|---|---|---|---|
| 01 | Go asset fetcher | Complete | `tools/fetch-assets/`, idempotent rerun checked |
| 02 | Frontend shell | Complete | Loader, hero, profile, letter, audio, toasts |
| 03 | Mini-games | Complete | Steal An Egg, Blox Fruit gacha, cake |
| 04 | Go API | Complete (code) | Echo API, tests pass with `-race`. Docker build unverified |
| 05 | Guestbook UI | Complete | Guestbook and leaderboard, offline and online paths tested at cook time |
| 06 | Deploy | Partial | Static site live on GitHub Pages. API deploy deferred |

## Next

1. **API live deploy.** Create Turso DB and Render service (see [deployment guide](./deployment-guide.md)). Set `API_BASE` in `site/js/config.js`, push both branches. Verify `remote_ip` in Render logs (open item C1 from the cook report) and CORS from the Pages origin.
2. **iOS device test.** Safari on a real iPhone: loader, audio unlock, mic in cake game, 375px layout, 44px targets. Not done.
3. **Lighthouse.** Run on the live URL, record scores here. Not done. No target set.

## Open items
- Docker image build not verified in this repo state.
- Turso store not verified live.
- Render cold start (~1 min) is accepted. Wake ping on load mitigates it.
- `api/server_test.go` is 393 LOC. Split if it grows further.
- Plan risk "Render signup may need card" is unverified. Fallback per plan: Cloud Run, or ship static-only.

## Out of scope (from plan)
3D avatar, accounts or auth, admin UI, i18n beyond Vietnamese, light theme, analytics.

## Related
- [Project overview / PDR](./project-overview-pdr.md)
- [Changelog](./project-changelog.md)
