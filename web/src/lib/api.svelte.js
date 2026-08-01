// Cliente HTTP hacia la API Go (auth Bearer, JSON).
const BASE = import.meta.env.VITE_API_BASE_URL || '';
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

export async function api(path, { method = 'GET', body } = {}) {
  const headers = { 'Content-Type': 'application/json' };
  if (A.token) headers.Authorization = 'Bearer ' + A.token;
  const res = await fetch(BASE + path, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined
  });
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
