import React, { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react';
import {
  API_URL, cloudApi, CloudApiError, isProActive, sessionStore, setSessionExpiredListener,
} from '../../utils/cloud';
import type { CloudUser, LoginInput, RegisterInput, Subscription } from '../../utils/cloud';
import CloudSyncController from './CloudSyncController';

interface CloudAccountState {
  /** False when EXPO_PUBLIC_API_URL is not set. */
  configured: boolean;
  /** True while restoring the stored session. */
  loading: boolean;
  user: CloudUser | null;
  subscription: Subscription | null;
  /** plan=pro, status=active and not expired, as reported by the server. */
  syncEnabled: boolean;
  login: (input: LoginInput) => Promise<void>;
  register: (input: RegisterInput) => Promise<void>;
  /** Activates Pro for the logged-in user, then refreshes the status. */
  activate: () => Promise<void>;
  cancel: () => Promise<void>;
  refresh: () => Promise<void>;
  logout: () => Promise<void>;
}

const noop = async () => {};
const CloudAccountContext = createContext<CloudAccountState>({
  configured: false, loading: true, user: null, subscription: null, syncEnabled: false,
  login: noop, register: noop, activate: noop, cancel: noop, refresh: noop, logout: noop,
});

export function CloudAccountProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<CloudUser | null>(null);
  const [subscription, setSubscription] = useState<Subscription | null>(null);
  const [loading, setLoading] = useState(true);
  const configured = API_URL.length > 0;

  const refresh = useCallback(async () => {
    try {
      setSubscription(await cloudApi.getSubscription());
    } catch (e) {
      // Offline or server down: keep what we had. Only a dropped session clears state.
      if (e instanceof CloudApiError && e.isUnauthorized) { setUser(null); setSubscription(null); }
    }
  }, []);

  useEffect(() => {
    setSessionExpiredListener(() => { setUser(null); setSubscription(null); });
    return () => setSessionExpiredListener(null);
  }, []);

  useEffect(() => {
    let alive = true;
    (async () => {
      try {
        const stored = configured && (await sessionStore.getToken()) ? await sessionStore.getUser() : null;
        if (!alive) return;
        setUser(stored);
        if (stored) await refresh();
      } finally {
        if (alive) setLoading(false);
      }
    })();
    return () => { alive = false; };
  }, [configured, refresh]);

  const value = useMemo<CloudAccountState>(() => ({
    configured,
    loading,
    user,
    subscription,
    syncEnabled: !!user && isProActive(subscription),
    login: async (input) => { const r = await cloudApi.login(input); setUser(r.user); await refresh(); },
    register: async (input) => { const r = await cloudApi.register(input); setUser(r.user); await refresh(); },
    activate: async () => { await cloudApi.activateSubscription(); await refresh(); },
    cancel: async () => { await cloudApi.cancelSubscription(); await refresh(); },
    refresh,
    logout: async () => { await cloudApi.logout(); setUser(null); setSubscription(null); },
  }), [configured, loading, user, subscription, refresh]);

  return (
    <CloudAccountContext.Provider value={value}>
      <CloudSyncController />
      {children}
    </CloudAccountContext.Provider>
  );
}

export function useCloudAccount() {
  return useContext(CloudAccountContext);
}
