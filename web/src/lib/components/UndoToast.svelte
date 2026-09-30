<script>
  import { i18n } from '$lib/i18n.svelte.js';
  import Toast from '$lib/components/ui/toast.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import { X } from 'lucide-svelte';

  let { message, secondsLeft, onUndo, onClose } = $props();
</script>

<div class="fixed bottom-[calc(6rem+env(safe-area-inset-bottom))] left-1/2 -translate-x-1/2 z-50">
  <Toast>
    <span class="text-sm">{message}</span>
    {#if secondsLeft > 0}
      <Button variant="ghost" size="sm" class="px-2 text-primary" onclick={onUndo}>
        {i18n.t('common.undo')}
      </Button>
    {/if}
    <Button variant="ghost" size="sm" class="h-11 w-11 p-0 text-muted-foreground" onclick={onClose} aria-label={i18n.t('common.close')}>
      <X size={16} />
    </Button>
  </Toast>
</div>
