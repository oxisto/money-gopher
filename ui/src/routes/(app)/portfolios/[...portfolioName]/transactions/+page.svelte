<script lang="ts">
	import PortfolioTransactionsTable from '$lib/components/PortfolioTransactionsTable.svelte';
	import type { PageData } from './$types';

	export let data: PageData;
</script>

<PortfolioTransactionsTable transactions={data.transactions} securities={data.securities} />
