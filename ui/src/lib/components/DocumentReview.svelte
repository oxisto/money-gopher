<script lang="ts">
	import { graphql } from '$houdini';
	import { formatMoney, formatDate } from '$lib/format';
	import Button from './ui/Button.svelte';
	import ErrorNote from './ui/ErrorNote.svelte';
	import SelectField from './ui/SelectField.svelte';

	interface StagedTx {
		id: string;
		type: string;
		time: Date;
		units: number;
		price: { amount: number; currency: string } | null;
		fees: { amount: number; currency: string };
		taxes: { amount: number; currency: string };
		cashDelta: { amount: number; currency: string };
		currency: string;
		securityHint: string | null;
		isin: string | null;
		security: { id: string; displayName: string } | null;
	}
	interface Document {
		id: string;
		stagedTransactions: StagedTx[];
	}
	interface Portfolio {
		id: string;
		displayName: string;
	}
	interface CashAccount {
		id: string;
		displayName: string;
		currency: string;
	}
	interface Security {
		id: string;
		displayName: string;
	}

	let {
		document: doc,
		portfolios,
		cashAccounts,
		securities,
		onimported
	}: {
		document: Document;
		portfolios: Portfolio[];
		cashAccounts: CashAccount[];
		securities: Security[];
		onimported: () => void;
	} = $props();

	const confirmImport = graphql(`
		mutation ConfirmImport($documentID: ID!, $input: ConfirmImportInput!) {
			confirmImport(documentID: $documentID, input: $input) {
				id
				type
			}
		}
	`);

	let portfolioID = $state(portfolios[0]?.id ?? '');
	let cashAccountID = $state(cashAccounts[0]?.id ?? '');
	// Per-transaction security overrides keyed by staged tx id.
	let securityOverrides = $state<Record<string, string>>({});

	let confirming = $state(false);
	let error = $state<string | null>(null);

	// Indicate whether the staged tx needs a portfolio (trade/delivery types).
	function needsPortfolio(type: string): boolean {
		return ['BUY', 'SELL', 'DELIVERY_INBOUND', 'DELIVERY_OUTBOUND', 'DIVIDEND'].includes(type);
	}

	// True when at least one staged tx needs a portfolio but none is selected.
	const portfolioRequired = $derived(
		doc.stagedTransactions.some((tx) => needsPortfolio(tx.type))
	);

	async function confirm() {
		if (portfolioRequired && !portfolioID) {
			error = 'Please select a portfolio.';
			return;
		}
		if (!cashAccountID) {
			error = 'Please select a cash account.';
			return;
		}

		confirming = true;
		error = null;

		const txOverrides = Object.entries(securityOverrides)
			.filter(([, secID]) => secID)
			.map(([id, secID]) => ({ id, securityID: secID }));

		const result = await confirmImport.mutate({
			documentID: doc.id,
			input: {
				portfolioID: portfolioRequired ? portfolioID : null,
				cashAccountID,
				transactions: txOverrides
			}
		});

		confirming = false;
		if (result.errors?.length) {
			error = result.errors[0].message;
		} else {
			onimported();
		}
	}
</script>

<div class="flex flex-col gap-4">
	<!-- Global defaults -->
	<div class="grid gap-4 sm:grid-cols-2">
		{#if portfolioRequired}
			<SelectField bind:value={portfolioID} label="Portfolio (applies to trades & deliveries)">
				<option value="">— select —</option>
				{#each portfolios as p (p.id)}
					<option value={p.id}>{p.displayName}</option>
				{/each}
			</SelectField>
		{/if}
		<SelectField bind:value={cashAccountID} label="Cash account">
			<option value="">— select —</option>
			{#each cashAccounts as a (a.id)}
				<option value={a.id}>{a.displayName} ({a.currency})</option>
			{/each}
		</SelectField>
	</div>

	<!-- Staged transactions table -->
	<div class="overflow-x-auto rounded-md border border-gray-200">
		<table class="min-w-full text-sm">
			<thead class="bg-gray-50 text-xs text-gray-500">
				<tr>
					<th class="px-3 py-2 text-left font-medium">Date</th>
					<th class="px-3 py-2 text-left font-medium">Type</th>
					<th class="px-3 py-2 text-left font-medium">Security</th>
					<th class="px-3 py-2 text-right font-medium">Units</th>
					<th class="px-3 py-2 text-right font-medium">Price</th>
					<th class="px-3 py-2 text-right font-medium">Fees</th>
					<th class="px-3 py-2 text-right font-medium">Cash Δ</th>
				</tr>
			</thead>
			<tbody class="divide-y divide-gray-100 bg-white">
				{#each doc.stagedTransactions as tx (tx.id)}
					<tr>
						<td class="px-3 py-2 text-gray-600 whitespace-nowrap">{formatDate(tx.time)}</td>
						<td class="px-3 py-2">
							<span class="rounded bg-gray-100 px-1.5 py-0.5 text-xs font-mono">{tx.type}</span>
						</td>
						<td class="px-3 py-2">
							{#if tx.security}
								<span class="text-gray-900">{tx.security.displayName}</span>
								{#if tx.isin}
									<span class="ml-1 text-xs text-gray-400">{tx.isin}</span>
								{/if}
							{:else if tx.securityHint || tx.isin}
								<!-- No auto-match: offer a dropdown to pick manually -->
								<div class="flex flex-col gap-1">
									{#if tx.securityHint}
										<span class="text-xs text-gray-400 italic">"{tx.securityHint}"</span>
									{/if}
									<select
										bind:value={securityOverrides[tx.id]}
										class="rounded border border-gray-200 px-2 py-1 text-xs"
									>
										<option value="">— unmatched —</option>
										{#each securities as s (s.id)}
											<option value={s.id}>{s.displayName}</option>
										{/each}
									</select>
								</div>
							{:else}
								<span class="text-xs text-gray-400">—</span>
							{/if}
						</td>
						<td class="px-3 py-2 text-right tabular-nums text-gray-700">
							{tx.units > 0 ? tx.units.toLocaleString() : '—'}
						</td>
						<td class="px-3 py-2 text-right tabular-nums text-gray-700">
							{tx.price ? formatMoney(tx.price) : '—'}
						</td>
						<td class="px-3 py-2 text-right tabular-nums text-gray-700">
							{formatMoney(tx.fees)}
						</td>
						<td class="px-3 py-2 text-right tabular-nums {tx.cashDelta.amount >= 0 ? 'text-green-700' : 'text-red-700'}">
							{tx.cashDelta.amount !== 0 ? formatMoney(tx.cashDelta) : '—'}
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>

	<div class="flex items-center gap-4">
		<Button onclick={confirm} disabled={confirming}>
			{confirming ? 'Importing…' : `Import ${doc.stagedTransactions.length} transaction${doc.stagedTransactions.length !== 1 ? 's' : ''}`}
		</Button>
		<ErrorNote message={error} />
	</div>
</div>
