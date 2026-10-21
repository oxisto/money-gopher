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
		// committed transactions from before — confirmImport reconciles each
		// staged row against its already-committed counterpart from the same
		// document (matched by type/time/units/cashDelta) instead of
		// duplicating it, so this is safe to auto-confirm too.

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
		UPLOADED: 'bg-gold/15 text-gold',
		PARSED: 'bg-accent-soft text-accent-fg',
		FAILED: 'bg-neg-soft text-neg',
		SKIPPED: 'bg-hover text-muted',
		IMPORTED: 'bg-pos-soft text-pos'
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
	<p class="text-sm text-muted">No documents yet. Upload a bank statement or CSV to get started.</p>
{:else}
	<label class="mb-3 flex w-fit items-center gap-2 text-sm text-muted">
		<input type="checkbox" bind:checked={hideImported} class="rounded border-line-strong" />
		Hide imported
	</label>

	{#if visibleDocuments.length === 0}
		<p class="text-sm text-muted">
			All documents are imported.
			<button class="text-accent-fg hover:underline" onclick={() => (hideImported = false)}>
				Show them
			</button>
		</p>
	{/if}

	<div class="flex flex-col gap-4">
		{#each visibleDocuments as doc (doc.id)}
			<div class="rounded-md border border-line bg-surface p-4">
				<div class="flex items-start justify-between gap-4">
					<div class="flex-1 min-w-0">
						<div class="flex items-center gap-2 flex-wrap">
							<span class="rounded-full px-2 py-0.5 text-xs font-medium {stateBadge[doc.state] ?? 'bg-hover text-muted'}">
								{stateLabel[doc.state] ?? doc.state}
							</span>
							{#if doc.detectedBank}
								<span class="text-xs text-muted">{doc.detectedBank}</span>
							{/if}
						</div>

						{#if doc.stagedTransactions.length > 0}
							{@const tx = doc.stagedTransactions[0]}
							<div class="mt-1 flex flex-wrap items-baseline gap-x-3 gap-y-0.5">
								<span class="text-sm font-medium text-fg">{formatDate(tx.time)}</span>
								<span class="rounded bg-hover px-1.5 py-0.5 text-xs font-mono">{tx.type}</span>
								{#if tx.security}
									<span class="text-sm text-fg">{tx.security.displayName}</span>
								{:else if tx.securityHint}
									<span class="text-sm text-muted italic">{tx.securityHint}</span>
								{:else if tx.isin}
									<span class="text-sm text-muted font-mono">{tx.isin}</span>
								{/if}
								{#if tx.units > 0}
									<span class="text-sm text-muted tabular-nums">{tx.units.toLocaleString()} units</span>
								{/if}
								{#if tx.price}
									<span class="text-sm text-muted tabular-nums">@ {formatPrice(tx.price)}</span>
								{/if}
								{#if tx.taxes.amount !== 0}
									<span class="text-xs text-muted tabular-nums">tax {formatMoney(tx.taxes)}</span>
								{/if}
								{#if tx.cashDelta.amount !== 0}
									<span class="text-sm font-medium {tx.cashDelta.amount >= 0 ? 'text-pos' : 'text-neg'}">
										{formatMoney(tx.cashDelta)}
									</span>
								{/if}
								{#if tx.calculatedCashDelta}
									<span
										class="text-xs {tx.calculatedCashDelta.amount !== tx.cashDelta.amount
											? 'font-medium text-gold'
											: 'text-muted'}"
									>
										calc: {formatMoney(tx.calculatedCashDelta)}
									</span>
								{/if}
							</div>
						{:else if doc.state === 'FAILED' && doc.error}
							<p class="mt-1 text-xs text-neg">{doc.error}</p>
						{:else if doc.transactionDate}
							<p class="mt-1 text-sm text-muted">{formatDate(doc.transactionDate)}</p>
						{/if}

						<p class="mt-0.5 text-xs text-muted truncate">{doc.filename}</p>
					</div>
					<div class="flex items-center gap-2 flex-shrink-0">
						{#if doc.state === 'PARSED'}
							{@const match = autoMatch(doc)}
							{#if match}
								<button
									onclick={() => quickImport(doc, match)}
									disabled={quickImporting[doc.id]}
									class="rounded-md border border-pos/30 bg-pos-soft px-3 py-1.5 text-sm font-medium text-pos hover:bg-pos-soft disabled:opacity-50"
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
								class="rounded px-2 py-1.5 text-xs text-muted hover:text-fg hover:bg-hover disabled:opacity-50"
								title="Re-run extraction and parsing with the current importer code"
							>
								{reprocessing[doc.id] ? 'Re-analyzing…' : 'Re-analyze'}
							</button>
						{/if}
						<button
							onclick={() => del(doc.id)}
							class="rounded px-2 py-1.5 text-xs text-muted hover:text-neg hover:bg-neg-soft"
						>
							Delete
						</button>
					</div>
				</div>

				{#if doc.transactions.length > 0}
					<p class="mt-2 rounded bg-accent-soft px-2 py-1.5 text-xs text-accent-fg">
						Already has {doc.transactions.length} committed transaction{doc.transactions.length === 1
							? ''
							: 's'} from this document. Confirming the staged transactions below will update
						{doc.transactions.length === 1 ? 'it' : 'them'} in place (matched by type/time/units/amount)
						rather than creating a duplicate.
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
					<div class="mt-4 border-t border-line pt-4">
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
