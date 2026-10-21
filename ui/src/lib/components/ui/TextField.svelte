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
		<span class="mb-1 block text-xs font-medium text-muted">{label}</span>
	{/if}
	<input
		{type}
		{placeholder}
		{required}
		{maxlength}
		{step}
		bind:value
		class="h-9 w-full px-3 py-1.5"
	/>
</label>
