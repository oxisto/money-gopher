<script lang="ts">
	import Button from './ui/Button.svelte';
	import ErrorNote from './ui/ErrorNote.svelte';

	let { onuploaded }: { onuploaded: () => void } = $props();

	let fileInput = $state<HTMLInputElement | null>(null);
	let uploading = $state(false);
	let error = $state<string | null>(null);

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		const file = fileInput?.files?.[0];
		if (!file) return;

		uploading = true;
		error = null;

		const body = new FormData();
		body.append('file', file);

		try {
			const res = await fetch('/upload', { method: 'POST', body });
			if (!res.ok) {
				error = await res.text();
			} else {
				if (fileInput) fileInput.value = '';
				onuploaded();
			}
		} catch (e) {
			error = String(e);
		} finally {
			uploading = false;
		}
	}
</script>

<form onsubmit={submit} class="rounded-md border border-gray-200 bg-white p-4">
	<div class="flex items-center gap-4">
		<input
			type="file"
			bind:this={fileInput}
			accept=".csv,.pdf,text/csv,application/pdf"
			class="block text-sm text-gray-700 file:mr-4 file:rounded file:border-0 file:bg-gray-100 file:px-3 file:py-1.5 file:text-sm file:font-medium hover:file:bg-gray-200"
			required
		/>
		<Button type="submit" disabled={uploading}>
			{uploading ? 'Uploading…' : 'Upload'}
		</Button>
	</div>
	<p class="mt-2 text-xs text-gray-400">Supported: CSV (money-gopher format) and PDF bank statements.</p>
	<ErrorNote message={error} />
</form>
