<script>
  import { i18n } from '$lib/i18n.svelte.js';
  import { api } from '$lib/api.svelte.js';
  import { S, create, update, remove } from '$lib/stores.svelte.js';
  import { money, pct } from '$lib/format.js';

  const now = new Date();
  let year = $state(now.getFullYear());
  let month = $state(now.getMonth() + 1);
  let list = $state([]);
  let showForm = $state(false);
  let editing = $state(null);
  let catId = $state('');
  let amount = $state('');
  let err = $state('');

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

  async function del(b) {
    if (!confirm(i18n.t('budgets.deleteConfirm'))) return;
    await remove('budgets', b.budget_id);
    await load();
  }

  function catName(id) {
    return S.db.categories.find((c) => c.category_id === id)?.name || '—';
  }
</script>

<div class="page-head">
  <h1 class="headline">{i18n.t('budgets.title')}</h1>
  <button class="btn btn-primary" onclick={openNew}>+ {i18n.t('budgets.newBudget')}</button>
</div>

<div class="month-nav">
  <button class="icon-btn" onclick={() => (month = month === 1 ? (year -= 1, 12) : month - 1)}>‹</button>
  <div class="month-label">{year}-{String(month).padStart(2, '0')}</div>
  <button class="icon-btn" onclick={() => (month = month === 12 ? (year += 1, 1) : month + 1)}>›</button>
</div>

{#if list.length === 0}
  <div class="card empty">
    <p class="meta">{i18n.t('budgets.empty')}</p>
    <button class="btn btn-primary" onclick={openNew}>{i18n.t('budgets.newBudget')}</button>
  </div>
{:else}
  <div class="card list-card">
    {#each list as b (b.budget_id)}
      <div class="budget-row" onclick={() => openEdit(b)}>
        <div class="row-body">
          <div class="row-title">{b.category_name || catName(b.category_id)}</div>
          <div class="bar-track" style="margin-top:6px">
            <div
              class="bar-fill"
              class:over={b.spent > b.amount}
              style="width:{pct(b.spent, b.amount)}%;background:{b.category_color || '#5b7cf6'}"
            ></div>
          </div>
          <div class="row-sub" style="margin-top:4px">
            {money(b.spent)} / {money(b.amount)}
            {#if b.spent > b.amount}
              <span class="badge-danger">{i18n.t('budgets.exceeded')}</span>
            {/if}
          </div>
        </div>
      </div>
    {/each}
  </div>
{/if}

{#if showForm}
  <div class="overlay" onclick={() => (showForm = false)}>
    <div class="sheet" onclick={(e) => e.stopPropagation()}>
      <h2 class="title" style="margin-bottom:16px">
        {editing ? i18n.t('budgets.editBudget') : i18n.t('budgets.newBudget')}
      </h2>
      <div class="form-field">
        <label>{i18n.t('budgets.category')}</label>
        <select bind:value={catId}>
          <option value="">{i18n.t('budgets.allCategories')}</option>
          {#each S.db.categories.filter((c) => c.type === 'expense') as c (c.category_id)}
            <option value={c.category_id}>{c.name}</option>
          {/each}
        </select>
      </div>
      <div class="form-field">
        <label>{i18n.t('budgets.amount')}</label>
        <input bind:value={amount} inputmode="decimal" />
      </div>
      {#if err}
        <p class="error-text">{err}</p>
      {/if}
      <div class="inline-flex" style="margin-top:8px">
        <button class="btn btn-cancel" onclick={() => (showForm = false)}>{i18n.t('common.cancel')}</button>
        <button class="btn btn-primary" onclick={save}>{i18n.t('common.save')}</button>
      </div>
      {#if editing}
        <button class="btn btn-danger" style="margin-top:12px" onclick={() => { del(editing); showForm = false; }}>
          {i18n.t('common.delete')}
        </button>
      {/if}
    </div>
  </div>
{/if}
