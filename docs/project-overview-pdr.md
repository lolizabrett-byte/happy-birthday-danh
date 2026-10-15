# Project Overview / PDR

Roblox-themed birthday site for **Lưu Thành Danh** (Roblox: SaveAsDanh), turning 20 on 2026-10-15. Signed "from kho".
Live: https://lolizabrett-byte.github.io/happy-birthday-danh/

## Goal
One link, phone or PC → Roblox-style join screen → Danh's real avatar levels up to Lv 20 → play mini-games → read the letter from kho → leave a wish in the guestbook.

## Features

| Feature | Where | Needs API |
|---|---|---|
| Join screen ("Joining server…"), Play unlocks audio | `site/js/loader.js` | no |
| 8-bit birthday song (WebAudio, no audio files) | `site/js/audio-8bit.js` | no |
| Avatar level-up Lv 19 → 20, "Badge Awarded" toast, confetti | `hero.js`, `toast.js`, `confetti.js` | no |
| Profile card: real bio and "Currently Wearing" items | `hero.js`, `assets/data/wearing.json` | no |
| Steal An Egg: steal eggs from the conveyor, 45 s, 3 lives | `site/js/games/steal-an-egg.js` | scores only |
| Blox Fruit gacha: weighted roll, one wish per fruit | `site/js/games/fruit-gacha.js` | no |
| Birthday cake: blow out 6 candles (tap or mic) | `site/js/games/cake.js` | no |
| Letter from kho, typed as Roblox chat | `site/js/letter.js` | no |
| Guestbook (wishes) | `site/js/guestbook.js` | yes, seed wishes otherwise |
| Egg leaderboard (top 10, local best when offline) | `site/js/leaderboard.js` | yes, local otherwise |

Deviation from plan: 6 candles, not 7 (cook report).

## Non-goals
- 3D avatar (needs Roblox auth)
- Accounts or auth
- Admin or moderation UI
- i18n: Vietnamese only, game terms in English
- Light theme (site is dark only)
- Analytics

## Constraints
- Static frontend, no build step. Only external asset: Google Fonts (Bungee, Chakra Petch).
- Site works with API off (`API_BASE = ''`). See [system-architecture](./system-architecture.md#offline-fallback).
- Mobile-first at 375px, touch targets ≥ 44px, `prefers-reduced-motion` honored (`site/css/base.css`).
- Public writes are the only security surface: validation, 4 KB body limit, 5 req/min/IP, honeypot field.

## Success metrics

| Metric | Status |
|---|---|
| Site loads on gh-pages URL (HTTP 200) | Reported at cook time (cook report), not re-checked here |
| Go API tests pass with `-race` | Reported at cook time: 67 PASS, 0 FAIL (cook report) |
| Games and guestbook e2e green on live URL | Reported at cook time (cook report) |
| API live on Render + Turso | Not done (deferred) |
| iOS Safari device test | Unverified |
| Lighthouse score | Unverified, no target recorded |
| Danh opens it on his birthday and reads the letter | Outcome, not measurable in repo |

## Related docs
- [System architecture](./system-architecture.md)
- [Code standards](./code-standards.md)
- [Codebase summary](./codebase-summary.md)
- [Deployment guide](./deployment-guide.md)
- [Roadmap](./development-roadmap.md) · [Changelog](./project-changelog.md)
