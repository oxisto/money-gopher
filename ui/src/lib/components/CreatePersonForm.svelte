<script lang="ts">
	import { refreshAll } from '$app/navigation';
	import { graphql, mutate } from '#lib/graphql/client.js';
	import Button from './ui/Button.svelte';
	import ErrorNote from './ui/ErrorNote.svelte';
	import TextField from './ui/TextField.svelte';

	let { oncreated }: { oncreated?: () => void } = $props();

	const createPerson = graphql(`
		mutation CreatePerson($displayName: String!) {
			createPerson(displayName: $displayName) {
				id
			}
		}
	`);

	let name = $state('');
	let error = $state<string | null>(null);

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		if (!name.trim()) return;

		const result = await mutate(createPerson, { displayName: name.trim() });
		error = result.error;
		if (!error) {
			name = '';
			// Reloads this page's list and the sidebar's person switcher.
			await refreshAll();
			oncreated?.();
		}
	}
</script>

<form onsubmit={submit} class="flex flex-col gap-3">
	<p class="text-xs text-muted">
		A person is whose money it is — you, a partner, a child, or a shared household. Accounts belong to a person.
	</p>
	<TextField bind:value={name} label="Name" placeholder="e.g. Jane Doe" required />
	<ErrorNote message={error} />
	<div class="flex justify-end"><Button type="submit">Create person</Button></div>
</form>
