<script lang="ts">
  import { onMount } from 'svelte';
  import type { Snippet } from 'svelte';
  import BackLink from '$lib/components/back-link.svelte';

  let {
    title,
    updated,
    summary = '',
    children
  }: {
    title: string;
    updated: string;
    summary?: string;
    children?: Snippet;
  } = $props();

  let body: HTMLElement;
  let sections = $state<{ id: string; text: string }[]>([]);
  let activeId = $state('');
  let tocOpen = $state(false);

  onMount(() => {
    const headings = Array.from(body.querySelectorAll('h2'));
    sections = headings.map((h, i) => {
      const id = `seccion-${i + 1}`;
      h.id = id;
      return { id, text: h.textContent?.trim() ?? '' };
    });
    if (!('IntersectionObserver' in window)) return;
    const observer = new IntersectionObserver(
      (entries) => {
        const visible = entries.filter((e) => e.isIntersecting).sort((a, b) => a.boundingClientRect.top - b.boundingClientRect.top);
        if (visible.length) activeId = visible[0].target.id;
      },
      { rootMargin: '-72px 0px -70% 0px' }
    );
    headings.forEach((h) => observer.observe(h));
    return () => observer.disconnect();
  });
</script>

<div class="mx-auto max-w-5xl px-4 py-8">
  <BackLink href="/" class="mb-3 print:hidden">Volver</BackLink>

  <header class="mb-8 border-b border-border pb-6">
    <h1 class="text-2xl font-bold tracking-tight md:text-3xl">{title}</h1>
    <p class="mt-3 inline-flex items-center gap-2 rounded-full bg-primary/10 px-3 py-1 text-xs font-medium text-primary">
      Última actualización: {updated}
    </p>
    {#if summary}
      <p class="mt-4 max-w-2xl text-sm leading-relaxed text-muted-foreground">{summary}</p>
    {/if}
  </header>

  <div class="grid gap-8 md:grid-cols-[230px_minmax(0,1fr)] md:gap-12">
    {#if sections.length}
      <details class="rounded-xl border border-border bg-card print:hidden md:hidden" bind:open={tocOpen}>
        <summary class="flex h-11 cursor-pointer list-none items-center justify-between px-4 text-sm font-semibold">
          Contenido
          <span class="text-xs font-normal text-muted-foreground">{sections.length} secciones</span>
        </summary>
        <nav aria-label="Contenido" class="border-t border-border p-2">
          <ul class="m-0 list-none p-0">
            {#each sections as s (s.id)}
              <li>
                <a
                  href={`#${s.id}`}
                  onclick={() => (tocOpen = false)}
                  class="flex min-h-11 items-center rounded-md px-3 text-sm {activeId === s.id ? 'bg-primary/10 font-medium text-primary' : 'text-muted-foreground'}"
                >{s.text}</a>
              </li>
            {/each}
          </ul>
        </nav>
      </details>

      <aside class="hidden print:hidden md:block">
        <nav aria-label="Contenido" class="sticky top-6 max-h-[calc(100dvh-3rem)] overflow-y-auto pr-2">
          <p class="mb-2 px-3 text-xs font-semibold uppercase tracking-wider text-muted-foreground">Contenido</p>
          <ul class="m-0 list-none p-0">
            {#each sections as s (s.id)}
              <li>
                <a
                  href={`#${s.id}`}
                  class="block rounded-md px-3 py-1.5 text-[13px] leading-snug transition-colors hover:bg-accent {activeId === s.id ? 'bg-primary/10 font-medium text-primary' : 'text-muted-foreground hover:text-foreground'}"
                >{s.text}</a>
              </li>
            {/each}
          </ul>
        </nav>
      </aside>
    {/if}

    <article bind:this={body} class="legal-body min-w-0 md:col-start-2 md:row-start-1">
      {@render children?.()}
    </article>
  </div>
</div>

<style>
  .legal-body :global(h2) {
    margin: 2.25rem 0 0.75rem;
    padding-top: 1.5rem;
    border-top: 1px solid hsl(var(--border));
    font-size: 1.125rem;
    line-height: 1.4;
    font-weight: 700;
    letter-spacing: -0.01em;
    scroll-margin-top: 1.5rem;
  }
  .legal-body :global(h2:first-child) {
    margin-top: 0;
    padding-top: 0;
    border-top: 0;
  }
  .legal-body :global(p) {
    margin: 0.75rem 0;
    font-size: 0.9375rem;
    line-height: 1.75;
    color: hsl(var(--foreground) / 0.88);
  }
  .legal-body :global(strong) {
    font-weight: 600;
    color: hsl(var(--foreground));
  }
  .legal-body :global(ul),
  .legal-body :global(ol) {
    margin: 0.75rem 0;
    padding-left: 1.25rem;
  }
  .legal-body :global(ul) {
    list-style: disc;
  }
  .legal-body :global(ol) {
    list-style: decimal;
  }
  .legal-body :global(li) {
    margin: 0.4rem 0;
    padding-left: 0.25rem;
    font-size: 0.9375rem;
    line-height: 1.7;
    color: hsl(var(--foreground) / 0.88);
  }
  .legal-body :global(li::marker) {
    color: hsl(var(--primary));
  }
  .legal-body :global(a) {
    color: hsl(var(--primary));
    text-decoration: underline;
    text-underline-offset: 3px;
  }
  .legal-body :global(table) {
    width: 100%;
    margin: 1rem 0;
    border-collapse: separate;
    border-spacing: 0;
    overflow: hidden;
    border: 1px solid hsl(var(--border));
    border-radius: 0.75rem;
    font-size: 0.8125rem;
    line-height: 1.5;
  }
  .legal-body :global(th) {
    padding: 0.625rem 0.75rem;
    text-align: left;
    font-weight: 600;
    background: hsl(var(--muted));
  }
  .legal-body :global(td) {
    padding: 0.625rem 0.75rem;
    border-top: 1px solid hsl(var(--border));
    vertical-align: top;
  }
  .legal-body :global(.overflow-x-auto) {
    margin: 1rem 0;
  }
  .legal-body :global(.overflow-x-auto table) {
    margin: 0;
    min-width: 32rem;
  }
</style>
