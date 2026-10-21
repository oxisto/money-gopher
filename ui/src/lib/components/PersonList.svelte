<script lang="ts">
	import type { PeopleResult } from '$lib/graphql/documents';
	import PersonRow from './PersonRow.svelte';

	let {
		persons,
		allUsers
	}: {
		persons: PeopleResult['persons'];
		allUsers: PeopleResult['users'];
	} = $props();
</script>

{#if persons.length === 0}
	<p class="rounded-xl border border-dashed border-line px-4 py-8 text-center text-sm text-muted">No people yet.</p>
{:else}
	<ul class="divide-y divide-line rounded-xl border border-line">
		{#each persons as person (person.id)}
			<PersonRow {person} {allUsers} />
		{/each}
	</ul>
{/if}
