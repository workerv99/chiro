// Stores globales: sesión + datos por tabla + operaciones CRUD.
import { api, setToken, A } from './api.svelte.js';

export const S = $state({
  user: null,
  busy: false,
  error: '',
  lastSync: 0,
  db: {
    accounts: [],
    categories: [],
    persons: [],
    loans: [],
    installments: [],
    payments: [],
    budgets: [],
    piggy: [],
    bills: [],
    tags: [],
    expense_tag: []
  },
  summary: { income: 0, expense: 0, balance: 0 },
  monthExpenses: [],
  view: { year: 0, month: 0 }
});

const TABLE_PATHS = {
  accounts: '/api/accounts',
  categories: '/api/categories',
  persons: '/api/persons',
  payments: '/api/payments',
  budgets: '/api/budgets',
  piggy: '/api/piggy',
  bills: '/api/bills',
  tags: '/api/tags'
};

const PK = {
  accounts: 'account_id',
  categories: 'category_id',
  persons: 'person_id',
  payments: 'payment_id',
  budgets: 'budget_id',
  piggy: 'piggy_bank_id',
  bills: 'bill_id',
  tags: 'tag_id'
};

export async function login(email, password) {
  S.busy = true;
  S.error = '';
  try {
    const data = await api('/api/auth/login', { method: 'POST', body: { email, password } });
    setToken(data.token);
    S.user = data.user;
  } catch (e) {
    S.error = e.message;
    throw e;
  } finally {
    S.busy = false;
  }
}

export async function register(name, email, password) {
  S.busy = true;
  S.error = '';
  try {
    const data = await api('/api/auth/register', { method: 'POST', body: { name, email, password } });
    setToken(data.token);
    S.user = data.user;
  } catch (e) {
    S.error = e.message;
    throw e;
  } finally {
    S.busy = false;
  }
}

export function logout() {
  setToken('');
  S.user = null;
  for (const k of Object.keys(S.db)) S.db[k] = [];
  S.summary = { income: 0, expense: 0, balance: 0 };
  S.monthExpenses = [];
}

export async function me() {
  if (!A.token) return null;
  try {
    S.user = await api('/api/auth/me');
    return S.user;
  } catch {
    return null;
  }
}

export async function fetchAll() {
  const paths = { ...TABLE_PATHS, loans: '/api/loans' };
  const entries = await Promise.all(
    Object.entries(paths).map(async ([key, path]) => [key, await api(path)])
  );
  for (const [key, rows] of entries) S.db[key] = rows ?? [];
  S.lastSync = Date.now();
}

export async function loadMonth(year, month) {
  S.view.year = year;
  S.view.month = month;
  const [sum, exps] = await Promise.all([
    api(`/api/summary?year=${year}&month=${month}`),
    api(`/api/expenses?year=${year}&month=${month}`)
  ]);
  S.summary = sum ?? { income: 0, expense: 0, balance: 0 };
  S.monthExpenses = exps ?? [];
}

// Recarga el mes visible actual (tras crear/editar/borrar un movimiento).
export function refreshMonth() {
  if (S.view.year) return loadMonth(S.view.year, S.view.month);
  return Promise.resolve();
}

// ── CRUD genérico por tabla ───────────────────────────────────────────────────
export async function create(table, row) {
  const data = await api(TABLE_PATHS[table], { method: 'POST', body: row });
  S.db[table] = [data, ...S.db[table]];
  return data;
}

export async function update(table, id, row) {
  const data = await api(`${TABLE_PATHS[table]}/${id}`, { method: 'PUT', body: row });
  S.db[table] = S.db[table].map((r) => (r[PK[table]] === id ? data : r));
  return data;
}

export async function remove(table, id) {
  await api(`${TABLE_PATHS[table]}/${id}`, { method: 'DELETE' });
  S.db[table] = S.db[table].filter((r) => r[PK[table]] !== id);
}

export function resetError() {
  S.error = '';
}

// ── Transacciones ─────────────────────────────────────────────────────────────
export async function createExpense(row) {
  return api('/api/expenses', { method: 'POST', body: row });
}

export async function updateExpense(id, row) {
  return api(`/api/expenses/${id}`, { method: 'PUT', body: row });
}

export async function deleteExpense(id) {
  await api(`/api/expenses/${id}`, { method: 'DELETE' });
  S.monthExpenses = S.monthExpenses.filter((e) => e.expense_id !== id);
}

export async function createTransfer(row) {
  return api('/api/expenses/transfer', { method: 'POST', body: row });
}

// ── Préstamos ─────────────────────────────────────────────────────────────────
export async function createLoan(row) {
  const data = await api('/api/loans', { method: 'POST', body: row });
  await fetchAll();
  return data;
}

export async function updateLoan(loanId, row) {
  const data = await api(`/api/loans/${loanId}`, { method: 'PUT', body: row });
  await fetchAll();
  return data;
}

export async function loanInstallments(loanId) {
  return api(`/api/loans/${loanId}/installments`);
}

export async function payInstallment(id, { amount, date }) {
  return api(`/api/installments/${id}/pay`, { method: 'POST', body: { amount, date } });
}

export async function cascadeInstallment(id, { amount, date }) {
  return api(`/api/installments/${id}/cascade`, { method: 'POST', body: { amount, date } });
}

export async function unpayInstallment(id) {
  return api(`/api/installments/${id}/unpay`, { method: 'POST' });
}

// ── Personas ──────────────────────────────────────────────────────────────────
export function genId(prefix) {
  return `${prefix}_${Date.now().toString(36)}${Math.random().toString(36).slice(2, 7)}`;
}

// La API no expone POST/PUT de personas; se crean/editan vía /api/sync.
export async function savePerson(row) {
  if (!row.person_id) row.person_id = genId('per');
  await api('/api/sync', {
    method: 'POST',
    body: { tables: { person: [{ ...row, deleted: 0 }] }, since: 0 }
  });
  const data = await api('/api/persons');
  S.db.persons = data ?? [];
  return row;
}

// ── Facturas recurrentes ──────────────────────────────────────────────────────
export async function payBill(id) {
  const res = await api(`/api/bills/${id}/pay`, { method: 'POST', body: {} });
  refreshBills();
  return res;
}

export async function skipBill(id, nextDate, frequency) {
  const res = await api(`/api/bills/${id}/skip`, { method: 'POST', body: { next_date: nextDate, frequency } });
  refreshBills();
  return res;
}

export async function dueBills() {
  return api('/api/bills/due');
}

async function refreshBills() {
  try {
    const data = await api('/api/bills');
    S.db.bills = data ?? [];
  } catch { /* ignore */ }
}
