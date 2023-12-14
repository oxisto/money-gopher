<script lang="ts">
	import PortfolioBreadcrumb from '$lib/components/PortfolioBreadcrumb.svelte';
	import PortfolioPositionsTable from '$lib/components/PortfolioPositionsTable.svelte';
	import type { PageData } from './$types';

	export let data: PageData;
</script>

<PortfolioBreadcrumb portfolio={data.portfolio} snapshot={data.snapshot} />
<PortfolioPositionsTable snapshot={data.snapshot} />

<a href="/portfolios/{data.portfolio.name}/transactions/add">Add transaction</a>
