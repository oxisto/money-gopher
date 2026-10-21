<script lang="ts">
	import CashAccountList from '$lib/components/CashAccountList.svelte';
	import PortfolioList from '$lib/components/PortfolioList.svelte';
	import Section from '$lib/components/ui/Section.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	const Dashboard = $derived(data.Dashboard);
</script>

{#if $Dashboard.fetching && !$Dashboard.data}
	<p class="text-sm text-muted">Loading…</p>
{:else if $Dashboard.errors?.length}
	<p class="text-sm text-neg">Failed to load: {$Dashboard.errors[0].message}</p>
{:else if $Dashboard.data}
	<div class="grid gap-8 md:grid-cols-2">
		<Section title="Portfolios">
			<PortfolioList
				portfolios={$Dashboard.data.portfolios}
				cashAccounts={$Dashboard.data.cashAccounts}
			/>
		</Section>
		<Section title="Cash accounts">
			<CashAccountList accounts={$Dashboard.data.cashAccounts} />
		</Section>
	</div>
{/if}
