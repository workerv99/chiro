import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('./api.svelte.js', () => ({
  api: vi.fn(),
  BASE: '',
  setToken: vi.fn(),
  A: { token: '' }
}));

import { api } from './api.svelte.js';
import { S, remove } from './stores.svelte.js';

// Every table the generic store operations accept, with the field that holds its id.
// A table missing from the store's PK map makes remove() silently keep the row.
const TABLES = [
  ['accounts', 'account_id'],
  ['categories', 'category_id'],
  ['persons', 'person_id'],
  ['payments', 'payment_id'],
  ['budgets', 'budget_id'],
  ['piggy', 'piggy_bank_id'],
  ['bills', 'bill_id'],
  ['tags', 'tag_id'],
  ['loans', 'loan_id']
];

describe('remove', () => {
  beforeEach(() => {
    api.mockReset();
  });

  it.each(TABLES)('drops the deleted %s row by %s', async (table, pk) => {
    S.db[table] = [{ [pk]: 'a' }, { [pk]: 'b' }];
    api.mockResolvedValue(null);

    await remove(table, 'a');

    expect(S.db[table]).toEqual([{ [pk]: 'b' }]);
    expect(api).toHaveBeenCalledTimes(1);
    expect(api.mock.calls[0][0]).toMatch(/\/a$/);
    expect(api.mock.calls[0][1]).toEqual({ method: 'DELETE' });
  });

  it('keeps the row and rethrows when the server rejects the delete', async () => {
    S.db.loans = [{ loan_id: 'a' }, { loan_id: 'b' }];
    api.mockRejectedValue(new Error('Error 500'));

    await expect(remove('loans', 'a')).rejects.toThrow('Error 500');

    expect(S.db.loans).toEqual([{ loan_id: 'a' }, { loan_id: 'b' }]);
  });
});
