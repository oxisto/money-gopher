<script lang="ts">
	import type { PortfolioDetail$result } from '$houdini';
	import { formatMoney, formatPercent } from '$lib/format';

	type Snapshot = NonNullable<PortfolioDetail$result['portfolio']>['snapshot'];

	let { snapshot }: { snapshot: Snapshot } = $props();

	const cards = $derived([
		{ label: 'Market value', value: formatMoney(snapshot.totalMarketValue), tone: '' },
		{
			label: 'Profit / loss',
			value: `${formatMoney(snapshot.totalProfitOrLoss)} (${formatPercent(snapshot.totalGains)})`,
			tone: snapshot.totalProfitOrLoss.amount < 0 ? 'text-red-600' : 'text-green-700'
		},
		{
			label: 'Time-weighted return (all time)',
			value: formatPercent(snapshot.performance.timeWeightedReturn),
			tone: snapshot.performance.timeWeightedReturn < 0 ? 'text-red-600' : 'text-green-700'
		}
	]);
</script>

<div class="grid gap-4 sm:grid-cols-3">
	{#each cards as card (card.label)}
		<div class="rounded-md border border-gray-200 bg-white px-4 py-3">
			<div class="text-xs text-gray-500 uppercase">{card.label}</div>
			<div class="mt-1 text-lg font-semibold tabular-nums {card.tone}">{card.value}</div>
		</div>
	{/each}
</div>
