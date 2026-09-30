<script>
  import { i18n } from '$lib/i18n.svelte.js';
  import { api } from '$lib/api.svelte.js';
  import { S } from '$lib/stores.svelte.js';
  import { onMount } from 'svelte';
  import Card from '$lib/components/ui/card.svelte';
  import Button from '$lib/components/ui/button.svelte';

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

<h1 class="text-xl md:text-2xl font-bold mb-4">Dashboard Admin</h1>

{#if loading}
  <p class="text-sm text-muted-foreground py-8 text-center">{i18n.t('common.loading')}</p>
{:else if err}
  <Card class="p-4 text-center">
    <p class="text-sm text-destructive">{err}</p>
  </Card>
{:else}
  <div class="grid grid-cols-2 md:grid-cols-3 gap-3 mb-3">
    <Card class="p-4">
      <p class="text-xs font-bold text-muted-foreground uppercase mb-1">Usuarios</p>
      <p class="text-2xl font-extrabold">{stats.total_users}</p>
    </Card>
    <Card class="p-4">
      <p class="text-xs font-bold text-muted-foreground uppercase mb-1">Gastos totales</p>
      <p class="text-2xl font-extrabold">{stats.total_expenses}</p>
    </Card>
    <Card class="p-4">
      <p class="text-xs font-bold text-muted-foreground uppercase mb-1">Préstamos</p>
      <p class="text-2xl font-extrabold">{stats.total_loans}</p>
    </Card>
  </div>

  <div class="grid grid-cols-2 gap-3 mb-4">
    <Card class="p-4">
      <p class="text-xs font-bold text-muted-foreground uppercase mb-1">Pro</p>
      <p class="text-2xl font-extrabold text-green-500">{stats.pro_users}</p>
    </Card>
    <Card class="p-4">
      <p class="text-xs font-bold text-muted-foreground uppercase mb-1">Free</p>
      <p class="text-2xl font-extrabold">{stats.free_users}</p>
    </Card>
  </div>

  <Card class="overflow-hidden">
    <h3 class="font-bold p-4 pb-0">Usuarios</h3>
    {#each users as u (u.user_id)}
      <div class="flex items-center gap-3 px-4 py-3 border-t first:border-t-0">
        <div class="h-2.5 w-2.5 rounded-full" style="background:{u.plan === 'pro' ? '#22c55e' : '#94a3b8'}"></div>
        <div class="flex-1 min-w-0">
          <p class="text-sm font-semibold truncate" title={u.name || u.email}>{u.name || u.email}</p>
          <p class="text-xs text-muted-foreground truncate" title={u.email}>
            {u.email}
            {#if u.role === 'admin'}
              <span class="inline-flex items-center rounded-full bg-primary/10 px-2 py-0.5 text-xs font-bold text-primary ml-1">admin</span>
            {/if}
            <span class="inline-flex items-center rounded-full bg-secondary px-2 py-0.5 text-xs font-bold text-secondary-foreground ml-1">{u.plan}</span>
          </p>
        </div>
        <Button variant="outline" size="sm" onclick={() => toggleStatus(u)}>
          {u.status === 'active' ? 'Desactivar' : 'Activar'}
        </Button>
      </div>
    {/each}
  </Card>
{/if}
