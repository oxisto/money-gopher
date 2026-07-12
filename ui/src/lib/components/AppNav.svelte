<script lang="ts">
	import { page } from '$app/state';

	const nav = [
		{ href: '/', label: 'Dashboard' },
		{ href: '/securities', label: 'Securities' },
		{ href: '/imports', label: 'Imports' }
	];

	function isActive(href: string): boolean {
		return href === '/' ? page.url.pathname === '/' : page.url.pathname.startsWith(href);
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
		<a
			href="/auth/logout"
			class="text-xs text-gray-400 hover:text-gray-700"
			title="Sign out"
		>Sign out</a>
	</div>
</header>
