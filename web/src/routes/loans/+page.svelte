<script>
  import { page } from '$app/state';
  import { goto } from '$app/navigation';
  import { i18n } from '$lib/i18n.svelte.js';
  import { S } from '$lib/stores.svelte.js';
  import { money } from '$lib/format.js';
  import Button from '$lib/components/ui/button.svelte';
  import Card from '$lib/components/ui/card.svelte';
  import LoanFormDialog from '$lib/components/loan-form-dialog.svelte';
  import { Plus } from 'lucide-svelte';

  let showForm = $state(false);
  let initialPersonId = $state('');

  function openNew() {
    initialPersonId = '';
    showForm = true;
  }

  // Deep link: /loans?new=1&person=<id> opens the dialog with that lender preselected.
  $effect(() => {
    const params = page.url.searchParams;
    if (params.get('new') === '1') {
      initialPersonId = params.get('person') || '';
      showForm = true;
    }
  });

  // Land on the lender's panel, not the main loans list.
  async function onSaved(personId) {
    await goto(`/loans/person/${personId}`);
  }

  const totalLoans = $derived(S.db.loans.reduce((sum, l) => sum + l.total_amount, 0));
  const totalPaid = $derived(S.db.loans.reduce((sum, l) => sum + l.total_paid, 0));
  const totalPending = $derived(totalLoans - totalPaid);

  const groupedPersons = $derived.by(() => {
    const groups = {};
    for (const l of S.db.loans) {
      const pid = l.person_id || 'unknown';
      if (!groups[pid]) {
        groups[pid] = { person_id: pid, person_name: l.person_name || '—', count: 0, total: 0, paid: 0 };
      }
      groups[pid].count++;
      groups[pid].total += l.total_amount;
      groups[pid].paid += l.total_paid;
    }
    return Object.values(groups).sort((a, b) => a.person_name.localeCompare(b.person_name));
  });
</script>

<svelte:head><title>{i18n.t('loans.title')} · Chiro</title></svelte:head>

<div class="flex items-center justify-between mb-4">
  <h1 class="text-xl md:text-2xl font-bold">{i18n.t('loans.title')}</h1>
</div>

{#if S.db.loans.length > 0}
  <div class="grid grid-cols-2 gap-3 mb-4">
    <Card class="p-4">
      <p class="text-xs font-bold text-muted-foreground uppercase tracking-wider mb-1">Total prestado</p>
      <p class="text-lg font-bold">{money(totalLoans)}</p>
    </Card>
    <Card class="p-4">
      <p class="text-xs font-bold text-muted-foreground uppercase tracking-wider mb-1">Pendiente</p>
      <p class="text-lg font-bold text-destructive">{money(totalPending)}</p>
    </Card>
  </div>
{/if}

{#if S.db.loans.length === 0}
  <Card class="flex flex-col items-center gap-3 py-8">
    <p class="text-sm text-muted-foreground">{i18n.t('loans.empty')}</p>
    <Button onclick={openNew}>{i18n.t('loans.newLoan')}</Button>
  </Card>
{:else}
  <Card class="overflow-hidden">
    {#each groupedPersons as group (group.person_id)}
      {@const groupRemaining = group.total - group.paid}
      <a class="flex items-center gap-3 px-4 py-3 hover:bg-muted/50 transition-colors border-t first:border-t-0" href={`/loans/person/${group.person_id}`}>
        <div class="h-2.5 w-2.5 rounded-full bg-primary"></div>
        <div class="flex-1 min-w-0">
          <p class="text-sm font-semibold">{group.person_name}</p>
          <p class="text-xs text-muted-foreground">
            {group.count} préstamo{group.count > 1 ? 's' : ''}
          </p>
        </div>
        <p class="text-sm font-bold text-destructive">{money(groupRemaining)}</p>
      </a>
    {/each}
  </Card>
{/if}

<Button
  size="icon"
  class="fixed right-5 bottom-[calc(5rem+env(safe-area-inset-bottom))] sm:bottom-6 h-14 w-14 rounded-2xl shadow-lg z-40"
  onclick={openNew}
  aria-label={i18n.t('loans.newLoan')}
>
  <Plus class="h-6 w-6" />
</Button>

<LoanFormDialog bind:open={showForm} {initialPersonId} {onSaved} />
