<script lang="ts">
	import { formatMoney, formatMoneyCompact, type Money } from '#lib/format.js';

	interface Point {
		month: Date | string;
		expected: Money;
		contributed: Money;
		p10: Money;
		p25: Money;
		p50: Money;
		p75: Money;
		p90: Money;
	}
	interface Actual {
		time: Date | string;
		netWorth: Money;
	}

	let {
		projection,
		actual,
		startingAmount
	}: { projection: Point[]; actual: Actual[]; startingAmount: Money } = $props();

	// Drawn in a fixed 1000×H box stretched to the container; labels and
	// the hover layer are HTML positioned in percent, as the net-worth
	// chart does it.
	const W = 1000;
	const H = 280;
	const PAD = 10;

	const ms = (t: Date | string) => new Date(t).getTime();
	const t0 = $derived(ms(projection[0].month));
	const t1 = $derived(ms(projection[projection.length - 1].month));
	const putIn = (p: Point) => startingAmount.amount + p.contributed.amount;

	const scale = $derived.by(() => {
		const values = [
			...projection.flatMap((p) => [p.p10.amount, p.p90.amount, p.expected.amount, putIn(p)]),
			...actual.map((a) => a.netWorth.amount)
		];
		const lo = Math.min(...values);
		const hi = Math.max(...values);
		const pad = (hi - lo) * 0.08 || Math.abs(hi) * 0.05 || 100;
		return { lo: lo >= 0 ? Math.max(0, lo - pad) : lo - pad, hi: hi + pad };
	});

	const x = (t: number) => (t1 === t0 ? 0 : ((t - t0) / (t1 - t0)) * W);
	const y = (v: number) => PAD + (1 - (v - scale.lo) / (scale.hi - scale.lo)) * (H - 2 * PAD);
	const pt = (t: number, v: number) => `${x(t).toFixed(1)},${y(v).toFixed(1)}`;
	const line = (points: { t: number; v: number }[]) =>
		points.map((p, i) => `${i ? 'L' : 'M'}${pt(p.t, p.v)}`).join(' ');
	const band = (upper: (p: Point) => number, lower: (p: Point) => number) => {
		const up = projection.map((p, i) => `${i ? 'L' : 'M'}${pt(ms(p.month), upper(p))}`);
		const down = [...projection].reverse().map((p) => `L${pt(ms(p.month), lower(p))}`);
		return `${up.join(' ')} ${down.join(' ')} Z`;
	};

	const outer = $derived(band((p) => p.p90.amount, (p) => p.p10.amount));
	const inner = $derived(band((p) => p.p75.amount, (p) => p.p25.amount));
	const expected = $derived(line(projection.map((p) => ({ t: ms(p.month), v: p.expected.amount }))));
	const contributions = $derived(line(projection.map((p) => ({ t: ms(p.month), v: putIn(p) }))));
	const actualLine = $derived(line(actual.map((a) => ({ t: ms(a.time), v: a.netWorth.amount }))));
	const currency = $derived(projection[0].expected.currency);

	const yLabels = $derived(
		[scale.hi, (scale.hi + scale.lo) / 2, scale.lo].map((v) => formatMoneyCompact({ amount: v, currency }))
	);
	const xLabels = $derived.by(() => {
		const years = (t1 - t0) / (365.25 * 864e5);
		const fmt = new Intl.DateTimeFormat(undefined, years > 3 ? { year: 'numeric', timeZone: 'UTC' } : { month: 'short', year: '2-digit', timeZone: 'UTC' });
		return [0, 0.25, 0.5, 0.75, 1].map((f) => fmt.format(new Date(t0 + f * (t1 - t0))));
	});

	let hover = $state<number | null>(null);
	function track(event: PointerEvent) {
		const rect = (event.currentTarget as HTMLElement).getBoundingClientRect();
		const f = Math.min(1, Math.max(0, (event.clientX - rect.left) / rect.width));
		hover = Math.round(f * (projection.length - 1));
	}
	const hovered = $derived(hover === null ? null : projection[hover]);
	// The actual nearest to the hovered month, when one lies within it.
	const hoveredActual = $derived.by(() => {
		if (!hovered || actual.length === 0) return null;
		const t = ms(hovered.month);
		const nearest = actual.reduce((best, a) => (Math.abs(ms(a.time) - t) < Math.abs(ms(best.time) - t) ? a : best));
		return Math.abs(ms(nearest.time) - t) <= 32 * 864e5 ? nearest : null;
	});
	const monthFmt = new Intl.DateTimeFormat(undefined, { month: 'long', year: 'numeric', timeZone: 'UTC' });
</script>

<div class="flex flex-col gap-2">
	<div
		class="relative h-[280px] touch-none"
		role="img"
		aria-label="Net worth projected over the plan's horizon, with the range of simulated outcomes and the actual net worth so far"
		onpointermove={track}
		onpointerleave={() => (hover = null)}
	>
		<svg viewBox="0 0 {W} {H}" preserveAspectRatio="none" class="block h-full w-full" aria-hidden="true">
			{#each [PAD, H / 2, H - PAD] as gy (gy)}
				<line x1="0" x2={W} y1={gy} y2={gy} class="stroke-line" vector-effect="non-scaling-stroke" />
			{/each}
			{#if scale.lo < 0}
				<line x1="0" x2={W} y1={y(0)} y2={y(0)} class="stroke-line-strong" vector-effect="non-scaling-stroke" />
			{/if}
			<path d={outer} class="fill-accent/10" />
			<path d={inner} class="fill-accent/20" />
			<path d={contributions} fill="none" class="stroke-muted" stroke-width="1.5" stroke-dasharray="1.5 3" stroke-linecap="round" vector-effect="non-scaling-stroke" />
			<path d={expected} fill="none" class="stroke-accent-fg" stroke-width="1.5" stroke-dasharray="4 4" vector-effect="non-scaling-stroke" />
			{#if actual.length > 1}
				<path d={actualLine} fill="none" class="stroke-accent" stroke-width="2" vector-effect="non-scaling-stroke" />
			{/if}
		</svg>

		<div class="num pointer-events-none absolute inset-y-0 right-0 flex flex-col justify-between text-[11px] text-muted">
			{#each yLabels as label, i (i)}
				<span class="bg-surface pl-1">{label}</span>
			{/each}
		</div>

		{#if hovered}
			{@const left = (x(ms(hovered.month)) / W) * 100}
			<div class="pointer-events-none absolute inset-y-0 w-px bg-line-strong" style="left: {left}%"></div>
			<div
				class="pointer-events-none absolute size-2.5 -translate-x-1/2 -translate-y-1/2 rounded-full border-2 border-surface bg-accent-fg"
				style="left: {left}%; top: {(y(hovered.expected.amount) / H) * 100}%"
			></div>
			{#if hoveredActual}
				<div
					class="pointer-events-none absolute size-2.5 -translate-x-1/2 -translate-y-1/2 rounded-full border-2 border-surface bg-accent"
					style="left: {(x(ms(hoveredActual.time)) / W) * 100}%; top: {(y(hoveredActual.netWorth.amount) / H) * 100}%"
				></div>
			{/if}
			<div
				class="num pointer-events-none absolute top-2 z-10 w-60 rounded-lg border border-line bg-surface px-3 py-2 text-xs shadow-lg"
				style={left > 60 ? `right: ${100 - left + 1.5}%` : `left: ${left + 1.5}%`}
			>
				<div class="mb-1 font-medium">{monthFmt.format(new Date(hovered.month))}</div>
				{#if hoveredActual}
					<div class="flex justify-between gap-3"><span class="text-muted">Actual</span>{formatMoney(hoveredActual.netWorth)}</div>
				{/if}
				<div class="flex justify-between gap-3"><span class="text-muted">Expected</span>{formatMoney(hovered.expected)}</div>
				<div class="flex justify-between gap-3"><span class="text-muted">Likely range</span>{formatMoneyCompact(hovered.p25)} – {formatMoneyCompact(hovered.p75)}</div>
				<div class="flex justify-between gap-3"><span class="text-muted">10–90 %</span>{formatMoneyCompact(hovered.p10)} – {formatMoneyCompact(hovered.p90)}</div>
				<div class="flex justify-between gap-3"><span class="text-muted">Put in</span>{formatMoney({ amount: putIn(hovered), currency })}</div>
			</div>
		{/if}
	</div>

	<div class="flex justify-between text-[11px] text-muted">
		{#each xLabels as label, i (i)}
			<span>{label}</span>
		{/each}
	</div>
	<div class="flex flex-wrap gap-4 text-xs text-muted">
		{#if actual.length > 1}
			<span class="flex items-center gap-1.5"><span class="h-0.5 w-3.5 bg-accent"></span>Actual</span>
		{/if}
		<span class="flex items-center gap-1.5"><span class="w-3.5 border-t-2 border-dashed border-accent-fg"></span>Expected</span>
		<span class="flex items-center gap-1.5"><span class="h-2.5 w-3.5 rounded-sm bg-accent/20"></span>Likely range (25–75 %)</span>
		<span class="flex items-center gap-1.5"><span class="h-2.5 w-3.5 rounded-sm bg-accent/10"></span>10–90 % of simulations</span>
		<span class="flex items-center gap-1.5"><span class="w-3.5 border-t-2 border-dotted border-muted"></span>Put in (start + contributions)</span>
	</div>
</div>
