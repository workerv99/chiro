// Formato de dinero y fechas. Las fechas se almacenan como YYYY-MM-DD en el
// backend y se muestran como DD/MM/YYYY en la UI (port de las pantallas).
export function money(amount, currency = 'USD') {
  const n = Number(amount) || 0;
  try {
    return new Intl.NumberFormat(undefined, {
      style: 'currency',
      currency,
      minimumFractionDigits: 2,
      maximumFractionDigits: 2
    }).format(n);
  } catch {
    return n.toFixed(2);
  }
}

export function signed(amount, currency) {
  const n = Number(amount) || 0;
  const prefix = n > 0 ? '+' : '';
  return prefix + money(n, currency);
}

// YYYY-MM-DD -> DD/MM/YYYY
export function toDisplay(iso) {
  if (!iso) return '';
  const [y, m, d] = iso.split('-');
  if (!y || !m || !d) return iso;
  return `${d}/${m}/${y}`;
}

// YYYY-MM-DD (or a full ISO timestamp) -> { year, month, day } as zero-padded
// strings. Splits the string instead of using new Date(), which parses
// date-only strings as UTC and shifts the day in negative-offset timezones.
export function isoParts(iso) {
  const m = /^(\d{4})-(\d{2})-(\d{2})/.exec(String(iso || '').slice(0, 10));
  if (!m) return { year: '', month: '', day: '' };
  return { year: m[1], month: m[2], day: m[3] };
}

// DD/MM/YYYY -> YYYY-MM-DD
export function toISO(display) {
  if (!display) return '';
  const parts = display.trim().split('/');
  if (parts.length !== 3) return display;
  const [d, m, y] = parts;
  const dd = d.padStart(2, '0');
  const mm = m.padStart(2, '0');
  if (!/^\d{4}$/.test(y)) return display;
  return `${y}-${mm}-${dd}`;
}

export function todayISO() {
  const now = new Date();
  const mm = String(now.getMonth() + 1).padStart(2, '0');
  const dd = String(now.getDate()).padStart(2, '0');
  return `${now.getFullYear()}-${mm}-${dd}`;
}

export function monthLabel(year, month) {
  const mm = String(month).padStart(2, '0');
  return `${year}-${mm}`;
}

// color de categoría con fallback
export function colorOf(cat) {
  return (cat && cat.color) || '#5B7CF6';
}

export function initials(name) {
  if (!name) return '?';
  const parts = name.trim().split(/\s+/);
  return parts[0][0] + (parts.length > 1 ? parts[1][0] : '');
}

export function pct(spent, limit) {
  if (!limit || limit <= 0) return 0;
  return Math.min(100, Math.round((spent / limit) * 100));
}

// Parses user-typed decimals; accepts both "12.5" and "12,5" (comma keypads).
export function parseDecimal(value) {
  const n = parseFloat(String(value ?? "").trim().replace(",", "."));
  return Number.isFinite(n) ? n : NaN;
}

// Friendly "Browser on OS" label from a User-Agent string. Returns '' when the
// agent is empty or unrecognised so callers can show a localized fallback.
export function deviceLabel(ua) {
  const s = String(ua || '');
  if (!s) return '';
  let os = '';
  if (/Android/i.test(s)) os = 'Android';
  else if (/iPhone|iPad|iPod/i.test(s)) os = 'iOS';
  else if (/Windows/i.test(s)) os = 'Windows';
  else if (/Mac OS X|Macintosh/i.test(s)) os = 'macOS';
  else if (/CrOS/i.test(s)) os = 'ChromeOS';
  else if (/Linux|X11/i.test(s)) os = 'Linux';
  let browser = '';
  if (/Edg(e|A|iOS)?\//i.test(s)) browser = 'Edge';
  else if (/OPR\/|Opera/i.test(s)) browser = 'Opera';
  else if (/SamsungBrowser/i.test(s)) browser = 'Samsung Internet';
  else if (/Firefox\/|FxiOS/i.test(s)) browser = 'Firefox';
  else if (/Chrome\/|CriOS/i.test(s)) browser = 'Chrome';
  else if (/Safari\//i.test(s)) browser = 'Safari';
  if (browser && os) return browser + ' on ' + os;
  return browser || os;
}

export function isMobileAgent(ua) {
  return /Android|iPhone|iPad|iPod|Mobile/i.test(String(ua || ''));
}
