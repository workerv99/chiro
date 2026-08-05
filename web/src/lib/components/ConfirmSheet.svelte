<script>
  import { i18n } from '$lib/i18n.svelte.js';
  import Dialog from '$lib/components/ui/dialog.svelte';
  import DialogHeader from '$lib/components/ui/dialog-header.svelte';
  import DialogTitle from '$lib/components/ui/dialog-title.svelte';
  import DialogDescription from '$lib/components/ui/dialog-description.svelte';
  import DialogFooter from '$lib/components/ui/dialog-footer.svelte';
  import Button from '$lib/components/ui/button.svelte';

  let { open = $bindable(false), title, message, confirmLabel, cancelLabel, danger = false, onConfirm, onCancel } = $props();

  function handleConfirm() {
    onConfirm();
    open = false;
  }

  function handleCancel() {
    onCancel();
    open = false;
  }
</script>

<Dialog bind:open>
  <DialogHeader>
    <DialogTitle>{title}</DialogTitle>
    <DialogDescription>{message}</DialogDescription>
  </DialogHeader>
  <DialogFooter>
    <Button variant="outline" onclick={handleCancel}>
      {cancelLabel || i18n.t('common.cancel')}
    </Button>
    <Button variant={danger ? 'destructive' : 'default'} onclick={handleConfirm}>
      {confirmLabel || i18n.t('common.save')}
    </Button>
  </DialogFooter>
</Dialog>
