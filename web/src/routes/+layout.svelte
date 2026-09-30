<script>
  import '../app.css';
  import { page } from '$app/state';
  import { goto } from '$app/navigation';
  import { i18n } from '$lib/i18n.svelte.js';
  import { S, me, fetchAll, logout, loadMonth, fetchSubscription } from '$lib/stores.svelte.js';
  import { A, loadToken } from '$lib/api.svelte.js';
  import CookieBanner from '$lib/components/CookieBanner.svelte';
  import { initTheme } from '$lib/theme.svelte.js';
  import { Receipt, ChartPie, PiggyBank, HandCoins, Settings } from 'lucide-svelte';

  let { children } = $props();
  let ready = $state(false);

  $effect(() => initTheme());

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
    { href: '/dashboard', label: i18n.t('tabs.expenses'), icon: Receipt },
    { href: '/stats', label: i18n.t('tabs.stats'), icon: ChartPie },
    { href: '/budgets', label: i18n.t('tabs.budgets'), icon: PiggyBank },
    { href: '/loans', label: i18n.t('tabs.loans'), icon: HandCoins },
    { href: '/config', label: i18n.t('tabs.config'), icon: Settings }
  ]);

  function isActive(href) {
    return page.url.pathname === href || page.url.pathname.startsWith(href + '/');
  }
</script>

{#if !ready}
  <div class="flex items-center justify-center min-h-dvh text-muted-foreground">
    {i18n.t('common.loading')}
  </div>
{:else if isPublicPage}
  {@render children?.()}
  <CookieBanner />
{:else if S.user}
  <nav class="hidden sm:flex sticky top-0 z-20 items-center gap-1 px-3 py-2 bg-background/80 backdrop-blur-xl border-b border-border">
    <div class="flex items-center gap-2 mr-3">
      <div class="w-7 h-7 rounded-lg bg-primary/10 border border-primary/20 text-primary font-bold text-xs flex items-center justify-center">C</div>
      <span class="font-bold text-sm hidden sm:block">Chiro</span>
    </div>
    {#each routes as r (r.href)}
      <a
        href={r.href}
        class="flex-1 sm:flex-none sm:min-w-[80px] flex items-center justify-center h-11 rounded-full text-xs font-semibold transition-colors {page.url.pathname === r.href ? 'bg-primary/10 text-primary' : 'text-muted-foreground hover:bg-accent hover:text-foreground'}"
        aria-current={page.url.pathname === r.href ? 'page' : undefined}
      >
        {r.label}
      </a>
    {/each}
    {#if S.user?.role === 'admin'}
      <a
        href="/admin"
        class="flex-1 sm:flex-none sm:min-w-[80px] flex items-center justify-center h-11 rounded-full text-xs font-semibold transition-colors {page.url.pathname === '/admin' ? 'bg-primary/10 text-primary' : 'text-muted-foreground hover:bg-accent hover:text-foreground'}"
        aria-current={page.url.pathname === '/admin' ? 'page' : undefined}
      >
        Admin
      </a>
    {/if}
    <div class="flex-1 sm:flex-none sm:ml-auto">
      <button class="w-full sm:w-auto flex items-center justify-center h-11 px-3 rounded-full text-xs font-semibold text-muted-foreground hover:bg-accent hover:text-foreground transition-colors" onclick={() => { logout(); goto('/login'); }}>
        {i18n.t('common.logout')}
      </button>
    </div>
  </nav>
  <nav
    class="sm:hidden fixed inset-x-0 bottom-0 z-30 flex bg-background/90 backdrop-blur-xl border-t border-border pb-[env(safe-area-inset-bottom)]"
    aria-label={i18n.t('common.mainNav')}
  >
    {#each routes as r (r.href)}
      <a
        href={r.href}
        class="flex-1 flex flex-col items-center justify-center gap-0.5 min-h-14 text-[11px] font-medium transition-colors {isActive(r.href) ? 'text-primary' : 'text-muted-foreground'}"
        aria-current={isActive(r.href) ? 'page' : undefined}
      >
        <r.icon class="h-5 w-5" />
        <span>{r.label}</span>
      </a>
    {/each}
  </nav>
  <main class="max-w-[760px] mx-auto px-4 pb-[calc(10.5rem+env(safe-area-inset-bottom))] sm:pb-[calc(7rem+env(safe-area-inset-bottom))] pt-4 sm:pt-4">
    {@render children?.()}
  </main>
  <CookieBanner />
{/if}
