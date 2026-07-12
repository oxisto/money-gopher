<script lang="ts">
	import UploadForm from '$lib/components/UploadForm.svelte';
	import DocumentList from '$lib/components/DocumentList.svelte';
	import Section from '$lib/components/ui/Section.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	const Imports = $derived(data.Imports);

	function refresh() {
		Imports.fetch({ policy: 'NetworkOnly' });
	}
</script>

{#if $Imports.fetching && !$Imports.data}
	<p class="text-sm text-gray-500">Loading…</p>
{:else if $Imports.errors?.length}
	<p class="text-sm text-red-600">Failed to load: {$Imports.errors[0].message}</p>
{:else if $Imports.data}
	<div class="flex flex-col gap-8">
		<Section title="Upload document">
			<UploadForm onuploaded={refresh} />
		</Section>

		<Section title="Documents">
			<DocumentList
				documents={$Imports.data.documents}
				portfolios={$Imports.data.portfolios}
				cashAccounts={$Imports.data.cashAccounts}
				securities={$Imports.data.securities}
				onchanged={refresh}
			/>
		</Section>
	</div>
{/if}
