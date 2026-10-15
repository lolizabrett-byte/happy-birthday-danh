// The letter from kho, typed out as Roblox chat when it scrolls into view.
import { playSfx } from './audio-8bit.js';

const LINES = [
  { who: 'system', text: '[Server] SaveAsDanh đã lên Level 20! 🎉' },
  { who: 'kho', text: 'Danh ơi! Chúc mừng sinh nhật tuổi 20 nha ông bạn đại học kiêm đồng đội Roblox số 1 của tui 🎂' },
  { who: 'kho', text: 'Cảm ơn ông vì bao nhiêu trận Steal An Egg cướp trứng tới khuya, bao nhiêu lần cày Blox Fruits chung sau giờ học. Mấy lúc đó vui thiệt sự.' },
  { who: 'kho', text: 'Chúc Danh có một bữa tiệc sinh nhật thật hạnh phúc và ấm áp, bên cạnh gia đình và những người thương Danh nhất.' },
  { who: 'kho', text: 'Tuổi 20 rồi, chúc ông level up không chỉ trong game mà cả ngoài đời: mỗi ngày phát triển bản thân thêm một chút, học hành suôn sẻ, ngày càng vững vàng.' },
  { who: 'kho', text: 'Chúc Danh luôn giữ lửa đam mê, theo đuổi tới cùng và hoàn thành mọi ước mơ, nguyện vọng của mình. Kiên trì như cách ông roll ra trái Mythical vậy đó 😆' },
  { who: 'kho', text: 'Sau này có bận cỡ nào thì cũng nhớ chừa một slot trong server cho tui vào chơi chung nha!' },
  { who: 'kho', text: 'Happy Birthday, SaveAsDanh! GG WP tuổi 19, welcome tuổi 20 🥳' },
  { who: 'sign', text: '— from kho' },
];

const reduceMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

function lineEl({ who }) {
  const li = document.createElement('li');
  li.className = 'chat__line';
  if (who === 'system') li.classList.add('chat__line--system');
  if (who === 'sign') li.classList.add('chat__line--sign');
  if (who === 'kho') {
    const name = document.createElement('span');
    name.className = 'chat__who chat__who--kho';
    name.textContent = '[kho]:';
    li.append(name);
  }
  const body = document.createElement('span');
  li.append(body);
  return { li, body };
}

async function typeAll(log) {
  for (const line of LINES) {
    const { li, body } = lineEl(line);
    log.append(li);
    if (reduceMotion) {
      body.textContent = line.text;
      continue;
    }
    const caret = document.createElement('span');
    caret.className = 'chat__caret';
    li.append(caret);
    const chars = Array.from(line.text); // keeps emoji / Vietnamese combining marks intact
    for (let i = 0; i < chars.length; i++) {
      body.textContent += chars[i];
      if (i % 6 === 0) playSfx('spin');
      await sleep(22);
    }
    caret.remove();
    await sleep(380);
  }
}

export function initLetter() {
  const log = document.getElementById('chat-log');
  if (!log) return;
  // Screen readers get the whole letter up front; the typed copy is visual only.
  const full = document.createElement('div');
  full.className = 'sr-only';
  full.textContent = LINES.map((l) => (l.who === 'kho' ? `kho: ${l.text}` : l.text)).join(' ');
  log.before(full);
  log.setAttribute('aria-hidden', 'true');
  const io = new IntersectionObserver((entries) => {
    if (!entries.some((e) => e.isIntersecting)) return;
    io.disconnect();
    typeAll(log);
  }, { threshold: 0.35 });
  io.observe(log);
}
