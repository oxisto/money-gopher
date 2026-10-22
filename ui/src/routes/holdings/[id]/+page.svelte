<script lang="ts">
	import { refreshAll } from '$app/navigation';
	import PriceChart from '#lib/components/holdings/PriceChart.svelte';
	import { holdingsOf } from '#lib/components/holdings/HoldingsTable.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import Dialog from '#lib/components/ui/Dialog.svelte';
	import ErrorNote from '#lib/components/ui/ErrorNote.svelte';
	import Icon from '#lib/components/ui/Icon.svelte';
	import TextField from '#lib/components/ui/TextField.svelte';
	import { graphql, mutate } from '#lib/graphql/client.js';
	import SelectField from '#lib/components/ui/SelectField.svelte';
	import { asKind, handValuedKinds, kindLabel, type SecurityKind } from '#lib/securities.js';
	import { describeTransaction, formatUnits } from '#lib/describe.js';
	import {
		formatDate,
		formatMoney,
		formatPercent,
		formatPrice,
		formatSignedMoney,
		toneOf,
		type Money
	} from '#lib/format.js';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	const d = $derived(data);
	const security = $derived(d.security);
	const listing = $derived(security?.listings[0] ?? null);
	// Without a provider, prices come from whoever owns the asset: a
	// valuation of the office room, entered here.
	const valuedByHand = $derived(!!listing && !listing.quoteProvider);

	const recordQuote = graphql(`
		mutation RecordValuation($listingID: ID!, $time: Time!, $price: Decimal!) {
			recordQuote(listingID: $listingID, time: $time, price: $price) {
				time
			}
		}
	`);
	// A personal asset is edited here, not on the Securities page, which
	// lists the shared instruments only.
	const owned = $derived(!!security?.owner);

	const updateAsset = graphql(`
		mutation UpdateAsset($id: ID!, $input: UpdateSecurityInput!) {
			updateSecurity(id: $id, input: $input) {
				id
			}
		}
	`);
	let editOpen = $state(false);
	let editName = $state('');
	let editKind = $state('');
	let editError = $state<string | null>(null);
	let savingEdit = $state(false);

	function openEdit() {
		editName = security?.displayName ?? '';
		editKind = security?.kind ?? '';
		editError = null;
		editOpen = true;
	}

	async function saveEdit(event: SubmitEvent) {
		event.preventDefault();
		if (!security || !editName.trim()) return;
		savingEdit = true;
		const result = await mutate(updateAsset, {
			id: security.id,
			input: { displayName: editName.trim(), kind: asKind(editKind) }
		});
		savingEdit = false;
		editError = result.error;
		if (!editError) {
			editOpen = false;
			await refreshAll();
		}
	}

	let valuationOpen = $state(false);
	let valuationDate = $state(new Date().toISOString().slice(0, 10));
	let valuationPrice = $state('');
	let valuationError = $state<string | null>(null);
	let saving = $state(false);

	async function saveValuation(event: SubmitEvent) {
		event.preventDefault();
		if (!listing) return;
		saving = true;
		const result = await mutate(recordQuote, {
			listingID: listing.id,
			time: new Date(valuationDate + 'T12:00:00Z'),
			price: valuationPrice.replace(',', '.').trim()
		});
		saving = false;
		valuationError = result.error;
		if (!valuationError) {
			valuationOpen = false;
			valuationPrice = '';
			await refreshAll();
		}
	}
	const isin = $derived(security?.identifiers.find((i) => i.kind === 'ISIN')?.value ?? null);

	const holding = $derived.by(() => {
		if (!security) return null;
		const positions = d.portfolios.flatMap((p) => p.snapshot.positions).filter((p) => p.security.id === security.id);
		return positions.length ? holdingsOf(positions)[0] : null;
	});

	const history = $derived(
		d.portfolios
			.flatMap((p) => p.transactions.map((t) => ({ ...t, portfolio: p.displayName })))
			.filter((t) => t.security?.id === security?.id)
			.sort((a, b) => b.time.getTime() - a.time.getTime())
	);

	const currency = $derived(holding?.marketValue.currency ?? listing?.currency ?? 'EUR');
	const total = (pick: (t: (typeof history)[number]) => number): Money => ({
		amount: history.reduce((s, t) => s + pick(t), 0),
		currency
	});
	const dividends = $derived(total((t) => (t.type === 'DIVIDEND' ? t.cashDelta.amount : 0)));
	const realized = $derived(total((t) => (t.type === 'SELL' ? (t.gain?.amount ?? 0) : 0)));
	const totalReturn = $derived({
		amount: (holding?.profitOrLoss.amount ?? 0) + realized.amount + dividends.amount,
		currency
	});
	const portfolioValue = $derived(d.portfolios.reduce((s, p) => s + p.snapshot.totalMarketValue.amount, 0));

	const position = $derived.by(() => {
		const rows: { label: string; value: string; sub?: string; tone?: string }[] = [];
		if (holding) {
			rows.push(
				{ label: 'Units', value: formatUnits(holding.units) },
				{ label: 'Average cost', value: formatPrice(holding.purchasePrice, 2) },
				{ label: 'Cost basis', value: formatMoney(holding.purchaseValue) },
				{ label: 'Market value', value: formatMoney(holding.marketValue) },
				{
					label: 'Unrealized',
					value: formatSignedMoney(holding.profitOrLoss),
					sub: formatPercent(holding.gains),
					tone: toneOf(holding.profitOrLoss.amount)
				}
			);
		} else {
			rows.push({ label: 'Units', value: '0', sub: 'no longer held' });
		}
		rows.push(
			{ label: 'Realized from sales', value: formatSignedMoney(realized), tone: toneOf(realized.amount) },
			{
				label: 'Dividends received',
				value: formatSignedMoney(dividends),
				sub: `${history.filter((t) => t.type === 'DIVIDEND').length} payments`,
				tone: toneOf(dividends.amount)
			},
			{
				label: 'Total return',
				value: formatSignedMoney(totalReturn),
				sub: 'unrealized + realized + dividends',
				tone: toneOf(totalReturn.amount)
			}
		);
		if (holding && portfolioValue) {
			rows.push({ label: 'Weight in portfolios', value: `${((holding.marketValue.amount / portfolioValue) * 100).toFixed(1)}%` });
		}
		return rows;
	});

	const chart = $derived({
		quotes: (listing?.quotes ?? []).map((q) => ({ time: q.time, price: Number(q.price.amount) })),
		trades: history
			.filter((t) => (t.type === 'BUY' || t.type === 'SELL') && t.price)
			.map((t) => ({
				time: t.time,
				price: Number(t.price!.amount),
				kind: t.type === 'BUY' ? ('buy' as const) : ('sell' as const),
				label: `${describeTransaction(t).title} @ ${formatPrice(t.price!, 4)} · ${formatDate(t.time)}`
			})),
		dividends: history.filter((t) => t.type === 'DIVIDEND').map((t) => t.time)
	});

	let filter = $state<'all' | 'trades' | 'dividends'>('all');
	const shownHistory = $derived(
		history.filter((t) =>
			filter === 'all' ? true : filter === 'dividends' ? t.type === 'DIVIDEND' : t.type !== 'DIVIDEND'
		)
	);
</script>

{#if !security}
	<p class="text-sm text-muted">Security not found.</p>
{:else}
	<nav aria-label="Breadcrumb" class="mb-4 flex gap-1.5 text-[13px] text-muted">
		<a href="/holdings" class="hover:text-fg">Holdings</a>
		<span aria-hidden="true">/</span>
		<span class="text-fg">{security.displayName}</span>
	</nav>

	<div class="mb-6 flex flex-wrap items-end gap-4">
		<div class="min-w-64 flex-1">
			<h1 class="text-2xl font-semibold tracking-tight">{security.displayName}</h1>
			<p class="mt-0.5 text-[13px] text-muted">
				<!-- ING listings use the ISIN as their ticker; don't print it twice. -->
				{[
					security.kind ? kindLabel[security.kind as SecurityKind] : null,
					isin,
					listing && !valuedByHand
						? `${listing.exchange ?? (listing.ticker !== isin ? listing.ticker : 'traded')} in ${listing.currency}`
						: null,
					listing?.quoteProvider
						? `quotes: ${listing.quoteProvider}`
						: listing
							? `valued by hand in ${listing.currency}`
							: 'no price series'
				]
					.filter(Boolean)
					.join(' · ')}
				{#if security.underlyingSecurity}
					· derives from
					<a href="/holdings/{security.underlyingSecurity.id}" class="text-fg hover:underline">
						{security.underlyingSecurity.displayName}
					</a>
				{/if}
			</p>
			{#if listing?.latestQuote}
				<div class="num mt-3 flex flex-wrap items-baseline gap-2.5">
					<span class="text-3xl font-semibold tracking-tight">{formatPrice(listing.latestQuote.price, 4)}</span>
					<span class="text-xs text-muted">
						as of {new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' }).format(
							listing.latestQuote.time
						)}
					</span>
				</div>
			{/if}
		</div>
		{#if valuedByHand}
			<Button onclick={() => (valuationOpen = true)}>Record valuation</Button>
		{/if}
		{#if owned}
			<Button variant="secondary" onclick={openEdit}>Edit asset</Button>
		{:else}
			<a
				href="/securities"
				class="inline-flex h-8.5 items-center rounded-lg border border-line bg-surface px-3 text-sm font-medium shadow-xs hover:bg-raised"
			>
				Edit security
			</a>
		{/if}
	</div>

	<Dialog bind:open={editOpen} title="Edit asset">
		<form onsubmit={saveEdit} class="flex flex-col gap-3">
			<TextField bind:value={editName} label="Name" required />
			<SelectField bind:value={editKind} label="Kind">
				{#each handValuedKinds as [value, label] (value)}
					<option {value}>{label}</option>
				{/each}
			</SelectField>
			<p class="text-xs text-muted">
				The cost basis is the transfer-in in the history below; open it to change the amount or the
				date. Valuations are recorded with the button next to this one.
			</p>
			<ErrorNote message={editError} />
			<div class="mt-2 flex justify-end gap-2">
				<Button variant="secondary" onclick={() => (editOpen = false)}>Cancel</Button>
				<Button type="submit" disabled={savingEdit || !editName.trim()}>
					{savingEdit ? 'Saving…' : 'Save'}
				</Button>
			</div>
		</form>
	</Dialog>

	<Dialog bind:open={valuationOpen} title="Record valuation">
		<form onsubmit={saveValuation} class="flex flex-col gap-3">
			<p class="text-xs text-muted">
				What one unit is worth on that date, in {listing?.currency}. The chart and the market
				value follow the newest valuation; the date lets you record the past too.
			</p>
			<div class="grid grid-cols-2 gap-3">
				<TextField bind:value={valuationDate} label="Date" type="date" required />
				<TextField bind:value={valuationPrice} label="Value per unit" placeholder="e.g. 42000" required />
			</div>
			<ErrorNote message={valuationError} />
			<div class="mt-2 flex justify-end gap-2">
				<Button variant="secondary" onclick={() => (valuationOpen = false)}>Cancel</Button>
				<Button type="submit" disabled={saving || !valuationPrice.trim()}>
					{saving ? 'Saving…' : 'Save valuation'}
				</Button>
			</div>
		</form>
	</Dialog>

	<div class="flex flex-col gap-6">
		<div class="flex flex-wrap items-start gap-5">
			<section class="min-w-0 flex-[2_1_34rem] rounded-xl border border-line p-4">
				<PriceChart
					quotes={chart.quotes}
					trades={chart.trades}
					dividends={chart.dividends}
					averageCost={holding ? Number(holding.purchasePrice.amount) : null}
					{currency}
				/>
			</section>

			<section class="num min-w-0 flex-[1_1_18rem] rounded-xl border border-line p-4">
				<h2 class="mb-1 text-sm font-semibold">Your position</h2>
				<dl>
					{#each position as row (row.label)}
						<div class="grid grid-cols-[1fr_auto] gap-3 border-t border-line py-2 text-[13px] first:border-t-0">
							<dt class="text-muted">{row.label}</dt>
							<dd class="text-right {row.tone ?? ''}">
								{row.value}
								{#if row.sub}<span class="block text-[11px] {row.tone ? '' : 'text-muted'}">{row.sub}</span>{/if}
							</dd>
						</div>
					{/each}
				</dl>
			</section>
		</div>

		<section class="flex flex-col gap-3">
			<div class="flex flex-wrap items-center gap-3">
				<h2 class="flex-1 text-sm font-semibold">History</h2>
				<div role="radiogroup" aria-label="Filter history" class="flex gap-1.5">
					{#each [['all', 'All'], ['trades', 'Trades'], ['dividends', 'Dividends']] as [value, label] (value)}
						<button
							type="button"
							role="radio"
							aria-checked={filter === value}
							onclick={() => (filter = value as typeof filter)}
							class="h-7.5 rounded-full border px-3 text-xs font-medium
								{filter === value ? 'border-accent bg-accent-soft text-accent-fg' : 'border-line text-muted hover:text-fg'}"
						>
							{label}
						</button>
					{/each}
				</div>
			</div>

			{#if shownHistory.length}
				<div class="overflow-x-auto rounded-xl border border-line">
					<table class="w-full min-w-[46rem] text-[13px]">
						<thead class="bg-raised text-xs text-muted">
							<tr>
								<th scope="col" class="py-2.5 pr-3 pl-4 text-left font-medium">Date</th>
								<th scope="col" class="px-3 py-2.5 text-left font-medium">Event</th>
								<th scope="col" class="px-3 py-2.5 text-right font-medium">Price</th>
								<th scope="col" class="px-3 py-2.5 text-right font-medium">Fees &amp; tax</th>
								<th scope="col" class="px-3 py-2.5 text-right font-medium">Cash</th>
								<th scope="col" class="py-2.5 pr-4 pl-3 text-right font-medium"><span class="sr-only">Details</span></th>
							</tr>
						</thead>
						<tbody>
							{#each shownHistory as t (t.id)}
								{@const feesAndTax = t.fees.amount + t.taxes.amount}
								<tr class="border-t border-line hover:bg-raised">
									<td class="py-2 pr-3 pl-4 whitespace-nowrap">{formatDate(t.time)}</td>
									<td class="px-3 py-2">
										<span class="flex items-center gap-2">
											<span class="text-muted"><Icon name={describeTransaction(t).icon} size={14} /></span>
											{describeTransaction(t).title}
										</span>
										{#if t.type === 'SELL' && t.gain}
											<span class="ml-6 text-[11px] {toneOf(t.gain.amount)}">realized {formatSignedMoney(t.gain)}</span>
										{/if}
									</td>
									<td class="px-3 py-2 text-right">{t.price ? formatPrice(t.price, 4) : '—'}</td>
									<td class="px-3 py-2 text-right text-muted">
										{feesAndTax ? formatMoney({ amount: feesAndTax, currency: t.fees.currency }) : '—'}
									</td>
									<td class="px-3 py-2 text-right font-medium {t.cashDelta.amount > 0 ? 'text-pos' : ''}">
										{t.cashDelta.amount ? formatSignedMoney(t.cashDelta) : '—'}
									</td>
									<td class="py-2 pr-4 pl-3 text-right">
										<a href="/transactions/{t.id}" class="text-xs text-muted underline hover:text-fg">Details</a>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{:else}
				<p class="rounded-xl border border-dashed border-line px-4 py-6 text-center text-sm text-muted">Nothing here.</p>
			{/if}
		</section>
	</div>
{/if}
