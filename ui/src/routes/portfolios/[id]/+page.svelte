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
	import StatStrip from '#lib/components/ui/StatStrip.svelte';
	import TextField from '#lib/components/ui/TextField.svelte';
	import { formatMoney, formatPercent, formatSignedMoney, toneOf } from '#lib/format.js';
	import { graphql, mutate } from '#lib/graphql/client.js';
	import type { FeedTransaction } from '#lib/transaction-types.js';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let tab = $state<'holdings' | 'transactions'>('holdings');
	let addOpen = $state(false);

	const portfolio = $derived(data.portfolio);

	const updatePortfolio = graphql(`
		mutation RenamePortfolio($id: ID!, $displayName: String!) {
			updatePortfolio(id: $id, input: { displayName: $displayName }) {
				id
				displayName
			}
		}
	`);
	let renameOpen = $state(false);
	let newName = $state('');
	let renameError = $state<string | null>(null);

	function openRename() {
		newName = portfolio?.displayName ?? '';
		renameError = null;
		renameOpen = true;
	}

	async function saveName(event: SubmitEvent) {
		event.preventDefault();
		if (!portfolio || !newName.trim()) return;
		const result = await mutate(updatePortfolio, { id: portfolio.id, displayName: newName.trim() });
		renameError = result.error;
		if (!renameError) {
			renameOpen = false;
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
		<Button variant="secondary" onclick={openRename}>Rename</Button>
		<Button onclick={() => (addOpen = true)}><Icon name="plus" size={15} /> Add transaction</Button>
	</div>

	<Dialog bind:open={renameOpen} title="Rename portfolio">
		<form onsubmit={saveName} class="flex flex-col gap-3">
			<TextField bind:value={newName} label="Name" required />
			<ErrorNote message={renameError} />
			<div class="mt-2 flex justify-end gap-2">
				<Button variant="secondary" onclick={() => (renameOpen = false)}>Cancel</Button>
				<Button type="submit" disabled={!newName.trim()}>Save</Button>
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
				{ value: 'transactions', label: `Transactions · ${portfolio.transactions.length}` }
			]}
		/>

		{#if tab === 'holdings'}
			<HoldingsTable holdings={holdingsOf(snap.positions)} />
		{:else}
			<ActivityFeed
				transactions={portfolio.transactions as FeedTransaction[]}
				securities={[...data.securities, ...data.assets]}
				cashAccounts={data.cashAccounts}
				onchanged={refresh}
			/>
		{/if}
	</div>
{:else}
	<p class="text-sm text-muted">Portfolio not found.</p>
{/if}
