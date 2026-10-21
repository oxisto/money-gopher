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
		const result = await mutate(createSecurityMutation, {
			input: {
				displayName: createDialog.name,
				identifiers: createDialog.isin ? [{ kind: 'ISIN', value: createDialog.isin }] : []
			}
		});
		creatingDialog = false;
		if (result.error) {
			error = result.error;
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

		const result = await mutate(confirmImport, {
			documentID: doc.id,
			input: {
				portfolioID: portfolioRequired ? portfolioID : null,
				cashAccountID,
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
				<div class="flex-1 min-w-[180px]">
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
						{#each doc.stagedTransactions as tx (tx.id)}
							<tr>
								<td class="py-2 pr-2 pl-3 text-muted whitespace-nowrap">{formatDate(tx.time)}</td>
								<td class="px-2 py-2 whitespace-nowrap">
									<span class="flex items-center gap-1.5">
										<span class="text-muted"><Icon name={describeTransaction(tx).icon} size={13} /></span>
										{labelOf(tx.type)}
									</span>
								</td>
								<td class="px-2 py-2 max-w-[220px]">
									{#if tx.security}
										<span class="text-fg">{tx.security.displayName}</span>
										{#if tx.isin}
											<span class="ml-1 text-xs text-muted">{tx.isin}</span>
										{/if}
									{:else if tx.securityHint || tx.isin}
										<div class="flex flex-col gap-1">
											{#if tx.securityHint}
												<span class="text-xs text-muted italic truncate">"{tx.securityHint}"</span>
											{/if}
											<div class="flex items-center gap-1">
												<select
													bind:value={securityOverrides[tx.id]}
													class="h-7 min-w-0 flex-1 py-0 pr-7 pl-2 text-xs"
												>
													<option value="">— unmatched —</option>
													{#each allSecurities as s (s.id)}
														<option value={s.id}>{s.displayName}</option>
													{/each}
												</select>
												{#if tx.isin && !securityOverrides[tx.id]}
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
	title="Add security"
	onclose={() => (createDialog = null)}
>
	{#if createDialog}
		<div class="flex flex-col gap-3">
			<TextField label="Name" bind:value={createDialog.name} placeholder="Security name" required />
			{#if createDialog.isin}
				<div>
					<span class="mb-1 block text-xs font-medium text-muted">ISIN</span>
					<p class="num text-sm">{createDialog.isin}</p>
				</div>
			{/if}
			<div class="mt-2 flex justify-end gap-2">
				<Button variant="secondary" onclick={() => (createDialog = null)}>Cancel</Button>
				<Button onclick={confirmCreateSecurity} disabled={creatingDialog || !createDialog.name}>
					{creatingDialog ? 'Creating…' : 'Create security'}
				</Button>
			</div>
		</div>
	{/if}
</Dialog>
