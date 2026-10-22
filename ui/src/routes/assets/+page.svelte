<script lang="ts">
	import AssetForm from '#lib/components/AssetForm.svelte';
	import { holdingsOf } from '#lib/components/holdings/HoldingsTable.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import Dialog from '#lib/components/ui/Dialog.svelte';
	import Icon from '#lib/components/ui/Icon.svelte';
	import { formatDate, formatMoney, formatSignedMoney, toneOf } from '#lib/format.js';
	import { kindLabel, type AssetKind } from '#lib/assets.js';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let addOpen = $state(false);

	// The asset's position, merged across portfolios like Holdings does; an
	// asset with none yet (its transfer-in deleted) still lists, at nothing.
	const holdings = $derived(
		new Map(holdingsOf(data.portfolios.flatMap((p) => p.snapshot.positions)).map((h) => [h.assetId, h]))
	);
	const rows = $derived(
		data.assets.map((asset) => ({
			asset,
			holding: holdings.get(asset.id) ?? null,
			valuedAt: asset.listings[0]?.latestQuote?.time ?? null
		}))
	);
</script>

<div class="mb-6 flex flex-wrap items-center gap-3">
	<div class="flex-1">
		<h1 class="text-lg font-semibold">Assets</h1>
		<p class="text-[13px] text-muted">
			Yours alone: property, vehicles, loans you gave. Valued by hand, counted in your net worth.
		</p>
	</div>
	<Button onclick={() => (addOpen = true)}><Icon name="plus" size={15} /> Book asset</Button>
</div>

<Dialog bind:open={addOpen} title="Book an asset" wide>
	<AssetForm cashAccounts={data.cashAccounts} oncreated={() => (addOpen = false)} />
</Dialog>

{#if rows.length === 0}
	<div class="rounded-xl border border-dashed border-line p-8 text-center text-sm text-muted">
		Nothing booked yet. The office room you rent out, the car, money you lent: book it with what it
		cost and what it is worth, and it joins your holdings.
	</div>
{:else}
	<div class="overflow-x-auto rounded-xl border border-line">
		<table class="num w-full text-sm">
			<thead class="bg-raised text-xs text-muted">
				<tr>
					<th scope="col" class="py-2.5 pr-3 pl-4 text-left font-medium">Asset</th>
					<th scope="col" class="px-3 py-2.5 text-right font-medium">Cost basis</th>
					<th scope="col" class="px-3 py-2.5 text-right font-medium">Value</th>
					<th scope="col" class="px-3 py-2.5 text-right font-medium">Unrealized</th>
					<th scope="col" class="py-2.5 pr-4 pl-3 text-right font-medium">Last valuation</th>
				</tr>
			</thead>
			<tbody>
				{#each rows as { asset, holding, valuedAt } (asset.id)}
					<tr class="border-t border-line hover:bg-raised">
						<td class="py-2.5 pr-3 pl-4">
							<a href="/holdings/{asset.id}" class="font-medium hover:underline">{asset.displayName}</a>
							<div class="text-[11px] text-muted">
								{asset.kind ? kindLabel[asset.kind as AssetKind] : 'Asset'}
								{#if asset.listings[0]}
									· in {asset.listings[0].currency}
								{/if}
							</div>
						</td>
						<td class="px-3 py-2.5 text-right">{holding ? formatMoney(holding.purchaseValue) : '—'}</td>
						<td class="px-3 py-2.5 text-right">{holding ? formatMoney(holding.marketValue) : '—'}</td>
						<td class="px-3 py-2.5 text-right {holding ? toneOf(holding.profitOrLoss.amount) : ''}">
							{holding ? formatSignedMoney(holding.profitOrLoss) : '—'}
						</td>
						<td class="py-2.5 pr-4 pl-3 text-right text-muted">
							{valuedAt ? formatDate(valuedAt) : 'at cost'}
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
{/if}
