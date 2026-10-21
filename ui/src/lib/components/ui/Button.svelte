<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		type = 'button',
		variant = 'primary',
		disabled = false,
		onclick,
		children
	}: {
		type?: 'button' | 'submit';
		variant?: 'primary' | 'secondary' | 'danger';
		disabled?: boolean;
		onclick?: () => void;
		children: Snippet;
	} = $props();

	const variants = {
		primary: 'border-transparent bg-accent text-on-accent hover:bg-accent-hover',
		secondary: 'border-line bg-surface text-fg shadow-xs hover:bg-raised',
		danger: 'border-line bg-surface text-neg shadow-xs hover:bg-neg-soft'
	};
</script>

<button
	{type}
	{disabled}
	{onclick}
	class="inline-flex h-8.5 items-center justify-center gap-1.5 rounded-lg border px-3 text-sm font-medium disabled:opacity-50 {variants[
		variant
	]}"
>
	{@render children()}
</button>
