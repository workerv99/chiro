<script lang="ts">
  import { cn } from "$lib/utils.js";
  import type { Snippet } from "svelte";

  let {
    class: className = "",
    open = $bindable(false),
    children,
    ...restProps
  }: {
    class?: string;
    open?: boolean;
    children?: Snippet;
    [key: string]: any;
  } = $props();

  let dialogEl = $state<HTMLElement | undefined>();

  $effect(() => {
    if (open) queueMicrotask(() => dialogEl?.focus());
  });
</script>

{#if open}
  <button
    type="button"
    class="fixed inset-0 z-50 border-0 bg-black/80 p-0"
    aria-label="Cerrar diálogo"
    onclick={() => open = false}
  ></button>
  <div
    bind:this={dialogEl}
    role="dialog"
    aria-modal="true"
    aria-label="Diálogo"
    tabindex="-1"
    onkeydown={(event) => { if (event.key === 'Escape') open = false; }}
    class={cn(
      "fixed left-[50%] top-[50%] z-50 grid w-full max-w-lg translate-x-[-50%] translate-y-[-50%] gap-4 border bg-background p-6 shadow-lg sm:rounded-lg",
      className
    )}
    {...restProps}
  >
    {@render children?.()}
  </div>
{/if}
