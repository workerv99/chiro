// Cliente HTTP hacia la API Go (auth Bearer, JSON).
export const BASE = import.meta.env.VITE_API_BASE_URL || '';
const TOKEN_KEY = 'chiro_token';

export const A = $state({ token: '' });

export function loadToken() {
  try {
    A.token = localStorage.getItem(TOKEN_KEY) || '';
  } catch { /* ignore */ }
}

export function setToken(t) {
  A.token = t;
  try {
    if (t) localStorage.setItem(TOKEN_KEY, t);
    else localStorage.removeItem(TOKEN_KEY);
  } catch { /* ignore */ }
}

// Auth endpoints never trigger a silent refresh on 401 (avoids loops and
// keeps wrong-credentials errors visible).
const NO_REFRESH = ['/api/auth/login', '/api/auth/register', '/api/auth/refresh'];

let refreshing = null;

// Exchanges the current (possibly expired) token for a fresh one. Single-flight:
// concurrent callers share one in-flight request. Resolves true on success.
// A definitive rejection (401/403) clears the stored token; transient failures
// (network, 5xx) leave it in place so a later attempt can still succeed.
export function refreshToken() {
  if (!A.token) return Promise.resolve(false);
  if (!refreshing) {
    const old = A.token;
    refreshing = (async () => {
      try {
        const res = await fetch(BASE + '/api/auth/refresh', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json', Authorization: 'Bearer ' + old }
        });
        if (res.status === 401 || res.status === 403) {
          setToken('');
          return false;
        }
        if (!res.ok) return false;
        const data = await res.json().catch(() => null);
        if (!data || !data.token) return false;
        setToken(data.token);
        return true;
      } catch {
        return false;
      } finally {
        refreshing = null;
      }
    })();
  }
  return refreshing;
}

async function request(path, { method, body, signal }) {
  const headers = { 'Content-Type': 'application/json' };
  if (A.token) headers.Authorization = 'Bearer ' + A.token;
  return fetch(BASE + path, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
    signal
  });
}

export async function api(path, { method = 'GET', body, signal } = {}) {
  let res = await request(path, { method, body, signal });
  if (res.status === 401 && !NO_REFRESH.includes(path) && A.token) {
    // Retry exactly once with the refreshed token.
    if (await refreshToken()) res = await request(path, { method, body, signal });
  }
  if (res.status === 401) {
    setToken('');
    throw new Error('unauthorized');
  }
  const data = await res.json().catch(() => null);
  if (!res.ok) {
    throw new Error(data && data.error ? data.error : 'Error ' + res.status);
  }
  return data;
}
