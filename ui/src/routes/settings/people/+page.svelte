<script lang="ts">
	import CreatePersonForm from '$lib/components/CreatePersonForm.svelte';
	import PersonList from '$lib/components/PersonList.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Dialog from '$lib/components/ui/Dialog.svelte';
	import Icon from '$lib/components/ui/Icon.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let addOpen = $state(false);
</script>

<div class="mb-6 flex flex-wrap items-center gap-3">
	<div class="flex-1">
		<h1 class="text-lg font-semibold">People &amp; access</h1>
		<p class="text-[13px] text-muted">
			Each person has their own accounts. Switch between them at the top of the sidebar.
		</p>
	</div>
	<Button onclick={() => (addOpen = true)}><Icon name="plus" size={15} /> Add person</Button>
</div>

<Dialog bind:open={addOpen} title="Add person">
	<CreatePersonForm oncreated={() => (addOpen = false)} />
</Dialog>

<div class="flex flex-col gap-6">
	<PersonList persons={data.persons} allUsers={data.users} />

	<section class="flex flex-col gap-2">
		<h2 class="text-sm font-semibold">Logins</h2>
		<p class="text-[13px] text-muted">
			<span class="text-fg">{data.users.map((u) => u.displayName).join(', ')}</span>
			{#if data.users.length === 1}
				— the only login, so it sees every person above. Other logins show up here after their first sign-in, and
				each person can then be shared with them.
			{/if}
		</p>
	</section>
</div>
