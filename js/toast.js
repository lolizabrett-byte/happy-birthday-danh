// Roblox-style "Badge Awarded" toasts, shown one after another.
const SHOW_MS = 3800;
const MAX_VISIBLE = 2;
const queue = [];
let visible = 0;

const MEDAL_SVG =
  '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M8 2h8l-2 6h-4zM12 9a6.5 6.5 0 1 1 0 13 6.5 6.5 0 0 1 0-13z" fill="currentColor"/><path d="M12 12.2l1.2 2.4 2.6.4-1.9 1.8.5 2.6-2.4-1.3-2.4 1.3.5-2.6-1.9-1.8 2.6-.4z" fill="#ffc83d"/></svg>';

/** Queue a badge toast. Text is set via textContent (safe for any input). */
export function badge(title, subtitle = '', iconUrl = '') {
  queue.push({ title, subtitle, iconUrl });
  pump();
}

function pump() {
  const host = document.getElementById('toasts');
  if (!host) return;
  while (visible < MAX_VISIBLE && queue.length) show(host, queue.shift());
}

function show(host, { title, subtitle, iconUrl }) {
  visible++;
  const el = document.createElement('div');
  el.className = 'toast';
  el.setAttribute('role', 'status');

  const icon = document.createElement('div');
  icon.className = 'toast__icon';
  if (iconUrl) {
    const img = document.createElement('img');
    img.src = iconUrl;
    img.alt = '';
    icon.append(img);
  } else {
    icon.innerHTML = MEDAL_SVG; // static markup, no user data
  }

  const body = document.createElement('div');
  const label = document.createElement('p');
  label.className = 'toast__label';
  label.textContent = 'Badge Awarded!';
  const t = document.createElement('p');
  t.className = 'toast__title';
  t.textContent = title;
  body.append(label, t);
  if (subtitle) {
    const s = document.createElement('p');
    s.className = 'toast__sub';
    s.textContent = subtitle;
    body.append(s);
  }

  el.append(icon, body);
  host.append(el);

  setTimeout(() => {
    el.classList.add('is-leaving');
    el.addEventListener('animationend', () => {
      el.remove();
      visible--;
      pump();
    }, { once: true });
  }, SHOW_MS);
}
