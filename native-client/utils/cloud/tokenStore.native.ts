/** Native (Android/iOS): session kept in the OS keystore via expo-secure-store. */
import * as SecureStore from 'expo-secure-store';
import type { CloudUser, SessionStore } from './types';

const TOKEN_KEY = 'cloud_token';
const USER_KEY = 'cloud_user';

export const sessionStore: SessionStore = {
  getToken: () => SecureStore.getItemAsync(TOKEN_KEY),
  setToken: (token) => SecureStore.setItemAsync(TOKEN_KEY, token),
  async getUser() {
    const raw = await SecureStore.getItemAsync(USER_KEY);
    try { return raw ? (JSON.parse(raw) as CloudUser) : null; } catch { return null; }
  },
  setUser: (user) => SecureStore.setItemAsync(USER_KEY, JSON.stringify(user)),
  async clear() {
    await SecureStore.deleteItemAsync(TOKEN_KEY);
    await SecureStore.deleteItemAsync(USER_KEY);
  },
};
