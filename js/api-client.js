// Thin client for the birthday API. Never throws: every call resolves to
// { ok, status, data?, error? } so callers can always fall back gracefully.
import { API_BASE } from './config.js';

const WRITE_TIMEOUT_MS = 8000;
const READ_TIMEOUT_MS = 15000;
const WAKE_TIMEOUT_MS = 60000; // free-tier host can take ~1 min to wake up

export const apiEnabled = () => API_BASE !== '';

async function request(path, { method = 'GET', body, timeout = READ_TIMEOUT_MS, retry = method === 'GET' } = {}) {
  if (!apiEnabled()) return { ok: false, status: 0, error: 'offline' };
  const ctrl = new AbortController();
  const timer = setTimeout(() => ctrl.abort(), timeout);
  try {
    const res = await fetch(API_BASE + path, {
      method,
      signal: ctrl.signal,
      headers: body ? { 'Content-Type': 'application/json' } : undefined,
      body: body ? JSON.stringify(body) : undefined,
    });
    const data = await res.json().catch(() => null);
    return res.ok
      ? { ok: true, status: res.status, data }
      : { ok: false, status: res.status, error: data?.error || `HTTP ${res.status}` };
  } catch (err) {
    // While the host wakes up it answers without CORS headers, which surfaces
    // as a network error — one silent retry covers most of those.
    if (retry && err.name !== 'AbortError') return request(path, { method, body, timeout, retry: false });
    return { ok: false, status: 0, error: err.name === 'AbortError' ? 'timeout' : 'network' };
  } finally {
    clearTimeout(timer);
  }
}

export const ping = () => request('/healthz', { timeout: WAKE_TIMEOUT_MS });

export async function listWishes() {
  const r = await request('/api/wishes?limit=50');
  return r.ok ? { ...r, data: r.data?.wishes ?? [] } : r;
}

export const createWish = ({ name, message, website }) =>
  request('/api/wishes', { method: 'POST', body: { name, message, website }, timeout: WRITE_TIMEOUT_MS });

export async function listScores() {
  const r = await request('/api/scores?limit=10');
  return r.ok ? { ...r, data: r.data?.scores ?? [] } : r;
}

export const submitScore = ({ name, score }) =>
  request('/api/scores', { method: 'POST', body: { name, score }, timeout: WRITE_TIMEOUT_MS });
