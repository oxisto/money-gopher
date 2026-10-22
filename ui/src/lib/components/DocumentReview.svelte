<script lang="ts">
	import { graphql, mutate } from '#lib/graphql/client.js';
	import { describeTransaction, formatUnits, labelOf } from '#lib/describe.js';
	import { formatMoney, formatPrice, formatDate, formatSignedMoney } from '#lib/format.js';
	import Button from './ui/Button.svelte';
	import Dialog from './ui/Dialog.svelte';
	import ErrorNote from './ui/ErrorNote.svelte';
	import Icon from './ui/Icon.svelte';
	import SelectField from './ui/SelectField.svelte';
	import TextField from './ui/TextField.svelte';
	import { asKind, assetKinds } from '#lib/assets.js';

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
		assetHint: string | null;
		isin: string | null;
		assetKindHint: string | null;
		asset: { id: string; displayName: string } | null;
		accountHint?: string | null;
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
		kind?: string | null;
		reference?: string | null;
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

	const createAssetMutation = graphql(`
		mutation CreateAssetFromImport($input: CreateAssetInput!) {
			createAsset(input: $input) {
				id
				displayName
			}
		}
	`);

	// Pre-select the matched cash account if the server found one, otherwise fall back to first.
	let portfolioID = $state(portfolios[0]?.id ?? '');
	let cashAccountID = $state(suggestedCashAccount?.id ?? cashAccounts[0]?.id ?? '');

	// A document that bundles several statements (an LBS year-end bundle:
	// the Bausparvertrag and the loan) tags each row with the account it
	// belongs to; then one cash account is picked per section instead.
	const sections = $derived(
		[...new Set(doc.stagedTransactions.map((tx) => tx.accountHint).filter((h): h is string => !!h))]
	);
	const rows = $derived(
		sections.length > 0
			? [...doc.stagedTransactions].sort(
					(a, b) =>
						sections.indexOf(a.accountHint ?? '') - sections.indexOf(b.accountHint ?? '') ||
						a.time.getTime() - b.time.getTime()
				)
			: doc.stagedTransactions
	);
	// A hint is "<part> <contract number>". The number plus the kind the
	// part implies is exact: a Bausparvertrag and the loan against it
	// share the number but not the kind. Without a stored number, a name
	// that contains the part's word wins, then the kind alone.
	function guessAccount(hint: string): string {
		const [word, number] = hint.split(' ');
		const kind = /kredit|darlehen/i.test(word) ? 'LOAN' : /spar/i.test(word) ? 'SAVINGS' : null;
		const byReference = cashAccounts.filter((a) => number && a.reference === number);
		const exact = byReference.find((a) => !kind || (a.kind ?? 'CASH') === kind) ?? byReference[0];
		if (exact) return exact.id;
		const byName = cashAccounts.find((a) => a.displayName.toLowerCase().includes(word.toLowerCase()));
		if (byName) return byName.id;
		return (kind && cashAccounts.find((a) => a.kind === kind)?.id) ?? '';
	}
	// Every section has a key before its select renders: a bound select
	// with no value at all is an error, an empty string is "— select —".
	let sectionAccounts = $state<Record<string, string>>(
		Object.fromEntries(sections.map((s) => [s, guessAccount(s)]))
	);
	$effect.pre(() => {
		for (const s of sections) {
			if (sectionAccounts[s] === undefined) sectionAccounts[s] = guessAccount(s);
		}
	});
	function accountName(id: string): string {
		return cashAccounts.find((a) => a.id === id)?.displayName ?? '';
	}
	let assetOverrides = $state<Record<string, string>>({});
	let localSecurities = $state<Security[]>([]);
	let createDialog = $state<{
		txId: string;
		name: string;
		isin: string;
		kind: string;
		currency: string;
		quoteProvider: string;
	} | null>(null);
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
			name: tx.assetHint ?? tx.isin ?? '',
			isin: tx.isin ?? '',
			kind: tx.assetKindHint ?? '',
			// The document says what the asset trades in: a foreign bond's
			// price is in its own currency, everything else in the settlement
			// currency. Quotes by ISIN need no ticker, so an ISIN is enough
			// to get them.
			currency: tx.price?.currency ?? tx.currency,
			quoteProvider: tx.isin ? 'ing' : ''
		};
	}

	async function confirmCreateAsset() {
		if (!createDialog) return;
		creatingDialog = true;
		const result = await mutate(createAssetMutation, {
			input: {
				displayName: createDialog.name,
				kind: asKind(createDialog.kind),
				identifiers: createDialog.isin ? [{ kind: 'ISIN', value: createDialog.isin }] : [],
				listings: createDialog.isin
					? [
							{
								ticker: createDialog.isin,
								currency: createDialog.currency.toUpperCase(),
								quoteProvider: createDialog.quoteProvider || undefined
							}
						]
					: []
			}
		});
		creatingDialog = false;
		if (result.error) {
			error = result.error;
			return;
		}
		const newSec = result.data?.createAsset;
		if (newSec) {
			localSecurities = [...localSecurities, newSec];
			assetOverrides[createDialog.txId] = newSec.id;
		}
		createDialog = null;
	}

	async function confirm() {
		if (portfolioRequired && !portfolioID) {
			error = 'Please select a portfolio.';
			return;
		}
		if (sections.length > 0) {
			const missing = sections.find((s) => !sectionAccounts[s]);
			if (missing) {
				error = `Please select a cash account for ${missing}.`;
				return;
			}
		} else if (!cashAccountID) {
			error = 'Please select a cash account.';
			return;
		}

		confirming = true;
		error = null;

		const txOverrides = doc.stagedTransactions
			.map((tx) => ({
				id: tx.id,
				assetID: assetOverrides[tx.id] || undefined,
				cashAccountID: tx.accountHint ? sectionAccounts[tx.accountHint] : undefined
			}))
			.filter((ov) => ov.assetID || ov.cashAccountID);

		const result = await mutate(confirmImport, {
			documentID: doc.id,
			input: {
				portfolioID: portfolioRequired ? portfolioID : null,
				cashAccountID: sections.length > 0 ? sectionAccounts[sections[0]] : cashAccountID,
				transactions: txOverrides
			}
		});

		confirming = false;
		if (result.error) {
			error = result.error;
		} else {
			onimported();
		}
	}
</script>

<!-- Split-pane: PDF on the left, review form on the right -->
<div class="flex flex-col gap-4">
	<!-- PDF toggle (visible on narrow screens; on wide screens pane is always shown) -->
	<div class="flex items-center justify-between 2xl:hidden">
		<span class="text-xs text-muted">Source document</span>
		<button
			onclick={() => (showPDF = !showPDF)}
			class="text-xs text-muted hover:text-fg"
		>
			{showPDF ? 'Hide PDF' : 'Show PDF'}
		</button>
	</div>

	<div class="flex flex-col gap-6 2xl:flex-row 2xl:gap-4 2xl:h-[min(65vh,820px)]">
		<!-- PDF pane -->
		{#if showPDF}
			<!-- Stacked, the form comes first and the PDF below it. -->
			<div class="order-last flex-1 min-w-0 2xl:order-first 2xl:flex-none 2xl:w-[540px] 2xl:flex 2xl:flex-col">
				<div class="hidden 2xl:flex items-center justify-between mb-1">
					<span class="text-xs text-muted">Source document</span>
					<a href={pdfURL} target="_blank" rel="noopener" class="text-xs text-muted hover:text-fg">Open in new tab ↗</a>
				</div>
				<embed
					src={pdfURL}
					type="application/pdf"
					class="w-full flex-1 rounded-lg border border-line"
					style="height: min(65vh, 820px);"
				/>
				<a href={pdfURL} target="_blank" rel="noopener" class="mt-1 block text-xs text-muted hover:text-fg 2xl:hidden">Open in new tab ↗</a>
			</div>
		{/if}

		<!-- Review form -->
		<div class="flex-1 min-w-0 flex flex-col gap-4 2xl:overflow-y-auto">
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
				{#each sections as section (section)}
					<div class="flex-1 min-w-[180px]">
						<SelectField bind:value={sectionAccounts[section]} label="Cash account for {section}">
							<option value="">— select —</option>
							{#each cashAccounts as a (a.id)}
								<option value={a.id}>{a.displayName} ({a.currency})</option>
							{/each}
						</SelectField>
					</div>
				{/each}
				<div class="flex-1 min-w-[180px]" class:hidden={sections.length > 0}>
					<SelectField bind:value={cashAccountID} label="Cash account">
						<option value="">— select —</option>
						{#each cashAccounts as a (a.id)}
							<option value={a.id}>{a.displayName} ({a.currency})</option>
						{/each}
					</SelectField>
					{#if suggestedCashAccount && suggestedCashAccount.id === cashAccountID}
						<p class="mt-1 text-xs text-pos">✓ Auto-matched from document IBAN</p>
					{:else if suggestedCashAccount}
						<p class="mt-1 text-xs text-muted">
							Suggested: {suggestedCashAccount.displayName} —
							<button class="underline" onclick={() => (cashAccountID = suggestedCashAccount!.id)}>
								use it
							</button>
						</p>
					{/if}
				</div>
			</div>

			<!-- Staged transactions table -->
			<div class="overflow-x-auto rounded-xl border border-line">
				<table class="num min-w-full text-[13px]">
					<thead class="bg-raised text-xs text-muted">
						<tr>
							<th class="py-2 pr-2 pl-3 text-left font-medium">Date</th>
							<th class="px-2 py-2 text-left font-medium">Event</th>
							<th class="px-2 py-2 text-left font-medium">Security</th>
							<th class="px-2 py-2 text-right font-medium">Units</th>
							<th class="px-2 py-2 text-right font-medium">Price</th>
							<th class="px-2 py-2 text-right font-medium">Fees</th>
							<th class="px-2 py-2 text-right font-medium">Taxes</th>
							<th class="py-2 pr-3 pl-2 text-right font-medium">Cash</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-line">
						{#each rows as tx, i (tx.id)}
							{#if sections.length > 0 && (i === 0 || rows[i - 1].accountHint !== tx.accountHint)}
								<tr class="bg-raised">
									<td colspan="7" class="px-3 py-1.5 text-xs font-medium text-muted">
										{tx.accountHint}
										{#if sectionAccounts[tx.accountHint ?? '']}
											<span class="font-normal">→ {accountName(sectionAccounts[tx.accountHint ?? ''])}</span>
										{/if}
									</td>
								</tr>
							{/if}
							<tr>
								<td class="py-2 pr-2 pl-3 text-muted whitespace-nowrap">{formatDate(tx.time)}</td>
								<td class="px-2 py-2 whitespace-nowrap">
									<span class="flex items-center gap-1.5">
										<span class="text-muted"><Icon name={describeTransaction(tx).icon} size={13} /></span>
										{labelOf(tx.type)}
									</span>
								</td>
								<td class="px-2 py-2 max-w-[220px]">
									{#if tx.asset}
										<span class="text-fg">{tx.asset.displayName}</span>
										{#if tx.isin}
											<span class="ml-1 text-xs text-muted">{tx.isin}</span>
										{/if}
									{:else if tx.assetHint || tx.isin}
										<div class="flex flex-col gap-1">
											{#if tx.assetHint}
												<span class="text-xs text-muted italic truncate">"{tx.assetHint}"</span>
											{/if}
											<div class="flex items-center gap-1">
												<select
													bind:value={assetOverrides[tx.id]}
													class="h-7 min-w-0 flex-1 py-0 pr-7 pl-2 text-xs"
												>
													<option value="">— unmatched —</option>
													{#each allSecurities as s (s.id)}
														<option value={s.id}>{s.displayName}</option>
													{/each}
												</select>
												{#if tx.isin && !assetOverrides[tx.id]}
													<button
														onclick={() => openCreateDialog(tx)}
														class="h-7 shrink-0 rounded-md border border-accent/30 bg-accent-soft px-2 text-xs font-medium text-accent-fg"
													>
														Create
													</button>
												{/if}
											</div>
										</div>
									{:else}
										<span class="text-xs text-muted">—</span>
									{/if}
								</td>
								<td class="px-2 py-2 text-right tabular-nums text-fg">
									{tx.units > 0 ? formatUnits(tx.units) : '—'}
								</td>
								<td class="px-2 py-2 text-right tabular-nums text-fg">
									{tx.price ? formatPrice(tx.price) : '—'}
									{#if tx.fxRate}
										<div
											class="text-xs text-muted"
											title={`1 ${tx.currency} = ${tx.fxRate} ${tx.price?.currency}`}
										>
											FX {tx.fxRate}
										</div>
									{/if}
								</td>
								<td class="px-2 py-2 text-right tabular-nums text-fg">
									{tx.fees.amount !== 0 ? formatMoney(tx.fees) : '—'}
								</td>
								<td class="px-2 py-2 text-right tabular-nums text-fg">
									{tx.taxes.amount !== 0 ? formatMoney(tx.taxes) : '—'}
								</td>
								<td class="py-2 pr-3 pl-2 text-right">
									<div class="font-medium {tx.cashDelta.amount > 0 ? 'text-pos' : ''}">
										{tx.cashDelta.amount !== 0 ? formatSignedMoney(tx.cashDelta) : '—'}
									</div>
									{#if tx.calculatedCashDelta}
										{@const diff = tx.cashDelta.amount - tx.calculatedCashDelta.amount}
										<div
											class="text-[11px] {diff ? 'font-medium text-neg' : 'text-pos'}"
											title={diff
												? `The document says ${formatMoney(tx.cashDelta)}, units × price ± fees and taxes give ${formatMoney(tx.calculatedCashDelta)}. Usually a line item that isn't captured yet (a fee, a foreign tax, accrued interest).`
												: 'Matches units × price ± fees and taxes.'}
										>
											{diff ? `differs by ${formatMoney({ amount: Math.abs(diff), currency: tx.cashDelta.currency })}` : '✓ matches'}
										</div>
									{/if}
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

<Dialog
	open={createDialog !== null}
	title="Add asset"
	onclose={() => (createDialog = null)}
>
	{#if createDialog}
		<div class="flex flex-col gap-3">
			<TextField label="Name" bind:value={createDialog.name} placeholder="Security name" required />
			<SelectField bind:value={createDialog.kind} label="Kind">
				<option value="">Not set</option>
				{#each assetKinds as [value, label] (value)}
					<option {value}>{label}</option>
				{/each}
			</SelectField>
			{#if createDialog.isin}
				<div>
					<span class="mb-1 block text-xs font-medium text-muted">ISIN</span>
					<p class="num text-sm">{createDialog.isin}</p>
				</div>
				<div class="grid grid-cols-2 gap-3">
					<TextField
						label="Listing currency"
						bind:value={createDialog.currency}
						maxlength={3}
						required
					/>
					<SelectField bind:value={createDialog.quoteProvider} label="Quote provider">
						<option value="">None</option>
						<option value="ing">ING (by ISIN)</option>
					</SelectField>
				</div>
				<p class="text-xs text-muted">
					The currency the asset is quoted in, as the document states it. Quotes are converted
					into the portfolio's currency when they differ.
				</p>
			{/if}
			<div class="mt-2 flex justify-end gap-2">
				<Button variant="secondary" onclick={() => (createDialog = null)}>Cancel</Button>
				<Button
					onclick={confirmCreateAsset}
					disabled={creatingDialog ||
						!createDialog.name ||
						(!!createDialog.isin && createDialog.currency.trim().length !== 3)}
				>
					{creatingDialog ? 'Creating…' : 'Create asset'}
				</Button>
			</div>
		</div>
	{/if}
</Dialog>
