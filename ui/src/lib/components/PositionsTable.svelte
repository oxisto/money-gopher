<script lang="ts">
	import type { PortfolioDetail$result } from '$houdini';
	import { formatMoney } from '$lib/format';
	import GainBadge from './ui/GainBadge.svelte';

	type Positions = NonNullable<PortfolioDetail$result['portfolio']>['snapshot']['positions'];

	let { positions }: { positions: Positions } = $props();
</script>

{#if positions.length === 0}
	<p class="text-sm text-gray-500">No positions yet.</p>
{:else}
	<div class="overflow-x-auto rounded-md border border-gray-200 bg-white">
		<table class="w-full text-sm">
			<thead>
				<tr class="border-b border-gray-200 text-left text-xs text-gray-500 uppercase">
					<th class="px-4 py-2 font-medium">Security</th>
					<th class="px-4 py-2 text-right font-medium">Units</th>
					<th class="px-4 py-2 text-right font-medium">Purchase price</th>
					<th class="px-4 py-2 text-right font-medium">Market price</th>
					<th class="px-4 py-2 text-right font-medium">Market value</th>
					<th class="px-4 py-2 text-right font-medium">Profit / loss</th>
				</tr>
			</thead>
			<tbody class="divide-y divide-gray-100">
				{#each positions as position (position.security.id)}
					<tr>
						<td class="px-4 py-2 font-medium">{position.security.displayName}</td>
						<td class="px-4 py-2 text-right tabular-nums">{position.units}</td>
						<td class="px-4 py-2 text-right tabular-nums">
							{formatMoney(position.purchasePrice)}
						</td>
						<td class="px-4 py-2 text-right tabular-nums">
							{formatMoney(position.marketPrice)}
						</td>
						<td class="px-4 py-2 text-right tabular-nums">
							{formatMoney(position.marketValue)}
						</td>
						<td class="px-4 py-2 text-right tabular-nums">
							<span class={position.profitOrLoss.amount < 0 ? 'text-red-600' : 'text-green-700'}>
								{formatMoney(position.profitOrLoss)}
							</span>
							<GainBadge gains={position.gains} />
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
{/if}
