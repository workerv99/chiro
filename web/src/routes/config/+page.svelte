<script>
  import { goto } from '$app/navigation';
  import { i18n } from '$lib/i18n.svelte.js';
  import { S, logout, create, update, remove, savePerson, payBill, skipBill, dueBills, activatePro, fetchSubscription } from '$lib/stores.svelte.js';
  import { toDisplay, todayISO, money } from '$lib/format.js';
  import UndoToast from '$lib/components/UndoToast.svelte';
  import ConfirmSheet from '$lib/components/ConfirmSheet.svelte';

  let section = $state('accounts');
  let showForm = $state(false);
  let editing = $state(null);
  let err = $state('');
  let due = $state([]);
  let confirmDel = $state(null);
  let undo = $state(null);
  let amoled = $state(false);
  let showUpgrade = $state(false);

  let form = $state(emptyForm());

  async function handleUpgrade() {
    try {
      await activatePro();
      showUpgrade = false;
    } catch (e) {
      err = e.message;
    }
  }

  function emptyForm() {
    return { name: '', currency: 'USD', type: 'expense', color: '#5B7CF6', notes: '', target_amount: 0, current_amount: 0, amount: 0, next_date: '', frequency: 'monthly' };
  }

  $effect(() => {
    if (typeof document !== 'undefined') {
      const saved = localStorage.getItem('chiro_amoled');
      amoled = saved === 'true';
      document.documentElement.classList.toggle('amoled', amoled);
    }
  });

  function toggleAmoled() {
    amoled = !amoled;
    document.documentElement.classList.toggle('amoled', amoled);
    localStorage.setItem('chiro_amoled', amoled);
  }

  $effect(() => {
    if (section === 'bills') dueBills().then((d) => (due = d ?? [])).catch(() => {});
  });

  const items = $derived.by(() => {
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
    form = emptyForm();
    form.next_date = todayISO();
    err = '';
    showForm = true;
  }

  function openEdit(item) {
    editing = item;
    form = { ...emptyForm(), ...item };
    err = '';
    showForm = true;
  }

  async function save() {
    err = '';
    const name = form.name.trim();
    if (!name) return (err = i18n.t('common.required'));
    try {
      if (section === 'persons') {
        const body = { name, notes: form.notes.trim() || null };
        await savePerson(editing ? { ...body, person_id: editing.person_id } : body);
      } else if (section === 'bills' && !editing) {
        await create('bills', { name, amount: form.amount, next_date: form.next_date, frequency: 'monthly', type: 'expense' });
      } else if (section === 'bills' && editing) {
        await update('bills', editing.bill_id, {
          name, amount: form.amount, next_date: editing.next_date, frequency: editing.frequency || 'monthly',
          type: editing.type || 'expense', active: editing.active ?? 1,
          category_id: editing.category_id, account_id: editing.account_id, notes: editing.notes
        });
      } else {
        const body = sectionPayload(name);
        if (editing) await update(section, editing.key, body);
        else await create(section, body);
      }
      showForm = false;
    } catch (e) {
      err = e.message;
    }
    refreshDue();
  }

  function sectionPayload(name) {
    switch (section) {
      case 'accounts': return { name, currency: (form.currency || 'USD').trim() || 'USD' };
      case 'categories': return { name, type: form.type || 'expense' };
      case 'tags': return { name, color: form.color || 'var(--indigo)' };
      case 'piggy': return { name, target_amount: Number(form.target_amount) || 0, current_amount: Number(form.current_amount) || 0 };
      default: return { name };
    }
  }

  function askDelete(item) {
    confirmDel = item;
  }

  async function doDelete() {
    const item = confirmDel;
    confirmDel = null;
    if (!item) return;
    try {
      const snapshot = structuredClone(item);
      await remove(section, item.key);
      const label = i18n.t('common.deleted') || 'Eliminado';
      const msg = `${label}: ${item.label}`;
      undo = { section, item: snapshot, msg, t: 0 };
      const timer = setInterval(() => {
        undo = undo ? { ...undo, t: undo.t + 1 } : null;
        if (!undo || undo.t >= 5) { clearInterval(timer); undo = null; }
      }, 1000);
      refreshDue();
    } catch (e) {
      err = e.message;
    }
  }

  async function doUndo() {
    if (!undo) return;
    const { section: s, item } = undo;
    undo = null;
    try {
      await create(s, { ...item, deleted: 0 });
    } catch (e) {
      err = e.message;
    }
    refreshDue();
  }

  function focusInit(node) {
    node.focus();
  }

  function onKeydown(e) {
    if (e.key === 'Escape' && showForm) showForm = false;
    if (e.key === 'Escape' && confirmDel) confirmDel = null;
  }

  function refreshDue() {
    if (section === 'bills') dueBills().then((d) => (due = d ?? [])).catch(() => {});
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
  <div class="inline-flex">
    <button class="chip" class:active={amoled} onclick={toggleAmoled}>AMOLED</button>
    <button class="btn btn-primary" onclick={openNew}>+ {i18n.t('common.add')}</button>
  </div>
</div>

<div class="card" style="padding:16px;margin-bottom:14px">
  <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:12px">
    <div>
      <h3 class="card-title" style="margin-bottom:2px">Plan {S.subscription.plan === 'pro' ? 'Pro' : 'Free'}</h3>
      <p class="meta">
        {#if S.subscription.plan === 'pro'}
          Gastos ilimitados, cuentas ilimitadas, préstamos ilimitados
        {:else}
          {S.subscription.usage?.expenses_this_month || 0}/50 gastos · {S.subscription.usage?.total_accounts || 0}/3 cuentas · {S.subscription.usage?.total_loans || 0}/10 préstamos
        {/if}
      </p>
    </div>
    {#if S.subscription.plan === 'free'}
      <a href="#upgrade" class="btn btn-primary btn-small" onclick={(e) => { e.preventDefault(); showUpgrade = true; }}>Upgrade a Pro</a>
    {/if}
  </div>
  {#if S.subscription.plan === 'free'}
    <div class="progress-bar" style="margin-top:8px">
      <div class="progress-fill" style="width:{Math.min(100, ((S.subscription.usage?.expenses_this_month || 0) / 50) * 100)}%"></div>
    </div>
  {/if}
</div>

{#if showUpgrade}
  <div class="overlay" onclick={() => (showUpgrade = false)}>
    <div class="sheet" role="dialog" aria-modal="true" onclick={(e) => e.stopPropagation()}>
      <div class="handle"></div>
      <h2 class="title" style="margin-bottom:16px">Actualizar a Pro</h2>
      <div style="text-align:center;padding:16px 0">
        <div style="font-size:2rem;font-weight:800;color:var(--indigo);margin-bottom:8px">$4.99/mes</div>
        <p class="meta" style="margin-bottom:24px">Gastos, cuentas y préstamos ilimitados</p>
        <ul style="list-style:none;padding:0;text-align:left;max-width:280px;margin:0 auto 24px">
          <li style="padding:8px 0;border-bottom:1px solid var(--border)">✓ Gastos ilimitados por mes</li>
          <li style="padding:8px 0;border-bottom:1px solid var(--border)">✓ Cuentas ilimitadas</li>
          <li style="padding:8px 0;border-bottom:1px solid var(--border)">✓ Préstamos ilimitados</li>
          <li style="padding:8px 0;border-bottom:1px solid var(--border)">✓ Reportes PDF</li>
          <li style="padding:8px 0">✓ Soporte prioritario</li>
        </ul>
        <button class="btn btn-primary" style="width:100%" onclick={handleUpgrade}>Activar Pro (simulado)</button>
        <button class="btn btn-cancel" style="width:100%;margin-top:8px" onclick={() => (showUpgrade = false)}>Cancelar</button>
      </div>
    </div>
  </div>
{/if}

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
  {#if items.length === 0}
    <div class="empty"><p class="meta">{i18n.t('config.empty')}</p></div>
  {:else}
    {#each items as item (item.key)}
      <a class="row" href="#" onclick={(e) => { e.preventDefault(); openEdit(item); }}>
        <div class="cat-dot" style="background:{item.color || 'var(--indigo)'}"></div>
        <div class="row-body">
          <div class="row-title">{item.label}</div>
          <div class="row-sub">{item.sub}</div>
        </div>
        <button class="btn btn-small btn-cancel" onclick={(e) => { e.stopPropagation(); e.preventDefault(); askDelete(item); }}>{i18n.t('common.delete')}</button>
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
        <label class="eyebrow" for="cfg-name">{i18n.t('config.name')}</label>
        <input id="cfg-name" bind:value={form.name} use:focusInit />
      </div>

      {#if section === 'accounts'}
        <div class="form-field">
          <label class="eyebrow" for="cfg-currency">{i18n.t('config.currency')}</label>
          <input id="cfg-currency" bind:value={form.currency} placeholder="USD" />
        </div>
      {:else if section === 'categories'}
        <div class="form-field">
          <label class="eyebrow" for="cfg-type">{i18n.t('config.type')}</label>
          <select id="cfg-type" bind:value={form.type}>
            <option value="expense">{i18n.t('common.expense')}</option>
            <option value="income">{i18n.t('common.income')}</option>
          </select>
        </div>
      {:else if section === 'tags'}
        <div class="form-field">
          <label class="eyebrow" for="cfg-color">{i18n.t('config.color')}</label>
          <div class="inline-flex" style="align-items:center;gap:10px">
            <input id="cfg-color" bind:value={form.color} type="color" class="color-swatch" />
            <code class="meta">{form.color}</code>
          </div>
        </div>
      {:else if section === 'piggy'}
        <div class="grid2">
          <div class="form-field">
            <label class="eyebrow" for="cfg-target">{i18n.t('config.target')}</label>
            <input id="cfg-target" type="number" inputmode="decimal" step="0.01" min="0" bind:value={form.target_amount} />
          </div>
          <div class="form-field">
            <label class="eyebrow" for="cfg-current">{i18n.t('config.current')}</label>
            <input id="cfg-current" type="number" inputmode="decimal" step="0.01" min="0" bind:value={form.current_amount} />
          </div>
        </div>
      {:else if section === 'bills'}
        <div class="form-field">
          <label class="eyebrow" for="cfg-amount">{i18n.t('config.amount')}</label>
          <input id="cfg-amount" type="number" inputmode="decimal" step="0.01" min="0" bind:value={form.amount} />
        </div>
      {:else if section === 'persons'}
        <div class="form-field">
          <label class="eyebrow" for="cfg-notes">{i18n.t('config.notes')}</label>
          <input id="cfg-notes" bind:value={form.notes} />
        </div>
      {/if}

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

{#if confirmDel}
  <ConfirmSheet
    title={i18n.t('common.delete')}
    message={i18n.t('config.deleteConfirm') + ': ' + confirmDel.label}
    confirmLabel={i18n.t('common.delete')}
    danger
    onConfirm={doDelete}
    onCancel={() => (confirmDel = null)}
  />
{/if}

{#if undo}
  <UndoToast message={undo.msg} secondsLeft={5 - undo.t} onUndo={doUndo} onClose={() => (undo = null)} />
{/if}

<div class="card" style="padding:16px;margin-top:16px">
  <h3 class="card-title">{i18n.t('config.session')}</h3>
  {#if S.user}
    <p class="meta">{i18n.t('config.loggedAs')}: {S.user.name} ({S.user.email})</p>
  {/if}
  <button class="btn btn-cancel" onclick={() => { logout(); goto('/login'); }}>{i18n.t('common.logout')}</button>
</div>
