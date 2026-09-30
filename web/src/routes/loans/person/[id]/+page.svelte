<script>
  import { page } from '$app/state';
  import { i18n } from '$lib/i18n.svelte.js';
  import { S } from '$lib/stores.svelte.js';
  import { money } from '$lib/format.js';
  import Button from '$lib/components/ui/button.svelte';
  import Card from '$lib/components/ui/card.svelte';
  import { ArrowLeft, Check, Plus } from 'lucide-svelte';

  const personId = $derived(String(page.params.id));

  const person = $derived(S.db.persons.find(p => p.person_id === personId));
  const loans = $derived(S.db.loans.filter(l => l.person_id === personId));
  const totalRemaining = $derived(loans.reduce((sum, l) => sum + (l.total_amount - l.total_paid), 0));
  const totalPaid = $derived(loans.reduce((sum, l) => sum + l.total_paid, 0));
  const paidCount = $derived(loans.filter(l => l.is_paid).length);

  function remaining(l) {
    return l.total_amount - l.total_paid;
  }
</script>

<svelte:head><title>{person?.name || i18n.t('loans.title')} · Chiro</title></svelte:head>

<a class="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground font-semibold mb-2" href="/loans"><ArrowLeft size={16} /> {i18n.t('tabs.loans')}</a>
<div class="flex items-center justify-between gap-3 mb-4">
  <h1 class="text-xl md:text-2xl font-bold">{person?.name || '—'}</h1>
  <a href={`/loans?new=1&person=${personId}`}>
    <Button size="sm"><Plus size={16} /> {i18n.t('loans.newLoan')}</Button>
  </a>
</div>

<div class="grid grid-cols-3 gap-2 mb-4">
  <Card class="p-3 min-w-0">
    <p class="text-xs font-bold text-muted-foreground uppercase tracking-wider mb-1 truncate">{i18n.t('loans.remaining')}</p>
    <p class="text-sm font-bold text-destructive truncate">{money(totalRemaining)}</p>
  </Card>
  <Card class="p-3 min-w-0">
    <p class="text-xs font-bold text-muted-foreground uppercase tracking-wider mb-1 truncate">{i18n.t('loans.paidLabel')}</p>
    <p class="text-sm font-bold text-green-500 truncate">{money(totalPaid)}</p>
  </Card>
  <Card class="p-3 min-w-0">
    <p class="text-xs font-bold text-muted-foreground uppercase tracking-wider mb-1 truncate">{i18n.t('loans.progress')}</p>
    <p class="text-sm font-bold truncate">{paidCount}/{loans.length}</p>
  </Card>
</div>

{#if loans.length === 0}
  <Card class="flex flex-col items-center gap-3 py-8">
    <p class="text-sm text-muted-foreground">{i18n.t('loans.noLoansForPerson')}</p>
    <a href={`/loans?new=1&person=${personId}`}>
      <Button><Plus size={16} /> {i18n.t('loans.newLoan')}</Button>
    </a>
  </Card>
{:else}
  <Card class="overflow-hidden">
    {#each loans as l (l.loan_id)}
      {@const loanRemaining = l.total_amount - l.total_paid}
      <a class="flex items-center gap-3 px-4 py-3 hover:bg-muted/50 transition-colors border-t first:border-t-0" href={`/loans/${l.loan_id}`}>
        <div class="h-2.5 w-2.5 rounded-full" style="background:{l.is_paid ? '#22c55e' : '#f59e0b'}"></div>
        <div class="flex-1 min-w-0">
          <p class="text-sm font-semibold truncate" title={l.description || i18n.t('loans.loan')}>{l.description || i18n.t('loans.loan')}</p>
          <p class="text-xs text-muted-foreground">
            {i18n.t('loans.installments')}: {l.months || '?'}
          </p>
        </div>
        <div class="text-right">
          <p class="text-sm font-bold" class:text-destructive={!l.is_paid}>{#if l.is_paid}<Check size={16} class="inline text-green-500" />{:else}{money(loanRemaining)}{/if}</p>
          {#if l.is_paid}
            <p class="text-xs text-green-500 font-semibold">{i18n.t('loans.paid')}</p>
          {/if}
        </div>
      </a>
    {/each}
  </Card>
{/if}
