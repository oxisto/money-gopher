<script lang="ts">
	import type { Snippet } from 'svelte';
	import Icon from './Icon.svelte';

	let {
		open = $bindable(false),
		title,
		wide = false,
		children
	}: { open?: boolean; title: string; wide?: boolean; children: Snippet } = $props();

	let dialog: HTMLDialogElement;

	// The native <dialog> brings focus trapping, Escape and the backdrop for
	// free; this only keeps it in sync with `open`.
	$effect(() => {
		if (open && !dialog.open) {
			dialog.showModal();
		} else if (!open && dialog.open) {
			dialog.close();
		}
	});
</script>

<dialog
	bind:this={dialog}
	onclose={() => (open = false)}
	onclick={(e) => e.target === dialog && (open = false)}
	class="m-auto w-full {wide ? 'max-w-2xl' : 'max-w-md'} rounded-xl border border-line bg-surface p-0 text-fg shadow-2xl backdrop:bg-black/40"
>
	<div class="flex items-center justify-between border-b border-line px-5 py-3.5">
		<h2 class="font-semibold">{title}</h2>
		<button
			type="button"
			onclick={() => (open = false)}
			aria-label="Close"
			class="grid size-8 place-items-center rounded-lg text-muted hover:bg-hover hover:text-fg"
		>
			<Icon name="close" />
		</button>
	</div>
	<div class="px-5 py-4">
		{@render children()}
	</div>
</dialog>
