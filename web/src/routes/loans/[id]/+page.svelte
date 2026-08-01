<script>
  import { page } from '$app/state';
  import { goto } from '$app/navigation';
  import { i18n } from '$lib/i18n.svelte.js';
  import { api } from '$lib/api.svelte.js';
  import { payInstallment, cascadeInstallment, unpayInstallment, remove } from '$lib/stores.svelte.js';
  import { money, toDisplay, toISO, todayISO } from '$lib/format.js';

  const loanId = String(page.params.id);
  let loan = $state(null);
  let schedule = $state([]);
  let loading = $state(true);
  let payAmount = $state('');
  let payDate = $state(toDisplay(todayISO()));

  async function load() {
    const [loans, s] = await Promise.all([api('/api/loans'), api(`/api/loans/${loanId}/installments`)]);
    loan = loans.find((l) => l.loan_id === loanId) || null;
    schedule = s;
    if (nextPending) payAmount = String(nextPending.amount);
  }

  $effect(() => {
    loading = true;
    load().finally(() => (loading = false));
  });

  const nextPending = $derived(schedule.find((x) => !x.is_paid));
  const paidCount = $derived(schedule.filter((x) => x.is_paid).length);
  const remaining = $derived(loan ? loan.total_amount - loan.total_paid : 0);

  async function pay(next) {
    const target = nextPending;
    if (!target) return;
    const amt = next ? target.amount : parseFloat(payAmount);
    if (!amt || amt <= 0) return;
    await payInstallment(target.installment_id, { amount: amt, date: toISO(payDate) });
    await load();
  }

  async function cascade() {
    const target = nextPending;
    if (!target) return;
    const amt = parseFloat(payAmount);
    if (!amt || amt <= 0) return;
    await cascadeInstallment(target.installment_id, { amount: amt, date: toISO(payDate) });
    await load();
  }

  async function unpay() {
    const last = [...schedule].reverse().find((x) => x.is_paid);
    if (!last) return;
    await unpayInstallment(last.installment_id);
    await load();
  }

  async function del() {
    if (!confirm(i18n.t('loans.deleteConfirm'))) return;
    await remove('loans', loanId);
    goto('/loans');
  }
</script>

{#if loading || !loan}
  <p class="meta" style="padding:24px;text-align:center">{i18n.t('common.loading')}</p>
{:else}
  <div class="page-head">
    <div>
      <a class="back-link" href="/loans">← {i18n.t('tabs.loans')}</a>
      <h1 class="headline">{loan.person_name}</h1>
      <p class="meta">
        {#if loan.description}{loan.description}{/if}
        {#if loan.frequency}· {loan.frequency}{/if}
      </p>
    </div>
  </div>

  <div class="summary-cards">
    <div class="card stat-card">
      <div class="stat-label">{i18n.t('loans.remaining')}</div>
      <div class="stat-value">{money(remaining)}</div>
    </div>
    <div class="card stat-card">
      <div class="stat-label">{i18n.t('loans.paidLabel')}</div>
      <div class="stat-value">{money(loan.total_paid)}</div>
    </div>
    <div class="card stat-card">
      <div class="stat-label">{i18n.t('loans.progress')}</div>
      <div class="stat-value">{paidCount}/{schedule.length}</div>
    </div>
  </div>

  <div class="card" style="padding:16px;margin-top:14px">
    <div class="grid2" style="align-items:end">
      <div class="form-field">
        <label>{i18n.t('loans.paymentAmount')}</label>
        <input bind:value={payAmount} inputmode="decimal" />
      </div>
      <div class="form-field">
        <label>{i18n.t('loans.paymentDate')}</label>
        <input bind:value={payDate} placeholder={i18n.t('expenses.datePlaceholder')} />
      </div>
    </div>
    <div class="inline-flex" style="margin-top:8px;flex-wrap:wrap">
      <button class="btn btn-primary" onclick={() => pay(true)} disabled={!nextPending}>
        {i18n.t('loans.payNext')}
      </button>
      <button class="btn" onclick={() => pay(false)} disabled={!nextPending}>{i18n.t('loans.payNow')}</button>
      <button class="btn" onclick={cascade} disabled={!nextPending}>{i18n.t('loans.cascade')}</button>
      <button class="btn" onclick={unpay} disabled={paidCount === 0}>{i18n.t('loans.unpay')}</button>
    </div>
  </div>

  <div class="card list-card" style="margin-top:14px">
    <h3 class="card-title" style="padding:12px 16px 0">{i18n.t('loans.schedule')}</h3>
    {#each schedule as s (s.number)}
      <div class="row">
        <div class="cat-dot" style="background:{s.is_paid ? '#22c55e' : s.is_overdue ? '#f43f5e' : '#e2e8f0'}"></div>
        <div class="row-body">
          <div class="row-title">{i18n.t('loans.installmentN')} {s.number}</div>
          <div class="row-sub">{toDisplay(s.due_date)}</div>
        </div>
        <div class="row-right">
          <div class="row-amount">{money(s.amount)}</div>
          <div class="row-sub">
            {#if s.is_paid}
              <span class="badge-ok">{i18n.t('loans.paid')}</span>
            {:else if s.is_partial}
              <span class="badge-warn">{i18n.t('loans.partial')}</span>
            {:else}
              <span class="badge-pending">{i18n.t('loans.pending')}</span>
            {/if}
          </div>
        </div>
      </div>
    {/each}
  </div>

  <button class="btn btn-danger" style="margin:16px 0 32px" onclick={del}>{i18n.t('common.delete')}</button>
{/if}
