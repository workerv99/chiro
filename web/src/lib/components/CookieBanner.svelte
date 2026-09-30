<script>
  import { onMount } from 'svelte';

  let showBanner = $state(false);
  let showSettings = $state(false);
  let analyticsConsent = $state(false);
  let functionalConsent = $state(true);

  onMount(() => {
    const consent = localStorage.getItem('chiro_cookie_consent');
    if (!consent) {
      showBanner = true;
    }
  });

  function acceptAll() {
    analyticsConsent = true;
    functionalConsent = true;
    saveConsent();
  }

  function acceptNecessary() {
    analyticsConsent = false;
    functionalConsent = true;
    saveConsent();
  }

  function saveConsent() {
    localStorage.setItem('chiro_cookie_consent', JSON.stringify({
      analytics: analyticsConsent,
      functional: functionalConsent,
      timestamp: new Date().toISOString()
    }));
    showBanner = false;
    showSettings = false;
  }
</script>

{#if showBanner}
  <div class="fixed bottom-[calc(1.5rem+env(safe-area-inset-bottom))] left-1/2 -translate-x-1/2 z-50 w-[calc(100%-48px)] max-w-2xl bg-card border border-border rounded-2xl p-5 shadow-lg sm:flex sm:items-center sm:justify-between sm:gap-5">
    <p class="text-sm text-foreground mb-4 sm:mb-0 sm:flex-1">
      We use cookies to ensure you get the best experience on our website.
    </p>
    <div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:gap-3 sm:flex-shrink-0">
      <button class="px-4 py-2.5 text-sm font-medium bg-secondary text-secondary-foreground rounded-lg hover:bg-secondary/80 transition-colors" onclick={() => showSettings = !showSettings}>Settings</button>
      <button class="px-4 py-2.5 text-sm font-medium text-secondary-foreground hover:opacity-70 transition-opacity" onclick={acceptNecessary}>Accept necessary</button>
      <button class="px-5 py-2.5 text-sm font-medium bg-foreground text-background rounded-lg hover:opacity-90 transition-opacity" onclick={acceptAll}>Accept all</button>
    </div>

    {#if showSettings}
      <div class="flex flex-col gap-3 mt-4 pt-4 border-t border-border sm:flex-row sm:items-center sm:mt-4 sm:pt-4">
        <label class="flex items-center gap-2 text-xs text-secondary-foreground cursor-pointer">
          <input type="checkbox" checked disabled class="w-4 h-4 accent-primary" />
          <span>Necessary</span>
        </label>
        <label class="flex items-center gap-2 text-xs text-secondary-foreground cursor-pointer">
          <input type="checkbox" bind:checked={functionalConsent} class="w-4 h-4 accent-primary" />
          <span>Functional</span>
        </label>
        <label class="flex items-center gap-2 text-xs text-secondary-foreground cursor-pointer">
          <input type="checkbox" bind:checked={analyticsConsent} class="w-4 h-4 accent-primary" />
          <span>Analytics</span>
        </label>
        <button class="sm:ml-auto px-4 py-2 text-xs font-medium bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition-colors" onclick={saveConsent}>Save preferences</button>
      </div>
    {/if}
  </div>
{/if}
