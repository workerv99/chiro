<script>
  import { i18n } from '$lib/i18n.svelte.js';

  let { title, message, confirmLabel, cancelLabel, danger = false, onConfirm, onCancel } = $props();

  function onKeydown(e) {
    if (e.key === 'Escape') onCancel();
  }
</script>

<svelte:window onkeydown={onKeydown} />

<div class="overlay" onclick={onCancel}>
  <div class="sheet sheet-confirm" role="alertdialog" aria-modal="true" aria-labelledby="confirm-title" aria-describedby="confirm-msg" onclick={(e) => e.stopPropagation()}>
    <h2 class="title" id="confirm-title">{title}</h2>
    <p class="confirm-msg" id="confirm-msg">{message}</p>
    <div class="inline-flex" style="margin-top:16px;justify-content:flex-end">
      <button class="btn btn-cancel" onclick={onCancel}>{cancelLabel || i18n.t('common.cancel')}</button>
      <button class={danger ? 'btn btn-danger' : 'btn btn-primary'} onclick={onConfirm}>{confirmLabel || i18n.t('common.save')}</button>
    </div>
  </div>
</div>

<style>
  .sheet-confirm {
    max-width: 420px;
    border-radius: 18px;
  }
  .confirm-msg {
    color: var(--ink);
    font-size: 0.95rem;
    line-height: 1.5;
    margin-top: 8px;
  }
</style>
