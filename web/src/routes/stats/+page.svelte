<script>
  import { i18n } from '$lib/i18n.svelte.js';
  import { api } from '$lib/api.svelte.js';
  import { S, loadMonth } from '$lib/stores.svelte.js';
  import { money } from '$lib/format.js';
  import { ChevronLeft, ChevronRight } from 'lucide-svelte';

  const now = new Date();
  let year = $state(now.getFullYear());
  let month = $state(now.getMonth() + 1);
  let stats = $state({ months: [], breakdown: [], outstanding: 0 });
  let loading = $state(true);

  $effect(() => {
    loading = true;
    Promise.all([loadMonth(year, month), api(`/api/stats?year=${year}&month=${month}`)])
      .then(([, s]) => {
        stats = s;
      })
      .finally(() => (loading = false));
  });

  const breakdown = $derived((stats.breakdown || []).filter((b) => b.type === 'expense'));
  const maxCat = $derived(Math.max(1, ...breakdown.map((c) => c.total)));
  const months = $derived(stats.months || []);
  const maxMonth = $derived(Math.max(1, ...months.map((m) => m.expense)));
  const monthNames = $derived(
    Array.from({ length: 12 }, (_, i) => new Date(year, i, 1).toLocaleDateString(i18n.lang === 'en' ? 'en' : 'es', { month: 'short' }))
  );
</script>

<svelte:head><title>{i18n.t('stats.title')} · Chiro</title></svelte:head>

<div class="page-head">
  <h1 class="headline">{i18n.t('stats.title')}</h1>
</div>

  <div class="month-nav">
    <button class="icon-btn" onclick={() => (month = month === 1 ? (year -= 1, 12) : month - 1)} aria-label={i18n.t('common.prevMonth')}><ChevronLeft size={20} /></button>
    <span class="month-label">{monthLabel(year, month)}</span>
    <button class="icon-btn" onclick={() => (month = month === 12 ? (year += 1, 1) : month + 1)} aria-label={i18n.t('common.nextMonth')}><ChevronRight size={20} /></button>
  </div>

{#if loading}
  <p class="meta" style="padding:24px;text-align:center">{i18n.t('common.loading')}</p>
{:else}
  <div class="card" style="padding:16px">
    <h3 class="card-title">{i18n.t('stats.byCategory')}</h3>
    {#if breakdown.length === 0}
      <p class="meta" style="padding:12px 0;text-align:center">{i18n.t('stats.empty')}</p>
    {:else}
      {#each breakdown as c (c.category_name)}
        <div class="bar-row">
          <div class="bar-label">{c.category_name}</div>
          <div class="bar-track">
            <div class="bar-fill" style="transform:scaleX({Math.max(0.04, c.total / maxCat)});background:{c.category_color}"></div>
          </div>
          <div class="bar-value">{money(c.total)}</div>
        </div>
      {/each}
    {/if}
  </div>

  <div class="card" style="padding:16px;margin-top:14px">
    <h3 class="card-title">{i18n.t('stats.monthly')}</h3>
    <div class="bars12">
      {#each months as m (m.month)}
        <div class="col" title={monthNames[m.month - 1]}>
          <div class="col-bar" style="height:{m.expense > 0 ? Math.max(4, (m.expense / maxMonth) * 100) : 2}%"></div>
          <div class="col-label">{monthNames[m.month - 1]}</div>
        </div>
      {/each}
    </div>
  </div>

  <div class="card" style="padding:16px;margin-top:14px">
    <h3 class="card-title">{i18n.t('stats.summary')}</h3>
    <div class="stat-row"><span>{i18n.t('summary.income')}</span><span class="positive">{money(S.summary.income)}</span></div>
    <div class="stat-row"><span>{i18n.t('summary.expense')}</span><span class="negative">{money(S.summary.expense)}</span></div>
    <div class="stat-row"><span>{i18n.t('summary.balance')}</span><span class:negative={S.summary.balance < 0}>{money(S.summary.balance)}</span></div>
    {#if stats.outstanding > 0}
      <div class="stat-row"><span>{i18n.t('stats.outstanding')}</span><span>{money(stats.outstanding)}</span></div>
    {/if}
  </div>
{/if}
