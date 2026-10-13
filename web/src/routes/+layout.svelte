<script>
  import '../app.css';
  import { page } from '$app/state';
  import { goto } from '$app/navigation';
  import { i18n } from '$lib/i18n.svelte.js';
  import { S, me, fetchAll, logout, loadMonth, fetchSubscription } from '$lib/stores.svelte.js';
  import { A, loadToken, refreshToken } from '$lib/api.svelte.js';
  import CookieBanner from '$lib/components/CookieBanner.svelte';
  import { initTheme } from '$lib/theme.svelte.js';
  import { Receipt, ChartPie, PiggyBank, HandCoins, Settings, LogOut } from 'lucide-svelte';

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
      // Proactive refresh: keeps the session alive past the 24h token TTL.
      await refreshToken();
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
  // `match` lists every path prefix that belongs to a tab, so nested routes
  // (a lender's panel, an expense detail) keep their parent tab highlighted.
  const routes = $derived([
    { href: '/dashboard', match: ['/dashboard', '/expense'], label: i18n.t('tabs.expenses'), icon: Receipt },
    { href: '/stats', match: ['/stats'], label: i18n.t('tabs.stats'), icon: ChartPie },
    { href: '/budgets', match: ['/budgets'], label: i18n.t('tabs.budgets'), icon: PiggyBank },
    { href: '/loans', match: ['/loans'], label: i18n.t('tabs.loans'), icon: HandCoins },
    { href: '/config', match: ['/config'], label: i18n.t('tabs.config'), icon: Settings }
  ]);

  function isActive(match) {
    const path = page.url.pathname;
    return match.some((m) => path === m || path.startsWith(m + '/'));
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
  <nav
    class="hidden sm:block sticky top-0 z-20 bg-background/80 backdrop-blur-xl border-b border-border"
    aria-label={i18n.t('common.mainNav')}
  >
    <div class="mx-auto flex max-w-[760px] items-center gap-1 px-4 py-2">
      <a href="/dashboard" class="mr-2 flex h-11 items-center gap-2 rounded-lg md:mr-3" aria-label="Chiro">
        <span class="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-sm font-bold text-primary-foreground">C</span>
        <span class="hidden text-sm font-bold md:inline">Chiro</span>
      </a>
      {#each routes as r (r.href)}
        <a
          href={r.href}
          class="flex h-11 items-center justify-center rounded-full px-2 text-sm font-medium transition-colors md:px-3 {isActive(r.match) ? 'bg-primary/10 text-primary' : 'text-muted-foreground hover:bg-accent hover:text-foreground'}"
          aria-current={isActive(r.match) ? 'page' : undefined}
        >
          {r.label}
        </a>
      {/each}
      {#if S.user?.role === 'admin'}
        <a
          href="/admin"
          class="flex h-11 items-center justify-center rounded-full px-2 text-sm font-medium transition-colors md:px-3 {isActive(['/admin']) ? 'bg-primary/10 text-primary' : 'text-muted-foreground hover:bg-accent hover:text-foreground'}"
          aria-current={isActive(['/admin']) ? 'page' : undefined}
        >
          Admin
        </a>
      {/if}
      <button
        class="ml-auto flex h-11 items-center justify-center gap-2 rounded-full px-3 text-sm font-medium text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
        aria-label={i18n.t('common.logout')}
        onclick={() => { logout(); goto('/login'); }}
      >
        <LogOut class="h-4 w-4" />
        <span class="hidden md:inline">{i18n.t('common.logout')}</span>
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
        class="flex-1 flex flex-col items-center justify-center gap-0.5 min-h-14 text-xs transition-colors {isActive(r.match) ? 'font-semibold text-primary' : 'font-medium text-muted-foreground'}"
        aria-current={isActive(r.match) ? 'page' : undefined}
      >
        <span class="flex h-7 w-14 items-center justify-center rounded-full transition-colors {isActive(r.match) ? 'bg-primary/10' : ''}">
          <r.icon class="h-5 w-5" />
        </span>
        <span class="tracking-tight">{r.label}</span>
      </a>
    {/each}
  </nav>
  <main class="max-w-[760px] mx-auto px-4 pb-[calc(10.5rem+env(safe-area-inset-bottom))] sm:pb-[calc(7rem+env(safe-area-inset-bottom))] pt-4 sm:pt-4">
    {@render children?.()}
  </main>
  <CookieBanner />
{/if}
