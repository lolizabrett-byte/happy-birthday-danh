// Egg-thief leaderboard: top 10 from the API, local best when offline.
// Listens for `egg:gameover` from the Steal An Egg game.
import { listScores, submitScore } from './api-client.js';

const BEST_KEY = 'bd-best';
const NAME_MAX = 20;

const TEMPLATE = `
  <div class="lb">
    <h3 class="lb__title">🏆 Bảng xếp hạng trộm trứng</h3>
    <ol class="lb__list" data-ref="list"></ol>
    <p class="lb__note" data-ref="note"></p>
    <form class="lb__save" data-ref="save" hidden>
      <label for="lb-name">Lưu <b data-ref="score">0</b> điểm với tên</label>
      <div class="lb__row">
        <input id="lb-name" name="name" maxlength="${NAME_MAX}" autocomplete="nickname" required>
        <button class="btn btn--egg" type="submit">Lưu</button>
      </div>
      <p class="field__error" data-ref="err" aria-live="polite"></p>
    </form>
  </div>`;

const readBest = () => { try { return Number(localStorage.getItem(BEST_KEY)) || 0; } catch { return 0; } };
const writeBest = (v) => { try { localStorage.setItem(BEST_KEY, String(v)); } catch { /* storage blocked */ } };

export function mountLeaderboard(root, { wake, getName, saveName }) {
  root.innerHTML = TEMPLATE; // static markup only
  const $ = (r) => root.querySelector(`[data-ref="${r}"]`);
  const form = $('save');
  let online = false;
  let pending = 0;

  const showLocal = () => {
    const best = readBest();
    $('note').textContent = best ? `Kỷ lục trên máy bạn: ${best} điểm` : 'Chơi một ván Steal An Egg để lên bảng nhé!';
  };

  function renderScores(scores) {
    const list = $('list');
    list.replaceChildren();
    scores.forEach((s, i) => {
      const li = document.createElement('li');
      li.className = `lb__item${i < 3 ? ` lb__item--top${i + 1}` : ''}`;
      const rank = document.createElement('span');
      rank.className = 'lb__rank';
      rank.textContent = `#${i + 1}`;
      const name = document.createElement('span');
      name.className = 'lb__name';
      name.textContent = s.name;
      const pts = document.createElement('b');
      pts.textContent = s.score;
      li.append(rank, name, pts);
      list.append(li);
    });
    if (!scores.length) $('note').textContent = 'Chưa ai lên bảng, bạn là người đầu tiên đi!';
  }

  async function refresh() {
    const r = await listScores();
    if (r.ok) renderScores(r.data);
  }

  showLocal();
  wake.then((r) => {
    online = r.ok;
    if (online) refresh();
  });

  window.addEventListener('egg:gameover', (e) => {
    const score = e.detail?.score ?? 0;
    if (score > readBest()) writeBest(score);
    showLocal();
    if (!online || score <= 0) return;
    pending = score;
    $('score').textContent = score;
    form.name.value = getName().slice(0, NAME_MAX);
    $('err').textContent = '';
    form.hidden = false;
  });

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    const name = form.name.value.trim();
    if (!name) { $('err').textContent = 'Nhập tên để lưu điểm nha.'; return; }
    const btn = form.querySelector('button');
    btn.disabled = true;
    const r = await submitScore({ name, score: pending });
    btn.disabled = false;
    if (r.ok) {
      saveName(name);
      form.hidden = true;
      refresh();
    } else {
      $('err').textContent = r.status === 429 ? 'Lưu nhanh quá, đợi 1 phút nha.' : 'Không lưu được điểm, thử lại sau nhé.';
    }
  });
}
