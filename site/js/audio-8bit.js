// 8-bit audio synthesized with WebAudio — no audio files to download or license.
const MUTE_KEY = 'bd-muted';

let ctx = null;
let master = null;
let muted = readMuted();

function readMuted() {
  try { return localStorage.getItem(MUTE_KEY) === '1'; } catch { return false; }
}

/** Must be called from a user gesture (browsers block audio before one). */
export function unlockAudio() {
  // iOS: play through the silent switch like a game would.
  try { if (navigator.audioSession) navigator.audioSession.type = 'playback'; } catch { /* unsupported */ }
  if (!ctx) {
    const AC = window.AudioContext || window.webkitAudioContext;
    if (!AC) return;
    ctx = new AC();
    master = ctx.createGain();
    master.gain.value = muted ? 0 : 0.22;
    master.connect(ctx.destination);
  }
  if (ctx.state !== 'running') ctx.resume();
}

export function isMuted() { return muted; }

export function setMuted(m) {
  muted = m;
  try { localStorage.setItem(MUTE_KEY, m ? '1' : '0'); } catch { /* storage blocked */ }
  if (master) master.gain.setTargetAtTime(m ? 0 : 0.22, ctx.currentTime, 0.02);
}

// One enveloped oscillator note. `to` (optional) slides the pitch.
function tone(freq, start, dur, { type = 'square', vol = 0.5, to = null } = {}) {
  const osc = ctx.createOscillator();
  const gain = ctx.createGain();
  osc.type = type;
  osc.frequency.setValueAtTime(freq, start);
  if (to) osc.frequency.exponentialRampToValueAtTime(to, start + dur);
  gain.gain.setValueAtTime(0.0001, start);
  gain.gain.exponentialRampToValueAtTime(vol, start + 0.01);
  gain.gain.exponentialRampToValueAtTime(0.0001, start + dur);
  osc.connect(gain).connect(master);
  osc.start(start);
  osc.stop(start + dur + 0.02);
}

const NOTE = {
  A3: 220, C4: 261.63, D4: 293.66, E4: 329.63, F4: 349.23, G4: 392, A4: 440, B4: 493.88,
  C5: 523.25, D5: 587.33, E5: 659.25, F5: 698.46, G5: 783.99, A5: 880, B5: 987.77, C6: 1046.5, E6: 1318.5,
};

const SFX = {
  pop: (t) => tone(600, t, 0.08, { to: 950, vol: 0.4 }),
  coin: (t) => { tone(NOTE.B5, t, 0.07); tone(NOTE.E6, t + 0.07, 0.28); },
  levelup: (t) => ['C5', 'E5', 'G5', 'C6', 'G5', 'C6'].forEach((n, i) => tone(NOTE[n], t + i * 0.08, 0.14)),
  steal: (t) => tone(300, t, 0.16, { type: 'triangle', to: 1300, vol: 0.6 }),
  alarm: (t) => [0, 1, 2].forEach((i) => { tone(880, t + i * 0.2, 0.1); tone(660, t + i * 0.2 + 0.1, 0.1); }),
  spin: (t) => tone(1400, t, 0.03, { vol: 0.25 }),
  reveal: (t) => ['G4', 'C5', 'E5', 'G5', 'C6'].forEach((n, i) => tone(NOTE[n], t + i * 0.06, 0.35, { type: 'triangle', vol: 0.5 })),
};

/** name: 'pop'|'coin'|'levelup'|'steal'|'alarm'|'spin'|'reveal' */
export function playSfx(name) {
  if (!ctx || muted || !SFX[name]) return;
  if (ctx.state !== 'running') ctx.resume(); // iOS 'interrupted' after backgrounding
  SFX[name](ctx.currentTime + 0.01);
}

// Happy Birthday in C, 3/4. [note, beats]
const SONG = [
  ['G4', 0.75], ['G4', 0.25], ['A4', 1], ['G4', 1], ['C5', 1], ['B4', 2],
  ['G4', 0.75], ['G4', 0.25], ['A4', 1], ['G4', 1], ['D5', 1], ['C5', 2],
  ['G4', 0.75], ['G4', 0.25], ['G5', 1], ['E5', 1], ['C5', 1], ['B4', 1], ['A4', 1],
  ['F5', 0.75], ['F5', 0.25], ['E5', 1], ['C5', 1], ['D5', 1], ['C5', 3],
];
// Bass root per bar (3 beats each), pickup handled by starting one beat in.
const BASS = ['C4', 'G4', 'G4', 'C4', 'C4', 'F4', 'G4', 'C4'];

let songUntil = 0;

export function playBirthdaySong() {
  if (!ctx || muted) return;
  if (ctx.state !== 'running') ctx.resume();
  const now = ctx.currentTime;
  if (now < songUntil) return; // already playing
  const beat = 0.36;
  let t = now + 0.05;
  for (const [n, b] of SONG) {
    tone(NOTE[n], t, b * beat * 0.92, { vol: 0.42 });
    t += b * beat;
  }
  BASS.forEach((n, i) => tone(NOTE[n] / 2, now + 0.05 + beat + i * 3 * beat, 2.6 * beat, { type: 'triangle', vol: 0.5 }));
  songUntil = t;
}
