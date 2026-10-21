<script lang="ts">
	import { graphql, mutate } from '#lib/graphql/client.js';
	import DocumentReview from '#lib/components/DocumentReview.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import ErrorNote from '#lib/components/ui/ErrorNote.svelte';
	import Icon from '#lib/components/ui/Icon.svelte';
	import { describeTransaction } from '#lib/describe.js';
	import { formatDate, formatSignedMoney } from '#lib/format.js';

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

	let {
		documents,
		portfolios,
		cashAccounts,
		securities,
		onchanged
	}: {
		documents: Document[];
		portfolios: { id: string; displayName: string }[];
		cashAccounts: { id: string; displayName: string; currency: string }[];
		securities: { id: string; displayName: string }[];
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

	type Tab = 'review' | 'problems' | 'skipped' | 'imported';
	const byTab = $derived({
		review: documents.filter((d) => d.state === 'PARSED' || d.state === 'UPLOADED'),
		problems: documents.filter((d) => d.state === 'FAILED'),
		skipped: documents.filter((d) => d.state === 'SKIPPED'),
		imported: documents.filter((d) => d.state === 'IMPORTED')
	});
	// Open on whatever needs attention; otherwise on what already happened.
	let tab = $state<Tab>('review');
	let tabChosen = false;
	$effect(() => {
		if (tabChosen) return;
		tab = byTab.review.length ? 'review' : byTab.problems.length ? 'problems' : 'imported';
	});

	const tabs: { key: Tab; label: string; loud?: boolean }[] = [
		{ key: 'review', label: 'Needs review', loud: true },
		{ key: 'problems', label: 'Problems', loud: true },
		{ key: 'skipped', label: 'Skipped' },
		{ key: 'imported', label: 'Imported' }
	];

	let busy = $state<Record<string, string>>({});
	let errors = $state<Record<string, string>>({});
	let reviewing = $state<string | null>(null);
	let confirmingDelete = $state<string | null>(null);

	async function run(id: string, label: string, action: () => Promise<{ error: string | null }>) {
		busy = { ...busy, [id]: label };
		errors = { ...errors, [id]: '' };
		const result = await action();
		busy = { ...busy, [id]: '' };
		if (result.error) {
			errors = { ...errors, [id]: result.error };
			return false;
		}
		return true;
	}

	async function reanalyze(id: string) {
		if (await run(id, 'Re-analyzing…', () => mutate(reprocessDocument, { id }))) onchanged();
	}

	async function del(id: string) {
		confirmingDelete = null;
		if (await run(id, 'Deleting…', () => mutate(deleteDocument, { id }))) onchanged();
	}

	function needsPortfolio(type: string): boolean {
		return ['BUY', 'SELL', 'DELIVERY_INBOUND', 'DELIVERY_OUTBOUND', 'DIVIDEND'].includes(type);
	}

	// A document can skip manual review only when every decision the review
	// form would otherwise ask for already has an unambiguous answer: the
	// cash account (IBAN match, or there's only one to begin with), the
	// portfolio (only one exists, since portfolios are never auto-matched),
	// every transaction's security (already resolved from its ISIN), and
	// the parsed cash amount agrees with the one recomputed from its parts.
	// A reprocessed document can be PARSED again while still having
	// committed transactions; confirmImport reconciles those in place, so
	// that is safe to auto-confirm too.
	function autoMatch(doc: Document): { cashAccountID: string; portfolioID: string | null } | null {
		if (doc.state !== 'PARSED' || doc.stagedTransactions.length === 0) return null;

		const cashAccountID = doc.suggestedCashAccount?.id ?? (cashAccounts.length === 1 ? cashAccounts[0].id : null);
		if (!cashAccountID) return null;

		const portfolioRequired = doc.stagedTransactions.some((tx) => needsPortfolio(tx.type));
		const portfolioID = portfolioRequired ? (portfolios.length === 1 ? portfolios[0].id : null) : null;
		if (portfolioRequired && !portfolioID) return null;

		for (const tx of doc.stagedTransactions) {
			if (!tx.security && (tx.isin || tx.securityHint)) return null;
			if (tx.calculatedCashDelta && tx.calculatedCashDelta.amount !== tx.cashDelta.amount) return null;
		}
		return { cashAccountID, portfolioID };
	}

	function quickImport(doc: Document, match: NonNullable<ReturnType<typeof autoMatch>>) {
		return run(doc.id, 'Importing…', () =>
			mutate(confirmImport, {
				documentID: doc.id,
				input: { portfolioID: match.portfolioID, cashAccountID: match.cashAccountID, transactions: [] }
			})
		);
	}

	const ready = $derived(byTab.review.filter((d) => autoMatch(d)));
	let importingAll = $state(false);
	async function importAll() {
		importingAll = true;
		// One at a time, like uploads: each confirm touches shared balances.
		for (const doc of ready) {
			const match = autoMatch(doc);
			if (match) await quickImport(doc, match);
		}
		importingAll = false;
		onchanged();
	}

	function summary(doc: Document) {
		const staged = doc.stagedTransactions;
		const first = staged[0];
		const mismatches = staged.filter(
			(s) => s.calculatedCashDelta && s.calculatedCashDelta.amount !== s.cashDelta.amount
		).length;
		const unmatched = staged.filter((s) => !s.security && (s.isin || s.securityHint)).length;
		const amount = staged.length
			? staged.reduce((s, t) => s + t.cashDelta.amount, 0)
			: doc.transactions.reduce((s, t) => s + t.cashDelta.amount, 0);
		const d = first
			? describeTransaction({
					...first,
					security: first.security ?? (first.securityHint ? { displayName: first.securityHint } : null)
				})
			: null;
		return {
			title: staged.length > 1 ? `${staged.length} transactions` : (d?.title ?? doc.filename),
			icon: staged.length > 1 ? ('document' as const) : (d?.icon ?? ('document' as const)),
			meta: [
				doc.transactionDate ? formatDate(doc.transactionDate) : null,
				doc.suggestedCashAccount?.displayName,
				doc.detectedBank?.toUpperCase()
			]
				.filter(Boolean)
				.join(' · '),
			amount: { amount, currency: first?.cashDelta.currency ?? doc.transactions[0]?.cashDelta.currency ?? 'EUR' },
			mismatches,
			unmatched
		};
	}

	// Skip reasons come back as "<reason>: skipped".
	const reasons: Record<string, { title: string; why: string }> = {
		vorabpauschale: {
			title: 'Vorabpauschale',
			why: 'A tax pre-assessment on funds. No money moves, so there is nothing to book.'
		},
		storno: {
			title: 'Storno',
			why: 'Cancels an earlier statement; the corrected statement is imported instead.'
		}
	};
	const skippedGroups = $derived.by(() => {
		const groups = new Map<string, Document[]>();
		for (const doc of byTab.skipped) {
			const key = (doc.error ?? 'other').split(':')[0].trim().toLowerCase();
			groups.set(key, [...(groups.get(key) ?? []), doc]);
		}
		return [...groups].map(([key, docs]) => ({
			key,
			...(reasons[key] ?? { title: key, why: docs[0].error ?? '' }),
			docs
		}));
	});
	let openGroup = $state<string | null>(null);
</script>

{#snippet secondaryActions(doc: Document)}
	{#if doc.state !== 'UPLOADED'}
		<button
			type="button"
			onclick={() => reanalyze(doc.id)}
			disabled={!!busy[doc.id]}
			title="Re-run extraction and parsing with the current importer code"
			class="rounded-md px-2 py-1 text-xs text-muted hover:bg-hover hover:text-fg disabled:opacity-50"
		>
			Re-analyze
		</button>
	{/if}
	{#if confirmingDelete === doc.id}
		<button type="button" onclick={() => del(doc.id)} class="rounded-md bg-neg-soft px-2 py-1 text-xs font-medium text-neg">
			Really delete
		</button>
	{:else}
		<button
			type="button"
			onclick={() => (confirmingDelete = doc.id)}
			class="rounded-md px-2 py-1 text-xs text-muted hover:bg-neg-soft hover:text-neg"
		>
			Delete
		</button>
	{/if}
{/snippet}

<div class="flex flex-col gap-4">
	<div role="tablist" aria-label="Documents" class="flex gap-5 overflow-x-auto border-b border-line whitespace-nowrap">
		{#each tabs as t (t.key)}
			{@const count = byTab[t.key].length}
			<button
				type="button"
				role="tab"
				aria-selected={tab === t.key}
				onclick={() => {
					tab = t.key;
					tabChosen = true;
				}}
				class="-mb-px flex items-center gap-1.5 border-b-2 py-2 text-[13px] font-medium
					{tab === t.key ? 'border-accent text-fg' : 'border-transparent text-muted hover:text-fg'}"
			>
				{t.label}
				<span
					class="rounded-full px-1.5 text-[11px] font-semibold
						{t.loud && count ? 'bg-accent-soft text-accent-fg' : 'text-muted'}">{count}</span
				>
			</button>
		{/each}
	</div>

	{#if tab === 'review'}
		{#if byTab.review.length === 0}
			<p class="rounded-xl border border-dashed border-line px-4 py-8 text-center text-sm text-muted">
				All caught up. Drop new statements above.
			</p>
		{:else}
			{#if ready.length}
				<div class="flex flex-wrap items-center gap-3 rounded-lg bg-pos-soft px-4 py-2.5 text-[13px] text-pos">
					<Icon name="check" />
					<span class="flex-1">
						{ready.length === 1 ? '1 document matches' : `${ready.length} documents match`} exactly: account, portfolio,
						securities and amounts. Ready to book without review.
					</span>
					<Button onclick={importAll} disabled={importingAll}>
						{importingAll ? 'Importing…' : `Import ${ready.length}`}
					</Button>
				</div>
			{/if}

			<ul class="flex flex-col gap-2.5">
				{#each byTab.review as doc (doc.id)}
					{@const s = summary(doc)}
					{@const match = autoMatch(doc)}
					<li class="num rounded-xl border {s.mismatches ? 'border-neg/40' : 'border-line'}">
						<div class="flex flex-wrap items-center gap-3.5 px-4 py-3">
							<span class="grid size-8 place-items-center rounded-lg border border-line bg-raised text-muted">
								<Icon name={s.icon} size={15} />
							</span>
							<div class="min-w-56 flex-1">
								<div class="font-medium">{doc.state === 'UPLOADED' ? 'Processing…' : s.title}</div>
								<div class="truncate text-xs text-muted" title={doc.filename}>
									{[s.meta, doc.filename].filter(Boolean).join(' · ')}
								</div>
							</div>
							<div class="ml-auto text-right">
								<div class="font-semibold {s.amount.amount > 0 ? 'text-pos' : ''}">{formatSignedMoney(s.amount)}</div>
								<!-- Say what review is needed for, in the order it has to be fixed. -->
								<div class="text-xs {s.mismatches || s.unmatched ? 'text-neg' : match ? 'text-pos' : 'text-muted'}">
									{#if doc.state !== 'PARSED'}
										&nbsp;
									{:else if s.mismatches}
										{s.mismatches} {s.mismatches === 1 ? 'amount differs' : 'amounts differ'}
									{:else if s.unmatched}
										{s.unmatched} unknown {s.unmatched === 1 ? 'security' : 'securities'}
									{:else if !match}
										amounts match · choose the account
									{:else}
										✓ ready to import
									{/if}
								</div>
							</div>
							<div class="flex flex-wrap items-center gap-1.5">
								{@render secondaryActions(doc)}
								{#if doc.state === 'PARSED'}
									<Button
										variant={match ? 'secondary' : 'primary'}
										onclick={() => (reviewing = reviewing === doc.id ? null : doc.id)}
									>
										{reviewing === doc.id ? 'Close' : 'Review'}
									</Button>
									{#if match}
										<Button
											disabled={!!busy[doc.id]}
											onclick={async () => (await quickImport(doc, match)) && onchanged()}
										>
											{busy[doc.id] || 'Import'}
										</Button>
									{/if}
								{/if}
							</div>
						</div>
						{#if doc.transactions.length > 0}
							<p class="mx-4 mb-3 rounded-lg bg-accent-soft px-3 py-2 text-xs text-accent-fg">
								Re-analyzed after an earlier import: importing updates the {doc.transactions.length} existing
								{doc.transactions.length === 1 ? 'transaction' : 'transactions'} in place instead of adding duplicates.
							</p>
						{/if}
						{#if errors[doc.id]}<div class="px-4 pb-3"><ErrorNote message={errors[doc.id]} /></div>{/if}
						{#if reviewing === doc.id}
							<div class="border-t border-line p-4">
								<DocumentReview
									document={doc}
									suggestedCashAccount={doc.suggestedCashAccount}
									{portfolios}
									{cashAccounts}
									{securities}
									onimported={() => {
										reviewing = null;
										onchanged();
									}}
								/>
							</div>
						{/if}
					</li>
				{/each}
			</ul>
		{/if}
	{:else if tab === 'problems'}
		{#if byTab.problems.length === 0}
			<p class="rounded-xl border border-dashed border-line px-4 py-8 text-center text-sm text-muted">No problems.</p>
		{:else}
			<ul class="flex flex-col gap-2.5">
				{#each byTab.problems as doc (doc.id)}
					<li class="flex flex-wrap items-center gap-3 rounded-xl border border-neg/40 px-4 py-3">
						<div class="min-w-56 flex-1">
							<div class="truncate font-medium" title={doc.filename}>{doc.filename}</div>
							<div class="text-xs text-neg">{doc.error ?? 'Could not be read.'}</div>
						</div>
						{@render secondaryActions(doc)}
						{#if errors[doc.id]}<ErrorNote message={errors[doc.id]} />{/if}
					</li>
				{/each}
			</ul>
		{/if}
	{:else if tab === 'skipped'}
		<p class="text-[13px] text-muted">
			Recognised, but with nothing to book. They stay here so you can see they were handled.
		</p>
		{#each skippedGroups as g (g.key)}
			<section class="rounded-xl border border-line">
				<button
					type="button"
					onclick={() => (openGroup = openGroup === g.key ? null : g.key)}
					aria-expanded={openGroup === g.key}
					class="flex w-full flex-wrap items-center gap-3 px-4 py-3 text-left"
				>
					<span class="min-w-56 flex-1">
						<span class="block font-medium">{g.title}</span>
						<span class="block text-xs text-muted">{g.why}</span>
					</span>
					<span class="text-[13px] text-muted">{g.docs.length} documents {openGroup === g.key ? '▴' : '▾'}</span>
				</button>
				{#if openGroup === g.key}
					<ul class="border-t border-line">
						{#each g.docs as doc (doc.id)}
							<li class="flex flex-wrap items-center gap-2 border-t border-line px-4 py-2 first:border-t-0">
								<span class="min-w-0 flex-1 truncate text-[13px]" title={doc.filename}>{doc.filename}</span>
								{@render secondaryActions(doc)}
							</li>
						{/each}
					</ul>
				{/if}
			</section>
		{/each}
	{:else}
		<div class="overflow-x-auto rounded-xl border border-line">
			<table class="w-full min-w-[40rem] text-[13px]">
				<thead class="bg-raised text-xs text-muted">
					<tr>
						<th scope="col" class="py-2.5 pr-3 pl-4 text-left font-medium">Date</th>
						<th scope="col" class="px-3 py-2.5 text-left font-medium">Booked</th>
						<th scope="col" class="px-3 py-2.5 text-right font-medium">Cash</th>
						<th scope="col" class="py-2.5 pr-4 pl-3 text-right font-medium"><span class="sr-only">Actions</span></th>
					</tr>
				</thead>
				<tbody>
					{#each byTab.imported as doc (doc.id)}
						{@const s = summary(doc)}
						<tr class="border-t border-line hover:bg-raised">
							<td class="py-2 pr-3 pl-4 whitespace-nowrap text-muted">
								{doc.transactionDate ? formatDate(doc.transactionDate) : '—'}
							</td>
							<td class="max-w-0 px-3 py-2">
								{#if doc.transactions[0]}
									<a href="/transactions/{doc.transactions[0].id}" class="hover:underline">{s.title}</a>
								{:else}
									{s.title}
								{/if}
								<div class="truncate text-[11px] text-muted" title={doc.filename}>{doc.filename}</div>
							</td>
							<td class="px-3 py-2 text-right font-medium {s.amount.amount > 0 ? 'text-pos' : ''}">
								{formatSignedMoney(s.amount)}
							</td>
							<td class="py-2 pr-4 pl-3 text-right whitespace-nowrap">{@render secondaryActions(doc)}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>
