import { createCloudClient } from './client';
import { sessionStore } from './tokenStore';

export { CloudApiError } from './errors';
export { isProActive, subscriptionExpiryMs } from './entitlement';
export type { AuthResponse, CloudUser, LoginInput, RegisterInput, SessionStore, Subscription } from './types';
export { sessionStore };

/** Base URL of the Chiro backend; empty means cloud features are not configured. */
export const API_URL = process.env.EXPO_PUBLIC_API_URL ?? '';

let expiredListener: (() => void) | null = null;
/** Lets the provider react when a failed refresh dropped the session. */
export function setSessionExpiredListener(fn: (() => void) | null) { expiredListener = fn; }

export const cloudApi = createCloudClient({
  baseUrl: API_URL,
  store: sessionStore,
  onSessionExpired: () => expiredListener?.(),
});
