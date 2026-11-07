import type { Subscription } from './types';

function toMillis(v: string | number | null | undefined): number | null {
  if (v == null || v === '') return null;
  const ms = typeof v === 'number' ? v : Date.parse(v);
  return Number.isFinite(ms) ? ms : null;
}

/** Expiry in epoch ms, or null when unknown / not provided. */
export function subscriptionExpiryMs(sub: Subscription | null | undefined): number | null {
  return sub ? toMillis(sub.expires_at) : null;
}

/**
 * Cloud sync is enabled only for plan=pro, status=active and not expired.
 * A missing expiry is treated as not expired (the backend enforces it on /api/sync).
 */
export function isProActive(sub: Subscription | null | undefined, now: number = Date.now()): boolean {
  if (!sub) return false;
  if (sub.plan !== 'pro' || sub.status !== 'active') return false;
  const exp = toMillis(sub.expires_at);
  return exp === null || exp > now;
}
