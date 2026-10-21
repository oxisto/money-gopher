<script lang="ts">
	/** A row of key figures sharing one border, e.g. value / gain / return. */
	let {
		stats
	}: {
		stats: { label: string; value: string; sub?: string; tone?: string }[];
	} = $props();
</script>

<section
	class="grid grid-cols-[repeat(auto-fit,minmax(13rem,1fr))] gap-px overflow-hidden rounded-xl border border-line bg-line"
>
	{#each stats as stat (stat.label)}
		<div class="bg-surface px-4 py-3.5">
			<div class="text-xs text-muted">{stat.label}</div>
			<div class="num mt-1 text-xl font-semibold {stat.tone ?? ''}">{stat.value}</div>
			{#if stat.sub}
				<div class="num mt-0.5 text-xs text-muted">{stat.sub}</div>
			{/if}
		</div>
	{/each}
</section>
