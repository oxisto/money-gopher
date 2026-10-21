<script lang="ts">
	import { graphql, type People$result } from '$houdini';
	import Button from './ui/Button.svelte';
	import ErrorNote from './ui/ErrorNote.svelte';
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

<li class="flex flex-wrap items-center gap-3 px-4 py-3">
	<span class="grid size-8 shrink-0 place-items-center rounded-lg bg-accent-soft text-sm font-semibold text-accent-fg">
		{person.displayName.charAt(0)}
	</span>
	<div class="min-w-0 flex-1">
		{#if editing}
			<div class="flex flex-wrap items-center gap-2">
				<div class="min-w-48 flex-1"><TextField bind:value={name} /></div>
				<Button onclick={saveName}>Save</Button>
				<Button variant="secondary" onclick={cancelEdit}>Cancel</Button>
			</div>
		{:else}
			<div class="font-medium">{person.displayName}</div>
		{/if}

		<!-- With a single login, who has access is obvious and only noise. -->
		{#if allUsers.length > 1}
			<div class="mt-1.5 flex flex-wrap items-center gap-1.5 text-xs">
				<span class="text-muted">Access:</span>
				{#each person.users as user (user.id)}
					<span class="flex items-center gap-1 rounded-full border border-line bg-raised px-2 py-0.5">
						{user.displayName}
						{#if person.users.length > 1}
							<button
								type="button"
								onclick={() => revoke(user.id)}
								aria-label="Revoke access for {user.displayName}"
								class="text-muted hover:text-neg"
							>
								×
							</button>
						{/if}
					</span>
				{/each}
				{#if availableUsers.length > 0}
					<select
						bind:value={grantUserID}
						onchange={grant}
						aria-label="Grant access"
						class="h-6.5 rounded-full py-0 pr-7 pl-2.5 text-xs"
					>
						<option value="">+ Share with…</option>
						{#each availableUsers as user (user.id)}
							<option value={user.id}>{user.displayName}</option>
						{/each}
					</select>
				{/if}
			</div>
		{/if}
		<ErrorNote message={error} />
	</div>
	{#if !editing}
		<button
			type="button"
			onclick={() => (editing = true)}
			class="rounded-md px-2 py-1 text-xs text-muted hover:bg-hover hover:text-fg"
		>
			Rename
		</button>
	{/if}
</li>
