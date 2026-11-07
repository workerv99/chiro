/**
 * Cloud account: login / register, optionally followed by subscription activation.
 */
import { Ionicons } from '@expo/vector-icons';
import { useLocalSearchParams, useRouter } from 'expo-router';
import { useState } from 'react';
import {
  ActivityIndicator, KeyboardAvoidingView, Platform, Pressable, ScrollView, StyleSheet, Text, TextInput, View,
} from 'react-native';
import { useCloudAccount } from '../components/cloud/CloudAccountProvider';
import { cloudStrings } from '../components/cloud/strings';
import { useI18n } from '../utils/i18n';
import { FORM_MAX_W } from '../utils/responsive';
import { C, F, R, S } from '../utils/theme';

function Check({ checked, label, onPress }: { checked: boolean; label: string; onPress: () => void }) {
  return (
    <Pressable style={s.check} onPress={onPress} accessibilityRole="checkbox" accessibilityState={{ checked }}>
      <Ionicons name={checked ? 'checkbox' : 'square-outline'} size={22} color={checked ? C.accent : C.textDim} />
      <Text style={s.checkLabel}>{label}</Text>
    </Pressable>
  );
}

export default function CloudAccountScreen() {
  const { lang } = useI18n();
  const c = cloudStrings(lang);
  const router = useRouter();
  const { activate: activateParam } = useLocalSearchParams<{ activate?: string }>();
  const wantsActivate = activateParam === '1';
  const cloud = useCloudAccount();

  const [mode, setMode] = useState<'login' | 'register'>('login');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [name, setName] = useState('');
  const [terms, setTerms] = useState(false);
  const [privacy, setPrivacy] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const isRegister = mode === 'register';

  const submit = async () => {
    setError(null);
    if (!email.trim() || !password || (isRegister && !name.trim())) { setError(c.fillAll); return; }
    if (isRegister && !(terms && privacy)) { setError(c.acceptTerms); return; }
    setBusy(true);
    try {
      if (isRegister) {
        await cloud.register({ email: email.trim(), password, name: name.trim(), terms_accepted: terms, privacy_accepted: privacy });
      } else {
        await cloud.login({ email: email.trim(), password });
      }
      if (wantsActivate) await cloud.activate();
      router.back();
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setBusy(false);
    }
  };

  return (
    <KeyboardAvoidingView style={s.root} behavior={Platform.OS === 'ios' ? 'padding' : undefined}>
      <ScrollView contentContainerStyle={s.content} keyboardShouldPersistTaps="handled">
        <View style={s.form}>
          <Text style={s.title}>{c.screenTitle}</Text>
          {wantsActivate ? <Text style={s.hint}>{c.activateHint}</Text> : null}

          <View style={s.tabs}>
            {(['login', 'register'] as const).map(m => (
              <Pressable key={m} style={[s.tab, mode === m && s.tabActive]} onPress={() => { setMode(m); setError(null); }}>
                <Text style={[s.tabText, mode === m && { color: C.accent, fontWeight: '700' }]}>
                  {m === 'login' ? c.loginTab : c.registerTab}
                </Text>
              </Pressable>
            ))}
          </View>

          {isRegister ? (
            <TextInput style={s.input} value={name} onChangeText={setName} placeholder={c.name}
              placeholderTextColor={C.textDim} autoCapitalize="words" />
          ) : null}
          <TextInput style={s.input} value={email} onChangeText={setEmail} placeholder={c.email}
            placeholderTextColor={C.textDim} autoCapitalize="none" keyboardType="email-address" autoCorrect={false} />
          <TextInput style={s.input} value={password} onChangeText={setPassword} placeholder={c.password}
            placeholderTextColor={C.textDim} secureTextEntry autoCapitalize="none" onSubmitEditing={submit} />

          {isRegister ? (
            <>
              <Check checked={terms} label={c.terms} onPress={() => setTerms(v => !v)} />
              <Check checked={privacy} label={c.privacy} onPress={() => setPrivacy(v => !v)} />
            </>
          ) : null}

          {error ? <Text style={s.error}>{error}</Text> : null}

          <Pressable style={[s.submit, busy && { opacity: 0.6 }]} onPress={submit} disabled={busy}>
            {busy ? <ActivityIndicator color="#fff" /> : (
              <Text style={s.submitText}>
                {wantsActivate ? c.submitActivate : isRegister ? c.submitRegister : c.submitLogin}
              </Text>
            )}
          </Pressable>
        </View>
      </ScrollView>
    </KeyboardAvoidingView>
  );
}

const s = StyleSheet.create({
  root: { flex: 1, backgroundColor: C.bg },
  content: { padding: S.lg, alignItems: 'center' },
  form: { width: '100%', maxWidth: FORM_MAX_W },
  title: { fontSize: F.xl, fontWeight: '800', color: C.text, marginBottom: S.sm },
  hint: { fontSize: F.sm, color: C.textSec, marginBottom: S.md, lineHeight: 20 },
  tabs: { flexDirection: 'row', gap: S.sm, marginVertical: S.md },
  tab: { flex: 1, paddingVertical: 10, borderRadius: R.sm, alignItems: 'center', backgroundColor: C.surfaceUp },
  tabActive: { backgroundColor: C.accentBg },
  tabText: { fontSize: F.sm, color: C.textSec },
  input: {
    backgroundColor: C.surfaceUp, borderWidth: 1, borderColor: C.border, borderRadius: R.sm,
    paddingHorizontal: S.md, paddingVertical: 12, fontSize: F.base, color: C.text, marginBottom: S.sm,
  },
  check: { flexDirection: 'row', alignItems: 'center', gap: S.sm, paddingVertical: S.xs },
  checkLabel: { flex: 1, fontSize: F.sm, color: C.textSec },
  error: { color: C.danger, fontSize: F.sm, marginTop: S.sm },
  submit: { backgroundColor: C.accent, borderRadius: R.sm, paddingVertical: 14, alignItems: 'center', marginTop: S.lg },
  submitText: { fontSize: F.base, fontWeight: '700', color: '#fff' },
});
