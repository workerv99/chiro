import { useRouter } from 'expo-router';
import { useState } from 'react';
import { ActivityIndicator, Alert, Pressable, StyleSheet, Text, View } from 'react-native';
import { subscriptionExpiryMs } from '../../utils/cloud';
import { useI18n } from '../../utils/i18n';
import { C, F, R, S } from '../../utils/theme';
import { useCloudAccount } from './CloudAccountProvider';
import { cloudStrings } from './strings';
import { useCloudSync } from './useCloudSync';

/** "Cloud sync" card for Settings: offline-only CTA or account/subscription status. */
export default function CloudSyncSection() {
  const { lang } = useI18n();
  const c = cloudStrings(lang);
  const router = useRouter();
  const cloud = useCloudAccount();
  const sync = useCloudSync();
  const [busy, setBusy] = useState(false);

  const run = async (fn: () => Promise<void>) => {
    setBusy(true);
    try { await fn(); } catch (e) { Alert.alert(c.error, e instanceof Error ? e.message : String(e)); } finally { setBusy(false); }
  };

  const onActivate = () => {
    if (cloud.user) run(cloud.activate);
    else router.push('/cloud-account?activate=1');
  };
  const onCancel = () => Alert.alert(c.cancel, c.cancelConfirm, [
    { text: c.keep, style: 'cancel' },
    { text: c.cancel, style: 'destructive', onPress: () => run(cloud.cancel) },
  ]);

  const sub = cloud.subscription;
  const expiry = subscriptionExpiryMs(sub);
  const statusLabel = sub?.status === 'active' ? c.statusActive : sub?.status === 'cancelled' ? c.statusCancelled : c.statusOther;
  const statusColor = cloud.syncEnabled ? C.income : C.textSec;

  return (
    <View style={s.card}>
      <Text style={s.title}>{c.title}</Text>
      <Text style={s.desc}>{c.desc}</Text>

      {!cloud.configured ? (
        <Text style={s.muted}>{c.notConfigured}</Text>
      ) : !cloud.user ? (
        <>
          <View style={[s.badge, { backgroundColor: C.surfaceUp }]}><Text style={[s.badgeText, { color: C.textSec }]}>{c.offline}</Text></View>
          <Pressable style={s.primary} onPress={onActivate} disabled={busy}>
            <Text style={s.primaryText}>{c.activate}</Text>
          </Pressable>
        </>
      ) : (
        <>
          <Text style={s.row}>{c.account}: <Text style={s.val}>{cloud.user.email}</Text></Text>
          <Text style={s.row}>{c.plan}: <Text style={s.val}>{sub?.plan === 'pro' ? c.planPro : c.planFree}</Text></Text>
          <Text style={s.row}>{c.status}: <Text style={s.val}>{statusLabel}</Text></Text>
          {expiry ? <Text style={s.row}>{c.expires}: <Text style={s.val}>{new Date(expiry).toLocaleDateString()}</Text></Text> : null}
          <View style={[s.badge, { backgroundColor: cloud.syncEnabled ? C.incomeBg : C.surfaceUp }]}>
            <Text style={[s.badgeText, { color: statusColor }]}>{cloud.syncEnabled ? c.syncOn : c.syncOff}</Text>
          </View>

          {cloud.syncEnabled ? (
            <>
              <Text style={s.row}>
                {sync.status === 'syncing' ? c.syncRunning
                  : sync.status === 'error' ? c.syncFailed
                  : sync.status === 'paywalled' ? c.syncPaywalled
                  : c.syncIdle}
                {sync.status === 'error' && sync.error ? `: ${sync.error}` : ''}
              </Text>
              <Text style={s.row}>{c.lastSync}: <Text style={s.val}>{sync.lastSyncAt ? new Date(sync.lastSyncAt).toLocaleString() : c.never}</Text></Text>
              <Pressable style={s.outline} onPress={() => { void sync.syncNow(); }} disabled={sync.status === 'syncing'}>
                <Text style={s.outlineText}>{c.syncNow}</Text>
              </Pressable>
            </>
          ) : null}

          {!cloud.syncEnabled ? (
            <Pressable style={s.primary} onPress={onActivate} disabled={busy}><Text style={s.primaryText}>{c.activate}</Text></Pressable>
          ) : (
            <Pressable style={[s.outline, { borderColor: C.expense + '44' }]} onPress={onCancel} disabled={busy}>
              <Text style={[s.outlineText, { color: C.expense }]}>{c.cancel}</Text>
            </Pressable>
          )}
          <Pressable style={s.outline} onPress={() => run(cloud.refresh)} disabled={busy}><Text style={s.outlineText}>{c.refresh}</Text></Pressable>
          <Pressable style={s.outline} onPress={() => run(cloud.logout)} disabled={busy}><Text style={s.outlineText}>{c.logout}</Text></Pressable>
        </>
      )}
      {busy ? <ActivityIndicator color={C.accent} style={{ marginTop: S.md }} /> : null}
    </View>
  );
}

const s = StyleSheet.create({
  card: { backgroundColor: C.surface, borderRadius: R.md, borderWidth: 1, borderColor: C.border, padding: S.lg, marginBottom: S.lg },
  title: { fontSize: F.md, fontWeight: '700', color: C.text, marginBottom: 4 },
  desc: { fontSize: F.sm, color: C.textSec, lineHeight: 20, marginBottom: S.md },
  muted: { fontSize: F.sm, color: C.textDim },
  row: { fontSize: F.sm, color: C.textSec, marginBottom: S.xs },
  val: { color: C.text, fontWeight: '600' },
  badge: { alignSelf: 'flex-start', borderRadius: R.full, paddingHorizontal: S.md, paddingVertical: S['2xs'], marginVertical: S.sm },
  badgeText: { fontSize: F.xs, fontWeight: '700' },
  primary: { backgroundColor: C.accent, borderRadius: R.sm, paddingVertical: 14, alignItems: 'center', marginTop: S.sm },
  primaryText: { fontSize: F.base, fontWeight: '700', color: '#fff' },
  outline: { borderWidth: 1, borderColor: C.border, borderRadius: R.sm, paddingVertical: 12, alignItems: 'center', marginTop: S.sm },
  outlineText: { fontSize: F.sm, fontWeight: '600', color: C.textSec },
});
