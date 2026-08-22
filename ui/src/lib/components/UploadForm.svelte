<script lang="ts">
	import Button from './ui/Button.svelte';
	import ErrorNote from './ui/ErrorNote.svelte';

	let { onuploaded }: { onuploaded: () => void } = $props();

	let fileInput = $state<HTMLInputElement | null>(null);
	let uploading = $state(false);
	let progress = $state<{ done: number; total: number } | null>(null);
	let error = $state<string | null>(null);

	async function uploadOne(file: File): Promise<string | null> {
		const body = new FormData();
		body.append('file', file);

		try {
			const res = await fetch('/upload', { method: 'POST', body });
			if (!res.ok) return `${file.name}: ${await res.text()}`;
			return null;
		} catch (e) {
			return `${file.name}: ${String(e)}`;
		}
	}

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		const files = fileInput?.files;
		if (!files || files.length === 0) return;

		uploading = true;
		error = null;
		const total = files.length;
		progress = { done: 0, total };

		// Sequential, not Promise.all: the server processes each upload
		// against shared state (security/document detection), and one
		// failure shouldn't be hidden among concurrent requests.
		const failures: string[] = [];
		let done = 0;
		for (const file of files) {
			const failure = await uploadOne(file);
			if (failure) failures.push(failure);
			done += 1;
			progress = { done, total };
		}

		if (failures.length > 0) {
			error = failures.join('; ');
		}
		if (fileInput) fileInput.value = '';
		uploading = false;
		progress = null;
		onuploaded();
	}
</script>

<form onsubmit={submit} class="rounded-md border border-gray-200 bg-white p-4">
	<div class="flex items-center gap-4">
		<input
			type="file"
			multiple
			bind:this={fileInput}
			accept=".csv,.pdf,text/csv,application/pdf"
			class="block text-sm text-gray-700 file:mr-4 file:rounded file:border-0 file:bg-gray-100 file:px-3 file:py-1.5 file:text-sm file:font-medium hover:file:bg-gray-200"
			required
		/>
		<Button type="submit" disabled={uploading}>
			{#if uploading && progress}
				Uploading {progress.done}/{progress.total}…
			{:else}
				Upload
			{/if}
		</Button>
	</div>
	<p class="mt-2 text-xs text-gray-400">
		Supported: CSV (money-gopher format) and PDF bank statements. Select multiple files to upload
		them all at once.
	</p>
	<ErrorNote message={error} />
</form>
