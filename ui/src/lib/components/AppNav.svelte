<script lang="ts">
	import { page } from '$app/state';
	import { graphql } from '$houdini';
	import { invalidateAll } from '$app/navigation';

	interface Person {
		id: string;
		displayName: string;
	}

	interface Props {
		activePerson: Person | null;
		persons: Person[];
	}

	let { activePerson, persons }: Props = $props();

	const nav = [
		{ href: '/', label: 'Dashboard' },
		{ href: '/securities', label: 'Securities' },
		{ href: '/imports', label: 'Imports' },
		{ href: '/settings/people', label: 'People' }
	];

	function isActive(href: string): boolean {
		return href === '/' ? page.url.pathname === '/' : page.url.pathname.startsWith(href);
	}

	const switchPersonMutation = graphql(`
		mutation SwitchPerson($personID: ID!) {
			switchPerson(personID: $personID) {
				id
				displayName
			}
		}
	`);

	let open = $state(false);

	async function switchTo(id: string) {
		open = false;
		await switchPersonMutation.mutate({ personID: id });
		await invalidateAll();
	}
</script>

<header class="border-b border-gray-200 bg-white">
	<div class="mx-auto flex max-w-5xl items-center gap-6 px-4 py-3">
		<a href="/" class="flex items-center gap-2 font-semibold">
			<img src="/gopher.png" alt="Money Gopher" class="h-8 w-8" />
			Money Gopher
		</a>
		<nav class="flex flex-1 gap-1">
			{#each nav as { href, label } (href)}
				<a
					{href}
					class="rounded-md px-3 py-1.5 text-sm font-medium {isActive(href)
						? 'bg-gray-100 text-gray-900'
						: 'text-gray-500 hover:text-gray-900'}"
				>
					{label}
				</a>
			{/each}
		</nav>

		{#if persons.length > 1}
			<div class="relative">
				<button
					onclick={() => (open = !open)}
					class="flex items-center gap-1 rounded-md px-2 py-1 text-sm text-gray-600 hover:bg-gray-100"
				>
					<span>{activePerson?.displayName ?? 'Select person'}</span>
					<svg class="h-4 w-4 text-gray-400" viewBox="0 0 20 20" fill="currentColor">
						<path fill-rule="evenodd" d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z" clip-rule="evenodd" />
					</svg>
				</button>
				{#if open}
					<div class="absolute right-0 z-10 mt-1 w-48 rounded-md border border-gray-200 bg-white py-1 shadow-lg">
						{#each persons as p (p.id)}
							<button
								onclick={() => switchTo(p.id)}
								class="flex w-full items-center gap-2 px-4 py-2 text-left text-sm hover:bg-gray-50
									{p.id === activePerson?.id ? 'font-medium text-gray-900' : 'text-gray-700'}"
							>
								{p.displayName}
								{#if p.id === activePerson?.id}
									<svg class="ml-auto h-4 w-4 text-indigo-600" viewBox="0 0 20 20" fill="currentColor">
										<path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
									</svg>
								{/if}
							</button>
						{/each}
					</div>
				{/if}
			</div>
		{:else if activePerson}
			<span class="text-sm text-gray-500">{activePerson.displayName}</span>
		{/if}

		<a
			href="/auth/logout"
			class="text-xs text-gray-400 hover:text-gray-700"
			title="Sign out"
		>Sign out</a>
	</div>
</header>
