<script>
  import { goto } from '$app/navigation';
  import { i18n } from '$lib/i18n.svelte.js';
  import { S, logout, create, update, remove, savePerson, payBill, skipBill, dueBills, activatePro, fetchSubscription, deleteAccount, exportData } from '$lib/stores.svelte.js';
  import { toDisplay, todayISO, money } from '$lib/format.js';
  import UndoToast from '$lib/components/UndoToast.svelte';
  import ConfirmSheet from '$lib/components/ConfirmSheet.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Card from '$lib/components/ui/card.svelte';
  import Input from '$lib/components/ui/input.svelte';
  import Label from '$lib/components/ui/label.svelte';
  import Badge from '$lib/components/ui/badge.svelte';

  let section = $state('accounts');
  let showForm = $state(false);
  let editing = $state(null);
  let err = $state('');
  let due = $state([]);
  let confirmDel = $state(null);
  let undo = $state(null);
  let showUpgrade = $state(false);
  let confirmDeleteAccount = $state(false);

  let form = $state(emptyForm());

  async function handleUpgrade() {
    try {
      await activatePro();
      showUpgrade = false;
    } catch (e) {
      err = e.message;
    }
  }

  async function handleDeleteAccount() {
    try {
      await deleteAccount();
      goto('/login');
    } catch (e) {
      err = e.message;
    }
  }

  function emptyForm() {
    return { name: '', currency: 'USD', type: 'expense', color: '#5B7CF6', notes: '', target_amount: 0, current_amount: 0, amount: 0, next_date: '', frequency: 'monthly' };
  }

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
    if (!form.name.trim()) return (err = i18n.t('common.required'));
    try {
      if (editing) await update(section, editing.key, form);
      else await create(section, form);
      showForm = false;
    } catch (e) {
      err = e.message;
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

  function refreshDue() {
    if (section === 'bills') dueBills().then((d) => (due = d ?? [])).catch(() => {});
  }

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
</script>

<svelte:head><title>{i18n.t('config.title')} · Chiro</title></svelte:head>

<div class="flex items-center justify-between mb-4">
  <h1 class="text-2xl font-bold">{i18n.t('config.title')}</h1>
</div>

<div class="flex gap-2 overflow-x-auto mb-4 pb-2">
  {#each tabs as tab (tab.id)}
    <Button
      variant={section === tab.id ? 'default' : 'outline'}
      size="sm"
      onclick={() => (section = tab.id)}
    >
      {tab.label}
    </Button>
  {/each}
</div>

<div class="flex items-center justify-between mb-4">
  <div class="flex gap-2">
    <Button variant={i18n.lang === 'es' ? 'default' : 'outline'} size="sm" onclick={() => i18n.setLang('es')}>ES</Button>
    <Button variant={i18n.lang === 'en' ? 'default' : 'outline'} size="sm" onclick={() => i18n.setLang('en')}>EN</Button>
  </div>
  <Button onclick={openNew}>+ {i18n.t('common.add')}</Button>
</div>

<Card class="p-4 mb-4">
  <div class="flex justify-between items-start mb-3">
    <div>
      <h3 class="font-bold">Plan {S.subscription.plan === 'pro' ? 'Pro' : 'Free'}</h3>
      <p class="text-xs text-muted-foreground">
        {#if S.subscription.plan === 'pro'}
          Gastos ilimitados, cuentas ilimitadas, préstamos ilimitados
        {:else}
          {S.subscription.usage?.expenses_this_month || 0}/50 gastos · {S.subscription.usage?.total_accounts || 0}/3 cuentas · {S.subscription.usage?.total_loans || 0}/10 préstamos
        {/if}
      </p>
    </div>
    {#if S.subscription.plan === 'free'}
      <Button size="sm" onclick={() => (showUpgrade = true)}>Upgrade a Pro</Button>
    {/if}
  </div>
  {#if S.subscription.plan === 'free'}
    <div class="h-1.5 bg-border rounded-full overflow-hidden">
      <div class="h-full bg-primary rounded-full transition-all" style="width:{Math.min(100, ((S.subscription.usage?.expenses_this_month || 0) / 50) * 100)}%"></div>
    </div>
  {/if}
</Card>

{#if section === 'bills'}
  <Card class="p-4 mb-4">
    <h3 class="font-bold mb-3">{i18n.t('config.dueBills')}</h3>
    {#if due.length === 0}
      <p class="text-sm text-muted-foreground">{i18n.t('config.noDueBills')}</p>
    {:else}
      {#each due as b (b.bill_id)}
        <div class="flex items-center justify-between py-2 border-t first:border-t-0">
          <div>
            <p class="text-sm font-semibold">{b.name}</p>
            <p class="text-xs text-muted-foreground">{money(b.amount)} · {toDisplay(b.next_date)}</p>
          </div>
          <div class="flex gap-2">
            <Button size="sm" onclick={() => payBill(b.bill_id).then(refreshDue)}>Pagar</Button>
            <Button size="sm" variant="outline" onclick={() => skipBill(b.bill_id, b.next_date, b.frequency).then(refreshDue)}>Saltar</Button>
          </div>
        </div>
      {/each}
    {/if}
  </Card>
{/if}

<Card class="overflow-hidden">
  {#if items.length === 0}
    <div class="flex items-center justify-center py-8">
      <p class="text-sm text-muted-foreground">{i18n.t('config.empty')}</p>
    </div>
  {:else}
    {#each items as item (item.key)}
      <button class="flex items-center gap-3 w-full px-4 py-3 text-left hover:bg-muted/50 transition-colors border-t first:border-t-0" onclick={() => openEdit(item)}>
        <div class="h-2.5 w-2.5 rounded-full" style="background:{item.color || 'var(--primary)'}"></div>
        <div class="flex-1 min-w-0">
          <p class="text-sm font-semibold truncate">{item.label}</p>
          <p class="text-xs text-muted-foreground truncate">{item.sub}</p>
        </div>
        <Button variant="ghost" size="sm" class="text-destructive" onclick={(e) => { e.stopPropagation(); askDelete(item); }}>
          {i18n.t('common.delete')}
        </Button>
      </button>
    {/each}
  {/if}
</Card>

{#if showForm}
  <div class="fixed inset-0 z-50 bg-black/80 flex items-end sm:items-center justify-center">
    <div class="w-full max-w-md bg-background border rounded-t-2xl sm:rounded-2xl p-6 max-h-[92vh] overflow-y-auto" onclick={(e) => e.stopPropagation()}>
      <h2 class="text-lg font-bold mb-4">
        {editing ? i18n.t('config.editItem') : i18n.t('common.add')}
      </h2>

      <div class="space-y-4">
        <div class="space-y-2">
          <Label for="cfg-name">{i18n.t('config.name')}</Label>
          <Input id="cfg-name" bind:value={form.name} />
        </div>

        {#if section === 'accounts'}
          <div class="space-y-2">
            <Label for="cfg-currency">{i18n.t('config.currency')}</Label>
            <Input id="cfg-currency" bind:value={form.currency} placeholder="USD" />
          </div>
        {:else if section === 'categories'}
          <div class="grid grid-cols-2 gap-4">
            <div class="space-y-2">
              <Label for="cfg-type">{i18n.t('config.type')}</Label>
              <select id="cfg-type" bind:value={form.type} class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
                <option value="expense">{i18n.t('common.expense')}</option>
                <option value="income">{i18n.t('common.income')}</option>
              </select>
            </div>
            <div class="space-y-2">
              <Label for="cfg-color">{i18n.t('config.color')}</Label>
              <Input id="cfg-color" type="color" bind:value={form.color} class="h-10 p-1 cursor-pointer" />
            </div>
          </div>
        {:else if section === 'piggy'}
          <div class="grid grid-cols-2 gap-4">
            <div class="space-y-2">
              <Label for="cfg-target">{i18n.t('config.target')}</Label>
              <Input id="cfg-target" type="number" inputmode="decimal" step="0.01" min="0" bind:value={form.target_amount} />
            </div>
            <div class="space-y-2">
              <Label for="cfg-current">{i18n.t('config.current')}</Label>
              <Input id="cfg-current" type="number" inputmode="decimal" step="0.01" min="0" bind:value={form.current_amount} />
            </div>
          </div>
        {:else if section === 'bills'}
          <div class="space-y-2">
            <Label for="cfg-amount">{i18n.t('config.amount')}</Label>
            <Input id="cfg-amount" type="number" inputmode="decimal" step="0.01" min="0" bind:value={form.amount} />
          </div>
        {:else if section === 'persons'}
          <div class="space-y-2">
            <Label for="cfg-notes">{i18n.t('config.notes')}</Label>
            <Input id="cfg-notes" bind:value={form.notes} />
          </div>
        {/if}

        {#if err}
          <p class="text-sm text-destructive">{err}</p>
        {/if}

        <div class="flex gap-2 justify-end">
          <Button variant="outline" onclick={() => (showForm = false)}>{i18n.t('common.cancel')}</Button>
          <Button onclick={save}>{i18n.t('common.save')}</Button>
        </div>
      </div>
    </div>
  </div>
{/if}

<Card class="p-4 mt-4">
  <h3 class="font-bold mb-2">{i18n.t('config.session')}</h3>
  {#if S.user}
    <p class="text-sm text-muted-foreground mb-3">{i18n.t('config.loggedAs')}: {S.user.name} ({S.user.email})</p>
  {/if}
  <Button variant="outline" onclick={() => { logout(); goto('/login'); }}>{i18n.t('common.logout')}</Button>
</Card>

<Card class="p-4 mt-4 border-destructive/50">
  <h3 class="font-bold mb-2">Tus datos (GDPR)</h3>
  <p class="text-sm text-muted-foreground mb-3">Exportá o eliminá tu cuenta y datos personales.</p>
  <div class="flex gap-2">
    <Button variant="outline" size="sm" onclick={exportData}>Exportar datos</Button>
    <Button variant="destructive" size="sm" onclick={() => (confirmDeleteAccount = true)}>Eliminar cuenta</Button>
  </div>
</Card>

{#if confirmDel}
  <ConfirmSheet
    bind:open={() => confirmDel !== null, (v) => confirmDel = v ? confirmDel : null}
    title={i18n.t('common.delete')}
    message={i18n.t('config.deleteConfirm') + ': ' + confirmDel.label}
    confirmLabel={i18n.t('common.delete')}
    danger
    onConfirm={doDelete}
    onCancel={() => (confirmDel = null)}
  />
{/if}

{#if confirmDeleteAccount}
  <ConfirmSheet
    bind:open={() => confirmDeleteAccount, (v) => confirmDeleteAccount = v}
    title="Eliminar cuenta"
    message="Esta acción es permanente. Se eliminarán todos tus datos, gastos, préstamos y configuración. ¿Estás seguro?"
    confirmLabel="Eliminar mi cuenta"
    danger
    onConfirm={handleDeleteAccount}
    onCancel={() => (confirmDeleteAccount = false)}
  />
{/if}

{#if showUpgrade}
  <div class="fixed inset-0 z-50 bg-black/80 flex items-end sm:items-center justify-center">
    <div class="w-full max-w-md bg-background border rounded-t-2xl sm:rounded-2xl p-6" onclick={(e) => e.stopPropagation()}>
      <h2 class="text-lg font-bold mb-4">Actualizar a Pro</h2>
      <div class="text-center py-4">
        <p class="text-3xl font-extrabold text-primary mb-2">$4.99/mes</p>
        <p class="text-sm text-muted-foreground mb-6">Gastos, cuentas y préstamos ilimitados</p>
        <ul class="text-sm text-left space-y-2 mb-6">
          <li class="flex items-center gap-2"><span class="text-green-500">✓</span> Gastos ilimitados por mes</li>
          <li class="flex items-center gap-2"><span class="text-green-500">✓</span> Cuentas ilimitadas</li>
          <li class="flex items-center gap-2"><span class="text-green-500">✓</span> Préstamos ilimitados</li>
          <li class="flex items-center gap-2"><span class="text-green-500">✓</span> Reportes PDF</li>
          <li class="flex items-center gap-2"><span class="text-green-500">✓</span> Soporte prioritario</li>
        </ul>
        <Button class="w-full" onclick={handleUpgrade}>Activar Pro (simulado)</Button>
        <Button variant="ghost" class="w-full mt-2" onclick={() => (showUpgrade = false)}>Cancelar</Button>
      </div>
    </div>
  </div>
{/if}

{#if undo}
  <UndoToast message={undo.msg} secondsLeft={5 - undo.t} onUndo={doUndo} onClose={() => (undo = null)} />
{/if}
