<script lang="ts">
	import { graphql } from '$houdini';
	import { formatDate, formatMoney, formatPrice } from '$lib/format';
	import type { TransactionRow, SecurityOption, CashAccountOption } from '$lib/transaction-types';
	import TransactionEditForm from './TransactionEditForm.svelte';

	let {
		transactions,
		securities,
		cashAccounts,
		showBalance = false,
		onchanged
	}: {
		transactions: TransactionRow[];
		securities: SecurityOption[];
		cashAccounts: CashAccountOption[];
		// Only meaningful when transactions is a single cash account's full
		// history (as on the cash account detail page): the running sum
		// then ends exactly at its current balance. A portfolio's
		// transaction list may not include every movement on its
		// settlement account (another portfolio can share it, or it can
		// have deposits/withdrawals of its own), so the same running sum
		// there wouldn't be accurate — callers opt in explicitly instead
		// of this component guessing which case it's in.
		showBalance?: boolean;
		onchanged: () => void;
	} = $props();

	// transactions is already ordered oldest-first (the query it comes from
	// sorts by time), so a running total in that order is a running balance.
	const balances = $derived.by(() => {
		let running = 0;
		return transactions.map((tx) => (running += tx.cashDelta.amount));
	});

	const columnCount = $derived(showBalance ? 8 : 7);

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
					{#if showBalance}
						<th class="px-4 py-2 text-right font-medium">Balance</th>
					{/if}
					<th class="px-4 py-2"></th>
				</tr>
			</thead>
			<tbody class="divide-y divide-gray-100">
				{#each transactions as tx, i (tx.id)}
					<tr class={editingID === tx.id ? 'bg-gray-50' : ''}>
						<td class="px-4 py-2 whitespace-nowrap">{formatDate(tx.time)}</td>
						<td class="px-4 py-2">
							<span class="rounded bg-gray-100 px-1.5 py-0.5 text-xs font-mono">{tx.type}</span>
						</td>
						<td class="px-4 py-2">{tx.security?.displayName ?? '—'}</td>
						<td class="px-4 py-2 text-right tabular-nums">{tx.units || '—'}</td>
						<td class="px-4 py-2 text-right tabular-nums">
							{tx.price ? formatPrice(tx.price) : '—'}
							{#if tx.fxRate}
								<div
									class="text-xs text-gray-400"
									title={`1 ${tx.cashDelta.currency} = ${tx.fxRate} ${tx.price?.currency}`}
								>
									FX {tx.fxRate}
								</div>
							{/if}
						</td>
						<td class="px-4 py-2 text-right tabular-nums {tx.cashDelta.amount < 0 ? 'text-red-600' : 'text-green-700'}">
							{formatMoney(tx.cashDelta)}
						</td>
						{#if showBalance}
							<td class="px-4 py-2 text-right tabular-nums text-gray-600">
								{formatMoney({ amount: balances[i], currency: tx.cashDelta.currency })}
							</td>
						{/if}
						<td class="px-4 py-2 text-right whitespace-nowrap">
							{#if tx.document}
								<a
									href={`/documents/${tx.document.id}`}
									target="_blank"
									rel="noopener"
									title={`Source document: ${tx.document.filename}`}
									class="rounded px-2 py-1 text-xs text-gray-400 hover:text-gray-700 hover:bg-gray-100"
								>
									📄
								</a>
							{/if}
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
							<td colspan={columnCount} class="px-4 pb-2 text-xs text-red-600">{deleteErrors[tx.id]}</td>
						</tr>
					{/if}
					{#if editingID === tx.id}
						<tr>
							<td colspan={columnCount} class="px-4 py-3 bg-gray-50 border-b border-gray-200">
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
