<script>
  import '../app.css';
  import { page } from '$app/state';
  import { goto } from '$app/navigation';
  import { i18n } from '$lib/i18n.svelte.js';
  import { S, me, fetchAll, logout, loadMonth, fetchSubscription } from '$lib/stores.svelte.js';
  import { A, loadToken } from '$lib/api.svelte.js';

  let ready = $state(false);

  $effect(() => {
    if (typeof document !== 'undefined') document.documentElement.lang = i18n.lang;
  });

  $effect(() => {
    loadToken();
    if (!A.token) {
      if (page.url.pathname !== '/login' && page.url.pathname !== '/' && !page.url.pathname.startsWith('/legal')) {
        goto('/login');
      }
      ready = true;
      return;
    }
    (async () => {
      const u = await me();
      if (!u) {
        goto('/login');
        ready = true;
        return;
      }
      if (page.url.pathname === '/login' || page.url.pathname === '/') {
        goto('/dashboard');
      }
      try {
        await Promise.all([fetchAll(), loadMonth(new Date().getFullYear(), new Date().getMonth() + 1), fetchSubscription()]);
      } catch { /* ignore */ }
      ready = true;
    })();
  });

  const isPublicPage = $derived(
    page.url.pathname === '/login' ||
    page.url.pathname === '/' ||
    page.url.pathname.startsWith('/legal')
  );
  const routes = $derived([
    { href: '/dashboard', label: i18n.t('tabs.expenses') },
    { href: '/stats', label: i18n.t('tabs.stats') },
    { href: '/budgets', label: i18n.t('tabs.budgets') },
    { href: '/loans', label: i18n.t('tabs.loans') },
    { href: '/config', label: i18n.t('tabs.config') }
  ]);
</script>

{#if !ready}
  <div style="display:flex;align-items:center;justify-content:center;min-height:100vh;color:var(--ink-dim)">
    {i18n.t('common.loading')}
  </div>
{:else if isPublicPage}
  <slot />
{:else if S.user}
  <nav class="nav">
      <div class="brand">
        <div class="brand-mark">C</div>
        <div class="brand-name">Chiro</div>
      </div>
      {#each routes as r (r.href)}
        <a
          href={r.href}
          class={page.url.pathname === r.href ? 'active' : ''}
          aria-current={page.url.pathname === r.href ? 'page' : undefined}
        >
          <span class="nav-pill">{r.label}</span>
        </a>
      {/each}
      {#if S.user?.role === 'admin'}
        <a
          href="/admin"
          class={page.url.pathname === '/admin' ? 'active' : ''}
          aria-current={page.url.pathname === '/admin' ? 'page' : undefined}
        >
          <span class="nav-pill">Admin</span>
        </a>
      {/if}
      <div class="spacer"></div>
      {#if S.user}
        <a class="nav-logout" href="#" onclick={(e) => { e.preventDefault(); logout(); goto('/login'); }}>
          <span class="nav-pill">{i18n.t('common.logout')}</span>
        </a>
      {/if}
    </nav>
    <main class="page">
      <slot />
    </main>
  {/if}
