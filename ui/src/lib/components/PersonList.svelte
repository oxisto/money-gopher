<script lang="ts">
	import type { People$result } from '$houdini';
	import PersonRow from './PersonRow.svelte';

	let {
		persons,
		allUsers
	}: {
		persons: People$result['persons'];
		allUsers: People$result['users'];
	} = $props();
</script>

{#if persons.length === 0}
	<p class="text-sm text-gray-500">No persons yet.</p>
{:else}
	<ul class="divide-y divide-gray-200 rounded-md border border-gray-200 bg-white">
		{#each persons as person (person.id)}
			<PersonRow {person} {allUsers} />
		{/each}
	</ul>
{/if}
