<script lang="ts">
	import { onMount } from 'svelte';
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

	// transactions is already ordered newest-first (the query it comes from
	// sorts by time descending). The account's current balance is the sum of
	// every delta, i.e. exactly the balance right after the newest (first)
	// row; walking down the list and subtracting each row's own delta lands
	// on the balance right after each older row in turn. Computed over the
	// full, unfiltered list — a filtered-out row must not shift the balance
	// shown on the rows around it.
	const rows = $derived.by(() => {
		let running = transactions.reduce((sum, tx) => sum + tx.cashDelta.amount, 0);
		return transactions.map((tx) => {
			const balance = running;
			running -= tx.cashDelta.amount;
			return { tx, balance };
		});
	});

	let typeFilter = $state('all');
	let securityFilter = $state('all');

	const types = $derived(Array.from(new Set(transactions.map((tx) => tx.type))).sort());

	const securityOptions = $derived.by(() => {
		const byID = new Map<string, string>();
		for (const tx of transactions) {
			if (tx.security) byID.set(tx.security.id, tx.security.displayName);
		}
		return Array.from(byID, ([id, displayName]) => ({ id, displayName })).sort((a, b) =>
			a.displayName.localeCompare(b.displayName)
		);
	});

	const filteredRows = $derived(
		rows.filter(
			({ tx }) =>
				(typeFilter === 'all' || tx.type === typeFilter) &&
				(securityFilter === 'all' || tx.security?.id === securityFilter)
		)
	);

	const columnCount = $derived(showBalance ? 8 : 7);

	const deleteTransaction = graphql(`
		mutation DeleteTransaction($id: ID!) {
			deleteTransaction(id: $id)
		}
	`);

	let editingID = $state<string | null>(null);
	let deleteErrors = $state<Record<string, string>>({});

	// A link from elsewhere (e.g. the Realized Gains page) can point at one
	// specific row via #tx-<id>; scroll to it and flash it briefly so it's
	// findable in an otherwise-unremarkable table.
	let highlightedID = $state<string | null>(null);
	onMount(() => {
		const id = window.location.hash.replace(/^#tx-/, '');
		if (!id || !transactions.some((tx) => tx.id === id)) return;

		highlightedID = id;
		document.getElementById(`tx-${id}`)?.scrollIntoView({ behavior: 'smooth', block: 'center' });
		setTimeout(() => (highlightedID = null), 2500);
	});

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
	<p class="text-sm text-muted">No transactions yet.</p>
{:else}
	<div class="mb-3 flex flex-wrap items-center gap-2">
		<select
			bind:value={typeFilter}
			class="rounded-md border border-line-strong bg-surface px-2 py-1 text-xs text-fg"
		>
			<option value="all">All types</option>
			{#each types as t (t)}
				<option value={t}>{t}</option>
			{/each}
		</select>
		{#if securityOptions.length > 0}
			<select
				bind:value={securityFilter}
				class="rounded-md border border-line-strong bg-surface px-2 py-1 text-xs text-fg"
			>
				<option value="all">All securities</option>
				{#each securityOptions as s (s.id)}
					<option value={s.id}>{s.displayName}</option>
				{/each}
			</select>
		{/if}
		{#if typeFilter !== 'all' || securityFilter !== 'all'}
			<button
				onclick={() => {
					typeFilter = 'all';
					securityFilter = 'all';
				}}
				class="text-xs text-muted hover:text-fg"
			>
				Clear filters
			</button>
			<span class="text-xs text-muted">{filteredRows.length} of {transactions.length}</span>
		{/if}
	</div>

	{#if filteredRows.length === 0}
		<p class="text-sm text-muted">No transactions match the current filter.</p>
	{:else}
		<div class="overflow-x-auto rounded-md border border-line bg-surface">
			<table class="w-full text-sm">
				<thead>
					<tr class="border-b border-line text-left text-xs text-muted uppercase">
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
				<tbody class="divide-y divide-line">
					{#each filteredRows as { tx, balance } (tx.id)}
					<tr
						id={`tx-${tx.id}`}
						class={editingID === tx.id
							? 'bg-raised'
							: highlightedID === tx.id
								? 'bg-gold/10 transition-colors duration-1000'
								: ''}
					>
						<td class="px-4 py-2 whitespace-nowrap">{formatDate(tx.time)}</td>
						<td class="px-4 py-2">
							<span class="rounded bg-hover px-1.5 py-0.5 text-xs font-mono">{tx.type}</span>
						</td>
						<td class="px-4 py-2">{tx.security?.displayName ?? '—'}</td>
						<td class="px-4 py-2 text-right tabular-nums">{tx.units || '—'}</td>
						<td class="px-4 py-2 text-right tabular-nums">
							{tx.price ? formatPrice(tx.price) : '—'}
							{#if tx.fxRate}
								<div
									class="text-xs text-muted"
									title={`1 ${tx.cashDelta.currency} = ${tx.fxRate} ${tx.price?.currency}`}
								>
									FX {tx.fxRate}
								</div>
							{/if}
						</td>
						<td class="px-4 py-2 text-right tabular-nums {tx.cashDelta.amount < 0 ? 'text-neg' : 'text-pos'}">
							{formatMoney(tx.cashDelta)}
						</td>
						{#if showBalance}
							<td class="px-4 py-2 text-right tabular-nums text-muted">
								{formatMoney({ amount: balance, currency: tx.cashDelta.currency })}
							</td>
						{/if}
						<td class="px-4 py-2 text-right whitespace-nowrap">
							<a
								href={`/transactions/${tx.id}`}
								title="View transaction details"
								class="rounded px-2 py-1 text-xs text-muted hover:text-fg hover:bg-hover"
							>
								Details
							</a>
							<button
								onclick={() => (editingID = editingID === tx.id ? null : tx.id)}
								class="rounded px-2 py-1 text-xs text-muted hover:text-fg hover:bg-hover"
							>
								{editingID === tx.id ? 'Cancel' : 'Edit'}
							</button>
							<button
								onclick={() => del(tx.id)}
								class="rounded px-2 py-1 text-xs text-muted hover:text-neg hover:bg-neg-soft"
							>
								Delete
							</button>
						</td>
					</tr>
					{#if deleteErrors[tx.id]}
						<tr>
							<td colspan={columnCount} class="px-4 pb-2 text-xs text-neg">{deleteErrors[tx.id]}</td>
						</tr>
					{/if}
					{#if editingID === tx.id}
						<tr>
							<td colspan={columnCount} class="px-4 py-3 bg-raised border-b border-line">
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
{/if}
