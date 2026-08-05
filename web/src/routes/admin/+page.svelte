<script>
  import { i18n } from '$lib/i18n.svelte.js';
  import { api } from '$lib/api.svelte.js';
  import { S } from '$lib/stores.svelte.js';
  import { onMount } from 'svelte';

  let stats = $state(null);
  let users = $state([]);
  let loading = $state(true);
  let err = $state('');

  onMount(async () => {
    if (S.user?.role !== 'admin') {
      err = 'Acceso denegado';
      loading = false;
      return;
    }
    try {
      const [s, u] = await Promise.all([
        api('/api/admin/stats'),
        api('/api/admin/users')
      ]);
      stats = s;
      users = u ?? [];
    } catch (e) {
      err = e.message;
    } finally {
      loading = false;
    }
  });

  async function toggleStatus(user) {
    const newStatus = user.status === 'active' ? 'disabled' : 'active';
    await api(`/api/admin/users/${user.user_id}`, {
      method: 'PUT',
      body: { status: newStatus, role: user.role }
    });
    users = users.map(u => u.user_id === user.user_id ? { ...u, status: newStatus } : u);
  }
</script>

<svelte:head><title>Admin · Chiro</title></svelte:head>

<div class="page-head">
  <h1 class="headline">Dashboard Admin</h1>
</div>

{#if loading}
  <p class="meta" style="padding:24px;text-align:center">{i18n.t('common.loading')}</p>
{:else if err}
  <div class="card empty">
    <p class="error-text">{err}</p>
  </div>
{:else}
  <div class="summary-cards">
    <div class="card stat-card">
      <div class="stat-label">Usuarios</div>
      <div class="stat-value">{stats.total_users}</div>
    </div>
    <div class="card stat-card">
      <div class="stat-label">Gastos totales</div>
      <div class="stat-value">{stats.total_expenses}</div>
    </div>
    <div class="card stat-card">
      <div class="stat-label">Préstamos</div>
      <div class="stat-value">{stats.total_loans}</div>
    </div>
  </div>

  <div class="summary-cards" style="margin-top:10px">
    <div class="card stat-card">
      <div class="stat-label">Pro</div>
      <div class="stat-value positive">{stats.pro_users}</div>
    </div>
    <div class="card stat-card">
      <div class="stat-label">Free</div>
      <div class="stat-value">{stats.free_users}</div>
    </div>
  </div>

  <div class="card list-card" style="margin-top:14px">
    <h3 class="card-title" style="padding:12px 16px 0">Usuarios</h3>
    {#each users as u (u.user_id)}
      <div class="row">
        <div class="cat-dot" style="background:{u.plan === 'pro' ? 'var(--green)' : 'var(--ink-dim)'}"></div>
        <div class="row-body">
          <div class="row-title">{u.name || u.email}</div>
          <div class="row-sub">
            <span>{u.email}</span>
            {#if u.role === 'admin'}
              <span class="tag mini">admin</span>
            {/if}
            <span class="tag mini" class:active={u.plan === 'pro'}>{u.plan}</span>
          </div>
        </div>
        <div class="row-right">
          <button class="btn btn-small" onclick={() => toggleStatus(u)}>
            {u.status === 'active' ? 'Desactivar' : 'Activar'}
          </button>
        </div>
      </div>
    {/each}
  </div>
{/if}
