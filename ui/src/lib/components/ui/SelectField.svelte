<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		value = $bindable(''),
		label,
		required = false,
		children
	}: {
		value?: string;
		label?: string;
		required?: boolean;
		children: Snippet;
	} = $props();
</script>

<label class="block text-sm">
	{#if label}
		<span class="mb-1 block text-xs font-medium text-muted">{label}</span>
	{/if}
	<select
		{required}
		bind:value
		class="h-9 w-full px-3 py-1.5"
	>
		{@render children()}
	</select>
</label>
