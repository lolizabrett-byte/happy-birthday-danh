// Roblox-style join screen. Resolves when the player presses Play.
const MIN_MS = 1800;
const PRELOAD_TIMEOUT_MS = 4000; // a stalled image must never block the party
const PRELOAD = ['assets/img/avatar.png', 'assets/img/headshot.png'];
const STATUS = ['Đang tải avatar…', 'Đang thổi bóng bay…', 'Đang giấu trứng…', 'Đang kết nối tới server…'];
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

function preload(src) {
  return new Promise((resolve) => {
    const img = new Image();
    img.onload = img.onerror = resolve;
    img.src = src;
  });
}

/** onPlay runs synchronously inside the click, so audio unlock counts as a user gesture. */
export function runLoader({ onPlay } = {}) {
  const loader = document.getElementById('loader');
  const fill = document.getElementById('loader-fill');
  const bar = fill?.parentElement;
  const status = document.getElementById('loader-status');
  const play = document.getElementById('play-btn');
  if (!loader || !play) return Promise.resolve();

  // Keep focus and screen readers inside the dialog while it is up.
  const behind = document.querySelectorAll('main, .dock, .mute-btn, footer');
  behind.forEach((el) => { el.inert = true; });
  document.body.classList.add('is-locked');

  const start = performance.now();
  let loaded = 0;
  const jobs = PRELOAD.map((src) => preload(src).then(() => { loaded++; }));
  const assets = Promise.race([Promise.all(jobs), sleep(PRELOAD_TIMEOUT_MS)]);

  let statusIdx = 0;
  const statusTimer = setInterval(() => {
    statusIdx = (statusIdx + 1) % STATUS.length;
    status.textContent = STATUS[statusIdx];
  }, 600);

  // Bar shows the slower of real progress and the minimum time, so it never stalls at 100%.
  let raf = 0;
  const step = () => {
    const timePct = Math.min(1, (performance.now() - start) / MIN_MS);
    const pct = Math.round(Math.min(timePct, (loaded + 1) / (PRELOAD.length + 1)) * 100);
    fill.style.width = pct + '%';
    bar?.setAttribute('aria-valuenow', String(pct));
    raf = requestAnimationFrame(step);
  };
  raf = requestAnimationFrame(step);

  return Promise.all([assets, sleep(MIN_MS)]).then(() => {
    cancelAnimationFrame(raf);
    clearInterval(statusTimer);
    fill.style.width = '100%';
    bar?.setAttribute('aria-valuenow', '100');
    status.textContent = 'Server đã sẵn sàng! 🎉';
    play.hidden = false;
    play.focus();
    return new Promise((resolve) => {
      play.addEventListener('click', () => {
        onPlay?.();
        behind.forEach((el) => { el.inert = false; });
        loader.classList.add('is-hidden');
        document.body.classList.remove('is-locked');
        loader.addEventListener('transitionend', (e) => {
          if (e.target === loader) loader.remove(); // children's transitions bubble here too
        });
        resolve();
      }, { once: true });
    });
  });
}
