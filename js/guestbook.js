// Guestbook: seed wishes render instantly, live wishes merge in once the API
// wakes up. Every piece of user text goes through textContent.
import { apiEnabled, ping, listWishes, createWish } from './api-client.js';
import { badge } from './toast.js';
import { playSfx } from './audio-8bit.js';
import { mountLeaderboard } from './leaderboard.js';

const NAME_KEY = 'bd-name';
const NAME_MAX = 30;
const MSG_MAX = 280;
const COLORS = ['#335fff', '#ff5c8a', '#00b06f', '#b07cff', '#ff9a1f', '#1fb6c9'];

const TEMPLATE = `
  <p class="gb-status" data-ref="status" role="status"></p>
  <ol class="gb-list" data-ref="list" aria-label="Lời chúc"></ol>
  <form class="gb-form" data-ref="form" novalidate>
    <h3 class="gb-form__title">Viết lời chúc</h3>
    <div class="field">
      <label for="gb-name">Tên của bạn</label>
      <input id="gb-name" name="name" maxlength="${NAME_MAX}" autocomplete="nickname" required>
      <p class="field__error" data-ref="name-err" aria-live="polite"></p>
    </div>
    <div class="field">
      <label for="gb-msg">Lời chúc cho Danh</label>
      <textarea id="gb-msg" name="message" rows="4" maxlength="${MSG_MAX}" required></textarea>
      <p class="field__hint"><span data-ref="count">0</span>/${MSG_MAX}</p>
      <p class="field__error" data-ref="msg-err" aria-live="polite"></p>
    </div>
    <div class="gb-hp" aria-hidden="true">
      <label for="gb-website">Website</label>
      <input id="gb-website" name="website" tabindex="-1" autocomplete="off">
    </div>
    <button class="btn btn--blue btn--wide" type="submit" data-ref="submit">Gửi lời chúc 🎉</button>
  </form>`;

const runes = (s) => [...s.normalize('NFC')].length;
const store = {
  get: (k) => { try { return localStorage.getItem(k) || ''; } catch { return ''; } },
  set: (k, v) => { try { localStorage.setItem(k, v); } catch { /* storage blocked */ } },
};

function colorFor(name) {
  let h = 0;
  for (const ch of name) h = (h * 31 + ch.codePointAt(0)) >>> 0;
  return COLORS[h % COLORS.length];
}

function formatTime(iso) {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return '';
  const mins = Math.round((Date.now() - d.getTime()) / 60000);
  if (mins < 1) return 'vừa xong';
  if (mins < 60) return `${mins} phút trước`;
  if (mins < 24 * 60) return `${Math.round(mins / 60)} giờ trước`;
  return d.toLocaleDateString('vi-VN');
}

function wishEl(w, isNew = false) {
  const li = document.createElement('li');
  li.className = `gb-wish${isNew ? ' is-new' : ''}`;
  const avatar = document.createElement('span');
  avatar.className = 'gb-wish__avatar';
  avatar.style.background = colorFor(w.name);
  avatar.textContent = [...w.name][0]?.toUpperCase() || '?';
  avatar.setAttribute('aria-hidden', 'true');
  const body = document.createElement('div');
  body.className = 'gb-wish__body';
  const head = document.createElement('p');
  head.className = 'gb-wish__head';
  const name = document.createElement('b');
  name.textContent = w.name;
  const time = document.createElement('span');
  time.textContent = formatTime(w.createdAt);
  head.append(name, time);
  const msg = document.createElement('p');
  msg.className = 'gb-wish__msg';
  msg.textContent = w.message;
  body.append(head, msg);
  li.append(avatar, body);
  return li;
}

export async function mount(root) {
  root.innerHTML = TEMPLATE; // static markup only
  const $ = (r) => root.querySelector(`[data-ref="${r}"]`);
  const form = $('form');
  const list = $('list');
  const seen = new Set();
  const setStatus = (state, text) => { $('status').dataset.state = state; $('status').textContent = text; };

  const render = (wishes, { prepend = false, isNew = false } = {}) => {
    for (const w of wishes) {
      if (!w?.id || seen.has(w.id)) continue;
      seen.add(w.id);
      prepend ? list.prepend(wishEl(w, isNew)) : list.append(wishEl(w, isNew));
    }
  };

  form.name.value = store.get(NAME_KEY);
  form.message.addEventListener('input', () => { $('count').textContent = runes(form.message.value); });

  // One wake-up ping shared by the guestbook and the leaderboard.
  const wake = apiEnabled() ? ping() : Promise.resolve({ ok: false });
  const lb = document.getElementById('leaderboard');
  if (lb) mountLeaderboard(lb, { wake, getName: () => form.name.value.trim() || store.get(NAME_KEY), saveName: (n) => store.set(NAME_KEY, n) });

  try {
    const res = await fetch('assets/data/seed-wishes.json');
    if (res.ok) render(await res.json());
  } catch { /* seeds are a nice-to-have */ }

  if (!apiEnabled()) {
    setStatus('disabled', 'Sổ lưu bút online chưa mở. Lời chúc mẫu ở trên nè!');
    form.hidden = true;
    return;
  }

  setStatus('waking', 'Đang đánh thức server… (lần đầu có thể mất ~1 phút)');
  const alive = await wake;
  if (!alive.ok) {
    setStatus('offline', 'Server đang ngủ say. Bạn vẫn có thể thử gửi lời chúc, hoặc quay lại sau nhé.');
  } else {
    setStatus('online', 'Server online');
    const live = await listWishes();
    if (live.ok) render([...live.data].reverse(), { prepend: true }); // newest ends up on top
  }

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    const name = form.name.value.trim();
    const message = form.message.value.trim();
    $('name-err').textContent = name ? (runes(name) > NAME_MAX ? `Tối đa ${NAME_MAX} ký tự.` : '') : 'Nhập tên của bạn nha.';
    $('msg-err').textContent = message ? (runes(message) > MSG_MAX ? `Tối đa ${MSG_MAX} ký tự.` : '') : 'Viết vài chữ cho Danh nè.';
    if ($('name-err').textContent || $('msg-err').textContent) return;

    const btn = $('submit');
    btn.disabled = true;
    btn.textContent = 'Đang gửi…';
    const r = await createWish({ name, message, website: form.website.value });
    btn.disabled = false;
    btn.textContent = 'Gửi lời chúc 🎉';

    if (r.ok) {
      store.set(NAME_KEY, name);
      render([r.data], { prepend: true, isNew: true });
      form.message.value = '';
      $('count').textContent = '0';
      setStatus('online', 'Server online');
      playSfx('coin');
      badge('Wish Sent!', 'Lời chúc của bạn đã lên server 🎉');
    } else if (r.status === 429) {
      $('msg-err').textContent = 'Bạn gửi nhanh quá, đợi khoảng 1 phút rồi thử lại nha.';
    } else if (r.status === 400) {
      $('msg-err').textContent = `Lời chúc chưa hợp lệ: ${r.error}`;
    } else {
      $('msg-err').textContent = 'Không gửi được, server có thể đang khởi động. Thử lại sau ít phút nha!';
    }
  });
}
