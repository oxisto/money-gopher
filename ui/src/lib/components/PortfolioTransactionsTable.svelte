<script lang="ts">
	import PortfolioTransactionRow from '$lib/components/PortfolioTransactionRow.svelte';
	import TableSorter from '$lib/components/TableSorter.svelte';
	import type { PortfolioEvent, Security } from '$lib/gen/mgo_pb';

	export let transactions: PortfolioEvent[];
	export let securities: Security[];

	const sorters = new Map<string, (a: PortfolioEvent, b: PortfolioEvent) => number>();
		sorters.set('time', (a: PortfolioEvent, b: PortfolioEvent) => {
		return (a.time?.toDate() ?? 0 ) < (b.time?.toDate() ?? 0 )? -1 : 1;
	});
	sorters.set('securityName', (a: PortfolioEvent, b: PortfolioEvent) => {
		return a.securityName.localeCompare(b.securityName);
	});
	sorters.set('amount', (a: PortfolioEvent, b: PortfolioEvent) => {
		return a.amount - b.amount;
	});
	sorters.set('price', (a: PortfolioEvent, b: PortfolioEvent) => {
		return a.price - b.price;
	});
	sorters.set('fees', (a: PortfolioEvent, b: PortfolioEvent) => {
		return a.fees - b.fees;
	});
	sorters.set('taxes', (a: PortfolioEvent, b: PortfolioEvent) => {
		return a.taxes - b.taxes;
	});

	let sortBy = 'time';
	let asc = false;

	$: sorted = getPositions(transactions, sortBy, asc);

	function getPositions(
		transactions: PortfolioEvent[],
		sortBy: string,
		asc: boolean
	): PortfolioEvent[] {
		return transactions.sort((a: PortfolioEvent, b: PortfolioEvent) => {
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

	function securityFor(tx: PortfolioEvent) {
		return securities.find((sec) => sec.name == tx.securityName);
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
						active={sortBy == 'date'}
						column="date"
						on:changeDirection={toggleSortDirection}
						on:changeSortBy={(column) => changeSortBy('date')}
					>
						Date
					</TableSorter>
				</th>
				<th
					scope="col"
					class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-0"
				>
					<TableSorter
						active={sortBy == 'date'}
						column="date"
						on:changeDirection={toggleSortDirection}
						on:changeSortBy={(column) => changeSortBy('date')}
					>
						Type
					</TableSorter>
				</th>
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
					>
						Price
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
						on:changeSortBy={(column) => changeSortBy('marketValue')}
					>
						Fees
					</TableSorter>
				</th>
				<th
					scope="col"
					class="hidden px-3 py-3.5 text-right text-sm font-semibold text-gray-900 sm:table-cell"
				>
					<TableSorter
						active={sortBy == 'profit'}
						column="profit"
						on:changeDirection={toggleSortDirection}
						on:changeSortBy={(column) => changeSortBy('profit')}
					>
						Taxes
					</TableSorter>
				</th>
				<th
					scope="col"
					class="hidden px-3 py-3.5 text-right text-sm font-semibold text-gray-900 sm:table-cell"
				>
					<TableSorter
						active={sortBy == 'total'}
						column="total"
						on:changeDirection={toggleSortDirection}
						on:changeSortBy={(column) => changeSortBy('total')}
					>
						Total
					</TableSorter>
				</th>
			</tr>
		</thead>
		<tbody class="divide-y divide-gray-200">
			{#each sorted as tx (tx.name)}
				<PortfolioTransactionRow {tx} security={securityFor(tx)}/>
			{/each}
		</tbody>
	</table>
</div>
