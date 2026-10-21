<script lang="ts">
	import { refreshAll } from '$app/navigation';
	import SecurityForm from '#lib/components/SecurityForm.svelte';
	import SecurityList from '#lib/components/SecurityList.svelte';
	import UpdateQuotesButton from '#lib/components/UpdateQuotesButton.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import Dialog from '#lib/components/ui/Dialog.svelte';
	import Icon from '#lib/components/ui/Icon.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let addOpen = $state(false);

	function refresh() {
		refreshAll();
	}
</script>

<div class="mb-6 flex flex-wrap items-center gap-3">
	<div class="flex-1">
		<h1 class="text-lg font-semibold">Securities</h1>
		<p class="text-[13px] text-muted">Shared by everyone on this server, with where their quotes come from.</p>
	</div>
	<UpdateQuotesButton onupdated={refresh} />
	<Button onclick={() => (addOpen = true)}><Icon name="plus" size={15} /> Add security</Button>
</div>

<Dialog bind:open={addOpen} title="Add security" wide>
	<SecurityForm oncreated={() => (addOpen = false)} />
</Dialog>

<SecurityList securities={data.securities} />
