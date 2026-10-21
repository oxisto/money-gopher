<script lang="ts" generics="T extends string">
	let {
		options,
		value = $bindable(),
		label,
		onchange
	}: {
		options: { value: T; label: string }[];
		value: T;
		label: string;
		onchange?: (value: T) => void;
	} = $props();
</script>

<div role="radiogroup" aria-label={label} class="inline-flex gap-0.5 rounded-lg border border-line bg-raised p-0.5">
	{#each options as option (option.value)}
		<button
			type="button"
			role="radio"
			aria-checked={value === option.value}
			onclick={() => {
				value = option.value;
				onchange?.(option.value);
			}}
			class="h-6.5 rounded-md px-2.5 text-xs font-medium
				{value === option.value ? 'bg-surface text-fg shadow-xs ring-1 ring-line' : 'text-muted hover:text-fg'}"
		>
			{option.label}
		</button>
	{/each}
</div>
