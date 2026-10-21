<script lang="ts">
	import HoldingsTable, { holdingsOf } from '$lib/components/holdings/HoldingsTable.svelte';
	import TransactionForm from '$lib/components/TransactionForm.svelte';
	import TransactionTable from '$lib/components/TransactionTable.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Dialog from '$lib/components/ui/Dialog.svelte';
	import Icon from '$lib/components/ui/Icon.svelte';
	import Segmented from '$lib/components/ui/Segmented.svelte';
	import StatStrip from '$lib/components/ui/StatStrip.svelte';
	import { formatMoney, formatPercent, formatSignedMoney, toneOf } from '$lib/format';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	const PortfolioDetail = $derived(data.PortfolioDetail);

	let tab = $state<'holdings' | 'transactions'>('holdings');
	let addOpen = $state(false);

	function refresh() {
		PortfolioDetail.fetch({ policy: 'NetworkOnly' });
	}
</script>

<!-- Only the initial load blanks the page; refetches keep the old data
     (and the components' state) on screen. -->
{#if $PortfolioDetail.fetching && !$PortfolioDetail.data}
	<p class="text-sm text-muted">Loading…</p>
{:else if $PortfolioDetail.errors?.length}
	<p class="text-sm text-neg">Failed to load: {$PortfolioDetail.errors[0].message}</p>
{:else if $PortfolioDetail.data}
	{@const portfolio = $PortfolioDetail.data.portfolio}
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
			<Button onclick={() => (addOpen = true)}><Icon name="plus" size={15} /> Add transaction</Button>
		</div>

		<Dialog bind:open={addOpen} title="Add transaction" wide>
			<TransactionForm
				portfolioID={portfolio.id}
				defaultCashAccountID={portfolio.cashAccount.id}
				securities={$PortfolioDetail.data.securities}
				cashAccounts={$PortfolioDetail.data.cashAccounts}
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
				<TransactionTable
					transactions={portfolio.transactions}
					securities={$PortfolioDetail.data.securities}
					cashAccounts={$PortfolioDetail.data.cashAccounts}
					onchanged={refresh}
				/>
			{/if}
		</div>
	{:else}
		<p class="text-sm text-muted">Portfolio not found.</p>
	{/if}
{/if}
