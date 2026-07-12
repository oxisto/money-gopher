<script lang="ts">
	import SecurityForm from '$lib/components/SecurityForm.svelte';
	import SecurityList from '$lib/components/SecurityList.svelte';
	import UpdateQuotesButton from '$lib/components/UpdateQuotesButton.svelte';
	import Section from '$lib/components/ui/Section.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	const Securities = $derived(data.Securities);

	function refresh() {
		Securities.fetch({ policy: 'NetworkOnly' });
	}
</script>

<!-- Only the initial load blanks the page; refetches keep the old data
     (and the components' state) on screen. -->
{#if $Securities.fetching && !$Securities.data}
	<p class="text-sm text-gray-500">Loading…</p>
{:else if $Securities.errors?.length}
	<p class="text-sm text-red-600">Failed to load: {$Securities.errors[0].message}</p>
{:else if $Securities.data}
	<div class="flex flex-col gap-8">
		<Section title="Securities">
			<SecurityList securities={$Securities.data.securities} />
			<div class="mt-4">
				<UpdateQuotesButton onupdated={refresh} />
			</div>
		</Section>

		<Section title="Add security">
			<SecurityForm />
		</Section>
	</div>
{/if}
