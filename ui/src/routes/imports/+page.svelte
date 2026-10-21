<script lang="ts">
	import { refreshAll } from '$app/navigation';
	import DropZone from '#lib/components/imports/DropZone.svelte';
	import ImportInbox from '#lib/components/imports/ImportInbox.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	function refresh() {
		refreshAll();
	}
</script>

<div class="mb-6">
	<h1 class="text-lg font-semibold">Import</h1>
	<p class="text-[13px] text-muted">
		Bank statements become transactions here. Documents whose amounts check out can be booked in one click; anything
		else waits for your review.
	</p>
</div>

<div class="flex flex-col gap-6">
	<DropZone onuploaded={refresh} />
	<ImportInbox
		documents={data.documents}
		portfolios={data.portfolios}
		cashAccounts={data.cashAccounts}
		securities={data.securities}
		onchanged={refresh}
	/>
</div>
