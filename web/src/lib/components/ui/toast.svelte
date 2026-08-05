<script lang="ts">
  import { cn } from "$lib/utils.js";
  import type { Snippet } from "svelte";

  let {
    class: className = "",
    variant = "default",
    children,
    ...restProps
  }: {
    class?: string;
    variant?: "default" | "destructive";
    children?: Snippet;
    [key: string]: any;
  } = $props();

  const variants = {
    default: "border bg-background text-foreground",
    destructive: "destructive group border-destructive bg-destructive text-destructive-foreground"
  };
</script>

<div
  class={cn(
    "pointer-events-auto relative flex w-full items-center justify-between space-x-4 overflow-hidden rounded-md border p-6 pr-8 shadow-lg transition-all",
    variants[variant],
    className
  )}
  {...restProps}
>
  {@render children?.()}
</div>
