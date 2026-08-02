<script>
  import { goto } from '$app/navigation';
  import { i18n } from '$lib/i18n.svelte.js';
  import { S, logout, create, update, remove, savePerson, payBill, skipBill, dueBills } from '$lib/stores.svelte.js';
  import { toDisplay, todayISO, toISO, money } from '$lib/format.js';

  let section = $state('accounts');
  let showForm = $state(false);
  let editing = $state(null);
  let name = $state('');
  let extra = $state('');
  let extra2 = $state('');
  let err = $state('');
  let due = $state([]);

  $effect(() => {
    if (section === 'bills') dueBills().then((d) => (due = d ?? [])).catch(() => {});
  });

  const items = $derived(() => {
    switch (section) {
      case 'accounts': return S.db.accounts.map((a) => ({ ...a, key: a.account_id, label: a.name, sub: a.currency }));
      case 'categories': return S.db.categories.map((c) => ({ ...c, key: c.category_id, label: c.name, sub: c.type, color: c.color }));
      case 'persons': return S.db.persons.map((p) => ({ ...p, key: p.person_id, label: p.name, sub: p.notes || '' }));
      case 'tags': return S.db.tags.map((x) => ({ ...x, key: x.tag_id, label: x.name, sub: x.color, color: x.color }));
      case 'piggy': return S.db.piggy.map((p) => ({ ...p, key: p.piggy_bank_id, label: p.name, sub: `${money(p.target_amount)} · ${money(p.current_amount)}`, color: p.color }));
      case 'bills': return S.db.bills.map((b) => ({ ...b, key: b.bill_id, label: b.name, sub: `${b.frequency || 'monthly'} · ${toDisplay(b.next_date)}` }));
      default: return [];
    }
  });

  const tabs = $derived([
    { id: 'accounts', label: i18n.t('config.accounts') },
    { id: 'categories', label: i18n.t('config.categories') },
    { id: 'persons', label: i18n.t('config.persons') },
    { id: 'tags', label: i18n.t('config.tags') },
    { id: 'piggy', label: i18n.t('config.piggy') },
    { id: 'bills', label: i18n.t('config.bills') }
  ]);

  function openNew() {
    editing = null;
    name = '';
    extra = '';
    extra2 = '';
    err = '';
    showForm = true;
  }

  function openEdit(item) {
    editing = item;
    name = item.label;
    extra = item.sub || '';
    if (section === 'categories') extra = item.type;
    if (section === 'bills') extra = String(item.amount ?? '');
    if (section === 'piggy') { extra = String(item.target_amount ?? ''); extra2 = String(item.current_amount ?? ''); }
    err = '';
    showForm = true;
  }

  async function save() {
    err = '';
    if (!name.trim()) return (err = i18n.t('common.required'));
    try {
      const body = { name: name.trim() };
      if (section === 'accounts') body.currency = extra.trim() || 'USD';
      if (section === 'categories') body.type = extra || 'expense';
      if (section === 'persons') body.notes = extra.trim() || null;
      if (section === 'tags') body.color = extra.trim() || '#5B7CF6';
      if (section === 'piggy') {
        body.target_amount = parseFloat(extra) || 0;
        body.current_amount = parseFloat(extra2) || 0;
      }
      if (section === 'bills') {
        body.amount = parseFloat(extra) || 0;
        if (editing) {
          body.next_date = editing.next_date;
          body.frequency = editing.frequency || 'monthly';
          body.type = editing.type || 'expense';
          body.active = editing.active ?? 1;
          for (const k of ['category_id', 'account_id', 'notes']) {
            if (editing[k] != null) body[k] = editing[k];
          }
        }
      }
      if (editing) {
        if (section === 'persons') await savePerson({ ...body, person_id: editing.person_id });
        else await update(section, editing.key, body);
      } else if (section === 'persons') {
        await savePerson(body);
      } else if (section === 'bills') {
        await create(section, { ...body, next_date: toISO(todayISO()), frequency: 'monthly', type: 'expense' });
      } else {
        await create(section, body);
      }
      showForm = false;
    } catch (e) {
      err = e.message;
    }
    refreshDue();
  }

  async function del(item) {
    if (!confirm(i18n.t('config.deleteConfirm'))) return;
    await remove(section, item.key);
    refreshDue();
  }

  function refreshDue() {
    if (section === 'bills') dueBills().then((d) => (due = d ?? [])).catch(() => {});
  }

  function focusInit(node) {
    node.focus();
  }

  function onKeydown(e) {
    if (e.key === 'Escape' && showForm) showForm = false;
  }
</script>

<svelte:head><title>{i18n.t('config.title')} · Chiro</title></svelte:head>
<svelte:window onkeydown={onKeydown} />

<div class="page-head">
  <h1 class="headline">{i18n.t('config.title')}</h1>
</div>

<div class="tabs" role="tablist">
  {#each tabs as tab (tab.id)}
    <button class="tab" class:active={section === tab.id} role="tab" aria-selected={section === tab.id} onclick={() => (section = tab.id)}>{tab.label}</button>
  {/each}
</div>

<div class="page-head" style="margin-top:10px">
  <div class="lang-toggle">
    <button class="chip" class:active={i18n.lang === 'es'} onclick={() => i18n.setLang('es')}>ES</button>
    <button class="chip" class:active={i18n.lang === 'en'} onclick={() => i18n.setLang('en')}>EN</button>
  </div>
  <button class="btn btn-primary" onclick={openNew}>+ {i18n.t('common.add')}</button>
</div>

{#if section === 'bills'}
  <div class="card" style="padding:16px;margin-bottom:14px">
    <h3 class="card-title">{i18n.t('config.dueBills')}</h3>
    {#if due.length === 0}
      <p class="meta">{i18n.t('config.noDueBills')}</p>
    {:else}
      {#each due as b (b.bill_id)}
        <div class="row">
          <div class="row-body">
            <div class="row-title">{b.name}</div>
            <div class="row-sub">{money(b.amount)} · {toDisplay(b.next_date)}</div>
          </div>
          <div class="row-right">
            <button class="btn btn-small" onclick={() => payBill(b.bill_id).then(refreshDue)}>{i18n.t('common.pay')}</button>
            <button class="btn btn-small btn-cancel" onclick={() => skipBill(b.bill_id, b.next_date, b.frequency).then(refreshDue)}>{i18n.t('common.skip')}</button>
          </div>
        </div>
      {/each}
    {/if}
  </div>
{/if}

<div class="card list-card">
  {#if items().length === 0}
    <div class="empty"><p class="meta">{i18n.t('config.empty')}</p></div>
  {:else}
    {#each items() as item (item.key)}
      <a class="row" href="#" onclick={(e) => { e.preventDefault(); openEdit(item); }}>
        <div class="cat-dot" style="background:{item.color || 'var(--indigo)'}"></div>
        <div class="row-body">
          <div class="row-title">{item.label}</div>
          <div class="row-sub">{item.sub}</div>
        </div>
        <button class="btn btn-small btn-cancel" onclick={(e) => { e.stopPropagation(); e.preventDefault(); del(item); }}>{i18n.t('common.delete')}</button>
      </a>
    {/each}
  {/if}
</div>

{#if showForm}
  <div class="overlay" onclick={() => (showForm = false)}>
    <div class="sheet" role="dialog" aria-modal="true" aria-labelledby="cfg-form-title" onclick={(e) => e.stopPropagation()}>
      <h2 class="title" id="cfg-form-title" style="margin-bottom:16px">
        {editing ? i18n.t('config.editItem') : i18n.t('common.add')}
      </h2>
      <div class="form-field">
        <label for="cfg-name">{i18n.t('config.name')}</label>
        <input id="cfg-name" bind:value={name} use:focusInit />
      </div>
      {#if section === 'accounts'}
        <div class="form-field">
          <label for="cfg-currency">{i18n.t('config.currency')}</label>
          <input id="cfg-currency" bind:value={extra} placeholder="USD" />
        </div>
      {:else if section === 'categories'}
        <div class="form-field">
          <label for="cfg-type">{i18n.t('config.type')}</label>
          <select id="cfg-type" bind:value={extra}>
            <option value="expense">{i18n.t('common.expense')}</option>
            <option value="income">{i18n.t('common.income')}</option>
          </select>
        </div>
      {:else if section === 'tags'}
        <div class="form-field">
          <label for="cfg-color">{i18n.t('config.color')}</label>
          <input id="cfg-color" bind:value={extra} type="color" />
        </div>
      {:else if section === 'piggy'}
        <div class="grid2">
          <div class="form-field">
            <label for="cfg-target">{i18n.t('config.target')}</label>
            <input id="cfg-target" bind:value={extra} inputmode="decimal" />
          </div>
          <div class="form-field">
            <label for="cfg-current">{i18n.t('config.current')}</label>
            <input id="cfg-current" bind:value={extra2} inputmode="decimal" />
          </div>
        </div>
      {:else if section === 'bills'}
        <div class="form-field">
          <label for="cfg-amount">{i18n.t('config.amount')}</label>
          <input id="cfg-amount" bind:value={extra} inputmode="decimal" />
        </div>
      {:else if section === 'persons'}
        <div class="form-field">
          <label for="cfg-notes">{i18n.t('config.notes')}</label>
          <input id="cfg-notes" bind:value={extra} />
        </div>
      {/if}
      {#if err}
        <p class="error-text">{err}</p>
      {/if}
      <div class="inline-flex" style="margin-top:8px">
        <button class="btn btn-cancel" onclick={() => (showForm = false)}>{i18n.t('common.cancel')}</button>
        <button class="btn btn-primary" onclick={save}>{i18n.t('common.save')}</button>
      </div>
      {#if editing}
        <button class="btn btn-danger" style="margin-top:12px" onclick={() => { del(editing); showForm = false; }}>{i18n.t('common.delete')}</button>
      {/if}
    </div>
  </div>
{/if}

<div class="card" style="padding:16px;margin-top:16px">
  <h3 class="card-title">{i18n.t('config.session')}</h3>
  {#if S.user}
    <p class="meta">{i18n.t('config.loggedAs')}: {S.user.name} ({S.user.email})</p>
  {/if}
  <button class="btn btn-cancel" onclick={() => { logout(); goto('/login'); }}>{i18n.t('common.logout')}</button>
</div>
