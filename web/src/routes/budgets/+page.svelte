<script>
  import { i18n } from '$lib/i18n.svelte.js';
  import { api } from '$lib/api.svelte.js';
  import { S, create, update, remove } from '$lib/stores.svelte.js';
  import { money, pct } from '$lib/format.js';
  import ConfirmSheet from '$lib/components/ConfirmSheet.svelte';
  import { ChevronLeft, ChevronRight } from 'lucide-svelte';

  const now = new Date();
  let year = $state(now.getFullYear());
  let month = $state(now.getMonth() + 1);
  let list = $state([]);
  let showForm = $state(false);
  let editing = $state(null);
  let catId = $state('');
  let amount = $state('');
  let err = $state('');
  let confirmDel = $state(null);

  function load() {
    return api(`/api/budgets/progress?year=${year}&month=${month}`)
      .then((d) => (list = d ?? []))
      .catch(() => {});
  }

  $effect(() => {
    load();
  });

  function openNew() {
    editing = null;
    amount = '';
    catId = '';
    err = '';
    showForm = true;
  }

  function openEdit(b) {
    editing = b;
    amount = String(b.amount ?? '');
    catId = b.category_id || '';
    err = '';
    showForm = true;
  }

  async function save() {
    err = '';
    const amt = parseFloat(amount);
    if (!amt || amt <= 0) return (err = i18n.t('common.required'));
    try {
      const body = { category_id: catId || null, amount: amt, month, year };
      if (editing) {
        await update('budgets', editing.budget_id, body);
      } else {
        await create('budgets', body);
      }
      await load();
      showForm = false;
    } catch (e) {
      err = e.message;
    }
  }

  function askDelete(b) {
    confirmDel = b;
  }

  async function doDelete() {
    const b = confirmDel;
    confirmDel = null;
    if (!b) return;
    await remove('budgets', b.budget_id);
    await load();
  }

  function catName(id) {
    return S.db.categories.find((c) => c.category_id === id)?.name || '—';
  }

  function focusInit(node) {
    node.focus();
  }

  function onKeydown(e) {
    if (e.key === 'Escape' && showForm) showForm = false;
    if (e.key === 'Escape' && confirmDel) confirmDel = null;
  }
</script>

<svelte:head><title>{i18n.t('budgets.title')} · Chiro</title></svelte:head>
<svelte:window onkeydown={onKeydown} />

<div class="page-head">
  <h1 class="headline">{i18n.t('budgets.title')}</h1>
  <button class="btn btn-primary" onclick={openNew}>+ {i18n.t('budgets.newBudget')}</button>
</div>

  <div class="month-nav">
    <button class="icon-btn" onclick={() => (month = month === 1 ? (year -= 1, 12) : month - 1)} aria-label={i18n.t('common.prevMonth')}><ChevronLeft size={20} /></button>
    <span class="month-label">{monthLabel(year, month)}</span>
    <button class="icon-btn" onclick={() => (month = month === 12 ? (year += 1, 1) : month + 1)} aria-label={i18n.t('common.nextMonth')}><ChevronRight size={20} /></button>
  </div>

{#if list.length === 0}
  <div class="card empty">
    <p class="meta">{i18n.t('budgets.empty')}</p>
    <button class="btn btn-primary" onclick={openNew}>{i18n.t('budgets.newBudget')}</button>
  </div>
{:else}
  <div class="card list-card">
    {#each list as b (b.budget_id)}
      <div class="budget-row" onclick={() => openEdit(b)} role="button" tabindex="0" onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); openEdit(b); } }}>
        <div class="row-body">
          <div class="row-title">{b.category_name || catName(b.category_id)}</div>
          <div class="bar-track" style="margin-top:6px">
            <div
              class="bar-fill"
              class:over={b.spent > b.amount}
              style="transform:scaleX({pct(b.spent, b.amount) / 100});background:{b.category_color || 'var(--indigo)'}"
            ></div>
          </div>
          <div class="row-sub" style="margin-top:4px">
            {money(b.spent)} / {money(b.amount)}
            {#if b.spent > b.amount}
              <span class="badge-danger">{i18n.t('budgets.exceeded')}</span>
            {/if}
            <button class="btn btn-small btn-cancel" style="margin-left:auto" onclick={(e) => { e.stopPropagation(); askDelete(b); }}>{i18n.t('common.delete')}</button>
          </div>
        </div>
      </div>
    {/each}
  </div>
{/if}

{#if showForm}
  <div class="overlay" onclick={() => (showForm = false)}>
    <div class="sheet" role="dialog" aria-modal="true" aria-labelledby="bud-form-title" onclick={(e) => e.stopPropagation()}>
      <h2 class="title" id="bud-form-title" style="margin-bottom:16px">
        {editing ? i18n.t('budgets.editBudget') : i18n.t('budgets.newBudget')}
      </h2>
      <div class="form-field">
        <label for="bud-cat">{i18n.t('budgets.category')}</label>
        <select id="bud-cat" bind:value={catId}>
          <option value="">{i18n.t('budgets.allCategories')}</option>
          {#each S.db.categories.filter((c) => c.type === 'expense') as c (c.category_id)}
            <option value={c.category_id}>{c.name}</option>
          {/each}
        </select>
      </div>
      <div class="form-field">
        <label for="bud-amount">{i18n.t('budgets.amount')}</label>
        <input id="bud-amount" bind:value={amount} inputmode="decimal" use:focusInit />
      </div>
      {#if err}
        <p class="error-text">{err}</p>
      {/if}
      <div class="inline-flex" style="margin-top:8px">
        <button class="btn btn-cancel" onclick={() => (showForm = false)}>{i18n.t('common.cancel')}</button>
        <button class="btn btn-primary" onclick={save}>{i18n.t('common.save')}</button>
      </div>
      {#if editing}
        <button class="btn btn-danger" style="margin-top:12px" onclick={() => { confirmDel = editing; showForm = false; }}>
          {i18n.t('common.delete')}
        </button>
      {/if}
    </div>
  </div>
{/if}

{#if confirmDel}
  <ConfirmSheet
    title={i18n.t('common.delete')}
    message={i18n.t('budgets.deleteConfirm') + (confirmDel.category_name ? ': ' + confirmDel.category_name : '')}
    confirmLabel={i18n.t('common.delete')}
    danger
    onConfirm={doDelete}
    onCancel={() => (confirmDel = null)}
  />
{/if}
