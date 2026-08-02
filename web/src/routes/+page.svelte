<script>
  import { i18n } from '$lib/i18n.svelte.js';
  import { S, loadMonth } from '$lib/stores.svelte.js';
  import { money, signed, colorOf, monthLabel, toDisplay } from '$lib/format.js';
  import ExpenseModal from '$lib/components/ExpenseModal.svelte';

  const now = new Date();
  let year = $state(now.getFullYear());
  let month = $state(now.getMonth() + 1);
  let showModal = $state(false);
  let loading = $state(true);

  $effect(() => {
    loading = true;
    loadMonth(year, month).finally(() => (loading = false));
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

  const grouped = $derived(() => {
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

  function catOf(id) {
    return S.db.categories.find((c) => c.category_id === id);
  }
</script>

<svelte:head><title>{i18n.t('expenses.title')} · Chiro</title></svelte:head>

<div class="page-head">
  <div>
    <h1 class="headline">{i18n.t('expenses.title')}</h1>
  </div>
</div>

<div class="month-nav">
  <button class="icon-btn" onclick={() => shift(-1)} aria-label={i18n.t('common.prevMonth')}>‹</button>
  <button class="month-label" onclick={today} title={i18n.t('common.goToday')}>{monthLabel(year, month)}</button>
  <button class="icon-btn" onclick={() => shift(1)} aria-label={i18n.t('common.nextMonth')}>›</button>
</div>

<div class="summary-cards">
  <div class="card stat-card">
    <div class="stat-label">{i18n.t('summary.income')}</div>
    <div class="stat-value positive">{signed(S.summary.income)}</div>
  </div>
  <div class="card stat-card">
    <div class="stat-label">{i18n.t('summary.expense')}</div>
    <div class="stat-value negative">{money(S.summary.expense)}</div>
  </div>
  <div class="card stat-card">
    <div class="stat-label">{i18n.t('summary.balance')}</div>
    <div class="stat-value" class:negative={S.summary.balance < 0}>{signed(S.summary.balance)}</div>
  </div>
</div>

<div class="card list-card">
  {#if loading}
    <p class="meta" style="padding:24px;text-align:center">{i18n.t('common.loading')}</p>
  {:else if grouped().length === 0}
    <div class="empty">
      <p class="meta">{i18n.t('expenses.empty')}</p>
      <button class="btn btn-primary" onclick={() => (showModal = true)}>{i18n.t('expenses.newExpense')}</button>
    </div>
  {:else}
    {#each grouped() as g (g.date)}
      <div class="day-group">
        <div class="day-head">
          <span>{toDisplay(g.date)}</span>
          <span class="day-total" class:negative={g.total < 0}>{signed(g.total)}</span>
        </div>
        {#each g.rows as e (e.expense_id)}
          <a class="row" href={`/expense/${e.expense_id}`}>
            <div class="cat-dot" style="background:{colorOf(catOf(e.category_id))}"></div>
            <div class="row-body">
              <div class="row-title">{e.description}</div>
              <div class="row-sub">
                {#if e.transfer_pair_id}
                  <span class="tag mini">{i18n.t('common.transfer')}</span>
                {/if}
                {#if catOf(e.category_id)}
                  <span>{catOf(e.category_id).name}</span>
                {/if}
                {#each e.tags || [] as tg (tg.tag_id)}
                  <span class="tag mini">{tg.name}</span>
                {/each}
              </div>
            </div>
            <div class="row-amount" class:negative={e.type === 'expense'} class:positive={e.type === 'income'}>
              {signed(e.type === 'expense' ? -e.amount : e.amount)}
            </div>
          </a>
        {/each}
      </div>
    {/each}
  {/if}
</div>

{#if showModal}
  <ExpenseModal onClose={() => (showModal = false)} />
{/if}

<button class="fab" onclick={() => (showModal = true)} aria-label={i18n.t('expenses.newExpense')}>+</button>
