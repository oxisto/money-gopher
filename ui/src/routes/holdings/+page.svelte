<script lang="ts">
	import HoldingsTable, { holdingsOf } from '$lib/components/holdings/HoldingsTable.svelte';
	import StatStrip from '$lib/components/ui/StatStrip.svelte';
	import { formatMoney, formatPercent, formatSignedMoney, toneOf, type Money } from '$lib/format';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	const Holdings = $derived(data.Holdings);

	const portfolios = $derived($Holdings.data?.portfolios ?? []);
	const holdings = $derived(holdingsOf(portfolios.flatMap((p) => p.snapshot.positions)));

	const sum = (values: Money[]): Money => ({
		amount: values.reduce((s, m) => s + m.amount, 0),
		currency: values[0]?.currency ?? 'EUR'
	});

	const stats = $derived.by(() => {
		const value = sum(portfolios.map((p) => p.snapshot.totalMarketValue));
		const cost = sum(portfolios.map((p) => p.snapshot.totalPurchaseValue));
		const pol = sum(portfolios.map((p) => p.snapshot.totalProfitOrLoss));
		// Value-weighted across portfolios, like the overview.
		const twr = value.amount
			? portfolios.reduce(
					(s, p) => s + p.snapshot.performance.timeWeightedReturn * p.snapshot.totalMarketValue.amount,
					0
				) / value.amount
			: 0;
		return [
			{
				label: 'Market value',
				value: formatMoney(value),
				sub: `${holdings.length} positions · ${portfolios.length} ${portfolios.length === 1 ? 'portfolio' : 'portfolios'}`
			},
			{ label: 'Cost basis', value: formatMoney(cost), sub: 'FIFO, without fees and taxes' },
			{
				label: 'Unrealized gain',
				value: formatSignedMoney(pol),
				sub: cost.amount ? formatPercent(pol.amount / cost.amount) : undefined,
				tone: toneOf(pol.amount)
			},
			{ label: 'Return, all time', value: formatPercent(twr), sub: 'time-weighted', tone: toneOf(twr) }
		];
	});
</script>

<div class="mb-6 flex flex-wrap items-center gap-3">
	<h1 class="flex-1 text-lg font-semibold">Holdings</h1>
</div>

{#if $Holdings.fetching && !$Holdings.data}
	<p class="text-sm text-muted">Loading…</p>
{:else if $Holdings.errors?.length}
	<p class="text-sm text-neg">Failed to load: {$Holdings.errors[0].message}</p>
{:else}
	<div class="flex flex-col gap-6">
		<StatStrip {stats} />
		<HoldingsTable {holdings} />
	</div>
{/if}
