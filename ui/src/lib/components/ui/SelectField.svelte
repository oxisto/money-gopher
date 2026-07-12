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
		<span class="mb-1 block font-medium text-gray-700">{label}</span>
	{/if}
	<select
		{required}
		bind:value
		class="w-full rounded-md border border-gray-300 bg-white px-3 py-1.5 text-sm"
	>
		{@render children()}
	</select>
</label>
