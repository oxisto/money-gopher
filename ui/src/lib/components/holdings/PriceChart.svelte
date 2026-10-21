<script lang="ts">
	import Segmented from '#lib/components/ui/Segmented.svelte';
	import { formatPrice } from '#lib/format.js';

	interface Observation {
		time: Date;
		price: number;
	}

	let {
		quotes,
		trades,
		dividends,
		averageCost,
		currency
	}: {
		/** Market quotes, any order. */
		quotes: Observation[];
		/** Own buys and sells at their execution price. */
		trades: (Observation & { kind: 'buy' | 'sell'; label: string })[];
		dividends: Date[];
		averageCost: number | null;
		currency: string;
	} = $props();

	type Range = '3M' | '1Y' | 'MAX';
	let range = $state<Range>('MAX');

	const W = 1000;
	const H = 240;
	const PAD = 12;

	const sortedQuotes = $derived([...quotes].sort((a, b) => a.time.getTime() - b.time.getTime()));
	const firstQuote = $derived(sortedQuotes[0]?.time.getTime() ?? Infinity);

	const bounds = $derived.by(() => {
		const to = Date.now();
		const earliest = Math.min(firstQuote, ...trades.map((t) => t.time.getTime()));
		const from =
			range === '3M'
				? to - 91 * 864e5
				: range === '1Y'
					? to - 365 * 864e5
					: Number.isFinite(earliest)
						? earliest
						: to - 365 * 864e5;
		return { from, to };
	});

	// Quotes only exist since the quote updater started. Before that, the
	// prices of one's own trades are the only observations there are; they
	// are drawn dashed and joined onto the first quote, never passed off as
	// market data.
	const series = $derived.by(() => {
		const inWindow = (o: Observation) => o.time.getTime() >= bounds.from && o.time.getTime() <= bounds.to;
		const early = trades.filter((t) => t.time.getTime() < firstQuote).sort((a, b) => a.time.getTime() - b.time.getTime());
		const solid = sortedQuotes.filter(inWindow);
		const dashed = [...early, ...(sortedQuotes[0] ? [sortedQuotes[0]] : [])].filter(inWindow);
		return { solid, dashed };
	});

	const scale = $derived.by(() => {
		const prices = [
			...series.solid,
			...series.dashed,
			...trades.filter((t) => t.time.getTime() >= bounds.from)
		].map((o) => o.price);
		if (averageCost !== null) prices.push(averageCost);
		const lo = Math.min(...prices);
		const hi = Math.max(...prices);
		const pad = (hi - lo) * 0.1 || hi * 0.05 || 1;
		return { lo: lo - pad, hi: hi + pad };
	});

	const x = (t: Date) => ((t.getTime() - bounds.from) / (bounds.to - bounds.from || 1)) * W;
	const y = (p: number) => PAD + (1 - (p - scale.lo) / (scale.hi - scale.lo)) * (H - 2 * PAD);
	const path = (obs: Observation[]) =>
		obs.map((o, i) => `${i ? 'L' : 'M'}${x(o.time).toFixed(1)},${y(o.price).toFixed(1)}`).join(' ');

	const visibleTrades = $derived(trades.filter((t) => t.time.getTime() >= bounds.from));
	const visibleDividends = $derived(dividends.filter((d) => d.getTime() >= bounds.from));

	const xLabels = $derived.by(() => {
		const span = bounds.to - bounds.from;
		const fmt = new Intl.DateTimeFormat(
			undefined,
			span > 400 * 864e5 ? { month: 'short', year: 'numeric' } : { month: 'short', day: 'numeric' }
		);
		return [0, 1 / 3, 2 / 3, 1].map((f) => fmt.format(new Date(bounds.from + f * span)));
	});

	const money = (amount: number) => formatPrice({ amount: String(amount), currency }, 2);
	const firstQuoteDate = $derived(
		Number.isFinite(firstQuote) ? new Intl.DateTimeFormat(undefined, { dateStyle: 'medium' }).format(firstQuote) : null
	);
</script>

<div class="flex flex-col gap-3">
	<div class="flex flex-wrap items-center gap-3">
		<h2 class="flex-1 text-sm font-semibold">Price &amp; your trades</h2>
		<Segmented
			label="Time range"
			bind:value={range}
			options={[
				{ value: '3M', label: '3M' },
				{ value: '1Y', label: '1Y' },
				{ value: 'MAX', label: 'Max' }
			]}
		/>
	</div>

	{#if series.solid.length + series.dashed.length < 2}
		<div class="grid h-[240px] place-items-center rounded-lg border border-dashed border-line text-sm text-muted">
			No prices in this range yet.
		</div>
	{:else}
		<div class="relative h-[240px]" role="img" aria-label="Price history with your buys, sells and dividends">
			<svg viewBox="0 0 {W} {H}" preserveAspectRatio="none" class="block h-full w-full" aria-hidden="true">
				{#each [PAD, H / 2, H - PAD] as gy (gy)}
					<line x1="0" x2={W} y1={gy} y2={gy} class="stroke-line" vector-effect="non-scaling-stroke" />
				{/each}
				{#if averageCost !== null}
					<line
						x1="0"
						x2={W}
						y1={y(averageCost)}
						y2={y(averageCost)}
						class="stroke-muted"
						stroke-dasharray="4 4"
						vector-effect="non-scaling-stroke"
					/>
				{/if}
				<path
					d={path(series.dashed)}
					fill="none"
					class="stroke-accent"
					stroke-width="1.5"
					stroke-dasharray="2 4"
					vector-effect="non-scaling-stroke"
				/>
				<path d={path(series.solid)} fill="none" class="stroke-accent" stroke-width="1.75" vector-effect="non-scaling-stroke" />
			</svg>

			{#if averageCost !== null}
				<span
					class="num absolute right-0 -translate-y-full bg-surface px-1 text-[12px] text-muted"
					style="top: {(y(averageCost) / H) * 100}%"
				>
					avg cost {money(averageCost)}
				</span>
			{/if}

			{#each visibleTrades as t, i (i)}
				<span
					title={t.label}
					class="absolute grid size-4.5 -translate-x-1/2 -translate-y-1/2 place-items-center rounded-full text-[9px] font-bold ring-2 ring-surface
						{t.kind === 'buy' ? 'bg-accent text-on-accent' : 'border-2 border-accent bg-surface text-accent'}"
					style="left: {(x(t.time) / W) * 100}%; top: {(y(t.price) / H) * 100}%"
				>
					{t.kind === 'buy' ? 'B' : 'S'}
				</span>
			{/each}
			{#each visibleDividends as d, i (i)}
				<span
					title="Dividend"
					class="absolute bottom-0.5 h-2.5 w-[3px] -translate-x-1/2 rounded-[1px] bg-gold"
					style="left: {(x(d) / W) * 100}%"
				></span>
			{/each}
		</div>
		<div class="flex justify-between text-[12px] text-muted">
			{#each xLabels as label, i (i)}<span>{label}</span>{/each}
		</div>
	{/if}

	<div class="flex flex-wrap gap-x-4 gap-y-1 text-xs text-muted">
		<span class="flex items-center gap-1.5">
			<span class="grid size-4 place-items-center rounded-full bg-accent text-[9px] font-bold text-on-accent">B</span>Buy
		</span>
		<span class="flex items-center gap-1.5">
			<span class="grid size-4 place-items-center rounded-full border-2 border-accent text-[9px] font-bold text-accent">S</span
			>Sell
		</span>
		<span class="flex items-center gap-1.5"><span class="h-2.5 w-[3px] rounded-[1px] bg-gold"></span>Dividend</span>
		{#if averageCost !== null}
			<span class="flex items-center gap-1.5"><span class="w-3.5 border-t-2 border-dashed border-muted"></span>Your average cost</span>
		{/if}
		{#if firstQuoteDate && series.dashed.length > 1}
			<span class="flex items-center gap-1.5">
				<span class="w-3.5 border-t-2 border-dotted border-accent"></span>Before {firstQuoteDate}: your trade prices, no quotes
			</span>
		{/if}
	</div>
</div>
