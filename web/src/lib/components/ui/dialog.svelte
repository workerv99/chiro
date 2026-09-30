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
      "fixed inset-x-0 bottom-0 z-50 grid w-full max-h-[85dvh] gap-4 overflow-y-auto border bg-background p-6 pb-[calc(1.5rem+env(safe-area-inset-bottom))] shadow-lg rounded-t-xl sm:inset-x-auto sm:left-[50%] sm:top-[50%] sm:bottom-auto sm:max-w-lg sm:translate-x-[-50%] sm:translate-y-[-50%] sm:rounded-lg sm:pb-6",
      className
    )}
    {...restProps}
  >
    {@render children?.()}
  </div>
{/if}
