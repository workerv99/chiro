<script>
  import { i18n } from '$lib/i18n.svelte.js';

  let { message, secondsLeft, onUndo, onClose } = $props();
</script>

<div class="toast" role="status" aria-live="polite">
  <span class="toast-msg">{message}</span>
  {#if secondsLeft > 0}
    <button class="toast-undo" onclick={onUndo}>{i18n.t('common.undo')}</button>
  {/if}
  <button class="toast-close" onclick={onClose} aria-label={i18n.t('common.close')}>×</button>
</div>

<style>
  .toast {
    position: fixed;
    left: 50%;
    bottom: calc(80px + env(safe-area-inset-bottom));
    transform: translateX(-50%);
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 12px;
    padding: 10px 14px;
    display: inline-flex;
    align-items: center;
    gap: 12px;
    z-index: 80;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4);
    font-size: 0.88rem;
  }
  .toast-msg {
    color: var(--ink);
  }
  .toast-undo {
    color: var(--indigo);
    font-weight: 700;
    min-height: 32px;
    padding: 4px 8px;
  }
  .toast-close {
    color: var(--ink-dim);
    font-size: 1.1rem;
    padding: 0 4px;
  }
</style>
