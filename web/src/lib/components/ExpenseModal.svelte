<script>
  import { i18n } from '$lib/i18n.svelte.js';
  import { S, createExpense, updateExpense, createTransfer, deleteExpense, refreshMonth } from '$lib/stores.svelte.js';
  import { toDisplay, toISO, todayISO } from '$lib/format.js';
  import Dialog from '$lib/components/ui/dialog.svelte';
  import DialogHeader from '$lib/components/ui/dialog-header.svelte';
  import DialogTitle from '$lib/components/ui/dialog-title.svelte';
  import DialogFooter from '$lib/components/ui/dialog-footer.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Input from '$lib/components/ui/input.svelte';
  import Label from '$lib/components/ui/label.svelte';
  import Badge from '$lib/components/ui/badge.svelte';
  import ConfirmSheet from './ConfirmSheet.svelte';

  let { open = $bindable(false), expense = null, onClose } = $props();

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

  async function doDelete() {
    confirmDel = false;
    await deleteExpense(expense.expense_id);
    await refreshMonth();
    open = false;
    onClose?.();
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
      open = false;
      onClose?.();
    } catch (e) {
      err = e.message || i18n.t('expenses.createError');
    } finally {
      saving = false;
    }
  }
</script>

<Dialog bind:open>
  <DialogHeader>
    <DialogTitle>
      {expense ? i18n.t('expenses.editExpense') : i18n.t('expenses.newExpense')}
    </DialogTitle>
  </DialogHeader>

  <div class="flex gap-2 mb-4">
    <Button
      variant={type === 'expense' ? 'default' : 'outline'}
      size="sm"
      onclick={() => (type = 'expense')}
    >
      {i18n.t('common.expense')}
    </Button>
    <Button
      variant={type === 'income' ? 'default' : 'outline'}
      size="sm"
      onclick={() => (type = 'income')}
    >
      {i18n.t('common.income')}
    </Button>
    <Button
      variant={type === 'transfer' ? 'default' : 'outline'}
      size="sm"
      onclick={() => (type = 'transfer')}
    >
      {i18n.t('common.transfer')}
    </Button>
  </div>

  <div class="space-y-4">
    <div class="space-y-2">
      <Label for="em-desc">{i18n.t('expenses.description')}</Label>
      <Input id="em-desc" bind:value={description} placeholder={i18n.t('expenses.descriptionPlaceholder')} />
    </div>

    <div class="grid grid-cols-2 gap-4">
      <div class="space-y-2">
        <Label for="em-amount">{i18n.t('expenses.amount')}</Label>
        <Input id="em-amount" bind:value={amount} inputmode="decimal" placeholder={i18n.t('expenses.amountPlaceholder')} />
      </div>
      <div class="space-y-2">
        <Label for="em-date">{i18n.t('expenses.date')}</Label>
        <Input id="em-date" bind:value={date} placeholder={i18n.t('expenses.datePlaceholder')} />
      </div>
    </div>

    {#if type !== 'transfer'}
      <div class="space-y-2">
        <Label for="em-cat">{i18n.t('expenses.category')}</Label>
        <select
          id="em-cat"
          bind:value={categoryId}
          class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          <option value="">{i18n.t('expenses.noCategory')}</option>
          {#each cats as c (c.category_id)}
            <option value={c.category_id}>{c.name}</option>
          {/each}
        </select>
      </div>
    {/if}

    {#if type === 'transfer'}
      <div class="grid grid-cols-2 gap-4">
        <div class="space-y-2">
          <Label for="em-from">{i18n.t('expenses.accountFrom')}</Label>
          <select
            id="em-from"
            bind:value={accountId}
            class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          >
            <option value="">—</option>
            {#each S.db.accounts as a (a.account_id)}
              <option value={a.account_id}>{a.name}</option>
            {/each}
          </select>
        </div>
        <div class="space-y-2">
          <Label for="em-to">{i18n.t('expenses.accountTo')}</Label>
          <select
            id="em-to"
            bind:value={destAccountId}
            class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          >
            <option value="">—</option>
            {#each S.db.accounts as a (a.account_id)}
              <option value={a.account_id}>{a.name}</option>
            {/each}
          </select>
        </div>
      </div>
      <p class="text-sm text-muted-foreground">{i18n.t('expenses.transferHint')}</p>
    {:else}
      <div class="space-y-2">
        <Label for="em-account">{i18n.t('expenses.account')}</Label>
        <select
          id="em-account"
          bind:value={accountId}
          class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          <option value="">{i18n.t('expenses.noAccount')}</option>
          {#each S.db.accounts as a (a.account_id)}
            <option value={a.account_id}>{a.name}</option>
          {/each}
        </select>
      </div>
    {/if}

    <div class="space-y-2">
      <Label for="em-notes">{i18n.t('expenses.notes')}</Label>
      <textarea
        id="em-notes"
        bind:value={notes}
        rows="2"
        placeholder={i18n.t('expenses.notesPlaceholder')}
        class="flex min-h-[80px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
      ></textarea>
    </div>

    {#if S.db.tags.length > 0}
      <div class="space-y-2">
        <Label>{i18n.t('expenses.tags')}</Label>
        <div class="flex flex-wrap gap-2">
          {#each S.db.tags as tg (tg.tag_id)}
            <button
              onclick={() => toggleTag(tg.tag_id)}
              class="inline-flex items-center rounded-full border px-3 py-1 text-sm font-medium transition-colors hover:bg-accent"
              class:bg-primary={selectedTags.includes(tg.tag_id)}
              class:text-primary-foreground={selectedTags.includes(tg.tag_id)}
              class:border-primary={selectedTags.includes(tg.tag_id)}
              style={!selectedTags.includes(tg.tag_id) ? `color:${tg.color};border-color:${tg.color}` : ''}
            >
              {tg.name}
            </button>
          {/each}
        </div>
      </div>
    {/if}

    {#if err}
      <p class="text-sm text-destructive">{err}</p>
    {/if}
  </div>

  <DialogFooter class="mt-4">
    <Button variant="outline" onclick={() => { open = false; onClose?.(); }}>
      {i18n.t('common.cancel')}
    </Button>
    <Button onclick={save} disabled={saving}>
      {saving ? 'Guardando...' : i18n.t('common.save')}
    </Button>
  </DialogFooter>

  {#if expense}
    <div class="mt-2">
      <Button variant="destructive" class="w-full" onclick={() => (confirmDel = true)}>
        {i18n.t('common.delete')}
      </Button>
    </div>
  {/if}
</Dialog>

{#if confirmDel}
  <ConfirmSheet
    bind:open={() => confirmDel, (v) => confirmDel = v}
    title={i18n.t('common.delete')}
    message={expense.transfer_pair_id ? i18n.t('expenses.deleteTransferConfirm') : i18n.t('expenses.deleteConfirm')}
    confirmLabel={i18n.t('common.delete')}
    danger
    onConfirm={doDelete}
    onCancel={() => (confirmDel = false)}
  />
{/if}
