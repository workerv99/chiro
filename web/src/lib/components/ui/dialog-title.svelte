<script lang="ts">
  import { cn } from "$lib/utils.js";
  import type { Snippet } from "svelte";

  let {
    class: className = "",
    children,
    ...restProps
  }: {
    class?: string;
    children?: Snippet;
    [key: string]: any;
  } = $props();
</script>

<h2
  class={cn("text-lg font-semibold leading-none tracking-tight", className)}
  {...restProps}
>
  {@render children?.()}
</h2>
