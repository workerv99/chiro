<script>
  import { page } from '$app/state';
  import { i18n } from '$lib/i18n.svelte.js';
  import { S } from '$lib/stores.svelte.js';
  import { money } from '$lib/format.js';

  const personId = String(page.params.id);

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

<div class="page-head">
  <div>
    <a class="back-link" href="/loans">← {i18n.t('tabs.loans')}</a>
    <h1 class="headline">{person?.name || '—'}</h1>
  </div>
</div>

<div class="summary-cards">
  <div class="card stat-card">
    <div class="stat-label">{i18n.t('loans.remaining')}</div>
    <div class="stat-value negative">{money(totalRemaining)}</div>
  </div>
  <div class="card stat-card">
    <div class="stat-label">{i18n.t('loans.paidLabel')}</div>
    <div class="stat-value positive">{money(totalPaid)}</div>
  </div>
  <div class="card stat-card">
    <div class="stat-label">{i18n.t('loans.progress')}</div>
    <div class="stat-value">{paidCount}/{loans.length}</div>
  </div>
</div>

{#if loans.length === 0}
  <div class="card empty" style="margin-top:14px">
    <p class="meta">{i18n.t('loans.noLoansForPerson')}</p>
    <a class="btn btn-primary" href="/loans">{i18n.t('loans.newLoan')}</a>
  </div>
{:else}
  <div class="card list-card" style="margin-top:14px">
    {#each loans as l (l.loan_id)}
      <a class="row" href={`/loans/${l.loan_id}`}>
        <div class="cat-dot" style="background:{l.is_paid ? 'var(--green)' : 'var(--amber)'}"></div>
        <div class="row-body">
          <div class="row-title">{l.description || i18n.t('loans.loan')}</div>
          <div class="row-sub">
            {#if l.is_paid}
              <span class="tag mini">{i18n.t('loans.paid')}</span>
            {:else}
              <span>{i18n.t('loans.progress')}: {l.months || '?'}</span>
            {/if}
          </div>
        </div>
        <div class="row-amount" class:negative={!l.is_paid}>
          {l.is_paid ? '✓' : money(remaining(l))}
        </div>
      </a>
    {/each}
  </div>
{/if}
