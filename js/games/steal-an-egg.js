// Steal An Egg — Birthday Edition. Tap eggs on the conveyor while the base
// owner looks away; get caught 3 times or run out the clock and it's over.
// A rAF loop drives everything, so the game pauses itself in background tabs.
import { playSfx } from '../audio-8bit.js';
import { badge } from '../toast.js';
import { burst } from '../confetti.js';
import { EGG_TIERS, eggWish } from './birthday-wishes.js';

const ROUND_S = 45;
const MAX_SCORE = 1000;
const LANES = 2;
const TIER_RANK = Object.fromEntries(EGG_TIERS.map((t, i) => [t.id, i]));
const rand = (a, b) => a + Math.random() * (b - a);

const TEMPLATE = `
  <div class="egg-hud" aria-live="off">
    <span class="egg-hud__item">⏱ <b data-ref="time">${ROUND_S}</b>s</span>
    <span class="egg-hud__item">⭐ <b data-ref="score">0</b></span>
    <span class="egg-hud__item">🥚 <b data-ref="bag">0</b></span>
    <span class="egg-hud__hearts" data-ref="hearts" role="img" aria-label="3 mạng"></span>
  </div>
  <div class="egg-arena" data-ref="arena">
    <div class="egg-owner is-away" data-ref="owner" aria-live="assertive">
      <div class="egg-owner__head"><span class="egg-owner__eyes"></span></div>
      <span class="egg-owner__bubble" data-ref="bubble"></span>
      <span class="egg-owner__label">Base của Danh</span>
    </div>
    <div class="egg-lane" data-lane="0"></div>
    <div class="egg-lane egg-lane--rev" data-lane="1"></div>
    <div class="egg-overlay" data-ref="overlay">
      <p class="egg-overlay__title">Cướp trứng sinh nhật!</p>
      <p class="egg-overlay__text">Bấm vào trứng để steal. Khi chủ base quay lại (<b>!</b>), đừng bấm, bị bắt là mất mạng ❤️</p>
      <button class="btn btn--egg btn--xl" type="button" data-ref="start">Bắt đầu</button>
    </div>
  </div>
  <div class="egg-result" data-ref="result" hidden></div>`;

export function mount(root) {
  root.innerHTML = TEMPLATE; // static markup only
  const $ = (r) => root.querySelector(`[data-ref="${r}"]`);
  const lanes = [...root.querySelectorAll('.egg-lane')];
  const ui = { time: $('time'), score: $('score'), bag: $('bag'), hearts: $('hearts'), owner: $('owner'), bubble: $('bubble'), arena: $('arena'), overlay: $('overlay'), result: $('result') };
  let s = null;
  let raf = 0;

  function renderHearts() {
    const h = Math.max(0, s.hearts);
    ui.hearts.textContent = '❤️'.repeat(h) + '🖤'.repeat(3 - h);
    ui.hearts.setAttribute('aria-label', `${s.hearts} mạng`);
  }

  function setGuard(mode) {
    s.guard.mode = mode;
    s.caughtThisWatch = false;
    ui.owner.className = `egg-owner is-${mode}`;
    ui.arena.classList.toggle('is-watched', mode === 'watching');
    ui.bubble.textContent = mode === 'warning' ? '!' : mode === 'watching' ? '👀' : '';
    if (mode === 'warning') playSfx('pop');
    s.guard.t = mode === 'away' ? rand(2.2, 4.5) : mode === 'warning' ? 0.6 : 1.2;
  }

  function pickTier() {
    const pool = EGG_TIERS.filter((t) => t.id !== 'secret' || (!s.secretUsed && s.time < ROUND_S - 15));
    let r = Math.random() * pool.reduce((n, t) => n + t.weight, 0);
    return pool.find((t) => (r -= t.weight) < 0) || pool[0];
  }

  function spawn() {
    const laneIdx = (Math.random() * LANES) | 0;
    const lane = lanes[laneIdx];
    const w = lane.clientWidth;
    const tier = pickTier();
    if (tier.id === 'secret') s.secretUsed = true;
    const btn = document.createElement('button');
    btn.type = 'button';
    btn.className = `egg egg--${tier.id}`;
    btn.setAttribute('aria-label', `Trứng ${tier.label}, ${tier.pts} điểm`);
    lane.append(btn);
    const dir = laneIdx === 1 ? -1 : 1;
    const speedUp = 1 + (ROUND_S - s.time) / 90;
    const egg = { btn, tier, x: dir === 1 ? -60 : w + 4, v: (dir * (w + 64) / rand(3.2, 5.2)) * speedUp, w };
    btn.addEventListener('click', () => steal(egg));
    s.eggs.push(egg);
  }

  function steal(egg) {
    if (s.phase !== 'playing' || egg.gone || s.hearts <= 0) return;
    if (s.guard.mode === 'watching') {
      if (s.caughtThisWatch) return; // one penalty per look, so a double-tap can't cost two hearts
      s.caughtThisWatch = true;
      s.hearts--;
      renderHearts();
      playSfx('alarm');
      egg.btn.classList.add('is-caught');
      egg.btn.addEventListener('animationend', () => egg.btn.classList.remove('is-caught'), { once: true });
      ui.arena.classList.remove('flash');
      void ui.arena.offsetWidth; // restart the flash animation
      ui.arena.classList.add('flash');
      if (s.hearts <= 0) { cancelAnimationFrame(raf); end(); }
      return;
    }
    egg.gone = true;
    s.score += egg.tier.pts;
    s.bag.push(egg.tier.id);
    ui.score.textContent = s.score;
    ui.bag.textContent = s.bag.length;
    playSfx(TIER_RANK[egg.tier.id] >= TIER_RANK.legendary ? 'coin' : 'steal');
    egg.btn.classList.add('is-stolen');
    egg.btn.addEventListener('animationend', () => egg.btn.remove(), { once: true });
    if (egg.tier.id === 'secret') {
      const r = egg.btn.getBoundingClientRect();
      burst({ x: r.left + r.width / 2, y: r.top, count: 80 });
      badge('Secret Egg!', 'Danh đã nhận được Birthday Egg siêu hiếm 🎂');
    }
  }

  function tick(ts) {
    const dt = Math.min(0.1, (ts - s.last) / 1000);
    s.last = ts;
    s.time -= dt;
    ui.time.textContent = Math.max(0, Math.ceil(s.time));

    if ((s.spawnT -= dt) <= 0) { spawn(); s.spawnT = rand(0.55, 0.95); }
    if ((s.guard.t -= dt) <= 0) {
      setGuard({ away: 'warning', warning: 'watching', watching: 'away' }[s.guard.mode]);
    }
    s.eggs = s.eggs.filter((e) => {
      if (e.gone) return false;
      e.x += e.v * dt;
      if (e.x < -70 || e.x > e.w + 10) { e.btn.remove(); return false; }
      e.btn.style.setProperty('--x', `${e.x.toFixed(1)}px`); // CSS keyframes reuse --x
      return true;
    });

    if (s.time <= 0 || s.hearts <= 0) return end();
    raf = requestAnimationFrame(tick);
  }

  function start() {
    cancelAnimationFrame(raf);
    lanes.forEach((l) => l.replaceChildren());
    s = { phase: 'playing', time: ROUND_S, score: 0, hearts: 3, bag: [], eggs: [], guard: { mode: 'away', t: 0 }, spawnT: 0, secretUsed: false, last: performance.now() };
    setGuard('away');
    s.guard.t = 3; // grace period before the first look
    ui.score.textContent = '0';
    ui.bag.textContent = '0';
    renderHearts();
    ui.overlay.hidden = true;
    ui.result.hidden = true;
    ui.arena.tabIndex = -1;
    ui.arena.focus({ preventScroll: true }); // the Start button just vanished
    playSfx('pop');
    raf = requestAnimationFrame(tick);
  }

  function end() {
    if (s.phase === 'over') return;
    s.phase = 'over';
    s.eggs.forEach((e) => e.btn.remove());
    s.eggs = [];
    setGuard('away');
    const score = Math.min(MAX_SCORE, s.score);
    playSfx(s.hearts > 0 ? 'levelup' : 'alarm');
    renderResult(score);
    window.dispatchEvent(new CustomEvent('egg:gameover', { detail: { score, eggs: s.bag.length } }));
    if (s.bag.length >= 20) badge('Egg Thief Pro', `Steal ${s.bag.length} trứng trong 1 ván!`);
  }

  function renderResult(score) {
    const top = [...s.bag].sort((a, b) => TIER_RANK[b] - TIER_RANK[a]).slice(0, 3);
    ui.result.replaceChildren();
    const h = document.createElement('p');
    h.className = 'egg-result__title';
    h.textContent = s.hearts > 0 ? `Hết giờ! ${score} điểm · ${s.bag.length} trứng` : `Bị bắt rồi! ${score} điểm · ${s.bag.length} trứng`;
    const hint = document.createElement('p');
    hint.className = 'egg-result__hint';
    hint.textContent = top.length ? 'Bấm vào trứng để ấp nở và nhận lời chúc:' : 'Chưa steal được trứng nào, thử lại nha!';
    const grid = document.createElement('div');
    grid.className = 'hatch-grid';
    const used = new Set(); // no repeated wish within one hatch screen
    top.forEach((tierId) => grid.append(hatchCard(tierId, used)));
    const again = document.createElement('button');
    again.type = 'button';
    again.className = 'btn btn--egg';
    again.textContent = 'Chơi lại';
    again.addEventListener('click', start);
    ui.result.append(h, hint, grid, again);
    ui.result.hidden = false;
    // Only pull the view back if the player is still looking at the arena.
    const r = ui.arena.getBoundingClientRect();
    if (r.bottom > 0 && r.top < innerHeight) ui.result.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
    again.focus({ preventScroll: true });
  }

  function hatchCard(tierId, used) {
    const tier = EGG_TIERS[TIER_RANK[tierId]];
    const card = document.createElement('button');
    card.type = 'button';
    card.className = `hatch hatch--${tierId}`;
    card.setAttribute('aria-label', `Ấp trứng ${tier.label}`);
    const egg = document.createElement('span');
    egg.className = `egg egg--${tierId} hatch__egg`;
    const label = document.createElement('span');
    label.className = 'hatch__tier';
    label.textContent = tier.label;
    card.append(egg, label);
    card.addEventListener('click', () => {
      if (card.classList.contains('is-open')) return;
      const { pet, wish } = eggWish(tierId, used);
      card.classList.add('is-open');
      card.removeAttribute('aria-label'); // let the pet + wish text become the accessible name
      playSfx('reveal');
      const p = document.createElement('span');
      p.className = 'hatch__pet';
      p.textContent = pet;
      const w = document.createElement('span');
      w.className = 'hatch__wish';
      w.textContent = wish;
      card.append(p, w);
    }, { once: true });
    return card;
  }

  $('start').addEventListener('click', start);
}
