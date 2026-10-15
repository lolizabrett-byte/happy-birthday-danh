// Entry point: join screen → audio unlock → level-up, then mount each section.
// Game/guestbook modules load lazily and fail soft, so one broken module never kills the party.
import { runLoader } from './loader.js';
import { unlockAudio, playBirthdaySong, setMuted, isMuted } from './audio-8bit.js';
import { initTilt, runLevelUp, renderWearing } from './hero.js';
import { initLetter } from './letter.js';
import { burst } from './confetti.js';

// Reached only if every static import loaded; cancels the CSS loader failsafe.
document.documentElement.classList.add('booted');

const MODULES = [
  ['egg-game', './games/steal-an-egg.js', 'Mini-game Steal An Egg đang bảo trì 🛠️'],
  ['fruit-gacha', './games/fruit-gacha.js', 'Blox Fruit Dealer đang đi nhập hàng 🍎'],
  ['cake', './games/cake.js', 'Bánh kem đang được nướng… 🎂'],
  ['guestbook', './guestbook.js', 'Sổ lưu bút sắp mở 📖'],
];

function mountAll() {
  for (const [id, path, fallback] of MODULES) {
    const el = document.getElementById(id);
    if (!el) continue;
    import(path)
      .then((m) => m.mount(el))
      .catch((err) => {
        console.warn(`[mount] ${path} unavailable:`, err.message);
        const p = document.createElement('p');
        p.className = 'mount-fallback';
        p.textContent = fallback;
        el.replaceChildren(p);
      });
  }
}

function initMute() {
  const btn = document.getElementById('mute-btn');
  if (!btn) return;
  const sync = () => {
    btn.setAttribute('aria-pressed', String(isMuted()));
    btn.setAttribute('aria-label', isMuted() ? 'Bật âm thanh' : 'Tắt âm thanh');
  };
  sync();
  btn.addEventListener('click', () => {
    unlockAudio();
    setMuted(!isMuted());
    sync();
  });
}

function initDock() {
  const links = [...document.querySelectorAll('.dock__item')];
  const byId = new Map(links.map((a) => [a.dataset.nav, a]));
  const io = new IntersectionObserver((entries) => {
    for (const e of entries) {
      if (!e.isIntersecting) continue;
      links.forEach((a) => a.classList.remove('is-active'));
      byId.get(e.target.id)?.classList.add('is-active');
    }
  }, { rootMargin: '-45% 0px -50% 0px' });
  byId.forEach((_, id) => {
    const sec = document.getElementById(id);
    if (sec) io.observe(sec);
  });
}

async function start() {
  initMute();
  initDock();
  initTilt();
  renderWearing();
  initLetter();
  mountAll();

  await runLoader({ onPlay: unlockAudio });
  playBirthdaySong();
  burst();
  runLevelUp();
}

start();
