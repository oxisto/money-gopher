<script lang="ts">
	/**
	 * A small donut chart with a legend, built from a CSS conic-gradient —
	 * no charting library needed for three or four slices. Only positive
	 * values are drawn as slices (a negative amount, e.g. a trading loss,
	 * doesn't have a meaningful angle); the legend still lists every segment
	 * so nothing is silently hidden.
	 */
	let {
		segments,
		formatValue
	}: {
		segments: { label: string; value: number; color: string }[];
		formatValue: (value: number) => string;
	} = $props();

	const positive = $derived(segments.filter((s) => s.value > 0));
	const total = $derived(positive.reduce((sum, s) => sum + s.value, 0));

	const gradient = $derived.by(() => {
		if (total <= 0) return null;
		let angle = 0;
		const stops: string[] = [];
		for (const s of positive) {
			const start = (angle / total) * 360;
			angle += s.value;
			const end = (angle / total) * 360;
			stops.push(`${s.color} ${start}deg ${end}deg`);
		}
		return `conic-gradient(${stops.join(', ')})`;
	});
</script>

<div class="flex flex-wrap items-center gap-6">
	<div
		class="h-32 w-32 shrink-0 rounded-full"
		style={gradient ? `background: ${gradient}` : 'background: #e5e7eb'}
	>
		<div
			class="m-[22%] flex h-[56%] w-[56%] items-center justify-center rounded-full bg-surface text-center text-xs text-muted"
		>
			{#if total <= 0}
				No gains
			{/if}
		</div>
	</div>
	<ul class="flex flex-col gap-1.5 text-sm">
		{#each segments as s (s.label)}
			<li class="flex items-center gap-2">
				<span class="h-2.5 w-2.5 shrink-0 rounded-full" style={`background: ${s.color}`}></span>
				<span class="text-muted">{s.label}</span>
				<span class="ml-auto font-medium tabular-nums">{formatValue(s.value)}</span>
				{#if s.value > 0 && total > 0}
					<span class="w-10 text-right text-xs text-muted">
						{Math.round((s.value / total) * 100)}%
					</span>
				{/if}
			</li>
		{/each}
	</ul>
</div>
