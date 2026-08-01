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
