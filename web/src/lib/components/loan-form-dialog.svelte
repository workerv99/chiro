<script>
  import { untrack } from 'svelte';
  import { i18n } from '$lib/i18n.svelte.js';
  import { S, create, createLoan } from '$lib/stores.svelte.js';
  import { toDisplay, toISO, todayISO, parseDecimal, nextLoanTitle } from '$lib/format.js';
  import Button from '$lib/components/ui/button.svelte';
  import Card from '$lib/components/ui/card.svelte';
  import Input from '$lib/components/ui/input.svelte';
  import Label from '$lib/components/ui/label.svelte';
  import Dialog from '$lib/components/ui/dialog.svelte';
  import DialogHeader from '$lib/components/ui/dialog-header.svelte';
  import DialogTitle from '$lib/components/ui/dialog-title.svelte';
  import DialogFooter from '$lib/components/ui/dialog-footer.svelte';
  import { Plus } from 'lucide-svelte';

  // lockedPersonId: fixes the lender (no selector, no "new lender" form).
  // initialPersonId: preselects the lender when not locked.
  // onSaved(personId): called after the loan is created and the dialog closed.
  let { open = $bindable(false), lockedPersonId = '', initialPersonId = '', onSaved } = $props();

  let selectedPersonId = $state('');
  let amount = $state('');
  let rate = $state('');
  let interestType = $state('simple');
  let months = $state('12');
  let frequency = $state('monthly');
  let startDate = $state(toDisplay(todayISO()));
  let dueDate = $state('');
  let description = $state('');
  let customInstallment = $state('');
  let showPersonForm = $state(false);
  let personName = $state('');
  let personNotes = $state('');
  let err = $state('');

  const locked = $derived(!!lockedPersonId);
  const personId = $derived(locked ? lockedPersonId : selectedPersonId);
  const lockedPerson = $derived(S.db.persons.find((p) => p.person_id === lockedPersonId));
  const autoTitle = $derived(nextLoanTitle(S.db.loans, i18n.t('loans.loan')));

  function reset() {
    selectedPersonId = initialPersonId || '';
    amount = '';
    rate = '';
    interestType = 'simple';
    months = '12';
    frequency = 'monthly';
    startDate = toDisplay(todayISO());
    dueDate = '';
    description = '';
    customInstallment = '';
    showPersonForm = false;
    personName = '';
    personNotes = '';
    err = '';
  }

  // Reset the form every time the dialog goes from closed to open.
  let wasOpen = false;
  $effect(() => {
    const isOpen = open;
    if (isOpen && !wasOpen) untrack(reset);
    wasOpen = isOpen;
  });

  function openPersonForm() {
    personName = '';
    personNotes = '';
    err = '';
    showPersonForm = true;
  }

  async function savePerson() {
    err = '';
    if (!personName.trim()) return (err = i18n.t('common.required'));
    try {
      const person = await create('persons', { name: personName.trim(), notes: personNotes.trim() || null });
      selectedPersonId = person.person_id;
      showPersonForm = false;
    } catch (e) {
      err = e.message;
    }
  }

  async function save() {
    err = '';
    const p = parseDecimal(amount);
    if (!personId) return (err = i18n.t('loans.personRequired'));
    if (!p || p <= 0) return (err = i18n.t('loans.amountRequired'));
    const n = parseInt(months, 10);
    if (!n || n <= 0) return (err = i18n.t('loans.installmentsRequired'));
    const custom = parseDecimal(customInstallment) || 0;
    const pid = personId;
    try {
      await createLoan({
        person_id: pid,
        description: description.trim() || autoTitle,
        amount: p,
        date: toISO(startDate),
        interest_rate: rate ? parseDecimal(rate) : 0,
        interest_type: interestType,
        months: n,
        frequency,
        first_due_date: dueDate ? toISO(dueDate) : null,
        custom_installment: custom
      });
      open = false;
      await onSaved?.(pid);
    } catch (e) {
      err = e.message;
    }
  }
</script>

<Dialog bind:open>
  <DialogHeader>
    <DialogTitle>{i18n.t('loans.newLoan')}</DialogTitle>
  </DialogHeader>

  <div class="space-y-4">
    {#if locked}
      <div class="space-y-2">
        <Label>{i18n.t('loans.person')}</Label>
        <p class="flex h-11 w-full items-center rounded-md border border-input bg-muted/40 px-3 text-base text-muted-foreground">
          {lockedPerson?.name || '—'}
        </p>
      </div>
    {:else}
      <div class="space-y-2">
        <Label for="loan-person">{i18n.t('loans.person')}</Label>
        <select id="loan-person" bind:value={selectedPersonId} class="flex h-11 w-full rounded-md border border-input bg-background px-3 py-2 text-base">
          <option value="">—</option>
          {#each S.db.persons as p (p.person_id)}
            <option value={p.person_id}>{p.name}</option>
          {/each}
        </select>
        <Button type="button" variant="outline" size="sm" onclick={openPersonForm}>
          <Plus size={16} /> Nuevo prestamista
        </Button>
      </div>

      {#if showPersonForm}
        <Card class="p-3 space-y-3 border-primary/40">
          <p class="text-sm font-semibold">Nuevo prestamista</p>
          <div class="space-y-2">
            <Label for="new-person-name">Nombre</Label>
            <Input id="new-person-name" bind:value={personName} autofocus />
          </div>
          <div class="space-y-2">
            <Label for="new-person-notes">Notas (opcional)</Label>
            <Input id="new-person-notes" bind:value={personNotes} />
          </div>
          <div class="flex justify-end gap-2">
            <Button type="button" variant="ghost" size="sm" onclick={() => (showPersonForm = false)}>Cancelar</Button>
            <Button type="button" size="sm" onclick={savePerson}>Guardar prestamista</Button>
          </div>
        </Card>
      {/if}
    {/if}

    <div class="space-y-2">
      <Label for="loan-desc">{i18n.t('loans.description')}</Label>
      <Input id="loan-desc" bind:value={description} placeholder={autoTitle} />
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
      <select id="loan-freq" bind:value={frequency} class="flex h-11 w-full rounded-md border border-input bg-background px-3 py-2 text-base">
        <option value="weekly">{i18n.t('loans.weekly')}</option>
        <option value="biweekly">{i18n.t('loans.biweekly')}</option>
        <option value="monthly">{i18n.t('loans.monthly')}</option>
      </select>
    </div>

    <details class="group">
      <summary class="cursor-pointer text-sm font-semibold text-muted-foreground py-2">{i18n.t('common.moreOptions')}</summary>
      <div class="space-y-4 pt-2">
        <div class="grid grid-cols-2 gap-4">
          <div class="space-y-2">
            <Label for="loan-rate">{i18n.t('loans.rate')}</Label>
            <Input id="loan-rate" bind:value={rate} inputmode="decimal" placeholder="0" />
          </div>
          <div class="space-y-2">
            <Label for="loan-interest">{i18n.t('loans.interest')}</Label>
            <select id="loan-interest" bind:value={interestType} class="flex h-11 w-full rounded-md border border-input bg-background px-3 py-2 text-base">
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
    <Button variant="outline" onclick={() => (open = false)}>{i18n.t('common.cancel')}</Button>
    <Button onclick={save}>{i18n.t('common.save')}</Button>
  </DialogFooter>
</Dialog>
