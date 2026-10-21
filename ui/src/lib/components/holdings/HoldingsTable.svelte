<script lang="ts" module>
	import type { Money, Price } from '#lib/format.js';

	export interface Holding {
		securityId: string;
		name: string;
		isin: string | null;
		units: number;
		purchasePrice: Price;
		marketPrice: Price;
		purchaseValue: Money;
		marketValue: Money;
		profitOrLoss: Money;
		/** Unrealized gain as a fraction, e.g. 0.05 for +5 %. */
		gains: number;
	}

	interface Position {
		security: { id: string; displayName: string; identifiers: { kind: string; value: string }[] };
		units: number;
		purchasePrice: Price;
		marketPrice: Price;
		purchaseValue: Money;
		marketValue: Money;
		profitOrLoss: Money;
	}

	/**
	 * Merges the positions of several portfolios into one holding per
	 * security; units and values add up, the average cost is recomputed.
	 */
	export function holdingsOf(positions: Position[]): Holding[] {
		const bySecurity = new Map<string, Holding>();
		for (const p of positions) {
			const h = bySecurity.get(p.security.id);
			if (!h) {
				bySecurity.set(p.security.id, {
					securityId: p.security.id,
					name: p.security.displayName,
					isin: p.security.identifiers.find((i) => i.kind === 'ISIN')?.value ?? null,
					units: p.units,
					purchasePrice: p.purchasePrice,
					marketPrice: p.marketPrice,
					purchaseValue: { ...p.purchaseValue },
					marketValue: { ...p.marketValue },
					profitOrLoss: { ...p.profitOrLoss },
					gains: 0
				});
				continue;
			}
			h.units += p.units;
			h.purchaseValue.amount += p.purchaseValue.amount;
			h.marketValue.amount += p.marketValue.amount;
			h.profitOrLoss.amount += p.profitOrLoss.amount;
			h.purchasePrice = {
				amount: String(h.purchaseValue.amount / 100 / h.units),
				currency: h.purchasePrice.currency
			};
		}
		return [...bySecurity.values()].map((h) => ({
			...h,
			gains: h.purchaseValue.amount ? h.profitOrLoss.amount / h.purchaseValue.amount : 0
		}));
	}
</script>

<script lang="ts">
	import Icon from '#lib/components/ui/Icon.svelte';
	import { formatUnits } from '#lib/describe.js';
	import { formatMoney, formatPercent, formatPrice, formatSignedMoney, toneOf } from '#lib/format.js';

	let { holdings }: { holdings: Holding[] } = $props();

	type SortKey = 'name' | 'value' | 'return';
	let sort = $state<SortKey>('value');
	let descending = $state(true);
	let query = $state('');

	const total = $derived(holdings.reduce((s, h) => s + h.marketValue.amount, 0) || 1);
	const maxWeight = $derived(Math.max(...holdings.map((h) => h.marketValue.amount), 1) / total);

	const rows = $derived.by(() => {
		const q = query.trim().toLowerCase();
		const filtered = q
			? holdings.filter((h) => h.name.toLowerCase().includes(q) || h.isin?.toLowerCase().includes(q))
			: holdings;
		const key = {
			name: (h: Holding) => h.name.toLowerCase(),
			value: (h: Holding) => h.marketValue.amount,
			return: (h: Holding) => h.gains
		}[sort];
		const dir = descending ? -1 : 1;
		return [...filtered].sort((a, b) => (key(a) < key(b) ? -dir : key(a) > key(b) ? dir : 0));
	});

	const shown = $derived({
		value: rows.reduce((s, h) => s + h.marketValue.amount, 0),
		profitOrLoss: rows.reduce((s, h) => s + h.profitOrLoss.amount, 0),
		currency: holdings[0]?.marketValue.currency ?? 'EUR'
	});

	function sortBy(key: SortKey) {
		if (sort === key) {
			descending = !descending;
		} else {
			sort = key;
			// Names read A→Z; numbers are most interesting biggest first.
			descending = key !== 'name';
		}
	}
</script>

{#snippet header(key: SortKey, label: string, align: 'left' | 'right')}
	<th
		scope="col"
		aria-sort={sort === key ? (descending ? 'descending' : 'ascending') : 'none'}
		class="px-3 font-medium first:pl-4 {align === 'right' ? 'text-right' : 'text-left'}"
	>
		<button
			type="button"
			onclick={() => sortBy(key)}
			class="py-2.5 {sort === key ? 'text-fg' : 'hover:text-fg'}"
		>
			{label}{sort === key ? (descending ? ' ↓' : ' ↑') : ''}
		</button>
	</th>
{/snippet}

<div class="flex flex-col gap-3">
	<label class="relative block max-w-xs">
		<span class="sr-only">Filter holdings</span>
		<span class="pointer-events-none absolute inset-y-0 left-2.5 grid place-items-center text-muted">
			<Icon name="search" size={15} />
		</span>
		<input type="search" bind:value={query} placeholder="Filter by name or ISIN" class="h-9 w-full pl-8" />
	</label>

	{#if holdings.length === 0}
		<p class="rounded-xl border border-dashed border-line px-4 py-8 text-center text-sm text-muted">
			Nothing held yet. Buys from imported statements show up here.
		</p>
	{:else}
		<div class="overflow-x-auto rounded-xl border border-line">
			<table class="w-full min-w-[52rem] text-[13px]">
				<thead class="bg-raised text-xs text-muted">
					<tr>
						{@render header('name', 'Security', 'left')}
						<th scope="col" class="px-3 py-2.5 text-right font-medium">Units</th>
						<th scope="col" class="px-3 py-2.5 text-right font-medium">Avg cost</th>
						<th scope="col" class="px-3 py-2.5 text-right font-medium">Price</th>
						{@render header('value', 'Value', 'right')}
						{@render header('return', 'Unrealized', 'right')}
						<th scope="col" class="py-2.5 pr-4 pl-3 text-right font-medium">Weight</th>
					</tr>
				</thead>
				<tbody>
					{#each rows as h (h.securityId)}
						{@const weight = h.marketValue.amount / total}
						<tr class="border-t border-line hover:bg-raised">
							<td class="py-2 pr-3 pl-4">
								<a href="/holdings/{h.securityId}" class="font-medium hover:underline">{h.name}</a>
								{#if h.isin}<div class="text-[11px] text-muted">{h.isin}</div>{/if}
							</td>
							<td class="px-3 py-2 text-right">{formatUnits(h.units)}</td>
							<td class="px-3 py-2 text-right text-muted">{formatPrice(h.purchasePrice, 2)}</td>
							<td class="px-3 py-2 text-right">{formatPrice(h.marketPrice, 4)}</td>
							<td class="px-3 py-2 text-right font-medium">{formatMoney(h.marketValue)}</td>
							<td class="px-3 py-2 text-right {toneOf(h.profitOrLoss.amount)}">
								<div>{formatSignedMoney(h.profitOrLoss)}</div>
								<div class="text-[11px]">{formatPercent(h.gains)}</div>
							</td>
							<td class="py-2 pr-4 pl-3">
								<div class="flex items-center justify-end gap-2">
									<span class="flex h-1.5 w-16 overflow-hidden rounded-full bg-line" aria-hidden="true">
										<span class="bg-accent" style="width: {(weight / maxWeight) * 100}%"></span>
									</span>
									<span class="w-11 text-right">{(weight * 100).toFixed(1)}%</span>
								</div>
							</td>
						</tr>
					{/each}
				</tbody>
				<tfoot class="font-semibold">
					<tr class="border-t border-line-strong">
						<td class="py-2.5 pr-3 pl-4">
							{rows.length === holdings.length ? `${rows.length} positions` : `${rows.length} of ${holdings.length} positions`}
						</td>
						<td colspan="3"></td>
						<td class="px-3 py-2.5 text-right">{formatMoney({ amount: shown.value, currency: shown.currency })}</td>
						<td class="px-3 py-2.5 text-right {toneOf(shown.profitOrLoss)}">
							{formatSignedMoney({ amount: shown.profitOrLoss, currency: shown.currency })}
						</td>
						<td class="py-2.5 pr-4 pl-3 text-right">{((shown.value / total) * 100).toFixed(1)}%</td>
					</tr>
				</tfoot>
			</table>
		</div>
	{/if}
</div>
