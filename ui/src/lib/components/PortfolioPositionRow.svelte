<script lang="ts">
	import type { PortfolioPosition } from '$lib/gen/mgo_pb';
	import { currency } from '$lib/intl';
	import { ArrowDown, ArrowRight, ArrowUp } from '@steeze-ui/heroicons';
	import { Icon } from '@steeze-ui/svelte-icon';

	export let position: PortfolioPosition;

	$: perf = ((position.marketPrice - position.purchasePrice) / position.purchasePrice) * 100;

	function shorten(text: string): string {
		let max = 30;

		if (text.length > max) {
			return text.substring(0, max) + '...';
		} else {
			return text;
		}
	}
</script>

{#if position.security}
	<tr>
		<td class="truncate whitespace-nowrap py-2 pl-4 pr-3 text-sm font-medium sm:pl-0">
			<div class="text-gray-900">
				{shorten(position.security.displayName)}
			</div>
			<div class="text-gray-400">
				{position.security.name}
			</div>
		</td>
		<td class="hidden whitespace-nowrap px-3 py-2 text-right text-sm text-gray-500 md:table-cell">
			{Intl.NumberFormat('de', {
				maximumFractionDigits: 2
			}).format(position.amount)}
		</td>
		<td class="hidden whitespace-nowrap px-3 py-2 text-right text-sm lg:table-cell">
			<div class="text-gray-500">
				{currency(position.purchasePrice, 'EUR')}
			</div>
			<div class="text-gray-400">
				{currency(position.purchaseValue, 'EUR')}
			</div>
		</td>
		<td class="hidden whitespace-nowrap px-3 py-2 text-right text-sm sm:table-cell">
			<div class="text-gray-500">
				{currency(position.marketPrice, 'EUR')}
			</div>
			<div class="text-gray-400">
				{currency(position.marketValue, 'EUR')}
			</div>
		</td>
		<td
			class="{perf < 0
				? 'text-red-500'
				: perf <= 1
				? 'text-gray-500'
				: 'text-green-500'} whitespace-nowrap px-3 py-2 text-right text-sm"
		>
			<div>
				{Intl.NumberFormat('de', {
					maximumFractionDigits: 2
				}).format(perf)} %
				<Icon
					src={perf < 0 ? ArrowDown : perf < 1 ? ArrowRight : ArrowUp}
					class="float-right mt-0.5 h-4 w-4"
					aria-hidden="true"
				/>
			</div>
			<div class="pr-4">
				{currency(position.marketValue - position.purchaseValue, 'EUR')}
			</div>
		</td>
	</tr>
{/if}
