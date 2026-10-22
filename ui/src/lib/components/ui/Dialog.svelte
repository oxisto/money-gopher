<script lang="ts">
	import type { Snippet } from 'svelte';
	import Icon from './Icon.svelte';

	let {
		open = $bindable(false),
		title,
		wide = false,
		onclose,
		children
	}: { open?: boolean; title: string; wide?: boolean; onclose?: () => void; children: Snippet } = $props();

	function close() {
		open = false;
		onclose?.();
	}

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
	onclose={close}
	onclick={(e) => e.target === dialog && close()}
	class="m-auto flex w-full {wide ? 'max-w-2xl' : 'max-w-md'} max-h-[calc(100dvh-2rem)] flex-col rounded-xl border border-line bg-surface p-0 text-fg shadow-2xl backdrop:bg-black/40 not-open:hidden"
>
	<div class="flex shrink-0 items-center justify-between border-b border-line px-5 py-3.5">
		<h2 class="font-semibold">{title}</h2>
		<button
			type="button"
			onclick={close}
			aria-label="Close"
			class="grid size-8 place-items-center rounded-lg text-muted hover:bg-hover hover:text-fg"
		>
			<Icon name="close" />
		</button>
	</div>
	<!-- A form taller than the window scrolls here instead of overflowing
	     the viewport, where its bottom could not be reached. -->
	<div class="min-h-0 overflow-y-auto px-5 py-4">
		{@render children()}
	</div>
</dialog>
