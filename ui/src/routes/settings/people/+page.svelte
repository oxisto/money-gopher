<script lang="ts">
	import CreatePersonForm from '$lib/components/CreatePersonForm.svelte';
	import PersonList from '$lib/components/PersonList.svelte';
	import UserList from '$lib/components/UserList.svelte';
	import Section from '$lib/components/ui/Section.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	const People = $derived(data.People);
</script>

{#if $People.fetching && !$People.data}
	<p class="text-sm text-gray-500">Loading…</p>
{:else if $People.errors?.length}
	<p class="text-sm text-red-600">Failed to load: {$People.errors[0].message}</p>
{:else if $People.data}
	<div class="flex flex-col gap-8">
		<Section title="People">
			<PersonList persons={$People.data.persons} allUsers={$People.data.users} />
		</Section>

		<Section title="Add person">
			<CreatePersonForm />
		</Section>

		<Section title="Users">
			<UserList users={$People.data.users} />
		</Section>
	</div>
{/if}
