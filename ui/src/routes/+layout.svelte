<script lang="ts">
	import '../app.css';
	import favicon from '$lib/assets/favicon.svg';
	import AppNav from '$lib/components/AppNav.svelte';

	let { children } = $props();
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<title>Money Gopher</title>
</svelte:head>

<div class="min-h-screen bg-gray-50 text-gray-900">
	<AppNav />

	<main class="mx-auto max-w-7xl px-4 py-6">
		{@render children()}
	</main>
</div>
