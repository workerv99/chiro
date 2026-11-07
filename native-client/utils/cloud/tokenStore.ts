/** Web / Electron: SecureStore is unavailable, so the session lives in AsyncStorage. */
import AsyncStorage from '@react-native-async-storage/async-storage';
import type { CloudUser, SessionStore } from './types';

const TOKEN_KEY = 'cloud_token';
const USER_KEY = 'cloud_user';

export const sessionStore: SessionStore = {
  getToken: () => AsyncStorage.getItem(TOKEN_KEY),
  setToken: (token) => AsyncStorage.setItem(TOKEN_KEY, token),
  async getUser() {
    const raw = await AsyncStorage.getItem(USER_KEY);
    try { return raw ? (JSON.parse(raw) as CloudUser) : null; } catch { return null; }
  },
  setUser: (user) => AsyncStorage.setItem(USER_KEY, JSON.stringify(user)),
  clear: () => AsyncStorage.multiRemove([TOKEN_KEY, USER_KEY]),
};
