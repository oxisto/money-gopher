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
		suggestedCashAccount,
		portfolios,
		cashAccounts,
		securities,
		onimported
	}: {
		document: Document;
		suggestedCashAccount: CashAccount | null;
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

	const createSecurityMutation = graphql(`
		mutation CreateSecurityFromImport($input: CreateSecurityInput!) {
			createSecurity(input: $input) {
				id
				displayName
			}
		}
	`);

	// Pre-select the matched cash account if the server found one, otherwise fall back to first.
	let portfolioID = $state(portfolios[0]?.id ?? '');
	let cashAccountID = $state(suggestedCashAccount?.id ?? cashAccounts[0]?.id ?? '');
	let securityOverrides = $state<Record<string, string>>({});
	let localSecurities = $state<Security[]>([]);
	let createDialog = $state<{ txId: string; name: string; isin: string } | null>(null);
	let creatingDialog = $state(false);

	const allSecurities = $derived([...securities, ...localSecurities]);

	let confirming = $state(false);
	let error = $state<string | null>(null);
	let showPDF = $state(true);

	function needsPortfolio(type: string): boolean {
		return ['BUY', 'SELL', 'DELIVERY_INBOUND', 'DELIVERY_OUTBOUND', 'DIVIDEND'].includes(type);
	}

	const portfolioRequired = $derived(
		doc.stagedTransactions.some((tx) => needsPortfolio(tx.type))
	);

	const pdfURL = $derived(`/documents/${doc.id}`);

	function openCreateDialog(tx: StagedTx) {
		createDialog = {
			txId: tx.id,
			name: tx.securityHint ?? tx.isin ?? '',
			isin: tx.isin ?? ''
		};
	}

	async function confirmCreateSecurity() {
		if (!createDialog) return;
		creatingDialog = true;
		const result = await createSecurityMutation.mutate({
			input: {
				displayName: createDialog.name,
				identifiers: createDialog.isin ? [{ kind: 'ISIN', value: createDialog.isin }] : []
			}
		});
		creatingDialog = false;
		if (result.errors?.length) {
			error = result.errors[0].message;
			return;
		}
		const newSec = result.data?.createSecurity;
		if (newSec) {
			localSecurities = [...localSecurities, newSec];
			securityOverrides[createDialog.txId] = newSec.id;
		}
		createDialog = null;
	}

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

<!-- Split-pane: PDF on the left, review form on the right -->
<div class="flex flex-col gap-4">
	<!-- PDF toggle (visible on narrow screens; on wide screens pane is always shown) -->
	<div class="flex items-center justify-between lg:hidden">
		<span class="text-xs text-gray-500">Source document</span>
		<button
			onclick={() => (showPDF = !showPDF)}
			class="text-xs text-blue-600 hover:underline"
		>
			{showPDF ? 'Hide PDF' : 'Show PDF'}
		</button>
	</div>

	<div class="flex flex-col gap-6 lg:flex-row lg:gap-4 lg:h-[min(65vh,820px)]">
		<!-- PDF pane -->
		{#if showPDF}
			<div class="flex-1 min-w-0 lg:flex-none lg:w-[540px] lg:flex lg:flex-col">
				<div class="hidden lg:flex items-center justify-between mb-1">
					<span class="text-xs text-gray-500">Source document</span>
					<a href={pdfURL} target="_blank" class="text-xs text-blue-600 hover:underline">Open in new tab ↗</a>
				</div>
				<embed
					src={pdfURL}
					type="application/pdf"
					class="w-full flex-1 rounded-md border border-gray-200"
					style="height: min(65vh, 820px);"
				/>
				<a href={pdfURL} target="_blank" class="mt-1 block text-xs text-blue-600 hover:underline lg:hidden">Open in new tab ↗</a>
			</div>
		{/if}

		<!-- Review form -->
		<div class="flex-1 min-w-0 flex flex-col gap-4 lg:overflow-y-auto">
			<!-- Global account selectors -->
			<div class="flex flex-wrap gap-4">
				{#if portfolioRequired}
					<div class="flex-1 min-w-[180px]">
						<SelectField bind:value={portfolioID} label="Portfolio">
							<option value="">— select —</option>
							{#each portfolios as p (p.id)}
								<option value={p.id}>{p.displayName}</option>
							{/each}
						</SelectField>
					</div>
				{/if}
				<div class="flex-1 min-w-[180px]">
					<SelectField bind:value={cashAccountID} label="Cash account">
						<option value="">— select —</option>
						{#each cashAccounts as a (a.id)}
							<option value={a.id}>{a.displayName} ({a.currency})</option>
						{/each}
					</SelectField>
					{#if suggestedCashAccount && suggestedCashAccount.id === cashAccountID}
						<p class="mt-1 text-xs text-green-700">✓ Auto-matched from document IBAN</p>
					{:else if suggestedCashAccount}
						<p class="mt-1 text-xs text-amber-600">
							Suggested: {suggestedCashAccount.displayName} —
							<button class="underline" onclick={() => (cashAccountID = suggestedCashAccount!.id)}>
								use it
							</button>
						</p>
					{/if}
				</div>
			</div>

			<!-- Staged transactions table -->
			<div class="overflow-x-auto rounded-md border border-gray-200">
				<table class="min-w-full text-sm">
					<thead class="bg-gray-50 text-xs text-gray-500">
						<tr>
							<th class="px-2 py-1.5 text-left font-medium">Date</th>
							<th class="px-2 py-1.5 text-left font-medium">Type</th>
							<th class="px-2 py-1.5 text-left font-medium">Security</th>
							<th class="px-2 py-1.5 text-right font-medium">Units</th>
							<th class="px-2 py-1.5 text-right font-medium">Price</th>
							<th class="px-2 py-1.5 text-right font-medium">Fees</th>
							<th class="px-2 py-1.5 text-right font-medium">Cash Δ</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100 bg-white">
						{#each doc.stagedTransactions as tx (tx.id)}
							<tr>
								<td class="px-2 py-1.5 text-gray-600 whitespace-nowrap">{formatDate(tx.time)}</td>
								<td class="px-2 py-1.5">
									<span class="rounded bg-gray-100 px-1.5 py-0.5 text-xs font-mono">{tx.type}</span>
								</td>
								<td class="px-2 py-1.5 max-w-[220px]">
									{#if tx.security}
										<span class="text-gray-900">{tx.security.displayName}</span>
										{#if tx.isin}
											<span class="ml-1 text-xs text-gray-400">{tx.isin}</span>
										{/if}
									{:else if tx.securityHint || tx.isin}
										<div class="flex flex-col gap-1">
											{#if tx.securityHint}
												<span class="text-xs text-gray-400 italic truncate">"{tx.securityHint}"</span>
											{/if}
											<div class="flex items-center gap-1">
												<select
													bind:value={securityOverrides[tx.id]}
													class="min-w-0 flex-1 rounded border border-gray-200 px-1.5 py-0.5 text-xs"
												>
													<option value="">— unmatched —</option>
													{#each allSecurities as s (s.id)}
														<option value={s.id}>{s.displayName}</option>
													{/each}
												</select>
												{#if tx.isin && !securityOverrides[tx.id]}
													<button
														onclick={() => openCreateDialog(tx)}
														class="shrink-0 rounded border border-blue-200 bg-blue-50 px-1.5 py-0.5 text-xs text-blue-700 hover:bg-blue-100"
													>
														Create
													</button>
												{/if}
											</div>
										</div>
									{:else}
										<span class="text-xs text-gray-400">—</span>
									{/if}
								</td>
								<td class="px-2 py-1.5 text-right tabular-nums text-gray-700">
									{tx.units > 0 ? tx.units.toLocaleString() : '—'}
								</td>
								<td class="px-2 py-1.5 text-right tabular-nums text-gray-700">
									{tx.price ? formatMoney(tx.price) : '—'}
								</td>
								<td class="px-2 py-1.5 text-right tabular-nums text-gray-700">
									{tx.fees.amount !== 0 ? formatMoney(tx.fees) : '—'}
								</td>
								<td class="px-2 py-1.5 text-right tabular-nums {tx.cashDelta.amount >= 0 ? 'text-green-700' : 'text-red-700'}">
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
	</div>
</div>

{#if createDialog}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
		onclick={() => (createDialog = null)}
	>
		<div
			class="w-full max-w-sm rounded-lg bg-white p-5 shadow-xl"
			onclick={(e) => e.stopPropagation()}
		>
			<h3 class="mb-4 text-sm font-semibold text-gray-800">Add security</h3>
			<div class="flex flex-col gap-3">
				<div>
					<label class="mb-1 block text-xs text-gray-500">Name</label>
					<input
						bind:value={createDialog.name}
						class="w-full rounded border border-gray-200 px-3 py-1.5 text-sm focus:border-blue-400 focus:outline-none"
						placeholder="Security name"
					/>
				</div>
				{#if createDialog.isin}
					<div>
						<label class="mb-1 block text-xs text-gray-500">ISIN</label>
						<p class="font-mono text-sm text-gray-700">{createDialog.isin}</p>
					</div>
				{/if}
			</div>
			<div class="mt-5 flex items-center justify-end gap-2">
				<button
					onclick={() => (createDialog = null)}
					class="rounded px-3 py-1.5 text-sm text-gray-500 hover:bg-gray-50"
				>
					Cancel
				</button>
				<Button onclick={confirmCreateSecurity} disabled={creatingDialog || !createDialog.name}>
					{creatingDialog ? 'Creating…' : 'Create security'}
				</Button>
			</div>
		</div>
	</div>
{/if}
