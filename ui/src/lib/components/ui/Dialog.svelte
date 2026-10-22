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

	// Only a click outside the dialog's box closes it. A click that targets
	// the dialog element itself can also come from its own edge, e.g. on a
	// control half hidden behind the content's scroll clip.
	function onBackdropClick(e: MouseEvent) {
		const r = dialog.getBoundingClientRect();
		if (e.clientX < r.left || e.clientX > r.right || e.clientY < r.top || e.clientY > r.bottom) close();
	}

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
	onclick={onBackdropClick}
	class="m-auto flex w-full {wide ? 'max-w-2xl' : 'max-w-md'} flex-col rounded-xl border border-line bg-surface p-0 text-fg shadow-2xl backdrop:bg-black/40 not-open:hidden"
	style="max-height: calc(100dvh - 2rem)"
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
