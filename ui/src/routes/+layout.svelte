<script lang="ts">
	import '../app.css';
	import favicon from '$lib/assets/favicon.svg';
	import AppNav from '$lib/components/AppNav.svelte';
	import type { LayoutProps } from './$types';

	let { children, data }: LayoutProps = $props();
	const Layout = $derived(data.Layout);
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<title>Money Gopher</title>
</svelte:head>

<div class="min-h-screen bg-gray-50 text-gray-900">
	<AppNav
		activePerson={$Layout.data?.me?.activePerson ?? null}
		persons={$Layout.data?.me?.persons ?? []}
	/>

	<main class="mx-auto max-w-7xl px-4 py-6">
		{@render children()}
	</main>
</div>
