<script>
  import { i18n } from '$lib/i18n.svelte.js';
  import { api } from '$lib/api.svelte.js';
  import { S, create, update, remove } from '$lib/stores.svelte.js';
  import { money, pct } from '$lib/format.js';
  import ConfirmSheet from '$lib/components/ConfirmSheet.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Card from '$lib/components/ui/card.svelte';
  import Input from '$lib/components/ui/input.svelte';
  import Label from '$lib/components/ui/label.svelte';
  import Dialog from '$lib/components/ui/dialog.svelte';
  import DialogHeader from '$lib/components/ui/dialog-header.svelte';
  import DialogTitle from '$lib/components/ui/dialog-title.svelte';
  import DialogFooter from '$lib/components/ui/dialog-footer.svelte';
  import { ChevronLeft, ChevronRight, Plus } from 'lucide-svelte';

  const now = new Date();
  let year = $state(now.getFullYear());
  let month = $state(now.getMonth() + 1);
  let list = $state([]);
  let showForm = $state(false);
  let editing = $state(null);
  let catId = $state('');
  let amount = $state('');
  let err = $state('');
  let confirmDel = $state(null);

  function load() {
    return api(`/api/budgets/progress?year=${year}&month=${month}`)
      .then((d) => (list = d ?? []))
      .catch(() => {});
  }

  $effect(() => { load(); });

  function shift(delta) {
    let m = month + delta;
    let y = year;
    if (m < 1) { m = 12; y -= 1; }
    if (m > 12) { m = 1; y += 1; }
    year = y;
    month = m;
  }

  function openNew() {
    editing = null;
    amount = '';
    catId = '';
    err = '';
    showForm = true;
  }

  function openEdit(b) {
    editing = b;
    amount = String(b.amount ?? '');
    catId = b.category_id || '';
    err = '';
    showForm = true;
  }

  async function save() {
    err = '';
    const amt = parseFloat(amount);
    if (!amt || amt <= 0) return (err = i18n.t('common.required'));
    try {
      if (editing) await update('budgets', editing.budget_id, { amount: amt, category_id: catId });
      else await create('budgets', { amount: amt, category_id: catId, month, year });
      showForm = false;
      await load();
    } catch (e) {
      err = e.message;
    }
  }

  async function doDelete() {
    if (!confirmDel) return;
    await remove('budgets', confirmDel.budget_id);
    confirmDel = null;
    await load();
  }
</script>

<svelte:head><title>{i18n.t('budgets.title')} · Chiro</title></svelte:head>

<div class="flex items-center justify-between mb-4">
  <h1 class="text-2xl font-bold">{i18n.t('budgets.title')}</h1>
</div>

<div class="flex items-center justify-between bg-card rounded-lg border p-2 mb-4">
  <Button variant="ghost" size="icon" onclick={() => shift(-1)} aria-label={i18n.t('common.prevMonth')}>
    <ChevronLeft class="h-5 w-5" />
  </Button>
  <span class="text-sm font-bold">{new Date(year, month - 1).toLocaleDateString(i18n.lang === 'en' ? 'en' : 'es', { month: 'long', year: 'numeric' })}</span>
  <Button variant="ghost" size="icon" onclick={() => shift(1)} aria-label={i18n.t('common.nextMonth')}>
    <ChevronRight class="h-5 w-5" />
  </Button>
</div>

{#if list.length === 0}
  <Card class="flex flex-col items-center gap-3 py-8">
    <p class="text-sm text-muted-foreground">{i18n.t('budgets.empty')}</p>
    <Button onclick={openNew}>{i18n.t('budgets.newBudget')}</Button>
  </Card>
{:else}
  <Card class="overflow-hidden">
    {#each list as b (b.budget_id)}
      {@const progress = b.amount > 0 ? Math.min(100, (b.spent / b.amount) * 100) : 0}
      {@const exceeded = b.spent > b.amount}
      <button class="w-full text-left p-4 hover:bg-muted/50 transition-colors border-t first:border-t-0" onclick={() => openEdit(b)}>
        <div class="flex justify-between items-center mb-2">
          <span class="text-sm font-semibold">{b.category_name}</span>
          <span class="text-sm font-bold" class:text-destructive={exceeded}>{money(b.spent)} / {money(b.amount)}</span>
        </div>
        <div class="h-2 bg-border rounded-full overflow-hidden">
          <div class="h-full rounded-full transition-all" class:bg-destructive={exceeded} class:bg-primary={!exceeded} style="width:{progress}%"></div>
        </div>
      </button>
    {/each}
  </Card>
{/if}

<Button
  size="icon"
  class="fixed right-5 bottom-20 h-14 w-14 rounded-2xl shadow-lg z-40"
  onclick={openNew}
  aria-label={i18n.t('budgets.newBudget')}
>
  <Plus class="h-6 w-6" />
</Button>

<Dialog bind:open={showForm}>
  <DialogHeader>
    <DialogTitle>{editing ? i18n.t('budgets.editBudget') : i18n.t('budgets.newBudget')}</DialogTitle>
  </DialogHeader>

  <div class="space-y-4">
    <div class="space-y-2">
      <Label for="bud-cat">{i18n.t('budgets.category')}</Label>
      <select id="bud-cat" bind:value={catId} class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
        <option value="">{i18n.t('budgets.allCategories')}</option>
        {#each S.db.categories as c (c.category_id)}
          <option value={c.category_id}>{c.name}</option>
        {/each}
      </select>
    </div>

    <div class="space-y-2">
      <Label for="bud-amount">{i18n.t('budgets.amount')}</Label>
      <Input id="bud-amount" bind:value={amount} inputmode="decimal" />
    </div>

    {#if err}
      <p class="text-sm text-destructive">{err}</p>
    {/if}
  </div>

  <DialogFooter>
    <Button variant="outline" onclick={() => (showForm = false)}>{i18n.t('common.cancel')}</Button>
    <Button onclick={save}>{i18n.t('common.save')}</Button>
  </DialogFooter>
</Dialog>

{#if confirmDel}
  <ConfirmSheet
    bind:open={() => confirmDel !== null, (v) => confirmDel = v ? confirmDel : null}
    title={i18n.t('common.delete')}
    message={i18n.t('budgets.deleteConfirm')}
    confirmLabel={i18n.t('common.delete')}
    danger
    onConfirm={doDelete}
    onCancel={() => (confirmDel = null)}
  />
{/if}