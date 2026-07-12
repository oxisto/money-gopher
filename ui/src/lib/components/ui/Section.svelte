<script lang="ts">
	import type { Snippet } from 'svelte';

	let { title, children }: { title: string; children: Snippet } = $props();
</script>

<section>
	<h2 class="mb-3 text-lg font-semibold">{title}</h2>
	{@render children()}
</section>
