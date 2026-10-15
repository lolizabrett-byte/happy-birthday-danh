// Lightweight canvas confetti: blocky pieces, gravity, auto-stops when empty.
const COLORS = ['#335fff', '#ffc83d', '#ff5c8a', '#1fb6c9', '#5bd16e', '#ffffff', '#b07cff'];
const reduceMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

let canvas = null;
let ctx2d = null;
let pieces = [];
let running = false;
let lastTs = 0;

function setup() {
  canvas = document.getElementById('confetti');
  if (!canvas) return false;
  ctx2d = canvas.getContext('2d');
  resize();
  window.addEventListener('resize', resize);
  return true;
}

function resize() {
  const dpr = Math.min(window.devicePixelRatio || 1, 2);
  canvas.width = innerWidth * dpr;
  canvas.height = innerHeight * dpr;
  ctx2d.setTransform(dpr, 0, 0, dpr, 0, 0);
}

/** Burst from (x, y) in viewport px; defaults to top-center rain. */
export function burst({ x, y, count = 150 } = {}) {
  if (!canvas && !setup()) return;
  const n = reduceMotion ? Math.round(count / 5) : count;
  const fromTop = x === undefined;
  for (let i = 0; i < n; i++) {
    const angle = fromTop ? Math.PI / 2 + (Math.random() - 0.5) * 0.8 : Math.random() * Math.PI * 2;
    const speed = fromTop ? 2 + Math.random() * 4 : 4 + Math.random() * 8;
    pieces.push({
      x: fromTop ? Math.random() * innerWidth : x,
      y: fromTop ? -20 - Math.random() * 120 : y,
      vx: Math.cos(angle) * speed,
      vy: fromTop ? Math.random() * 2 : Math.sin(angle) * speed - 4,
      size: 6 + Math.random() * 7,
      rot: Math.random() * Math.PI,
      vr: (Math.random() - 0.5) * 0.3,
      color: COLORS[(Math.random() * COLORS.length) | 0],
      age: 0,
    });
  }
  if (!running) {
    running = true;
    lastTs = performance.now();
    requestAnimationFrame(tick);
  }
}

// Velocities are tuned in px per 60 Hz frame; k rescales them so 120 Hz screens fall at the same speed.
function tick(ts) {
  const k = Math.min(3, Math.max(0, (ts - lastTs) / (1000 / 60)));
  lastTs = ts;
  ctx2d.clearRect(0, 0, innerWidth, innerHeight);
  pieces = pieces.filter((p) => p.y < innerHeight + 40 && p.age < 10000);
  for (const p of pieces) {
    p.vy += 0.18 * k;
    p.vx *= 0.99 ** k;
    p.x += p.vx * k;
    p.y += p.vy * k;
    p.rot += p.vr * k;
    p.age += k * (1000 / 60);
    ctx2d.save();
    ctx2d.translate(p.x, p.y);
    ctx2d.rotate(p.rot);
    ctx2d.fillStyle = p.color;
    ctx2d.fillRect(-p.size / 2, -p.size / 2, p.size, p.size * 0.6);
    ctx2d.restore();
  }
  if (pieces.length) {
    requestAnimationFrame(tick);
  } else {
    running = false;
    ctx2d.clearRect(0, 0, innerWidth, innerHeight);
  }
}
