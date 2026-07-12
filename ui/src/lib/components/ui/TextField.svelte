<script lang="ts">
	let {
		value = $bindable(''),
		label,
		type = 'text',
		placeholder = '',
		required = false,
		maxlength,
		step
	}: {
		value?: string;
		label?: string;
		type?: 'text' | 'number' | 'datetime-local';
		placeholder?: string;
		required?: boolean;
		maxlength?: number;
		step?: string;
	} = $props();
</script>

<label class="block text-sm">
	{#if label}
		<span class="mb-1 block font-medium text-gray-700">{label}</span>
	{/if}
	<input
		{type}
		{placeholder}
		{required}
		{maxlength}
		{step}
		bind:value
		class="w-full rounded-md border border-gray-300 bg-white px-3 py-1.5 text-sm"
	/>
</label>
