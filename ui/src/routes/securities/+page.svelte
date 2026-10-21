<script lang="ts">
	import SecurityForm from '$lib/components/SecurityForm.svelte';
	import SecurityList from '$lib/components/SecurityList.svelte';
	import UpdateQuotesButton from '$lib/components/UpdateQuotesButton.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Dialog from '$lib/components/ui/Dialog.svelte';
	import Icon from '$lib/components/ui/Icon.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	const Securities = $derived(data.Securities);

	let addOpen = $state(false);

	function refresh() {
		Securities.fetch({ policy: 'NetworkOnly' });
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

<!-- Only the initial load blanks the page; refetches keep the old data
     (and the components' state) on screen. -->
{#if $Securities.fetching && !$Securities.data}
	<p class="text-sm text-muted">Loading…</p>
{:else if $Securities.errors?.length}
	<p class="text-sm text-neg">Failed to load: {$Securities.errors[0].message}</p>
{:else if $Securities.data}
	<SecurityList securities={$Securities.data.securities} />
{/if}
