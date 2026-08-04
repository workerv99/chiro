<script>
  import { i18n } from '$lib/i18n.svelte.js';
  import { S, createLoan } from '$lib/stores.svelte.js';
  import { money, toDisplay, toISO, todayISO } from '$lib/format.js';

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
        first_due_date: dueDate ? toISO(dueDate) : null
      });
      showForm = false;
    } catch (e) {
      err = e.message;
    }
  }

  function focusInit(node) {
    node.focus();
  }

  function onKeydown(e) {
    if (e.key === 'Escape' && showForm) showForm = false;
  }

  const totalLoans = $derived(() => S.db.loans.reduce((sum, l) => sum + l.total_amount, 0));
  const totalPaid = $derived(() => S.db.loans.reduce((sum, l) => sum + l.total_paid, 0));
  const totalPending = $derived(() => totalLoans() - totalPaid());

  const groupedPersons = $derived(() => {
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
<svelte:window onkeydown={onKeydown} />

<div class="page-head">
  <h1 class="headline">{i18n.t('loans.title')}</h1>
</div>

{#if S.db.loans.length > 0}
  <div class="summary-cards">
    <div class="card stat-card">
      <div class="stat-label">Total prestado</div>
      <div class="stat-value">{money(totalLoans())}</div>
    </div>
    <div class="card stat-card">
      <div class="stat-label">Pendiente</div>
      <div class="stat-value negative">{money(totalPending())}</div>
    </div>
  </div>
{/if}

{#if S.db.loans.length === 0}
  <div class="card empty">
    <p class="meta">{i18n.t('loans.empty')}</p>
    <button class="btn btn-primary" onclick={openNew}>{i18n.t('loans.newLoan')}</button>
  </div>
{:else}
  <div class="card list-card">
    {#each groupedPersons() as group (group.person_id)}
      {@const groupRemaining = group.total - group.paid}
      <a class="row" href={`/loans/person/${group.person_id}`}>
        <div class="cat-dot" style="background:var(--indigo)"></div>
        <div class="row-body">
          <div class="row-title">{group.person_name}</div>
          <div class="row-sub">
            <span>{group.count} préstamo{group.count > 1 ? 's' : ''}</span>
          </div>
        </div>
        <div class="row-amount negative">
          {money(groupRemaining)}
        </div>
      </a>
    {/each}
  </div>
{/if}

{#if showForm}
  <div class="overlay" onclick={() => (showForm = false)}>
    <div class="sheet" role="dialog" aria-modal="true" aria-labelledby="loan-form-title" onclick={(e) => e.stopPropagation()}>
      <h2 class="title" id="loan-form-title" style="margin-bottom:16px">{i18n.t('loans.newLoan')}</h2>

      <div class="form-field">
        <label class="eyebrow" for="loan-person">{i18n.t('loans.person')}</label>
        <select id="loan-person" bind:value={personId} use:focusInit>
          <option value="">—</option>
          {#each S.db.persons as p (p.person_id)}
            <option value={p.person_id}>{p.name}</option>
          {/each}
        </select>
      </div>
      <div class="grid2">
        <div class="form-field">
          <label class="eyebrow" for="loan-amount">{i18n.t('loans.principal')}</label>
          <input id="loan-amount" bind:value={amount} inputmode="decimal" />
        </div>
        <div class="form-field">
          <label class="eyebrow" for="loan-months">{i18n.t('loans.installments')}</label>
          <input id="loan-months" bind:value={months} inputmode="numeric" />
        </div>
      </div>
      <div class="form-field">
        <label class="eyebrow" for="loan-freq">{i18n.t('loans.frequency')}</label>
        <select id="loan-freq" bind:value={frequency}>
          <option value="weekly">{i18n.t('loans.weekly')}</option>
          <option value="biweekly">{i18n.t('loans.biweekly')}</option>
          <option value="monthly">{i18n.t('loans.monthly')}</option>
        </select>
      </div>

      <details class="more-options">
        <summary>{i18n.t('common.moreOptions')}</summary>
        <div class="form-field" style="margin-top:12px">
          <label class="eyebrow" for="loan-desc">{i18n.t('loans.description')}</label>
          <input id="loan-desc" bind:value={description} />
        </div>
        <div class="grid2">
          <div class="form-field">
            <label class="eyebrow" for="loan-rate">
              {i18n.t('loans.rate')}
              <span class="hint" title={i18n.t('loans.rateHint')}>?</span>
            </label>
            <input id="loan-rate" bind:value={rate} inputmode="decimal" placeholder="0" />
          </div>
          <div class="form-field">
            <label class="eyebrow" for="loan-interest">
              {i18n.t('loans.interest')}
              <span class="hint" title={i18n.t('loans.interestHint')}>?</span>
            </label>
            <select id="loan-interest" bind:value={interestType}>
              <option value="simple">{i18n.t('loans.interestSimple')}</option>
              <option value="compound">{i18n.t('loans.interestCompound')}</option>
            </select>
          </div>
        </div>
        <div class="grid2">
          <div class="form-field">
            <label class="eyebrow" for="loan-start">{i18n.t('loans.startDate')}</label>
            <input id="loan-start" bind:value={startDate} placeholder={i18n.t('expenses.datePlaceholder')} />
          </div>
          <div class="form-field">
            <label class="eyebrow" for="loan-due">{i18n.t('loans.dueDate')}</label>
            <input id="loan-due" bind:value={dueDate} placeholder={i18n.t('expenses.datePlaceholder')} />
          </div>
        </div>
      </details>

      {#if err}
        <p class="error-text" role="alert">{err}</p>
      {/if}
      <div class="inline-flex" style="margin-top:8px">
        <button class="btn btn-cancel" onclick={() => (showForm = false)}>{i18n.t('common.cancel')}</button>
        <button class="btn btn-primary" onclick={save}>{i18n.t('common.save')}</button>
      </div>
    </div>
  </div>
{/if}

<button class="fab" onclick={openNew} aria-label={i18n.t('loans.newLoan')}>+</button>
