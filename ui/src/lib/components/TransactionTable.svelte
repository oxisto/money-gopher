<script lang="ts">
	import { graphql } from '$houdini';
	import type { PortfolioDetail$result } from '$houdini';
	import { formatDate, formatMoney } from '$lib/format';
	import TransactionEditForm from './TransactionEditForm.svelte';

	type Transactions = NonNullable<PortfolioDetail$result['portfolio']>['transactions'];

	let {
		transactions,
		securities,
		cashAccounts,
		onchanged
	}: {
		transactions: Transactions;
		securities: PortfolioDetail$result['securities'];
		cashAccounts: PortfolioDetail$result['cashAccounts'];
		onchanged: () => void;
	} = $props();

	const deleteTransaction = graphql(`
		mutation DeleteTransaction($id: ID!) {
			deleteTransaction(id: $id)
		}
	`);

	let editingID = $state<string | null>(null);
	let deleteErrors = $state<Record<string, string>>({});

	async function del(id: string) {
		const result = await deleteTransaction.mutate({ id });
		if (result.errors?.length) {
			deleteErrors = { ...deleteErrors, [id]: result.errors[0].message };
		} else {
			onchanged();
		}
	}
</script>

{#if transactions.length === 0}
	<p class="text-sm text-gray-500">No transactions yet.</p>
{:else}
	<div class="overflow-x-auto rounded-md border border-gray-200 bg-white">
		<table class="w-full text-sm">
			<thead>
				<tr class="border-b border-gray-200 text-left text-xs text-gray-500 uppercase">
					<th class="px-4 py-2 font-medium">Date</th>
					<th class="px-4 py-2 font-medium">Type</th>
					<th class="px-4 py-2 font-medium">Security</th>
					<th class="px-4 py-2 text-right font-medium">Units</th>
					<th class="px-4 py-2 text-right font-medium">Price</th>
					<th class="px-4 py-2 text-right font-medium">Cash</th>
					<th class="px-4 py-2"></th>
				</tr>
			</thead>
			<tbody class="divide-y divide-gray-100">
				{#each transactions as tx (tx.id)}
					<tr class={editingID === tx.id ? 'bg-gray-50' : ''}>
						<td class="px-4 py-2 whitespace-nowrap">{formatDate(tx.time)}</td>
						<td class="px-4 py-2">
							<span class="rounded bg-gray-100 px-1.5 py-0.5 text-xs font-mono">{tx.type}</span>
						</td>
						<td class="px-4 py-2">{tx.security?.displayName ?? '—'}</td>
						<td class="px-4 py-2 text-right tabular-nums">{tx.units || '—'}</td>
						<td class="px-4 py-2 text-right tabular-nums">
							{tx.price ? formatMoney(tx.price) : '—'}
						</td>
						<td class="px-4 py-2 text-right tabular-nums {tx.cashDelta.amount < 0 ? 'text-red-600' : 'text-green-700'}">
							{formatMoney(tx.cashDelta)}
						</td>
						<td class="px-4 py-2 text-right whitespace-nowrap">
							<button
								onclick={() => (editingID = editingID === tx.id ? null : tx.id)}
								class="rounded px-2 py-1 text-xs text-gray-400 hover:text-gray-700 hover:bg-gray-100"
							>
								{editingID === tx.id ? 'Cancel' : 'Edit'}
							</button>
							<button
								onclick={() => del(tx.id)}
								class="rounded px-2 py-1 text-xs text-gray-400 hover:text-red-600 hover:bg-red-50"
							>
								Delete
							</button>
						</td>
					</tr>
					{#if deleteErrors[tx.id]}
						<tr>
							<td colspan="7" class="px-4 pb-2 text-xs text-red-600">{deleteErrors[tx.id]}</td>
						</tr>
					{/if}
					{#if editingID === tx.id}
						<tr>
							<td colspan="7" class="px-4 py-3 bg-gray-50 border-b border-gray-200">
								<TransactionEditForm
									transaction={tx}
									{securities}
									{cashAccounts}
									onsaved={() => { editingID = null; onchanged(); }}
									oncancel={() => (editingID = null)}
								/>
							</td>
						</tr>
					{/if}
				{/each}
			</tbody>
		</table>
	</div>
{/if}
