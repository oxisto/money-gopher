<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		title,
		href,
		linkLabel,
		aside,
		children
	}: {
		title: string;
		/** An optional "see all" link in the header. */
		href?: string;
		linkLabel?: string;
		/** Anything else for the header's right side, e.g. a count. */
		aside?: Snippet;
		children: Snippet;
	} = $props();
</script>

<section class="flex min-w-0 flex-col gap-3 rounded-xl border border-line p-4">
	<div class="flex items-baseline justify-between gap-3">
		<h2 class="text-sm font-semibold">{title}</h2>
		{#if aside}
			{@render aside()}
		{:else if href}
			<a {href} class="text-xs text-muted hover:text-fg">{linkLabel ?? 'View all'} →</a>
		{/if}
	</div>
	{@render children()}
</section>
