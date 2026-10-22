<script lang="ts">
	// An input beside its label rather than inside it: Safari does not
	// reliably toggle a custom-styled checkbox wrapped in a flex label
	// within a modal dialog.
	let { checked = $bindable(false), label }: { checked?: boolean; label: string } = $props();
	const id = $props.id();
</script>

<div class="flex items-center gap-2 text-sm">
	<input {id} type="checkbox" bind:checked class="cursor-pointer" />
	<label for={id} class="cursor-pointer select-none">{label}</label>
</div>
