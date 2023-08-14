<script lang="ts">
	import PortfolioBreadcrumb from '$lib/components/PortfolioBreadcrumb.svelte';
	import PortfolioPositionsTable from '$lib/components/PortfolioPositionsTable.svelte';
	import type { PageData } from './$types';

	export let data: PageData;
</script>

Hey from there
<PortfolioBreadcrumb portfolio={data.portfolio} snapshot={data.snapshot} />
<PortfolioPositionsTable snapshot={data.snapshot} />
