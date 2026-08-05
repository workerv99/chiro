<script>
  import { i18n } from '$lib/i18n.svelte.js';
  import { S, loadMonth, create } from '$lib/stores.svelte.js';
  import { api, A } from '$lib/api.svelte.js';
  import { money, signed, colorOf, monthLabel, toDisplay, todayISO } from '$lib/format.js';
  import ExpenseModal from '$lib/components/ExpenseModal.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Card from '$lib/components/ui/card.svelte';
  import { ChevronLeft, ChevronRight, Plus } from 'lucide-svelte';

  const now = new Date();
  let year = $state(now.getFullYear());
  let month = $state(now.getMonth() + 1);
  let showModal = $state(false);
  let loading = $state(true);
  let prevBalance = $state(null);
  let abortCtrl = null;
  let onboardStep = $state(0);
  let showOnboard = $state(false);

  $effect(() => {
    if (S.db.accounts.length === 0 && S.db.categories.length === 0) {
      showOnboard = true;
      onboardStep = 1;
    }
  });

  $effect(() => {
    if (abortCtrl) abortCtrl.abort();
    abortCtrl = new AbortController();
    const signal = abortCtrl.signal;
    loading = true;
    const py = month === 1 ? year - 1 : year;
    const pm = month === 1 ? 12 : month - 1;
    Promise.all([
      loadMonth(year, month),
      A.token ? api(`/api/summary?year=${py}&month=${pm}`, { signal }).catch(() => null) : Promise.resolve(null)
    ]).then(([, prev]) => {
      if (!signal.aborted) prevBalance = prev ? prev.balance : null;
    }).finally(() => { if (!signal.aborted) loading = false; });
  });

  function shift(delta) {
    let m = month + delta;
    let y = year;
    if (m < 1) { m = 12; y -= 1; }
    if (m > 12) { m = 1; y += 1; }
    year = y;
    month = m;
  }

  function today() {
    const d = new Date();
    year = d.getFullYear();
    month = d.getMonth() + 1;
  }

  async function createDefaultAccount() {
    await create('accounts', { name: 'Efectivo', currency: 'USD', account_type: 'asset' });
    await create('accounts', { name: 'Banco', currency: 'USD', account_type: 'asset' });
    onboardStep = 2;
  }

  async function createDefaultCategories() {
    const cats = [
      { name: 'Alimentación', color: '#FF6B6B', type: 'expense' },
      { name: 'Transporte', color: '#4ECDC4', type: 'expense' },
      { name: 'Salario', color: '#27AE60', type: 'income' }
    ];
    for (const c of cats) {
      await create('categories', c);
    }
    showOnboard = false;
    onboardStep = 0;
  }

  const grouped = $derived.by(() => {
    const byDate = {};
    for (const e of S.monthExpenses) {
      (byDate[e.date] ||= []).push(e);
    }
    return Object.entries(byDate)
      .map(([date, rows]) => ({
        date,
        total: rows.reduce((s, r) => s + (r.type === 'expense' ? -r.amount : r.amount), 0),
        rows
      }))
      .sort((a, b) => (a.date < b.date ? 1 : -1));
  });

  const catMap = $derived.by(() => {
    const map = {};
    for (const c of S.db.categories) {
      map[c.category_id] = c;
    }
    return map;
  });

  function catOf(id) {
    return catMap[id];
  }

  const delta = $derived(prevBalance != null ? S.summary.balance - prevBalance : null);
</script>

<svelte:head><title>{i18n.t('expenses.title')} · Chiro</title></svelte:head>

<div class="flex items-center justify-between mb-4">
  <h1 class="text-2xl font-bold">{i18n.t('expenses.title')}</h1>
</div>

{#if showOnboard}
  <Card class="p-6 mb-4 bg-gradient-to-br from-primary/10 to-transparent border-primary/20">
    <h2 class="text-lg font-bold mb-1">Bienvenido a Chiro</h2>
    <p class="text-sm text-muted-foreground mb-4">Configurá tu cuenta en 2 pasos rápidos.</p>

    {#if onboardStep === 1}
      <div class="flex items-start gap-3 mb-4">
        <span class="flex h-8 w-8 items-center justify-center rounded-full bg-primary text-primary-foreground text-sm font-bold">1</span>
        <div>
          <p class="font-semibold">Crear cuentas</p>
          <p class="text-sm text-muted-foreground">Efectivo y Banco para empezar</p>
        </div>
      </div>
      <Button class="w-full" onclick={createDefaultAccount}>Crear cuentas por defecto</Button>
      <Button variant="ghost" class="w-full mt-2" onclick={() => { showOnboard = false; onboardStep = 0; }}>
        Saltar por ahora
      </Button>
    {:else if onboardStep === 2}
      <div class="flex items-start gap-3 mb-4">
        <span class="flex h-8 w-8 items-center justify-center rounded-full bg-primary text-primary-foreground text-sm font-bold">2</span>
        <div>
          <p class="font-semibold">Crear categorías</p>
          <p class="text-sm text-muted-foreground">Alimentación, Transporte y Salario</p>
        </div>
      </div>
      <Button class="w-full" onclick={createDefaultCategories}>Crear categorías por defecto</Button>
      <Button variant="ghost" class="w-full mt-2" onclick={() => { showOnboard = false; onboardStep = 0; }}>
        Configurar manualmente
      </Button>
    {/if}
  </Card>
{/if}

<div class="flex items-center justify-between bg-card rounded-lg border p-2 mb-4">
  <Button variant="ghost" size="icon" onclick={() => shift(-1)} aria-label={i18n.t('common.prevMonth')}>
    <ChevronLeft class="h-5 w-5" />
  </Button>
  <button class="text-sm font-bold" onclick={today} title={i18n.t('common.goToday')}>
    {monthLabel(year, month)}
  </button>
  <Button variant="ghost" size="icon" onclick={() => shift(1)} aria-label={i18n.t('common.nextMonth')}>
    <ChevronRight class="h-5 w-5" />
  </Button>
</div>

<Card class="p-6 mb-4 bg-gradient-to-br from-primary/10 to-transparent border-primary/20">
  <p class="text-xs font-bold text-muted-foreground uppercase tracking-wider mb-1">{i18n.t('summary.balance')}</p>
  <p class="text-3xl font-extrabold" class:text-destructive={S.summary.balance < 0} class:text-green-500={S.summary.balance >= 0}>
    {signed(S.summary.balance)}
  </p>
  {#if delta != null}
    <p class="text-sm font-bold mt-1" class:text-green-500={delta > 0} class:text-destructive={delta < 0}>
      {delta >= 0 ? '↑' : '↓'} {signed(Math.abs(delta))} {i18n.t('summary.vsLastMonth')}
    </p>
  {/if}
</Card>

<div class="grid grid-cols-2 gap-3 mb-4">
  <Card class="p-4">
    <p class="text-xs font-bold text-muted-foreground uppercase tracking-wider mb-1">{i18n.t('summary.income')}</p>
    <p class="text-lg font-bold text-green-500">{signed(S.summary.income)}</p>
  </Card>
  <Card class="p-4">
    <p class="text-xs font-bold text-muted-foreground uppercase tracking-wider mb-1">{i18n.t('summary.expense')}</p>
    <p class="text-lg font-bold text-destructive">{money(S.summary.expense)}</p>
  </Card>
</div>

<Card class="overflow-hidden">
  {#if loading}
    <p class="text-sm text-muted-foreground py-8 text-center">{i18n.t('common.loading')}</p>
  {:else if grouped().length === 0}
    <div class="flex flex-col items-center gap-3 py-8">
      <p class="text-sm text-muted-foreground">{i18n.t('expenses.empty')}</p>
      <Button onclick={() => (showModal = true)}>{i18n.t('expenses.newExpense')}</Button>
    </div>
  {:else}
    {#each grouped as g (g.date)}
      <div class="border-t first:border-t-0">
        <div class="flex justify-between items-center px-4 py-2 bg-muted/50">
          <span class="text-xs font-bold text-muted-foreground uppercase">{toDisplay(g.date)}</span>
          <span class="text-xs font-bold" class:text-green-500={g.total >= 0} class:text-destructive={g.total < 0}>
            {signed(g.total)}
          </span>
        </div>
        {#each g.rows as e (e.expense_id)}
          <a class="flex items-center gap-3 px-4 py-3 hover:bg-muted/50 transition-colors border-t first:border-t-0" href={`/expense/${e.expense_id}`}>
            <div class="h-2.5 w-2.5 rounded-sm" style="background:{colorOf(catOf(e.category_id))}"></div>
            <div class="flex-1 min-w-0">
              <p class="text-sm font-semibold truncate">{e.description}</p>
              <p class="text-xs text-muted-foreground flex gap-1.5 items-center flex-wrap">
                {#if e.transfer_pair_id}
                  <span class="inline-flex items-center rounded-full bg-primary/10 px-2 py-0.5 text-[10px] font-bold text-primary">{i18n.t('common.transfer')}</span>
                {/if}
                {#if catOf(e.category_id)}
                  <span>{catOf(e.category_id).name}</span>
                {/if}
                {#each e.tags || [] as tg (tg.tag_id)}
                  <span class="inline-flex items-center rounded-full bg-primary/10 px-2 py-0.5 text-[10px] font-bold text-primary">{tg.name}</span>
                {/each}
              </p>
            </div>
            <p class="text-sm font-bold" class:text-green-500={e.type === 'income'} class:text-destructive={e.type === 'expense'}>
              {signed(e.type === 'expense' ? -e.amount : e.amount)}
            </p>
          </a>
        {/each}
      </div>
    {/each}
  {/if}
</Card>

{#if showModal}
  <ExpenseModal bind:open={showModal} onClose={() => (showModal = false)} />
{/if}

<Button
  size="icon"
  class="fixed right-5 bottom-20 h-14 w-14 rounded-2xl shadow-lg z-40"
  onclick={() => (showModal = true)}
  aria-label={i18n.t('expenses.newExpense')}
>
  <Plus class="h-6 w-6" />
</Button>
