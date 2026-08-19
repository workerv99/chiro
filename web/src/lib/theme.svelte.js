const STORAGE_KEY = 'chiro_theme';
const validThemes = new Set(['light', 'dark', 'system']);

function initialTheme() {
  try {
    const saved = localStorage.getItem(STORAGE_KEY);
    return validThemes.has(saved) ? saved : 'system';
  } catch {
    return 'system';
  }
}

export const theme = $state({
  preference: initialTheme(),
  resolved: 'light'
});

function systemTheme() {
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

function applyTheme() {
  const resolved = theme.preference === 'system' ? systemTheme() : theme.preference;
  theme.resolved = resolved;
  document.documentElement.classList.toggle('dark', resolved === 'dark');
  document.documentElement.style.colorScheme = resolved;
}

export function initTheme() {
  if (typeof window === 'undefined') return;
  applyTheme();
  const media = window.matchMedia('(prefers-color-scheme: dark)');
  const onChange = () => {
    if (theme.preference === 'system') applyTheme();
  };
  media.addEventListener?.('change', onChange);
  return () => media.removeEventListener?.('change', onChange);
}

export function setThemePreference(preference) {
  if (!validThemes.has(preference)) return;
  theme.preference = preference;
  try {
    localStorage.setItem(STORAGE_KEY, preference);
  } catch { /* ignore */ }
  applyTheme();
}
