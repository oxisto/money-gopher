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
		return (a.purchaseValue?.value ?? 0) - (b.purchaseValue?.value ?? 0);
	});
	sorters.set('marketValue', (a: PortfolioPosition, b: PortfolioPosition) => {
		return (a.marketValue?.value ?? 0) - (b.marketValue?.value ?? 0);
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

	function toggleSortDirection() {
		asc = !asc;
	}

	function changeSortBy(column: string) {
		sortBy = column;
	}
</script>

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
						on:changeDirection={toggleSortDirection}
						on:changeSortBy={(column) => changeSortBy('displayName')}
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
						on:changeDirection={toggleSortDirection}
						on:changeSortBy={(column) => changeSortBy('amount')}>Amount</TableSorter
					>
				</th>
				<th
					scope="col"
					class="hidden px-3 py-3.5 text-right text-sm font-semibold text-gray-900 lg:table-cell"
				>
					<TableSorter
						active={sortBy == 'purchaseValue'}
						column="purchaseValue"
						on:changeDirection={toggleSortDirection}
						on:changeSortBy={(column) => changeSortBy('purchaseValue')}
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
						on:changeDirection={toggleSortDirection}
						on:changeSortBy={(column) => changeSortBy('marketValue')}>Market Value</TableSorter
					>
				</th>
				<th scope="col" class="px-3 py-3.5 text-right text-sm font-semibold text-gray-900">
					<TableSorter
						active={sortBy == 'profit'}
						column="profit"
						on:changeDirection={toggleSortDirection}
						on:changeSortBy={(column) => changeSortBy('profit')}>Profit/Loss</TableSorter
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
			<tr>
				<th
					scope="col"
					class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-0"
					>Total Assets</th
				>
				<th
					scope="col"
					class="hidden px-3 py-3.5 text-right text-sm font-semibold text-gray-900 md:table-cell"
				></th>
				<th
					scope="col"
					class="hidden px-3 py-3.5 text-right text-sm font-semibold text-gray-900 lg:table-cell"
				>
					{currency(snapshot.totalPurchaseValue)}
				</th>
				<th
					scope="col"
					class="hidden px-3 py-3.5 text-right text-sm font-semibold text-gray-900 sm:table-cell"
				>
					{currency(snapshot.totalMarketValue)}
				</th>
				<th
					scope="col"
					class="{snapshot.totalGains < 0
						? 'text-red-500'
						: snapshot.totalGains <= 0.01
							? 'text-gray-500'
							: 'text-green-500'} px-3 py-3.5 text-right text-sm font-semibold"
				>
					<div>
						{Intl.NumberFormat(navigator.language, {
							maximumFractionDigits: 2
						}).format(snapshot.totalGains * 100)} %
						<Icon
							src={snapshot.totalGains < 0
								? ArrowDown
								: snapshot.totalGains < 0.01
									? ArrowRight
									: ArrowUp}
							class="float-right mt-0.5 h-4 w-4"
							aria-hidden="true"
						/>
					</div>
					<div class="pr-4">
						{currency(snapshot.totalProfitOrLoss)}
					</div>
				</th>
			</tr>
			<tr>
				<th
					scope="col"
					class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-0"
				>
					Cash Value
				</th>
				<th></th>
				<th></th>
				<th
					scope="col"
					class="{(snapshot.cash?.value ?? 0) < 0
						? 'text-red-500'
						: ''} px-3 py-3.5 text-right text-sm font-semibold text-gray-900 lg:table-cell"
				>
					{currency(snapshot.cash)}
				</th>
			</tr>
			<tr>
				<th
					scope="col"
					class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-0"
				>
					Total Portfolio Value
				</th>
				<th></th>
				<th></th>
				<th
					scope="col"
					class="{(snapshot.cash?.value ?? 0) < 0
						? 'text-red-500'
						: ''} px-3 py-3.5 text-right text-sm font-semibold text-gray-900 lg:table-cell"
				>
					{currency(snapshot.totalPortfolioValue)}
				</th>
			</tr>
		</tfoot>
	</table>
</div>
