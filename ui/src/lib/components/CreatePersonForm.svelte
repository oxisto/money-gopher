<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { cache, graphql } from '$houdini';
	import Button from './ui/Button.svelte';
	import ErrorNote from './ui/ErrorNote.svelte';
	import TextField from './ui/TextField.svelte';

	const createPerson = graphql(`
		mutation CreatePerson($displayName: String!) {
			createPerson(displayName: $displayName) {
				...All_Persons_insert
			}
		}
	`);

	let name = $state('');
	let error = $state<string | null>(null);

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		if (!name.trim()) return;

		const result = await createPerson.mutate({ displayName: name.trim() });
		error = result.errors?.[0]?.message ?? null;
		if (!error) {
			name = '';
			// The new person needs to show up in the sidebar's switcher too, which
			// reads from the root layout's `me.persons` — a separate query
			// from this page's `All_Persons` list, so the ...All_Persons_insert
			// fragment above doesn't reach it. Same fix as switchPerson's.
			cache.markStale();
			await invalidateAll();
		}
	}
</script>

<form onsubmit={submit} class="flex items-end gap-2 rounded-md border border-line bg-surface p-4">
	<TextField bind:value={name} label="Name" placeholder="e.g. Jane Doe" required />
	<Button type="submit">Create person</Button>
</form>
<ErrorNote message={error} />
