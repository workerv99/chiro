import { CloudApiError, errorFromResponse, NETWORK_ERROR_MESSAGE, TIMEOUT_ERROR_MESSAGE } from './errors';
import { singleFlight } from './singleFlight';
import type {
  AuthResponse, LoginInput, RegisterInput, SessionStore, Subscription,
} from './types';

export const REQUEST_TIMEOUT_MS = 15_000;

export interface ClientOptions {
  baseUrl: string;
  store: SessionStore;
  fetchImpl?: typeof fetch;
  timeoutMs?: number;
  /** Called after a failed refresh cleared the session (token no longer valid). */
  onSessionExpired?: () => void;
}

interface RequestOptions {
  method?: 'GET' | 'POST';
  body?: unknown;
  /** Attach the Bearer token and handle 401 with refresh + retry. Default true. */
  auth?: boolean;
}

/** Framework-free API client: no React Native imports, so it can be unit tested. */
export function createCloudClient(opts: ClientOptions) {
  const { store, onSessionExpired } = opts;
  const baseUrl = opts.baseUrl.replace(/\/+$/, '');
  const timeoutMs = opts.timeoutMs ?? REQUEST_TIMEOUT_MS;
  const doFetch: typeof fetch = opts.fetchImpl ?? ((...a) => fetch(...a));

  async function send(path: string, o: RequestOptions, token: string | null): Promise<unknown> {
    const headers: Record<string, string> = { Accept: 'application/json' };
    if (o.body !== undefined) headers['Content-Type'] = 'application/json';
    if (token) headers.Authorization = `Bearer ${token}`;

    const ctrl = new AbortController();
    const timer = setTimeout(() => ctrl.abort(), timeoutMs);
    let res: Response;
    try {
      res = await doFetch(`${baseUrl}${path}`, {
        method: o.method ?? 'GET',
        headers,
        body: o.body !== undefined ? JSON.stringify(o.body) : undefined,
        signal: ctrl.signal,
      });
    } catch (e) {
      const aborted = (e as { name?: string } | null)?.name === 'AbortError';
      throw new CloudApiError(aborted ? -1 : 0, aborted ? TIMEOUT_ERROR_MESSAGE : NETWORK_ERROR_MESSAGE);
    } finally {
      clearTimeout(timer);
    }

    let parsed: unknown = null;
    try { parsed = await res.json(); } catch { /* empty or non-JSON body */ }
    if (!res.ok) throw errorFromResponse(res.status, parsed);
    return parsed;
  }

  async function saveAuth(r: AuthResponse): Promise<AuthResponse> {
    await store.setToken(r.token);
    await store.setUser(r.user);
    return r;
  }

  /** Exchanges the current JWT for a new one. Concurrent callers share one request. */
  const refresh = singleFlight(async (): Promise<string> => {
    const old = await store.getToken();
    if (!old) throw new CloudApiError(401, 'Sesión expirada');
    const r = await send('/api/auth/refresh', { method: 'POST' }, old);
    await saveAuth(r as AuthResponse);
    return (r as AuthResponse).token;
  });

  async function expire(): Promise<void> {
    await store.clear();
    onSessionExpired?.();
  }

  async function request<T>(path: string, o: RequestOptions = {}): Promise<T> {
    if (o.auth === false) return (await send(path, o, null)) as T;

    const token = await store.getToken();
    try {
      return (await send(path, o, token)) as T;
    } catch (e) {
      if (!(e instanceof CloudApiError) || e.status !== 401 || !token) throw e;
    }
    // 401: one single-flight refresh, one retry, otherwise drop the session.
    try {
      const fresh = await refresh();
      return (await send(path, o, fresh)) as T;
    } catch (e) {
      if (e instanceof CloudApiError && e.status === 401) {
        await expire();
        throw e;
      }
      throw e;
    }
  }

  return {
    register: async (input: RegisterInput) =>
      saveAuth(await request<AuthResponse>('/api/auth/register', { method: 'POST', body: input, auth: false })),
    login: async (input: LoginInput) =>
      saveAuth(await request<AuthResponse>('/api/auth/login', { method: 'POST', body: input, auth: false })),
    refresh,
    /** Best effort: the local session is always cleared, even if the server call fails. */
    logout: async () => {
      try {
        if (await store.getToken()) await send('/api/auth/logout', { method: 'POST' }, await store.getToken());
      } catch { /* server-side logout is best effort */ }
      await store.clear();
    },
    getSubscription: () => request<Subscription>('/api/subscription'),
    activateSubscription: () => request<{ ok: boolean }>('/api/subscription/activate', { method: 'POST' }),
    cancelSubscription: () => request<{ ok: boolean }>('/api/subscription/cancel', { method: 'POST' }),
    request,
  };
}

export type CloudClient = ReturnType<typeof createCloudClient>;
