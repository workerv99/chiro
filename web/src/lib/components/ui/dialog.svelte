<script lang="ts">
  import { cn } from "$lib/utils.js";
  import type { Snippet } from "svelte";

  let {
    open = $bindable(false),
    class: className = "",
    children,
    ...restProps
  }: {
    open?: boolean;
    class?: string;
    children?: Snippet;
    [key: string]: any;
  } = $props();

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === "Escape") open = false;
  }
</script>

{#if open}
  <div
    class="fixed inset-0 z-50 bg-black/80 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0"
    onclick={() => (open = false)}
    onkeydown={handleKeydown}
    role="presentation"
  >
    <div
      class={cn(
        "fixed left-[50%] top-[50%] z-50 grid w-full max-w-lg translate-x-[-50%] translate-y-[-50%] gap-4 border bg-background p-6 shadow-lg duration-200 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[state=closed]:slide-out-to-left-1/2 data-[state=closed]:slide-out-to-top-[48%] data-[state=open]:slide-in-from-left-1/2 data-[state=open]:slide-in-from-top-[48%] sm:rounded-lg",
        className
      )}
      role="dialog"
      aria-modal="true"
      onclick={(e) => e.stopPropagation()}
      {...restProps}
    >
      {@render children?.()}
    </div>
  </div>
{/if}
