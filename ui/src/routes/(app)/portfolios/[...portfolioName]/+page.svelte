<script lang="ts">
	import PortfolioPositionsTable from '$lib/components/PortfolioPositionsTable.svelte';
	import type { PageData } from './$types';

	export let data: PageData;
</script>

<PortfolioPositionsTable snapshot={data.snapshot} />

<a href="/portfolios/{data.portfolio.name}/transactions/">Show transactions list</a>

<a href="/portfolios/{data.portfolio.name}/transactions/add">Add transaction</a>
