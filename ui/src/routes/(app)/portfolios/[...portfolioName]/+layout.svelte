<script lang="ts">
	import PortfolioBreadcrumb from '$lib/components/PortfolioBreadcrumb.svelte';
	import type { LayoutData } from './$types';

	export let data: LayoutData;
</script>

<PortfolioBreadcrumb portfolio={data.portfolio} snapshot={data.snapshot} />
<slot />
