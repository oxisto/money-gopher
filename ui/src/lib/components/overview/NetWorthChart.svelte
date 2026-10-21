<script lang="ts">
	import { formatMoney, formatMoneyCompact, type Money } from '$lib/format';

	interface Point {
		time: Date;
		netWorth: Money;
		invested: Money;
	}

	let { points }: { points: Point[] } = $props();

	// The SVG is stretched to the container (preserveAspectRatio="none"), so
	// it is drawn in a fixed 1000×H box; anything that must not stretch —
	// labels, the hover dot — is HTML positioned in percent on top.
	const W = 1000;
	const H = 260;
	const PAD = 10;

	const scale = $derived.by(() => {
		const values = points.flatMap((p) => [p.netWorth.amount, p.invested.amount]);
		const lo = Math.min(...values);
		const hi = Math.max(...values);
		const pad = (hi - lo) * 0.08 || Math.abs(hi) * 0.05 || 100;
		// Padding must not invent a negative axis for values that never are.
		return { lo: lo >= 0 ? Math.max(0, lo - pad) : lo - pad, hi: hi + pad };
	});

	const x = (i: number) => (points.length < 2 ? 0 : (i / (points.length - 1)) * W);
	const y = (v: number) => PAD + (1 - (v - scale.lo) / (scale.hi - scale.lo)) * (H - 2 * PAD);
	const path = (values: number[]) =>
		values.map((v, i) => `${i ? 'L' : 'M'}${x(i).toFixed(1)},${y(v).toFixed(1)}`).join(' ');

	const line = $derived(path(points.map((p) => p.netWorth.amount)));
	const invested = $derived(path(points.map((p) => p.invested.amount)));
	const currency = $derived(points[0]?.netWorth.currency ?? 'EUR');

	const yLabels = $derived(
		[scale.hi, (scale.hi + scale.lo) / 2, scale.lo].map((v) =>
			formatMoneyCompact({ amount: v, currency })
		)
	);

	const xLabels = $derived.by(() => {
		if (points.length < 2) return [];
		const span = points[points.length - 1].time.getTime() - points[0].time.getTime();
		const fmt = new Intl.DateTimeFormat(
			undefined,
			span > 400 * 864e5 ? { month: 'short', year: 'numeric' } : { month: 'short', day: 'numeric' }
		);
		return [0, 0.25, 0.5, 0.75, 1].map((f) => fmt.format(points[Math.round(f * (points.length - 1))].time));
	});

	let hover = $state<number | null>(null);

	function track(event: PointerEvent) {
		const rect = (event.currentTarget as HTMLElement).getBoundingClientRect();
		const f = Math.min(1, Math.max(0, (event.clientX - rect.left) / rect.width));
		hover = Math.round(f * (points.length - 1));
	}

	const hovered = $derived(hover === null ? null : points[hover]);
	const dateFmt = new Intl.DateTimeFormat(undefined, { dateStyle: 'medium' });
</script>

<div class="flex flex-col gap-2">
	<div
		class="relative h-[260px] touch-none"
		role="img"
		aria-label="Net worth over time, compared with invested capital"
		onpointermove={track}
		onpointerleave={() => (hover = null)}
	>
		<svg viewBox="0 0 {W} {H}" preserveAspectRatio="none" class="block h-full w-full" aria-hidden="true">
			{#each [PAD, H / 2, H - PAD] as gy (gy)}
				<line x1="0" x2={W} y1={gy} y2={gy} class="stroke-line" vector-effect="non-scaling-stroke" />
			{/each}
			<path d="{line} L{W},{H} L0,{H} Z" class="fill-accent-soft" />
			<path
				d={invested}
				fill="none"
				class="stroke-muted"
				stroke-width="1.5"
				stroke-dasharray="4 4"
				vector-effect="non-scaling-stroke"
			/>
			<path d={line} fill="none" class="stroke-accent" stroke-width="2" vector-effect="non-scaling-stroke" />
		</svg>

		<div class="num pointer-events-none absolute inset-y-0 right-0 flex flex-col justify-between text-[11px] text-muted">
			{#each yLabels as label, i (i)}
				<span class="bg-surface pl-1">{label}</span>
			{/each}
		</div>

		{#if hovered && hover !== null}
			{@const left = (x(hover) / W) * 100}
			<div class="pointer-events-none absolute inset-y-0 w-px bg-line-strong" style="left: {left}%"></div>
			<div
				class="pointer-events-none absolute size-2.5 -translate-x-1/2 -translate-y-1/2 rounded-full border-2 border-surface bg-accent"
				style="left: {left}%; top: {(y(hovered.netWorth.amount) / H) * 100}%"
			></div>
			<div
				class="num pointer-events-none absolute top-2 z-10 w-52 rounded-lg border border-line bg-surface px-3 py-2 text-xs shadow-lg"
				style={left > 60 ? `right: ${100 - left + 1.5}%` : `left: ${left + 1.5}%`}
			>
				<div class="mb-1 font-medium">{dateFmt.format(hovered.time)}</div>
				<div class="flex justify-between gap-3">
					<span class="text-muted">Net worth</span>{formatMoney(hovered.netWorth)}
				</div>
				<div class="flex justify-between gap-3">
					<span class="text-muted">Invested</span>{formatMoney(hovered.invested)}
				</div>
				<div class="flex justify-between gap-3">
					<span class="text-muted">Earned</span>
					<span class={hovered.netWorth.amount < hovered.invested.amount ? 'text-neg' : 'text-pos'}>
						{formatMoney({
							amount: hovered.netWorth.amount - hovered.invested.amount,
							currency: hovered.netWorth.currency
						})}
					</span>
				</div>
			</div>
		{/if}
	</div>

	<div class="flex justify-between text-[11px] text-muted">
		{#each xLabels as label, i (i)}
			<span>{label}</span>
		{/each}
	</div>
	<div class="flex flex-wrap gap-4 text-xs text-muted">
		<span class="flex items-center gap-1.5"><span class="h-0.5 w-3.5 bg-accent"></span>Net worth</span>
		<span class="flex items-center gap-1.5">
			<span class="w-3.5 border-t-2 border-dashed border-muted"></span>Invested (deposits − withdrawals)
		</span>
	</div>
</div>
