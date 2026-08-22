<script lang="ts">
	import { graphql, type People$result } from '$houdini';
	import Button from './ui/Button.svelte';
	import ErrorNote from './ui/ErrorNote.svelte';
	import SelectField from './ui/SelectField.svelte';
	import TextField from './ui/TextField.svelte';

	let {
		person,
		allUsers
	}: {
		person: People$result['persons'][number];
		allUsers: People$result['users'];
	} = $props();

	const updatePerson = graphql(`
		mutation UpdatePerson($id: ID!, $displayName: String!) {
			updatePerson(id: $id, displayName: $displayName) {
				id
				displayName
			}
		}
	`);

	const grantPersonAccess = graphql(`
		mutation GrantPersonAccess($personID: ID!, $userID: ID!) {
			grantPersonAccess(personID: $personID, userID: $userID) {
				id
				users {
					id
					displayName
				}
			}
		}
	`);

	const revokePersonAccess = graphql(`
		mutation RevokePersonAccess($personID: ID!, $userID: ID!) {
			revokePersonAccess(personID: $personID, userID: $userID) {
				id
				users {
					id
					displayName
				}
			}
		}
	`);

	let editing = $state(false);
	let name = $state(person.displayName);
	let error = $state<string | null>(null);
	let grantUserID = $state('');

	const availableUsers = $derived(
		allUsers.filter((u) => !person.users.some((pu) => pu.id === u.id))
	);

	async function saveName() {
		if (!name.trim()) return;
		const result = await updatePerson.mutate({ id: person.id, displayName: name.trim() });
		error = result.errors?.[0]?.message ?? null;
		if (!error) editing = false;
	}

	function cancelEdit() {
		name = person.displayName;
		editing = false;
		error = null;
	}

	async function grant() {
		if (!grantUserID) return;
		const result = await grantPersonAccess.mutate({
			personID: person.id,
			userID: grantUserID
		});
		error = result.errors?.[0]?.message ?? null;
		if (!error) grantUserID = '';
	}

	async function revoke(userID: string) {
		const result = await revokePersonAccess.mutate({ personID: person.id, userID });
		error = result.errors?.[0]?.message ?? null;
	}
</script>

<li class="px-4 py-3">
	<div class="flex items-center justify-between gap-2">
		{#if editing}
			<div class="flex flex-1 items-center gap-2">
				<TextField bind:value={name} />
				<Button type="button" onclick={saveName}>Save</Button>
				<Button type="button" onclick={cancelEdit}>Cancel</Button>
			</div>
		{:else}
			<span class="text-sm font-medium">{person.displayName}</span>
			<button
				onclick={() => (editing = true)}
				class="text-xs text-gray-400 hover:text-gray-700"
			>
				Rename
			</button>
		{/if}
	</div>

	<div class="mt-2 flex flex-wrap items-center gap-1.5">
		{#each person.users as user (user.id)}
			<span
				class="flex items-center gap-1 rounded bg-gray-100 px-1.5 py-0.5 text-xs text-gray-600"
			>
				{user.displayName}
				{#if person.users.length > 1}
					<button
						onclick={() => revoke(user.id)}
						title="Revoke access"
						class="text-gray-400 hover:text-red-600"
					>
						×
					</button>
				{/if}
			</span>
		{/each}
	</div>

	{#if availableUsers.length > 0}
		<div class="mt-2 flex items-center gap-2">
			<SelectField bind:value={grantUserID}>
				<option value="">Grant access to…</option>
				{#each availableUsers as user (user.id)}
					<option value={user.id}>{user.displayName}</option>
				{/each}
			</SelectField>
			<Button type="button" onclick={grant}>Grant</Button>
		</div>
	{/if}

	<ErrorNote message={error} />
</li>
