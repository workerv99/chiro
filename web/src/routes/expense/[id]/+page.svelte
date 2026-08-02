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

<div class="page-head">
  <a class="back-link" href="/">← {i18n.t('expenses.title')}</a>
  <h1 class="headline">{i18n.t('expenses.editExpense')}</h1>
</div>

{#if loading}
  <p class="meta" style="padding:24px;text-align:center">{i18n.t('common.loading')}</p>
{:else if expense}
  <ExpenseModal
    expense={expense}
    onClose={() => {
      goto('/');
    }}
  />
{/if}
