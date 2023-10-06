<script lang="ts">
	import type { PortfolioSnapshot } from '$lib/gen/mgo_pb';
	import { currency } from '$lib/intl';
	import { ArrowTrendingUp, ArrowTrendingDown } from '@steeze-ui/heroicons';
	import { Icon } from '@steeze-ui/svelte-icon';

	export let snapshot: PortfolioSnapshot;
	export let icon: boolean = true;

	$: perf =
		((snapshot.totalMarketValue - snapshot.totalPurchaseValue) / snapshot.totalPurchaseValue) * 100;
	$: perfAbs = snapshot.totalMarketValue - snapshot.totalPurchaseValue;
</script>

<div
	class="{perf < 0 ? 'text-red-400' : 'text-green-400'} 
               flex items-center"
>
	{#if icon}
		<Icon
			src={perf > 0 ? ArrowTrendingUp : ArrowTrendingDown}
			class="{perf < 0 ? 'text-red-400' : 'text-green-400'} 
                 mr-1.5 h-5 w-5 flex-shrink-0"
			aria-hidden="true"
		/>
	{/if}
	{Intl.NumberFormat(navigator.language, { maximumFractionDigits: 2 }).format(perf)} % ({currency(
		perfAbs,
		'EUR'
	)})
</div>
