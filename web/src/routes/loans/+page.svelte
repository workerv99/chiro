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

  function remaining(l) {
    return l.total_amount - l.total_paid;
  }

  function statusLabel(l) {
    if (l.is_paid) return i18n.t('loans.paid');
    return i18n.t('loans.remaining') + ': ' + money(remaining(l));
  }

  function focusInit(node) {
    node.focus();
  }

  function onKeydown(e) {
    if (e.key === 'Escape' && showForm) showForm = false;
  }

  const groupedLoans = $derived(() => {
    const groups = {};
    for (const l of S.db.loans) {
      const pid = l.person_id || 'unknown';
      if (!groups[pid]) {
        groups[pid] = { person_name: l.person_name || '—', loans: [], total: 0, paid: 0 };
      }
      groups[pid].loans.push(l);
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
  <button class="btn btn-primary" onclick={openNew}>+ {i18n.t('loans.newLoan')}</button>
</div>

{#if S.db.loans.length === 0}
  <div class="card empty">
    <p class="meta">{i18n.t('loans.empty')}</p>
    <button class="btn btn-primary" onclick={openNew}>{i18n.t('loans.newLoan')}</button>
  </div>
{:else}
  <div class="card list-card">
    {#each groupedLoans() as group (group.person_name)}
      {@const groupRemaining = group.total - group.paid}
      <div class="day-group">
        <div class="day-head">
          <span>{group.person_name}</span>
          <span class="day-total negative">{money(groupRemaining)}</span>
        </div>
        {#each group.loans as l (l.loan_id)}
          <a class="row" href={`/loans/${l.loan_id}`}>
            <div class="cat-dot" style="background:{l.is_paid ? 'var(--green)' : 'var(--amber)'}"></div>
            <div class="row-body">
              <div class="row-title">{l.description || i18n.t('loans.loan')}</div>
              <div class="row-sub">
                {#if l.is_paid}
                  <span class="tag mini">{i18n.t('loans.paid')}</span>
                {/if}
              </div>
            </div>
            <div class="row-amount" class:negative={!l.is_paid}>
              {l.is_paid ? '✓' : money(remaining(l))}
            </div>
          </a>
        {/each}
      </div>
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
