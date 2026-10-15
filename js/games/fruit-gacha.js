// Blox Fruit Dealer gacha: a reel decelerates onto a weighted-random fruit,
// each fruit carries a birthday wish. Collected fruits persist per browser.
import { playSfx } from '../audio-8bit.js';
import { badge } from '../toast.js';
import { burst } from '../confetti.js';
import { FRUIT_WISHES, FRUIT_DEFAULT_WISH, FRUIT_WEIGHTS } from './birthday-wishes.js';

const STORE_KEY = 'bd-fruits';
const REEL_LEN = 40;
const LAND_AT = 34;
const SPIN_MS = window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 300 : 3800;

const TEMPLATE = `
  <div class="gacha">
    <div class="gacha__window">
      <div class="gacha__marker" aria-hidden="true"></div>
      <div class="gacha__strip" data-ref="strip"></div>
    </div>
    <button class="btn btn--blue btn--xl gacha__spin" type="button" data-ref="spin">Roll trái ác quỷ</button>
    <div class="gacha__reveal" data-ref="reveal" aria-live="polite" hidden></div>
    <div class="gacha__inv">
      <p class="gacha__inv-title">Túi trái của Danh <span data-ref="count"></span></p>
      <ul class="gacha__inv-grid" data-ref="inv"></ul>
    </div>
  </div>`;

function loadCollected() {
  try { return new Set(JSON.parse(localStorage.getItem(STORE_KEY) || '[]')); } catch { return new Set(); }
}
function saveCollected(set) {
  try { localStorage.setItem(STORE_KEY, JSON.stringify([...set])); } catch { /* storage blocked */ }
}

export function pickFruit(fruits) {
  const weights = fruits.map((f) => FRUIT_WEIGHTS[f.rarity] ?? 10);
  let r = Math.random() * weights.reduce((a, b) => a + b, 0);
  return fruits[weights.findIndex((w) => (r -= w) < 0)] ?? fruits[0];
}

function fruitTile(f, cls) {
  const el = document.createElement('div');
  el.className = `${cls} rarity--${f.rarity.toLowerCase()}`;
  const img = document.createElement('img');
  img.src = f.img;
  img.alt = '';
  img.width = 96;
  img.height = 96;
  img.onerror = () => img.replaceWith(Object.assign(document.createElement('span'), { className: 'fruit-fallback', textContent: f.name[0] }));
  el.append(img);
  return el;
}

export async function mount(root) {
  root.innerHTML = TEMPLATE; // static markup only
  const $ = (r) => root.querySelector(`[data-ref="${r}"]`);
  const strip = $('strip');
  const spinBtn = $('spin');
  const reveal = $('reveal');

  const res = await fetch('assets/data/fruits.json');
  if (!res.ok) throw new Error(`fruits.json ${res.status}`);
  const fruits = await res.json();
  if (!fruits.length) throw new Error('no fruits');

  const collected = loadCollected();
  let spinning = false;

  function renderInventory() {
    const inv = $('inv');
    inv.replaceChildren();
    for (const f of fruits) {
      const li = document.createElement('li');
      const has = collected.has(f.id);
      li.className = `gacha__slot${has ? ' is-owned' : ''}`;
      li.title = has ? f.name : '???';
      li.append(fruitTile(f, 'gacha__slot-tile'));
      const sr = document.createElement('span');
      sr.className = 'sr-only';
      sr.textContent = has ? f.name : 'Chưa có';
      li.append(sr);
      inv.append(li);
    }
    $('count').textContent = `(${collected.size}/${fruits.length})`;
  }

  function buildStrip(result) {
    strip.style.transition = 'none';
    strip.style.transform = 'translateX(0)';
    strip.replaceChildren();
    for (let i = 0; i < REEL_LEN; i++) {
      const f = i === LAND_AT ? result : fruits[(Math.random() * fruits.length) | 0];
      strip.append(fruitTile(f, 'gacha__item'));
    }
    void strip.offsetWidth; // commit the reset before animating
  }

  function showReveal(f) {
    reveal.replaceChildren();
    const card = fruitTile(f, 'gacha__card');
    const info = document.createElement('div');
    const rarity = document.createElement('p');
    rarity.className = `gacha__rarity rarity--${f.rarity.toLowerCase()}`;
    rarity.textContent = f.rarity;
    const name = document.createElement('p');
    name.className = 'gacha__name';
    name.textContent = `${f.name} Fruit`;
    const wish = document.createElement('p');
    wish.className = 'gacha__wish';
    wish.textContent = FRUIT_WISHES[f.id] || FRUIT_DEFAULT_WISH;
    info.append(rarity, name, wish);
    card.append(info);
    reveal.append(card);
    reveal.hidden = false;
  }

  function spin() {
    if (spinning) return;
    spinning = true;
    spinBtn.disabled = true;
    const result = pickFruit(fruits);
    buildStrip(result);

    const item = strip.children[LAND_AT];
    const itemW = item.offsetWidth + parseFloat(getComputedStyle(strip).columnGap || 0);
    const target = item.offsetLeft + item.offsetWidth / 2 - strip.parentElement.clientWidth / 2;
    const jitter = (Math.random() - 0.5) * item.offsetWidth * 0.5;
    strip.style.transition = `transform ${SPIN_MS}ms cubic-bezier(0.12, 0.8, 0.2, 1)`;
    strip.style.transform = `translateX(${-(target + jitter)}px)`;

    // Tick each time an item passes the marker, like a real reel.
    let lastIdx = 0;
    const t0 = performance.now();
    const watch = () => {
      const x = -new DOMMatrixReadOnly(getComputedStyle(strip).transform).m41;
      const idx = Math.floor(x / itemW);
      if (idx !== lastIdx) { lastIdx = idx; playSfx('spin'); }
      if (performance.now() - t0 < SPIN_MS) requestAnimationFrame(watch);
    };
    requestAnimationFrame(watch);

    setTimeout(() => {
      spinning = false;
      spinBtn.disabled = false;
      spinBtn.textContent = 'Roll tiếp';
      playSfx('reveal');
      showReveal(result);
      const isNew = !collected.has(result.id);
      collected.add(result.id);
      saveCollected(collected);
      renderInventory();
      if (result.rarity === 'Mythical') {
        const r = item.getBoundingClientRect();
        burst({ x: r.left + r.width / 2, y: r.top + r.height / 2, count: 70 });
      }
      if (isNew && collected.size === fruits.length) {
        badge('Fruit Collector', 'Danh đã sưu tầm đủ bộ trái ác quỷ!', result.img);
      }
    }, SPIN_MS + 80);
  }

  buildStrip(fruits[0]);
  renderInventory();
  spinBtn.addEventListener('click', spin);
}
