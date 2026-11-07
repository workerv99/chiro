import { useSyncExternalStore } from 'react';
import { cloudSyncNow, getCloudSyncState, subscribeCloudSync } from '../../utils/cloud/syncEngine';
import type { CloudSyncState } from '../../utils/cloud/syncEngine';

/** Live cloud sync status (idle / syncing / error / paywalled / disabled) and last sync time. */
export function useCloudSync(): CloudSyncState & { syncNow: () => Promise<void> } {
  const state = useSyncExternalStore(subscribeCloudSync, getCloudSyncState, getCloudSyncState);
  return { ...state, syncNow: () => cloudSyncNow({ manual: true }) };
}
