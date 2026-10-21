<script lang="ts">
	import ErrorNote from '#lib/components/ui/ErrorNote.svelte';
	import Icon from '#lib/components/ui/Icon.svelte';

	let { onuploaded }: { onuploaded: () => void } = $props();

	let dragging = $state(false);
	let progress = $state<{ done: number; total: number } | null>(null);
	let error = $state<string | null>(null);

	async function uploadOne(file: File): Promise<string | null> {
		const body = new FormData();
		body.append('file', file);
		try {
			const res = await fetch('/upload', { method: 'POST', body });
			return res.ok ? null : `${file.name}: ${await res.text()}`;
		} catch (e) {
			return `${file.name}: ${String(e)}`;
		}
	}

	// Choosing or dropping files uploads them right away; there is nothing
	// else to fill in, so a separate Upload button would only be a detour.
	async function upload(files: FileList | File[]) {
		const list = [...files].filter((f) => /\.(pdf|csv)$/i.test(f.name));
		if (list.length === 0) {
			error = 'Only PDF and CSV files can be imported.';
			return;
		}
		error = null;
		progress = { done: 0, total: list.length };

		// Sequential, not Promise.all: the server processes each upload
		// against shared state (security/document detection), and one
		// failure shouldn't be hidden among concurrent requests.
		const failures: string[] = [];
		for (const file of list) {
			const failure = await uploadOne(file);
			if (failure) failures.push(failure);
			progress = { done: progress.done + 1, total: list.length };
		}

		error = failures.length ? failures.join('; ') : null;
		progress = null;
		onuploaded();
	}

	function drop(event: DragEvent) {
		event.preventDefault();
		dragging = false;
		if (event.dataTransfer?.files.length) upload(event.dataTransfer.files);
	}
</script>

<label
	ondragover={(e) => {
		e.preventDefault();
		dragging = true;
	}}
	ondragleave={() => (dragging = false)}
	ondrop={drop}
	class="flex cursor-pointer flex-wrap items-center gap-4 rounded-xl border-[1.5px] border-dashed px-5 py-4 transition-colors
		{dragging ? 'border-accent bg-accent-soft' : 'border-line-strong bg-raised hover:border-accent'}"
>
	<span class="grid size-10 shrink-0 place-items-center rounded-lg bg-accent-soft text-accent-fg">
		<Icon name="upload" size={19} />
	</span>
	<span class="min-w-56 flex-1">
		{#if progress}
			<span class="block font-medium">Uploading {Math.min(progress.done + 1, progress.total)} of {progress.total}…</span>
			<span class="mt-1.5 block h-1 overflow-hidden rounded-full bg-line">
				<span class="block h-full bg-accent transition-all" style="width: {(progress.done / progress.total) * 100}%"></span>
			</span>
		{:else}
			<span class="block font-medium">
				Drop PDFs here, or <span class="text-accent-fg underline">browse</span>
			</span>
			<span class="block text-xs text-muted">As many as you like. Nothing is booked until you confirm it.</span>
		{/if}
	</span>
	<span class="flex flex-wrap gap-1.5 text-xs">
		<span class="rounded-full border border-line bg-surface px-2 py-0.5">ING securities statements</span>
		<span class="rounded-full border border-line bg-surface px-2 py-0.5">ING Kontoauszug</span>
		<span class="rounded-full border border-line bg-surface px-2 py-0.5">CSV</span>
	</span>
	<input
		type="file"
		multiple
		accept=".csv,.pdf,text/csv,application/pdf"
		class="sr-only"
		disabled={progress !== null}
		onchange={(e) => {
			const input = e.currentTarget;
			if (input.files?.length) upload(input.files).then(() => (input.value = ''));
		}}
	/>
</label>
<ErrorNote message={error} />
