<script>
  import { i18n } from '$lib/i18n.svelte.js';
  import { S, loadMonth } from '$lib/stores.svelte.js';
  import { api, A } from '$lib/api.svelte.js';
  import { money, signed, colorOf, monthLabel, toDisplay } from '$lib/format.js';
  import ExpenseModal from '$lib/components/ExpenseModal.svelte';

  const now = new Date();
  let year = $state(now.getFullYear());
  let month = $state(now.getMonth() + 1);
  let showModal = $state(false);
  let loading = $state(true);
  let prevBalance = $state(null);

  $effect(() => {
    loading = true;
    const py = month === 1 ? year - 1 : year;
    const pm = month === 1 ? 12 : month - 1;
    Promise.all([
      loadMonth(year, month),
      A.token ? api(`/api/summary?year=${py}&month=${pm}`).catch(() => null) : Promise.resolve(null)
    ]).then(([, prev]) => {
      prevBalance = prev ? prev.balance : null;
    }).finally(() => (loading = false));
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

  const delta = $derived(prevBalance != null ? S.summary.balance - prevBalance : null);
</script>

<svelte:head><title>{i18n.t('expenses.title')} · Chiro</title></svelte:head>

<div class="page-head">
  <div>
    <h1 class="headline">{i18n.t('expenses.title')}</h1>
  </div>
</div>

{#if S.db.accounts.length === 0 && S.db.categories.length === 0}
  <div class="card onboarding">
    <h2 class="title">{i18n.t('onboard.title')}</h2>
    <p class="meta">{i18n.t('onboard.sub')}</p>
    <ol class="onboard-steps">
      <li>
        <span class="step-num">1</span>
        <div>
          <strong>{i18n.t('onboard.step1Title')}</strong>
          <p class="meta">{i18n.t('onboard.step1Sub')}</p>
        </div>
        <a class="btn btn-small" href="/config">{i18n.t('onboard.go')}</a>
      </li>
      <li>
        <span class="step-num">2</span>
        <div>
          <strong>{i18n.t('onboard.step2Title')}</strong>
          <p class="meta">{i18n.t('onboard.step2Sub')}</p>
        </div>
      </li>
      <li>
        <span class="step-num">3</span>
        <div>
          <strong>{i18n.t('onboard.step3Title')}</strong>
          <p class="meta">{i18n.t('onboard.step3Sub')}</p>
        </div>
        <button class="btn btn-small btn-primary" onclick={() => (showModal = true)}>{i18n.t('onboard.go')}</button>
      </li>
    </ol>
  </div>
{/if}

<div class="month-nav">
  <button class="icon-btn" onclick={() => shift(-1)} aria-label={i18n.t('common.prevMonth')}>‹</button>
  <button class="month-label" onclick={today} title={i18n.t('common.goToday')}>{monthLabel(year, month)}</button>
  <button class="icon-btn" onclick={() => shift(1)} aria-label={i18n.t('common.nextMonth')}>›</button>
</div>

<div class="card balance-hero">
  <div class="balance-eyebrow">{i18n.t('summary.balance')}</div>
  <div class="balance-amount" class:negative={S.summary.balance < 0} aria-label={`${i18n.t('summary.balance')}: ${money(S.summary.balance)}`}>{signed(S.summary.balance)}</div>
  {#if delta != null}
    <div class="balance-delta" class:positive={delta > 0} class:negative={delta < 0}>
      {delta >= 0 ? '↑' : '↓'} {signed(Math.abs(delta))} {i18n.t('summary.vsLastMonth')}
    </div>
  {/if}
</div>

<div class="summary-secondary">
  <div class="card stat-card">
    <div class="stat-label">{i18n.t('summary.income')}</div>
    <div class="stat-value positive" aria-label={`${i18n.t('summary.income')}: ${money(S.summary.income)}`}>{signed(S.summary.income)}</div>
  </div>
  <div class="card stat-card">
    <div class="stat-label">{i18n.t('summary.expense')}</div>
    <div class="stat-value negative" aria-label={`${i18n.t('summary.expense')}: ${money(S.summary.expense)}`}>{money(S.summary.expense)}</div>
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
