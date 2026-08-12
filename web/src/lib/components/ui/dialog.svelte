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
</script>

{#if open}
  <div class="fixed inset-0 z-50 bg-black/80" onclick={() => open = false}></div>
  <div
    class={cn(
      "fixed left-[50%] top-[50%] z-50 grid w-full max-w-lg translate-x-[-50%] translate-y-[-50%] gap-4 border bg-background p-6 shadow-lg sm:rounded-lg",
      className
    )}
    {...restProps}
  >
    {@render children?.()}
  </div>
{/if}
