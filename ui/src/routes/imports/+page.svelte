<script lang="ts">
	import DropZone from '$lib/components/imports/DropZone.svelte';
	import ImportInbox from '$lib/components/imports/ImportInbox.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	const Imports = $derived(data.Imports);

	function refresh() {
		Imports.fetch({ policy: 'NetworkOnly' });
	}
</script>

<div class="mb-6">
	<h1 class="text-lg font-semibold">Import</h1>
	<p class="text-[13px] text-muted">
		Bank statements become transactions here. Documents whose amounts check out can be booked in one click; anything
		else waits for your review.
	</p>
</div>

{#if $Imports.fetching && !$Imports.data}
	<p class="text-sm text-muted">Loading…</p>
{:else if $Imports.errors?.length}
	<p class="text-sm text-neg">Failed to load: {$Imports.errors[0].message}</p>
{:else if $Imports.data}
	<div class="flex flex-col gap-6">
		<DropZone onuploaded={refresh} />
		<ImportInbox
			documents={$Imports.data.documents}
			portfolios={$Imports.data.portfolios}
			cashAccounts={$Imports.data.cashAccounts}
			securities={$Imports.data.securities}
			onchanged={refresh}
		/>
	</div>
{/if}
