<script lang="ts">
	import { graphql } from '$houdini';
	import { formatDate, formatMoney } from '$lib/format';
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
	}
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

	let reviewDocID = $state<string | null>(null);
	let deleteError = $state<Record<string, string>>({});

	async function del(id: string) {
		const result = await deleteDocument.mutate({ id });
		if (result.errors?.length) {
			deleteError = { ...deleteError, [id]: result.errors[0].message };
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
	<div class="flex flex-col gap-4">
		{#each documents as doc (doc.id)}
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
								{#if tx.cashDelta.amount !== 0}
									<span class="text-sm font-medium {tx.cashDelta.amount >= 0 ? 'text-green-700' : 'text-red-700'}">
										{formatMoney(tx.cashDelta)}
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
							<Button onclick={() => (reviewDocID = reviewDocID === doc.id ? null : doc.id)}>
								{reviewDocID === doc.id ? 'Close' : 'Review'}
							</Button>
						{/if}
						<button
							onclick={() => del(doc.id)}
							class="rounded px-2 py-1.5 text-xs text-gray-400 hover:text-red-500 hover:bg-red-50"
						>
							Delete
						</button>
					</div>
				</div>

				{#if deleteError[doc.id]}
					<ErrorNote message={deleteError[doc.id]} />
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
