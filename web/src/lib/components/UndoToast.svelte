<script>
  import { i18n } from '$lib/i18n.svelte.js';
  import Toast from '$lib/components/ui/toast.svelte';
  import Button from '$lib/components/ui/button.svelte';

  let { message, secondsLeft, onUndo, onClose } = $props();
</script>

<div class="fixed bottom-24 left-1/2 -translate-x-1/2 z-50">
  <Toast>
    <span class="text-sm">{message}</span>
    {#if secondsLeft > 0}
      <Button variant="ghost" size="sm" class="h-8 px-2 text-primary" onclick={onUndo}>
        {i18n.t('common.undo')}
      </Button>
    {/if}
    <Button variant="ghost" size="sm" class="h-8 w-8 p-0 text-muted-foreground" onclick={onClose} aria-label={i18n.t('common.close')}>
      ×
    </Button>
  </Toast>
</div>
