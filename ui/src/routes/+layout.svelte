<script lang="ts">
	import '../app.css';
	import { page } from '$app/state';
	import favicon from '#lib/assets/favicon.svg';
	import Sidebar from '#lib/components/Sidebar.svelte';
	import type { LayoutProps } from './$types';

	let { children, data }: LayoutProps = $props();

	// Below md the sidebar is a drawer behind the top bar's menu button.
	let menuOpen = $state(false);
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<title>Money Gopher</title>
</svelte:head>

<!-- /login has no session yet, so no person, accounts or navigation. -->
{#if page.url.pathname === '/login'}
	{@render children()}
{:else}
	<div class="min-h-screen md:flex">
		<header class="flex h-14 items-center gap-3 border-b border-line px-4 md:hidden">
			<button
				type="button"
				onclick={() => (menuOpen = !menuOpen)}
				aria-expanded={menuOpen}
				aria-controls="sidebar"
				aria-label="Menu"
				class="grid size-9 place-items-center rounded-lg text-muted hover:bg-hover hover:text-fg"
			>
				<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true">
					<path d={menuOpen ? 'M6 6l12 12M18 6L6 18' : 'M4 7h16M4 12h16M4 17h16'} />
				</svg>
			</button>
			<span class="flex items-center gap-2 font-semibold">
				<img src="/gopher.png" alt="" class="h-6 w-5 object-contain" />
				Money Gopher
			</span>
		</header>

		<aside
			id="sidebar"
			class="{menuOpen
				? 'fixed inset-x-0 top-14 bottom-0 z-30 overflow-y-auto bg-canvas'
				: 'hidden'} md:sticky md:top-0 md:block md:h-screen md:w-62 md:shrink-0 md:overflow-y-auto"
		>
			<Sidebar data={data.shell} onnavigate={() => (menuOpen = false)} />
		</aside>

		<main
			class="min-w-0 flex-1 border-line bg-surface md:m-2 md:ml-0 md:min-h-[calc(100vh-1rem)] md:rounded-xl md:border"
		>
			<div class="mx-auto max-w-6xl px-4 py-5 md:px-8 md:pb-12">
				{@render children()}
			</div>
		</main>
	</div>
{/if}
