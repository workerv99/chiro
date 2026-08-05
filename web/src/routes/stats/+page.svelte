<script>
  import { i18n } from '$lib/i18n.svelte.js';
  import { api } from '$lib/api.svelte.js';
  import { S, loadMonth } from '$lib/stores.svelte.js';
  import { money } from '$lib/format.js';
  import Button from '$lib/components/ui/button.svelte';
  import Card from '$lib/components/ui/card.svelte';
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

  function shift(delta) {
    let m = month + delta;
    let y = year;
    if (m < 1) { m = 12; y -= 1; }
    if (m > 12) { m = 1; y += 1; }
    year = y;
    month = m;
  }

  const breakdown = $derived((stats.breakdown || []).filter((b) => b.type === 'expense'));
  const maxCat = $derived(Math.max(1, ...breakdown.map((c) => c.total)));
  const months = $derived(stats.months || []);
  const maxMonth = $derived(Math.max(1, ...months.map((m) => m.expense)));
  const monthNames = $derived(
    Array.from({ length: 12 }, (_, i) => new Date(year, i, 1).toLocaleDateString(i18n.lang === 'en' ? 'en' : 'es', { month: 'short' }))
  );
</script>

<svelte:head><title>{i18n.t('stats.title')} · Chiro</title></svelte:head>

<div class="flex items-center justify-between mb-4">
  <h1 class="text-2xl font-bold">{i18n.t('stats.title')}</h1>
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

{#if loading}
  <p class="text-sm text-muted-foreground py-8 text-center">{i18n.t('common.loading')}</p>
{:else}
  <Card class="p-4 mb-4">
    <h3 class="font-bold mb-3">{i18n.t('stats.byCategory')}</h3>
    {#if breakdown.length === 0}
      <p class="text-sm text-muted-foreground text-center py-4">{i18n.t('stats.empty')}</p>
    {:else}
      <div class="space-y-3">
        {#each breakdown as c (c.category_name)}
          <div class="flex items-center gap-3">
            <span class="text-sm w-24 truncate">{c.category_name}</span>
            <div class="flex-1 h-2 bg-border rounded-full overflow-hidden">
              <div class="h-full rounded-full" style="width:{Math.max(4, (c.total / maxCat) * 100)}%;background:{c.category_color}"></div>
            </div>
            <span class="text-sm font-bold w-20 text-right">{money(c.total)}</span>
          </div>
        {/each}
      </div>
    {/if}
  </Card>

  <Card class="p-4 mb-4">
    <h3 class="font-bold mb-3">{i18n.t('stats.monthly')}</h3>
    <div class="flex items-end gap-1.5 h-32">
      {#each months as m (m.month)}
        <div class="flex-1 flex flex-col items-center justify-end gap-1" title={monthNames[m.month - 1]}>
          <div class="w-full rounded-sm bg-primary" style="height:{m.expense > 0 ? Math.max(4, (m.expense / maxMonth) * 100) : 2}%"></div>
          <span class="text-[10px] text-muted-foreground">{monthNames[m.month - 1]}</span>
        </div>
      {/each}
    </div>
  </Card>

  <Card class="p-4">
    <h3 class="font-bold mb-3">{i18n.t('stats.summary')}</h3>
    <div class="space-y-2">
      <div class="flex justify-between">
        <span class="text-sm">{i18n.t('summary.income')}</span>
        <span class="text-sm font-bold text-green-500">{money(S.summary.income)}</span>
      </div>
      <div class="flex justify-between">
        <span class="text-sm">{i18n.t('summary.expense')}</span>
        <span class="text-sm font-bold text-destructive">{money(S.summary.expense)}</span>
      </div>
      <div class="flex justify-between">
        <span class="text-sm">{i18n.t('summary.balance')}</span>
        <span class="text-sm font-bold" class:text-destructive={S.summary.balance < 0}>{money(S.summary.balance)}</span>
      </div>
      {#if stats.outstanding > 0}
        <div class="flex justify-between">
          <span class="text-sm">{i18n.t('stats.outstanding')}</span>
          <span class="text-sm font-bold">{money(stats.outstanding)}</span>
        </div>
      {/if}
    </div>
  </Card>
{/if}