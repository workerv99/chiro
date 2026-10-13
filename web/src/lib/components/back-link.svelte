<script lang="ts">
  import { cn } from "$lib/utils.js";
  import { ChevronLeft } from "lucide-svelte";
  import type { Snippet } from "svelte";

  let {
    href,
    class: className = "",
    children,
    ...restProps
  }: {
    href: string;
    class?: string;
    children?: Snippet;
    [key: string]: any;
  } = $props();
</script>

<a
  {href}
  class={cn(
    "-ml-1 mb-1 inline-flex h-11 max-w-full items-center gap-1 rounded-md pl-1 pr-3 text-sm font-medium text-foreground transition-colors hover:bg-accent focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
    className
  )}
  {...restProps}
>
  <ChevronLeft class="h-5 w-5 shrink-0 text-muted-foreground" />
  <span class="truncate">{@render children?.()}</span>
</a>
