<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		type = 'button',
		disabled = false,
		onclick,
		children
	}: {
		type?: 'button' | 'submit';
		disabled?: boolean;
		onclick?: () => void;
		children: Snippet;
	} = $props();
</script>

<button
	{type}
	{disabled}
	{onclick}
	class="rounded-md bg-blue-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
>
	{@render children()}
</button>
