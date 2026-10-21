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
	<p class="text-sm text-muted">No persons yet.</p>
{:else}
	<ul class="divide-y divide-line rounded-md border border-line bg-surface">
		{#each persons as person (person.id)}
			<PersonRow {person} {allUsers} />
		{/each}
	</ul>
{/if}
