<script lang="ts">
	import type { PortfolioPosition, PortfolioSnapshot } from '$lib/gen/mgo_pb';
	import { currency } from '$lib/intl';
	import { ArrowDown, ArrowRight, ArrowUp } from '@steeze-ui/heroicons';
	import { Icon } from '@steeze-ui/svelte-icon';
	import TableSorter from '$lib/components/TableSorter.svelte';
	import PortfolioPositionRow from '$lib/components/PortfolioPositionRow.svelte';

	export let snapshot: PortfolioSnapshot;

	const sorters = new Map<string, (a: PortfolioPosition, b: PortfolioPosition) => number>();
	sorters.set('displayName', (a: PortfolioPosition, b: PortfolioPosition) => {
		return a.security?.displayName.localeCompare(b.security?.displayName ?? '') ?? 0;
	});
	sorters.set('amount', (a: PortfolioPosition, b: PortfolioPosition) => {
		return a.amount - b.amount;
	});
	sorters.set('purchaseValue', (a: PortfolioPosition, b: PortfolioPosition) => {
		return a.purchaseValue - b.purchaseValue;
	});
	sorters.set('marketValue', (a: PortfolioPosition, b: PortfolioPosition) => {
		return a.marketValue - b.marketValue;
	});

	let sortBy = 'displayName';
	let asc = true;

	$: positions = getPositions(snapshot, sortBy, asc);

	function getPositions(
		snapshot: PortfolioSnapshot,
		sortBy: string,
		asc: boolean
	): PortfolioPosition[] {
		let positions = Object.values(snapshot.positions);
		return positions.sort((a: PortfolioPosition, b: PortfolioPosition) => {
			const sort = sorters.get(sortBy)?.call(null, a, b) ?? 0;
			return asc ? sort : -sort;
		});
	}

	$: perf =
		((snapshot.totalMarketValue - snapshot.totalPurchaseValue) / snapshot.totalPurchaseValue) * 100;

	function toggleSortDirection() {
		asc = !asc;
	}

	function changeSortBy(column: string) {
		sortBy = column;
	}
</script>

{{ asc }}
{{ sortBy }}
<div class="-mx-4 mt-8 sm:-mx-0">
	<table class="min-w-full divide-y divide-gray-300">
		<thead>
			<tr>
				<th
					scope="col"
					class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-0"
				>
					<TableSorter
						active={sortBy == 'displayName'}
						column="displayName"
						on:change-direction={toggleSortDirection}
						on:change-sort-by={(column) => changeSortBy('displayName')}
					>
						Name
					</TableSorter>
				</th>
				<th
					scope="col"
					class="hidden px-3 py-3.5 text-right text-sm font-semibold text-gray-900 md:table-cell"
				>
					<TableSorter
						active={sortBy == 'amount'}
						column="amount"
						on:change-direction={toggleSortDirection}
						on:change-sort-by={(column) => changeSortBy('amount')}>Amount</TableSorter
					>
				</th>
				<th
					scope="col"
					class="hidden px-3 py-3.5 text-right text-sm font-semibold text-gray-900 lg:table-cell"
				>
					<TableSorter
						active={sortBy == 'purchaseValue'}
						column="purchaseValue"
						on:change-direction={toggleSortDirection}
						on:change-sort-by={(column) => changeSortBy('purchaseValue')}
						>Purchase Value
					</TableSorter>
				</th>
				<th
					scope="col"
					class="hidden px-3 py-3.5 text-right text-sm font-semibold text-gray-900 sm:table-cell"
				>
					<TableSorter
						active={sortBy == 'marketValue'}
						column="marketValue"
						on:change-direction={toggleSortDirection}
						on:change-sort-by={(column) => changeSortBy('marketValue')}>Market Value</TableSorter
					>
				</th>
				<th scope="col" class="px-3 py-3.5 text-right text-sm font-semibold text-gray-900">
					<TableSorter
						active={sortBy == 'profit'}
						column="profit"
						on:change-direction={toggleSortDirection}
						on:change-sort-by={(column) => changeSortBy('profit')}>Profit/Loss</TableSorter
					>
				</th>
			</tr>
		</thead>
		<tbody class="divide-y divide-gray-200">
			{#each positions as position (position.security?.name)}
				<PortfolioPositionRow {position} />
			{/each}
		</tbody>
		<tfoot>
			<th scope="col" class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-0"
				>Total</th
			>
			<th
				scope="col"
				class="hidden px-3 py-3.5 text-right text-sm font-semibold text-gray-900 md:table-cell"
			></th>
			<th
				scope="col"
				class="hidden px-3 py-3.5 text-right text-sm font-semibold text-gray-900 lg:table-cell"
			>
				{currency(snapshot.totalPurchaseValue, 'EUR')}
			</th>
			<th
				scope="col"
				class="hidden px-3 py-3.5 text-right text-sm font-semibold text-gray-900 sm:table-cell"
			>
				{currency(snapshot.totalMarketValue, 'EUR')}
			</th>
			<th
				scope="col"
				class="{perf < 0
					? 'text-red-500'
					: perf <= 1
					? 'text-gray-500'
					: 'text-green-500'} px-3 py-3.5 text-right text-sm font-semibold"
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
					{currency(snapshot.totalMarketValue - snapshot.totalPurchaseValue, 'EUR')}
				</div>
			</th>
		</tfoot>
	</table>
</div>
