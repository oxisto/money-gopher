<script lang="ts">
	import { refreshAll } from '$app/navigation';
	import HoldingsTable, { holdingsOf } from '#lib/components/holdings/HoldingsTable.svelte';
	import TransactionForm from '#lib/components/TransactionForm.svelte';
	import ActivityFeed from '#lib/components/activity/ActivityFeed.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import Dialog from '#lib/components/ui/Dialog.svelte';
	import ErrorNote from '#lib/components/ui/ErrorNote.svelte';
	import Icon from '#lib/components/ui/Icon.svelte';
	import Segmented from '#lib/components/ui/Segmented.svelte';
	import SelectField from '#lib/components/ui/SelectField.svelte';
	import StatStrip from '#lib/components/ui/StatStrip.svelte';
	import TextField from '#lib/components/ui/TextField.svelte';
	import { formatMoney, formatPercent, formatSignedMoney, toneOf } from '#lib/format.js';
	import { graphql, mutate } from '#lib/graphql/client.js';
	import type { FeedTransaction } from '#lib/transaction-types.js';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let tab = $state<'holdings' | 'transactions' | 'return'>('holdings');
	// Valuation has no period; the return does. Three are fetched, one shows.
	let returnPeriod = $state<'ytd' | 'oneYear' | 'allTime'>('ytd');
	let addOpen = $state(false);

	const portfolio = $derived(data.portfolio);

	const updatePortfolio = graphql(`
		mutation UpdatePortfolio($id: ID!, $displayName: String!, $cashAccountID: ID!) {
			updatePortfolio(id: $id, input: { displayName: $displayName, cashAccountID: $cashAccountID }) {
				id
				displayName
				cashAccount { id displayName }
			}
		}
	`);
	let editOpen = $state(false);
	let newName = $state('');
	let newCashAccountID = $state('');
	let editError = $state<string | null>(null);

	function openEdit() {
		newName = portfolio?.displayName ?? '';
		newCashAccountID = portfolio?.cashAccount.id ?? '';
		editError = null;
		editOpen = true;
	}

	async function saveEdit(event: SubmitEvent) {
		event.preventDefault();
		if (!portfolio || !newName.trim() || !newCashAccountID) return;
		const result = await mutate(updatePortfolio, {
			id: portfolio.id,
			displayName: newName.trim(),
			cashAccountID: newCashAccountID
		});
		editError = result.error;
		if (!editError) {
			editOpen = false;
			// The sidebar lists portfolios by name too.
			await refreshAll();
		}
	}

	function refresh() {
		refreshAll();
	}
</script>

{#if portfolio}
	{@const snap = portfolio.snapshot}
	{@const ret = snap[returnPeriod]}
	{@const years = snap.returnsByYear}
	{@const scale = Math.max(1, ...years.map((y) => Math.abs(y.realized.amount) + Math.abs(y.income.amount) + Math.abs(y.costs.amount) + Math.abs(y.valueChange.amount)))}
	<div class="mb-6 flex flex-wrap items-center gap-3">
		<div class="flex-1">
			<h1 class="text-lg font-semibold">{portfolio.displayName}</h1>
			<p class="text-[13px] text-muted">
				Settles against
				<a href="/cash-accounts/{portfolio.cashAccount.id}" class="text-fg hover:underline">
					{portfolio.cashAccount.displayName}
				</a>
			</p>
		</div>
		<Button variant="secondary" onclick={openEdit}>Edit</Button>
		<Button onclick={() => (addOpen = true)}><Icon name="plus" size={15} /> Add transaction</Button>
	</div>

	<Dialog bind:open={editOpen} title="Edit portfolio">
		<form onsubmit={saveEdit} class="flex flex-col gap-3">
			<TextField bind:value={newName} label="Name" required />
			<SelectField bind:value={newCashAccountID} label="Settles against" required>
				{#each data.cashAccounts as a (a.id)}
					<option value={a.id}>{a.displayName}</option>
				{/each}
			</SelectField>
			<p class="text-xs text-muted">
				New trades and dividends default to this account. Transactions already booked keep theirs.
			</p>
			<ErrorNote message={editError} />
			<div class="mt-2 flex justify-end gap-2">
				<Button variant="secondary" onclick={() => (editOpen = false)}>Cancel</Button>
				<Button type="submit" disabled={!newName.trim() || !newCashAccountID}>Save</Button>
			</div>
		</form>
	</Dialog>

	<Dialog bind:open={addOpen} title="Add transaction" wide>
		<TransactionForm
			portfolioID={portfolio.id}
			defaultCashAccountID={portfolio.cashAccount.id}
			securities={[...data.securities, ...data.assets]}
			cashAccounts={data.cashAccounts}
			oncreated={() => {
				addOpen = false;
				refresh();
			}}
		/>
	</Dialog>

	<div class="flex flex-col gap-6">
		<StatStrip
			stats={[
				{
					label: 'Market value',
					value: formatMoney(snap.totalMarketValue),
					sub: `${snap.positions.length} positions`
				},
				{ label: 'Cost basis', value: formatMoney(snap.totalPurchaseValue), sub: 'FIFO, without fees and taxes' },
				{
					label: 'Unrealized gain',
					value: formatSignedMoney(snap.totalProfitOrLoss),
					sub: formatPercent(snap.totalGains),
					tone: toneOf(snap.totalProfitOrLoss.amount)
				},
				{
					label: 'Return, all time',
					value: formatPercent(snap.performance.timeWeightedReturn),
					sub: 'time-weighted',
					tone: toneOf(snap.performance.timeWeightedReturn)
				}
			]}
		/>

		<Segmented
			label="View"
			bind:value={tab}
			options={[
				{ value: 'holdings', label: `Holdings · ${snap.positions.length}` },
				{ value: 'transactions', label: `Transactions · ${portfolio.transactions.length}` },
				{ value: 'return', label: 'Return' }
			]}
		/>

		{#if tab === 'holdings'}
			<HoldingsTable holdings={holdingsOf(snap.positions)} />
		{:else if tab === 'transactions'}
			<ActivityFeed
				transactions={portfolio.transactions as FeedTransaction[]}
				securities={[...data.securities, ...data.assets]}
				cashAccounts={data.cashAccounts}
				onchanged={refresh}
			/>
		{:else}
			<div class="flex flex-col gap-4">
				<div class="flex flex-wrap items-center justify-between gap-3">
					<p class="text-sm text-muted">
						What the portfolio earned, in money: sales completed, income received, costs paid, and how
						the open positions moved. The valuation above has no period; this does.
					</p>
					<Segmented
						label="Period"
						bind:value={returnPeriod}
						options={[
							{ value: 'ytd', label: 'This year' },
							{ value: 'oneYear', label: '12 months' },
							{ value: 'allTime', label: 'All time' }
						]}
					/>
				</div>
			<StatStrip
				stats={[
					{ label: 'Realized', value: formatSignedMoney(ret.realized), sub: 'completed sales, FIFO', tone: toneOf(ret.realized.amount) },
					{ label: 'Income', value: formatSignedMoney(ret.income), sub: 'dividends, coupons, rent', tone: toneOf(ret.income.amount) },
					{ label: 'Costs', value: formatSignedMoney({ amount: -ret.costs.amount, currency: ret.costs.currency }), sub: 'fees and taxes' },
					{ label: 'Value change', value: formatSignedMoney(ret.valueChange), sub: 'open positions', tone: toneOf(ret.valueChange.amount) },
					{ label: 'Total', value: formatSignedMoney(ret.total), sub: `${formatPercent(ret.timeWeightedReturn)} time-weighted`, tone: toneOf(ret.total.amount) }
				]}
			/>
			{#if years.length > 1}
				<table class="num w-full text-[13px]">
					<thead class="text-xs text-muted">
						<tr>
							<th class="py-1 pr-3 text-left font-medium">Year</th>
							<th class="py-1 pr-3 text-left font-medium">
								<span class="mr-2 inline-block size-2 rounded-sm bg-pos"></span>realized
								<span class="mr-2 ml-3 inline-block size-2 rounded-sm bg-accent"></span>income
								<span class="mr-2 ml-3 inline-block size-2 rounded-sm bg-neg"></span>costs
								<span class="mr-2 ml-3 inline-block size-2 rounded-sm bg-gold"></span>value change
							</th>
							<th class="py-1 text-right font-medium">Total</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-line">
						{#each years as y (y.from)}
							<tr>
								<td class="py-1.5 pr-3 text-muted">{new Date(y.to).getUTCFullYear()}</td>
								<td class="py-1.5 pr-3">
									<div class="flex h-2.5 w-full gap-px overflow-hidden rounded-sm bg-raised" title={`realized ${formatSignedMoney(y.realized)} · income ${formatSignedMoney(y.income)} · costs ${formatMoney(y.costs)} · value change ${formatSignedMoney(y.valueChange)}`}>
										<div class="bg-pos" style="width: {(Math.abs(y.realized.amount) / scale) * 100}%"></div>
										<div class="bg-accent" style="width: {(Math.abs(y.income.amount) / scale) * 100}%"></div>
										<div class="bg-neg" style="width: {(Math.abs(y.costs.amount) / scale) * 100}%"></div>
										<div class="bg-gold" style="width: {(Math.abs(y.valueChange.amount) / scale) * 100}%"></div>
									</div>
								</td>
								<td class="py-1.5 text-right tabular-nums font-medium {y.total.amount > 0 ? 'text-pos' : y.total.amount < 0 ? 'text-neg' : ''}">{formatSignedMoney(y.total)}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			{/if}
			</div>
		{/if}
	</div>
{:else}
	<p class="text-sm text-muted">Portfolio not found.</p>
{/if}
