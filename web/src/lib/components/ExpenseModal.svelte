<script>
  import { i18n } from '$lib/i18n.svelte.js';
  import { S, createExpense, updateExpense, createTransfer, deleteExpense, refreshMonth } from '$lib/stores.svelte.js';
  import { toDisplay, toISO, todayISO } from '$lib/format.js';
  import ConfirmSheet from './ConfirmSheet.svelte';

  let { expense = null, onClose } = $props();

  let type = $state(expense?.transfer_pair_id ? 'transfer' : expense?.type || 'expense');
  let description = $state(expense?.description || '');
  let amount = $state(expense ? String(expense.amount) : '');
  let date = $state(expense ? toDisplay(expense.date) : toDisplay(todayISO()));
  let categoryId = $state(expense?.category_id || '');
  let accountId = $state(expense?.account_id || '');
  let destAccountId = $state(expense?.destination_account_id || '');
  let notes = $state(expense?.notes || '');
  let selectedTags = $state(expense?.tags || []);
  let saving = $state(false);
  let err = $state('');
  let confirmDel = $state(false);

  const cats = $derived(type === 'income' ? S.db.categories.filter((c) => c.type === 'income') : S.db.categories.filter((c) => c.type === 'expense'));

  function toggleTag(id) {
    if (selectedTags.includes(id)) selectedTags = selectedTags.filter((x) => x !== id);
    else selectedTags = [...selectedTags, id];
  }

  function focusInit(node) {
    node.focus();
  }

  function onKeydown(e) {
    if (e.key === 'Escape' && confirmDel) confirmDel = false;
    else if (e.key === 'Escape') onClose();
  }

  async function doDelete() {
    confirmDel = false;
    await deleteExpense(expense.expense_id);
    await refreshMonth();
    onClose();
  }

  async function save() {
    err = '';
    if (!description.trim()) return (err = i18n.t('expenses.descriptionRequired'));
    const amt = parseFloat(amount);
    if (!amt || amt <= 0) return (err = i18n.t('expenses.amountRequired'));
    const iso = toISO(date);
    if (!iso) return (err = i18n.t('expenses.dateRequired'));

    saving = true;
    try {
      if (type === 'transfer') {
        if (!accountId || !destAccountId) throw new Error(i18n.t('expenses.accountRequired'));
        await createTransfer({
          description,
          amount: amt,
          date: iso,
          account_id: accountId,
          destination_account_id: destAccountId,
          notes: notes || null
        });
      } else {
        const body = {
          description,
          amount: amt,
          date: iso,
          type,
          category_id: categoryId || null,
          account_id: accountId || null,
          notes: notes || null,
          tags: selectedTags
        };
        if (expense) await updateExpense(expense.expense_id, body);
        else await createExpense(body);
      }
      await refreshMonth();
      onClose();
    } catch (e) {
      err = e.message || i18n.t('expenses.createError');
    } finally {
      saving = false;
    }
  }

  async function removeExp() {
    confirmDel = true;
  }
</script>

<svelte:window onkeydown={onKeydown} />

<div class="overlay" onclick={onClose}>
  <div class="sheet" role="dialog" aria-modal="true" aria-labelledby="em-title" onclick={(e) => e.stopPropagation()}>
    <div class="handle"></div>
    <h2 class="title" id="em-title" style="margin-bottom:16px">
      {expense ? i18n.t('expenses.editExpense') : i18n.t('expenses.newExpense')}
    </h2>

    <div class="inline-flex" style="margin-bottom:14px;display:flex">
      <button class="chip" class:active={type === 'expense'} onclick={() => (type = 'expense')}>{i18n.t('common.expense')}</button>
      <button class="chip" class:active={type === 'income'} onclick={() => (type = 'income')}>{i18n.t('common.income')}</button>
      <button class="chip" class:active={type === 'transfer'} onclick={() => (type = 'transfer')}>{i18n.t('common.transfer')}</button>
    </div>

    <div class="form-field">
      <label for="em-desc">{i18n.t('expenses.description')}</label>
      <input id="em-desc" bind:value={description} placeholder={i18n.t('expenses.descriptionPlaceholder')} use:focusInit />
    </div>

    <div class="grid2">
      <div class="form-field">
        <label for="em-amount">{i18n.t('expenses.amount')}</label>
        <input id="em-amount" bind:value={amount} inputmode="decimal" placeholder={i18n.t('expenses.amountPlaceholder')} />
      </div>
      <div class="form-field">
        <label for="em-date">{i18n.t('expenses.date')}</label>
        <input id="em-date" bind:value={date} placeholder={i18n.t('expenses.datePlaceholder')} />
      </div>
    </div>

    {#if type !== 'transfer'}
      <div class="form-field">
        <label for="em-cat">{i18n.t('expenses.category')}</label>
        <select id="em-cat" bind:value={categoryId}>
          <option value="">{i18n.t('expenses.noCategory')}</option>
          {#each cats as c (c.category_id)}
            <option value={c.category_id}>{c.name}</option>
          {/each}
        </select>
      </div>
    {/if}

    {#if type === 'transfer'}
      <div class="grid2">
        <div class="form-field">
          <label for="em-from">{i18n.t('expenses.accountFrom')}</label>
          <select id="em-from" bind:value={accountId}>
            <option value="">—</option>
            {#each S.db.accounts as a (a.account_id)}
              <option value={a.account_id}>{a.name}</option>
            {/each}
          </select>
        </div>
        <div class="form-field">
          <label for="em-to">{i18n.t('expenses.accountTo')}</label>
          <select id="em-to" bind:value={destAccountId}>
            <option value="">—</option>
            {#each S.db.accounts as a (a.account_id)}
              <option value={a.account_id}>{a.name}</option>
            {/each}
          </select>
        </div>
      </div>
      <p class="meta">{i18n.t('expenses.transferHint')}</p>
    {:else}
      <div class="form-field">
        <label for="em-account">{i18n.t('expenses.account')}</label>
        <select id="em-account" bind:value={accountId}>
          <option value="">{i18n.t('expenses.noAccount')}</option>
          {#each S.db.accounts as a (a.account_id)}
            <option value={a.account_id}>{a.name}</option>
          {/each}
        </select>
      </div>
    {/if}

    <div class="form-field">
      <label for="em-notes">{i18n.t('expenses.notes')}</label>
      <textarea id="em-notes" bind:value={notes} rows="2" placeholder={i18n.t('expenses.notesPlaceholder')}></textarea>
    </div>

    {#if S.db.tags.length > 0}
      <div class="form-field">
        <label>{i18n.t('expenses.tags')}</label>
        <div class="inline-flex" style="flex-wrap:wrap">
          {#each S.db.tags as tg (tg.tag_id)}
            <button
              class="tag-chip"
              class:active={selectedTags.includes(tg.tag_id)}
              aria-pressed={selectedTags.includes(tg.tag_id)}
              style={selectedTags.includes(tg.tag_id) ? `color:${tg.color};border-color:${tg.color};background:${tg.color}22` : ''}
              onclick={() => toggleTag(tg.tag_id)}
            >{tg.name}</button>
          {/each}
        </div>
      </div>
    {/if}

    {#if err}
      <p class="error-text">{err}</p>
    {/if}

    <div class="inline-flex" style="margin-top:8px">
      <button class="btn btn-cancel" onclick={onClose}>{i18n.t('common.cancel')}</button>
      <button class="btn btn-primary" onclick={save} disabled={saving}>{i18n.t('common.save')}</button>
    </div>

    {#if expense}
      <button class="btn btn-danger" style="margin-top:12px" onclick={removeExp}>{i18n.t('common.delete')}</button>
    {/if}
  </div>
</div>

{#if confirmDel}
  <ConfirmSheet
    title={i18n.t('common.delete')}
    message={expense.transfer_pair_id ? i18n.t('expenses.deleteTransferConfirm') : i18n.t('expenses.deleteConfirm')}
    confirmLabel={i18n.t('common.delete')}
    danger
    onConfirm={doDelete}
    onCancel={() => (confirmDel = false)}
  />
{/if}
