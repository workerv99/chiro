<script>
  import { page } from '$app/state';
  import { goto } from '$app/navigation';
  import { i18n } from '$lib/i18n.svelte.js';
  import { api } from '$lib/api.svelte.js';
  import ExpenseModal from '$lib/components/ExpenseModal.svelte';

  const id = String(page.params.id);
  let expense = $state(null);
  let loading = $state(true);

  $effect(() => {
    loading = true;
    api(`/api/expenses/${id}`)
      .then((e) => (expense = e))
      .finally(() => (loading = false));
  });
</script>

<svelte:head><title>{i18n.t('expenses.editExpense')} · Chiro</title></svelte:head>

<a class="text-sm text-muted-foreground hover:text-foreground font-semibold mb-2 inline-block" href="/">← {i18n.t('expenses.title')}</a>
<h1 class="text-2xl font-bold mb-4">{i18n.t('expenses.editExpense')}</h1>

{#if loading}
  <p class="text-sm text-muted-foreground py-8 text-center">{i18n.t('common.loading')}</p>
{:else if expense}
  <ExpenseModal
    expense={expense}
    onClose={() => {
      goto('/');
    }}
  />
{/if}