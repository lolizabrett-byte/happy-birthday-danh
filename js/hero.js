// Hero (avatar tilt + Lv 19 → 20 level-up) and profile "Currently Wearing".
import { playSfx } from './audio-8bit.js';
import { badge } from './toast.js';
import { burst } from './confetti.js';

const reduceMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

export function initTilt() {
  const wrap = document.getElementById('hero-avatar');
  const hero = document.getElementById('home');
  if (!wrap || !hero || reduceMotion || !window.matchMedia('(pointer: fine)').matches) return;
  const img = wrap.querySelector('img');
  hero.addEventListener('pointermove', (e) => {
    const r = wrap.getBoundingClientRect();
    const dx = (e.clientX - (r.left + r.width / 2)) / r.width;
    const dy = (e.clientY - (r.top + r.height / 2)) / r.height;
    img.style.setProperty('--tilt-y', (dx * 18).toFixed(2) + 'deg');
    img.style.setProperty('--tilt-x', (-dy * 14).toFixed(2) + 'deg');
  });
  hero.addEventListener('pointerleave', () => {
    img.style.setProperty('--tilt-x', '0deg');
    img.style.setProperty('--tilt-y', '0deg');
  });
}

/** XP bar fills, Lv 19 flips to Lv 20, then fanfare. */
export function runLevelUp() {
  const fill = document.getElementById('level-fill');
  const label = document.getElementById('level-label');
  const xp = document.getElementById('level-xp');
  const burstText = document.getElementById('level-burst');
  if (!fill || !label) return;

  const fillMs = reduceMotion ? 0 : 1600;
  requestAnimationFrame(() => { fill.style.width = '100%'; });

  let shown = 0;
  const counter = setInterval(() => {
    shown = Math.min(365, shown + 13);
    xp.textContent = `XP ${shown} / 365 ngày`;
    if (shown === 365) clearInterval(counter);
  }, Math.max(16, fillMs / 30));

  setTimeout(() => {
    label.textContent = 'Lv 20';
    label.classList.add('is-up');
    burstText.textContent = 'LEVEL UP! Danh đã đạt Level 20 🎉';
    playSfx('levelup');
    const r = label.getBoundingClientRect();
    burst({ x: r.left + r.width / 2, y: r.top + r.height / 2, count: 90 });
    badge('LEVEL UP! → Lv 20', 'Chúc mừng sinh nhật tuổi 20, Danh!', 'assets/img/headshot.png');
  }, fillMs + 150);
}

export async function renderWearing() {
  const box = document.getElementById('wearing');
  const grid = document.getElementById('wearing-grid');
  if (!box || !grid) return;
  try {
    const res = await fetch('assets/data/wearing.json');
    if (!res.ok) return;
    const items = await res.json();
    if (!Array.isArray(items) || !items.length) return;
    for (const it of items) {
      const li = document.createElement('li');
      li.className = 'wearing__item';
      const img = document.createElement('img');
      img.src = it.img;
      img.alt = '';
      img.loading = 'lazy';
      img.width = 150;
      img.height = 150;
      const name = document.createElement('span');
      name.textContent = it.name;
      li.append(img, name);
      grid.append(li);
    }
    box.hidden = false;
  } catch {
    /* keep the section hidden */
  }
}
