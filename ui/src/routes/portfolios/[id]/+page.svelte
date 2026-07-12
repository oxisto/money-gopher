<script lang="ts">
	import PositionsTable from '$lib/components/PositionsTable.svelte';
	import SnapshotSummary from '$lib/components/SnapshotSummary.svelte';
	import TransactionForm from '$lib/components/TransactionForm.svelte';
	import TransactionTable from '$lib/components/TransactionTable.svelte';
	import Section from '$lib/components/ui/Section.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	const PortfolioDetail = $derived(data.PortfolioDetail);

	function refresh() {
		PortfolioDetail.fetch({ policy: 'NetworkOnly' });
	}
</script>

<!-- Only the initial load blanks the page; refetches keep the old data
     (and the components' state) on screen. -->
{#if $PortfolioDetail.fetching && !$PortfolioDetail.data}
	<p class="text-sm text-gray-500">Loading…</p>
{:else if $PortfolioDetail.errors?.length}
	<p class="text-sm text-red-600">Failed to load: {$PortfolioDetail.errors[0].message}</p>
{:else if $PortfolioDetail.data}
	{@const portfolio = $PortfolioDetail.data.portfolio}
	{#if portfolio}
		<div class="mb-6">
			<h1 class="text-2xl font-semibold">{portfolio.displayName}</h1>
			<p class="mt-1 text-sm text-gray-500">
				Settlement account: <span class="font-medium">{portfolio.cashAccount.displayName}</span>
			</p>
		</div>

		<div class="flex flex-col gap-8">
			<SnapshotSummary snapshot={portfolio.snapshot} />

			<Section title="Positions">
				<PositionsTable positions={portfolio.snapshot.positions} />
			</Section>

			<Section title="Transactions">
				<TransactionTable
					transactions={portfolio.transactions}
					securities={$PortfolioDetail.data.securities}
					cashAccounts={$PortfolioDetail.data.cashAccounts}
					onchanged={refresh}
				/>
			</Section>

			<Section title="Add transaction">
				<TransactionForm
					portfolioID={portfolio.id}
					defaultCashAccountID={portfolio.cashAccount.id}
					securities={$PortfolioDetail.data.securities}
					cashAccounts={$PortfolioDetail.data.cashAccounts}
					oncreated={refresh}
				/>
			</Section>
		</div>
	{:else}
		<p class="text-sm text-gray-500">Portfolio not found.</p>
	{/if}
{/if}
