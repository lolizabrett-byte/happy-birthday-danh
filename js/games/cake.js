// Birthday cake: blow out every candle (tap each flame, or opt in to the mic).
import { playSfx, playBirthdaySong } from '../audio-8bit.js';
import { badge } from '../toast.js';
import { burst } from '../confetti.js';

// "2" and "0" number candles in the middle, small candles around them.
const CANDLES = ['s', 's', '2', '0', 's', 's'];
const BLOW_RMS = 0.12;      // mic loudness that counts as blowing
const BLOW_HOLD_MS = 250;   // must stay loud this long
const MIC_TIMEOUT_MS = 20000;

const TEMPLATE = `
  <div class="cake-stage">
    <div class="cake" aria-label="Bánh sinh nhật">
      <div class="cake__candles" data-ref="candles"></div>
      <div class="cake__tier cake__tier--top"><span class="cake__drip"></span></div>
      <div class="cake__tier cake__tier--mid"><span class="cake__drip"></span><span class="cake__text">Danh 20</span></div>
      <div class="cake__tier cake__tier--base"><span class="cake__drip"></span></div>
      <div class="cake__plate"></div>
    </div>
    <p class="cake__status" data-ref="status" aria-live="polite"></p>
    <div class="cake__actions">
      <button class="btn btn--pink" type="button" data-ref="mic">🎤 Thổi bằng micro</button>
      <button class="btn" type="button" data-ref="relight" hidden>Thắp lại nến</button>
    </div>
    <p class="cake__wish" data-ref="wish" hidden>🎉 Điều ước đã được gửi lên server! Happy 20th, Danh!</p>
  </div>`;

export function mount(root) {
  root.innerHTML = TEMPLATE; // static markup only
  const $ = (r) => root.querySelector(`[data-ref="${r}"]`);
  const status = $('status');
  const micBtn = $('mic');
  const relight = $('relight');
  const wishEl = $('wish');
  const flames = [];
  let celebrated = false;
  let mic = null;
  let micPending = false; // permission prompt open: ignore extra clicks

  CANDLES.forEach((kind, i) => {
    const c = document.createElement('div');
    c.className = kind === 's' ? 'candle' : 'candle candle--num';
    if (kind !== 's') c.dataset.num = kind;
    const flame = document.createElement('button');
    flame.type = 'button';
    flame.className = 'flame';
    flame.setAttribute('aria-label', `Thổi nến ${i + 1}`);
    flame.addEventListener('click', () => blow(flame));
    c.append(flame);
    $('candles').append(c);
    flames.push(flame);
  });

  status.textContent = `Còn ${flames.length} ngọn nến. Bấm vào ngọn lửa để thổi!`;
  const lit = () => flames.filter((f) => !f.classList.contains('is-out'));

  function blow(flame) {
    if (flame.classList.contains('is-out')) return;
    flame.classList.add('is-out');
    flame.disabled = true;
    playSfx('pop');
    const left = lit().length;
    status.textContent = left ? `Còn ${left} ngọn nến…` : 'Tắt hết rồi! Ước gì được nấy 🎂';
    if (!left) celebrate();
  }

  function celebrate() {
    stopMic();
    relight.hidden = false;
    wishEl.hidden = false;
    playBirthdaySong();
    burst();
    if (!celebrated) badge('Wish Made', 'Ước gì được nấy!');
    celebrated = true;
  }

  relight.addEventListener('click', () => {
    flames.forEach((f) => { f.classList.remove('is-out'); f.disabled = false; });
    relight.hidden = true;
    wishEl.hidden = true;
    status.textContent = `Còn ${flames.length} ngọn nến. Thổi tiếp nào!`;
  });

  async function startMic() {
    if (!navigator.mediaDevices?.getUserMedia) {
      status.textContent = 'Trình duyệt không hỗ trợ micro, bấm vào ngọn lửa nhé!';
      return;
    }
    micPending = true;
    micBtn.disabled = true;
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      if (mic || !lit().length) { // candles were blown by hand while the prompt was open
        stream.getTracks().forEach((t) => t.stop());
        return;
      }
      const ac = new (window.AudioContext || window.webkitAudioContext)();
      await ac.resume(); // created after an await, so it may start suspended
      const analyser = ac.createAnalyser();
      analyser.fftSize = 1024;
      ac.createMediaStreamSource(stream).connect(analyser);
      const buf = new Float32Array(analyser.fftSize);
      mic = { stream, ac, analyser, buf, loudSince: 0, raf: 0, timer: setTimeout(stopMic, MIC_TIMEOUT_MS) };
      micBtn.textContent = '🎤 Đang nghe… thổi mạnh vào!';
      micBtn.setAttribute('aria-pressed', 'true');
      listen();
    } catch {
      status.textContent = 'Không dùng được micro, bấm vào ngọn lửa để thổi nha!';
    } finally {
      micPending = false;
      micBtn.disabled = false;
    }
  }

  function listen() {
    if (!mic) return;
    const { buf } = mic;
    mic.analyser.getFloatTimeDomainData(buf);
    const rms = Math.sqrt(buf.reduce((s, v) => s + v * v, 0) / buf.length);
    const now = performance.now();
    if (rms > BLOW_RMS) {
      if (!mic.loudSince) mic.loudSince = now;
      if (now - mic.loudSince > BLOW_HOLD_MS) {
        const remaining = lit();
        if (remaining.length) blow(remaining[(Math.random() * remaining.length) | 0]);
        if (mic) mic.loudSince = now; // keep blowing → keep putting candles out
      }
    } else if (mic) {
      mic.loudSince = 0;
    }
    if (mic) mic.raf = requestAnimationFrame(listen);
  }

  function stopMic() {
    if (!mic) return;
    cancelAnimationFrame(mic.raf);
    clearTimeout(mic.timer);
    mic.stream.getTracks().forEach((t) => t.stop());
    mic.ac.close();
    mic = null;
    micBtn.textContent = '🎤 Thổi bằng micro';
    micBtn.setAttribute('aria-pressed', 'false');
  }

  micBtn.addEventListener('click', () => {
    if (micPending) return;
    if (mic) stopMic();
    else startMic();
  });
}
