<script>
  import { i18n } from '$lib/i18n.svelte.js';
  import { S, createLoan } from '$lib/stores.svelte.js';
  import { money, toDisplay, toISO, todayISO } from '$lib/format.js';
  import Button from '$lib/components/ui/button.svelte';
  import Card from '$lib/components/ui/card.svelte';
  import Input from '$lib/components/ui/input.svelte';
  import Label from '$lib/components/ui/label.svelte';
  import Dialog from '$lib/components/ui/dialog.svelte';
  import DialogHeader from '$lib/components/ui/dialog-header.svelte';
  import DialogTitle from '$lib/components/ui/dialog-title.svelte';
  import DialogFooter from '$lib/components/ui/dialog-footer.svelte';
  import { Plus } from 'lucide-svelte';

  let showForm = $state(false);
  let personId = $state('');
  let amount = $state('');
  let rate = $state('');
  let interestType = $state('simple');
  let months = $state('12');
  let frequency = $state('monthly');
  let startDate = $state(toDisplay(todayISO()));
  let dueDate = $state('');
  let description = $state('');
  let customInstallment = $state('');
  let err = $state('');

  function openNew() {
    personId = '';
    amount = '';
    rate = '';
    interestType = 'simple';
    months = '12';
    frequency = 'monthly';
    startDate = toDisplay(todayISO());
    dueDate = '';
    description = '';
    customInstallment = '';
    err = '';
    showForm = true;
  }

  async function save() {
    err = '';
    const p = parseFloat(amount);
    if (!personId) return (err = i18n.t('loans.personRequired'));
    if (!p || p <= 0) return (err = i18n.t('loans.amountRequired'));
    if (!description.trim()) return (err = i18n.t('common.required'));
    const n = parseInt(months, 10);
    if (!n || n <= 0) return (err = i18n.t('loans.installmentsRequired'));
    const custom = parseFloat(customInstallment) || 0;
    try {
      await createLoan({
        person_id: personId,
        description: description.trim(),
        amount: p,
        date: toISO(startDate),
        interest_rate: rate ? parseFloat(rate) : 0,
        interest_type: interestType,
        months: n,
        frequency,
        first_due_date: dueDate ? toISO(dueDate) : null,
        custom_installment: custom
      });
      showForm = false;
    } catch (e) {
      err = e.message;
    }
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
  <h1 class="text-2xl font-bold">{i18n.t('loans.title')}</h1>
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
  class="fixed right-5 bottom-20 h-14 w-14 rounded-2xl shadow-lg z-40"
  onclick={openNew}
  aria-label={i18n.t('loans.newLoan')}
>
  <Plus class="h-6 w-6" />
</Button>

<Dialog bind:open={showForm}>
  <DialogHeader>
    <DialogTitle>{i18n.t('loans.newLoan')}</DialogTitle>
  </DialogHeader>

  <div class="space-y-4">
    <div class="space-y-2">
      <Label for="loan-person">{i18n.t('loans.person')}</Label>
      <select id="loan-person" bind:value={personId} class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
        <option value="">—</option>
        {#each S.db.persons as p (p.person_id)}
          <option value={p.person_id}>{p.name}</option>
        {/each}
      </select>
    </div>

    <div class="grid grid-cols-2 gap-4">
      <div class="space-y-2">
        <Label for="loan-amount">{i18n.t('loans.principal')}</Label>
        <Input id="loan-amount" bind:value={amount} inputmode="decimal" />
      </div>
      <div class="space-y-2">
        <Label for="loan-months">{i18n.t('loans.installments')}</Label>
        <Input id="loan-months" bind:value={months} inputmode="numeric" />
      </div>
    </div>

    <div class="space-y-2">
      <Label for="loan-custom">Cuota personalizada (opcional)</Label>
      <Input id="loan-custom" bind:value={customInstallment} inputmode="decimal" placeholder="Dejar vacío para cuota igual" />
      <p class="text-xs text-muted-foreground">Si definís un monto, el sistema calcula cuántas cuotas caben y el resto va en la última.</p>
    </div>

    <div class="space-y-2">
      <Label for="loan-freq">{i18n.t('loans.frequency')}</Label>
      <select id="loan-freq" bind:value={frequency} class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
        <option value="weekly">{i18n.t('loans.weekly')}</option>
        <option value="biweekly">{i18n.t('loans.biweekly')}</option>
        <option value="monthly">{i18n.t('loans.monthly')}</option>
      </select>
    </div>

    <details class="group">
      <summary class="cursor-pointer text-sm font-semibold text-muted-foreground py-2">{i18n.t('common.moreOptions')}</summary>
      <div class="space-y-4 pt-2">
        <div class="space-y-2">
          <Label for="loan-desc">{i18n.t('loans.description')}</Label>
          <Input id="loan-desc" bind:value={description} />
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div class="space-y-2">
            <Label for="loan-rate">{i18n.t('loans.rate')}</Label>
            <Input id="loan-rate" bind:value={rate} inputmode="decimal" placeholder="0" />
          </div>
          <div class="space-y-2">
            <Label for="loan-interest">{i18n.t('loans.interest')}</Label>
            <select id="loan-interest" bind:value={interestType} class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
              <option value="simple">{i18n.t('loans.interestSimple')}</option>
              <option value="compound">{i18n.t('loans.interestCompound')}</option>
            </select>
          </div>
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div class="space-y-2">
            <Label for="loan-start">{i18n.t('loans.startDate')}</Label>
            <Input id="loan-start" bind:value={startDate} placeholder={i18n.t('expenses.datePlaceholder')} />
          </div>
          <div class="space-y-2">
            <Label for="loan-due">{i18n.t('loans.dueDate')}</Label>
            <Input id="loan-due" bind:value={dueDate} placeholder={i18n.t('expenses.datePlaceholder')} />
          </div>
        </div>
      </div>
    </details>

    {#if err}
      <p class="text-sm text-destructive">{err}</p>
    {/if}
  </div>

  <DialogFooter>
    <Button variant="outline" onclick={() => (showForm = false)}>{i18n.t('common.cancel')}</Button>
    <Button onclick={save}>{i18n.t('common.save')}</Button>
  </DialogFooter>
</Dialog>
