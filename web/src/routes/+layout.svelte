<script>
  import '../app.css';
  import { page } from '$app/state';
  import { goto } from '$app/navigation';
  import { i18n } from '$lib/i18n.svelte.js';
  import { S, me, fetchAll, logout, loadMonth, fetchSubscription } from '$lib/stores.svelte.js';
  import { A, loadToken } from '$lib/api.svelte.js';
  import CookieBanner from '$lib/components/CookieBanner.svelte';

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
  <div class="flex items-center justify-center min-h-screen text-muted-foreground">
    {i18n.t('common.loading')}
  </div>
{:else if isPublicPage}
  <slot />
  <CookieBanner />
{:else if S.user}
  <nav class="sticky top-0 z-20 flex items-center gap-1 px-3 py-2 bg-background/80 backdrop-blur-xl border-b border-border">
    <div class="flex items-center gap-2 mr-3">
      <div class="w-7 h-7 rounded-lg bg-primary/10 border border-primary/20 text-primary font-bold text-xs flex items-center justify-center">C</div>
      <span class="font-bold text-sm hidden sm:block">Chiro</span>
    </div>
    {#each routes as r (r.href)}
      <a
        href={r.href}
        class="flex-1 sm:flex-none sm:min-w-[80px] flex items-center justify-center h-9 rounded-full text-xs font-semibold transition-colors {page.url.pathname === r.href ? 'bg-primary/10 text-primary' : 'text-muted-foreground hover:bg-accent hover:text-foreground'}"
        aria-current={page.url.pathname === r.href ? 'page' : undefined}
      >
        {r.label}
      </a>
    {/each}
    {#if S.user?.role === 'admin'}
      <a
        href="/admin"
        class="flex-1 sm:flex-none sm:min-w-[80px] flex items-center justify-center h-9 rounded-full text-xs font-semibold transition-colors {page.url.pathname === '/admin' ? 'bg-primary/10 text-primary' : 'text-muted-foreground hover:bg-accent hover:text-foreground'}"
        aria-current={page.url.pathname === '/admin' ? 'page' : undefined}
      >
        Admin
      </a>
    {/if}
    <div class="flex-1 sm:flex-none sm:ml-auto">
      <button class="w-full sm:w-auto flex items-center justify-center h-9 px-3 rounded-full text-xs font-semibold text-muted-foreground hover:bg-accent hover:text-foreground transition-colors" onclick={() => { logout(); goto('/login'); }}>
        {i18n.t('common.logout')}
      </button>
    </div>
  </nav>
  <main class="max-w-[760px] mx-auto px-4 pb-28 pt-4 sm:pt-4">
    <slot />
  </main>
  <CookieBanner />
{/if}