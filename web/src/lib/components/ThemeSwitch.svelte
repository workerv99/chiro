<script>
  import { Monitor, Moon, Sun } from 'lucide-svelte';
  import { i18n } from '$lib/i18n.svelte.js';
  import { setThemePreference, theme } from '$lib/theme.svelte.js';

  const options = [
    { value: 'light', icon: Sun, label: () => i18n.t('config.themeLight') },
    { value: 'dark', icon: Moon, label: () => i18n.t('config.themeDark') },
    { value: 'system', icon: Monitor, label: () => i18n.t('config.themeSystem') }
  ];
</script>

<div class="inline-flex rounded-xl border border-border bg-muted/50 p-1" role="group" aria-label={i18n.t('config.theme')}>
  {#each options as option (option.value)}
    <button
      type="button"
      class="inline-flex h-9 items-center gap-2 rounded-lg px-3 text-xs font-semibold transition-all {theme.preference === option.value ? 'bg-background text-foreground shadow-sm' : 'text-muted-foreground hover:text-foreground'}"
      aria-label={option.label()}
      aria-pressed={theme.preference === option.value}
      title={option.label()}
      onclick={() => setThemePreference(option.value)}
    >
      <option.icon size={16} />
      <span class="hidden sm:inline">{option.label()}</span>
    </button>
  {/each}
</div>
