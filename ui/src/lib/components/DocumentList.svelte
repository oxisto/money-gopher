<script lang="ts">
	import { graphql } from '$houdini';
	import { formatDate, formatMoney, formatPrice } from '$lib/format';
	import DocumentReview from './DocumentReview.svelte';
	import Button from './ui/Button.svelte';
	import ErrorNote from './ui/ErrorNote.svelte';

	interface Document {
		id: string;
		filename: string;
		state: string;
		detectedBank: string | null;
		error: string | null;
		transactionDate: Date | null;
		stagedTransactions: StagedTx[];
		suggestedCashAccount: { id: string; displayName: string; currency: string } | null;
		transactions: { id: string; type: string; cashDelta: { amount: number; currency: string } }[];
	}
	interface StagedTx {
		id: string;
		type: string;
		time: Date;
		units: number;
		price: { amount: string; currency: string } | null;
		fees: { amount: number; currency: string };
		taxes: { amount: number; currency: string };
		cashDelta: { amount: number; currency: string };
		calculatedCashDelta: { amount: number; currency: string } | null;
		fxRate: string | null;
		currency: string;
		securityHint: string | null;
		isin: string | null;
		security: { id: string; displayName: string } | null;
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
		documents,
		portfolios,
		cashAccounts,
		securities,
		onchanged
	}: {
		documents: Document[];
		portfolios: Portfolio[];
		cashAccounts: CashAccount[];
		securities: Security[];
		onchanged: () => void;
	} = $props();

	const deleteDocument = graphql(`
		mutation DeleteDocument($id: ID!) {
			deleteDocument(id: $id)
		}
	`);

	const confirmImport = graphql(`
		mutation ConfirmImportQuick($documentID: ID!, $input: ConfirmImportInput!) {
			confirmImport(documentID: $documentID, input: $input) {
				id
				type
			}
		}
	`);

	const reprocessDocument = graphql(`
		mutation ReprocessDocument($id: ID!) {
			reprocessDocument(id: $id) {
				id
			}
		}
	`);

	let reviewDocID = $state<string | null>(null);
	let deleteError = $state<Record<string, string>>({});
	let quickImporting = $state<Record<string, boolean>>({});
	let quickImportError = $state<Record<string, string>>({});
	let reprocessing = $state<Record<string, boolean>>({});
	let reprocessError = $state<Record<string, string>>({});
	let hideImported = $state(true);

	async function reanalyze(id: string) {
		reprocessing = { ...reprocessing, [id]: true };
		reprocessError = { ...reprocessError, [id]: '' };

		const result = await reprocessDocument.mutate({ id });

		reprocessing = { ...reprocessing, [id]: false };
		if (result.errors?.length) {
			reprocessError = { ...reprocessError, [id]: result.errors[0].message };
		} else {
			onchanged();
		}
	}

	const visibleDocuments = $derived(
		hideImported ? documents.filter((d) => d.state !== 'IMPORTED') : documents
	);

	async function del(id: string) {
		const result = await deleteDocument.mutate({ id });
		if (result.errors?.length) {
			deleteError = { ...deleteError, [id]: result.errors[0].message };
		} else {
			onchanged();
		}
	}

	function needsPortfolio(type: string): boolean {
		return ['BUY', 'SELL', 'DELIVERY_INBOUND', 'DELIVERY_OUTBOUND', 'DIVIDEND'].includes(type);
	}

	// A document can skip manual review only when every decision the review
	// form would otherwise ask for already has an unambiguous answer: the
	// cash account (IBAN match, or there's only one to begin with), the
	// portfolio (only one exists, since portfolios are never auto-matched),
	// and every transaction's security (already resolved from its ISIN).
	function autoMatch(doc: Document): { cashAccountID: string; portfolioID: string | null } | null {
		if (doc.stagedTransactions.length === 0) return null;
		// A reprocessed document can be PARSED again while still having
		// committed transactions from before — confirming these staged ones
		// too would duplicate them, not replace them.
		if (doc.transactions.length > 0) return null;

		const cashAccountID = doc.suggestedCashAccount?.id ?? (cashAccounts.length === 1 ? cashAccounts[0].id : null);
		if (!cashAccountID) return null;

		const portfolioRequired = doc.stagedTransactions.some((tx) => needsPortfolio(tx.type));
		const portfolioID = portfolioRequired ? (portfolios.length === 1 ? portfolios[0].id : null) : null;
		if (portfolioRequired && !portfolioID) return null;

		for (const tx of doc.stagedTransactions) {
			if (!tx.security && (tx.isin || tx.securityHint)) return null;
			// A mismatch between the parsed and calculated cash delta usually
			// means a line item on the document isn't being captured — skipping
			// review here would import a transaction nobody checked.
			if (tx.calculatedCashDelta && tx.calculatedCashDelta.amount !== tx.cashDelta.amount) return null;
		}

		return { cashAccountID, portfolioID };
	}

	async function quickImport(doc: Document, match: { cashAccountID: string; portfolioID: string | null }) {
		quickImporting = { ...quickImporting, [doc.id]: true };
		quickImportError = { ...quickImportError, [doc.id]: '' };

		const result = await confirmImport.mutate({
			documentID: doc.id,
			input: { portfolioID: match.portfolioID, cashAccountID: match.cashAccountID, transactions: [] }
		});

		quickImporting = { ...quickImporting, [doc.id]: false };
		if (result.errors?.length) {
			quickImportError = { ...quickImportError, [doc.id]: result.errors[0].message };
		} else {
			onchanged();
		}
	}

	const stateBadge: Record<string, string> = {
		UPLOADED: 'bg-yellow-100 text-yellow-800',
		PARSED: 'bg-blue-100 text-blue-800',
		FAILED: 'bg-red-100 text-red-800',
		SKIPPED: 'bg-gray-100 text-gray-500',
		IMPORTED: 'bg-green-100 text-green-800'
	};
	const stateLabel: Record<string, string> = {
		UPLOADED: 'Processing…',
		PARSED: 'Ready to review',
		FAILED: 'Failed',
		SKIPPED: 'Skipped',
		IMPORTED: 'Imported'
	};
</script>

{#if documents.length === 0}
	<p class="text-sm text-gray-400">No documents yet. Upload a bank statement or CSV to get started.</p>
{:else}
	<label class="mb-3 flex w-fit items-center gap-2 text-sm text-gray-600">
		<input type="checkbox" bind:checked={hideImported} class="rounded border-gray-300" />
		Hide imported
	</label>

	{#if visibleDocuments.length === 0}
		<p class="text-sm text-gray-400">
			All documents are imported.
			<button class="text-blue-600 hover:underline" onclick={() => (hideImported = false)}>
				Show them
			</button>
		</p>
	{/if}

	<div class="flex flex-col gap-4">
		{#each visibleDocuments as doc (doc.id)}
			<div class="rounded-md border border-gray-200 bg-white p-4">
				<div class="flex items-start justify-between gap-4">
					<div class="flex-1 min-w-0">
						<div class="flex items-center gap-2 flex-wrap">
							<span class="rounded-full px-2 py-0.5 text-xs font-medium {stateBadge[doc.state] ?? 'bg-gray-100 text-gray-600'}">
								{stateLabel[doc.state] ?? doc.state}
							</span>
							{#if doc.detectedBank}
								<span class="text-xs text-gray-400">{doc.detectedBank}</span>
							{/if}
						</div>

						{#if doc.stagedTransactions.length > 0}
							{@const tx = doc.stagedTransactions[0]}
							<div class="mt-1 flex flex-wrap items-baseline gap-x-3 gap-y-0.5">
								<span class="text-sm font-medium text-gray-800">{formatDate(tx.time)}</span>
								<span class="rounded bg-gray-100 px-1.5 py-0.5 text-xs font-mono">{tx.type}</span>
								{#if tx.security}
									<span class="text-sm text-gray-700">{tx.security.displayName}</span>
								{:else if tx.securityHint}
									<span class="text-sm text-gray-500 italic">{tx.securityHint}</span>
								{:else if tx.isin}
									<span class="text-sm text-gray-500 font-mono">{tx.isin}</span>
								{/if}
								{#if tx.units > 0}
									<span class="text-sm text-gray-600 tabular-nums">{tx.units.toLocaleString()} units</span>
								{/if}
								{#if tx.price}
									<span class="text-sm text-gray-600 tabular-nums">@ {formatPrice(tx.price)}</span>
								{/if}
								{#if tx.taxes.amount !== 0}
									<span class="text-xs text-gray-500 tabular-nums">tax {formatMoney(tx.taxes)}</span>
								{/if}
								{#if tx.cashDelta.amount !== 0}
									<span class="text-sm font-medium {tx.cashDelta.amount >= 0 ? 'text-green-700' : 'text-red-700'}">
										{formatMoney(tx.cashDelta)}
									</span>
								{/if}
								{#if tx.calculatedCashDelta}
									<span
										class="text-xs {tx.calculatedCashDelta.amount !== tx.cashDelta.amount
											? 'font-medium text-amber-700'
											: 'text-gray-400'}"
									>
										calc: {formatMoney(tx.calculatedCashDelta)}
									</span>
								{/if}
							</div>
						{:else if doc.state === 'FAILED' && doc.error}
							<p class="mt-1 text-xs text-red-600">{doc.error}</p>
						{:else if doc.transactionDate}
							<p class="mt-1 text-sm text-gray-500">{formatDate(doc.transactionDate)}</p>
						{/if}

						<p class="mt-0.5 text-xs text-gray-400 truncate">{doc.filename}</p>
					</div>
					<div class="flex items-center gap-2 flex-shrink-0">
						{#if doc.state === 'PARSED'}
							{@const match = autoMatch(doc)}
							{#if match}
								<button
									onclick={() => quickImport(doc, match)}
									disabled={quickImporting[doc.id]}
									class="rounded-md border border-green-200 bg-green-50 px-3 py-1.5 text-sm font-medium text-green-700 hover:bg-green-100 disabled:opacity-50"
									title="Every field on this document was auto-matched — import without reviewing"
								>
									{quickImporting[doc.id] ? 'Importing…' : '✓ Quick import'}
								</button>
							{/if}
							<Button onclick={() => (reviewDocID = reviewDocID === doc.id ? null : doc.id)}>
								{reviewDocID === doc.id ? 'Close' : 'Review'}
							</Button>
						{/if}
						{#if doc.state !== 'UPLOADED'}
							<button
								onclick={() => reanalyze(doc.id)}
								disabled={reprocessing[doc.id]}
								class="rounded px-2 py-1.5 text-xs text-gray-400 hover:text-gray-700 hover:bg-gray-100 disabled:opacity-50"
								title="Re-run extraction and parsing with the current importer code"
							>
								{reprocessing[doc.id] ? 'Re-analyzing…' : 'Re-analyze'}
							</button>
						{/if}
						<button
							onclick={() => del(doc.id)}
							class="rounded px-2 py-1.5 text-xs text-gray-400 hover:text-red-500 hover:bg-red-50"
						>
							Delete
						</button>
					</div>
				</div>

				{#if doc.transactions.length > 0}
					<p class="mt-2 rounded bg-amber-50 px-2 py-1.5 text-xs text-amber-800">
						Already has {doc.transactions.length} committed transaction{doc.transactions.length === 1
							? ''
							: 's'} from this document. Confirming the staged transactions below would duplicate
						{doc.transactions.length === 1 ? 'it' : 'them'} — delete the existing one{doc.transactions
							.length === 1
							? ''
							: 's'} first if you want to replace {doc.transactions.length === 1 ? 'it' : 'them'}.
					</p>
				{/if}

				{#if deleteError[doc.id]}
					<ErrorNote message={deleteError[doc.id]} />
				{/if}
				{#if quickImportError[doc.id]}
					<ErrorNote message={quickImportError[doc.id]} />
				{/if}
				{#if reprocessError[doc.id]}
					<ErrorNote message={reprocessError[doc.id]} />
				{/if}

				{#if reviewDocID === doc.id}
					<div class="mt-4 border-t border-gray-100 pt-4">
						<DocumentReview
							document={doc}
							suggestedCashAccount={doc.suggestedCashAccount}
							{portfolios}
							{cashAccounts}
							{securities}
							onimported={() => {
								reviewDocID = null;
								onchanged();
							}}
						/>
					</div>
				{/if}
			</div>
		{/each}
	</div>
{/if}
