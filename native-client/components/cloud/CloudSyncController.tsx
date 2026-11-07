import { useEffect, useRef } from 'react';
import { AppState } from 'react-native';
import { cloudSyncNow, configureCloudSync } from '../../utils/cloud/syncEngine';
import { useExpenses } from '../../utils/context';
import { useCloudAccount } from './CloudAccountProvider';
import { useCloudSync } from './useCloudSync';

/**
 * Connects the cloud sync engine to the account state: it syncs on login and when
 * the subscription becomes active (engine start), on app foreground, and refreshes
 * the subscription once when the server answers 402. Renders nothing.
 */
export default function CloudSyncController() {
  const { user, syncEnabled, refresh } = useCloudAccount();
  const { reloadData } = useExpenses();
  const { status } = useCloudSync();

  useEffect(() => {
    configureCloudSync({ enabled: syncEnabled, userId: user?.user_id ?? null, onRemoteApplied: reloadData });
  }, [syncEnabled, user?.user_id, reloadData]);

  useEffect(() => {
    const sub = AppState.addEventListener('change', (s) => {
      if (s === 'active') void cloudSyncNow({ manual: false });
    });
    return () => sub.remove();
  }, []);

  // The server said 402: re-read the subscription so the UI (and syncEnabled) catch up.
  const refreshed = useRef(false);
  useEffect(() => {
    if (status !== 'paywalled') { refreshed.current = false; return; }
    if (!refreshed.current) { refreshed.current = true; void refresh(); }
  }, [status, refresh]);

  return null;
}
