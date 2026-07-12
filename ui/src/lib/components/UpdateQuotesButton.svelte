<script lang="ts">
	import { graphql } from '$houdini';
	import Button from './ui/Button.svelte';
	import ErrorNote from './ui/ErrorNote.svelte';

	let { onupdated }: { onupdated: () => void } = $props();

	const triggerQuoteUpdate = graphql(`
		mutation TriggerQuoteUpdate {
			triggerQuoteUpdate {
				updatedListings
				errors
			}
		}
	`);

	let busy = $state(false);
	let message = $state<string | null>(null);
	let error = $state<string | null>(null);

	async function update() {
		busy = true;
		message = null;
		error = null;

		const result = await triggerQuoteUpdate.mutate(null);
		busy = false;

		if (result.errors?.length) {
			error = result.errors[0].message;
			return;
		}

		const data = result.data?.triggerQuoteUpdate;
		if (data) {
			message = `Updated ${data.updatedListings} listing${data.updatedListings === 1 ? '' : 's'}.`;
			error = data.errors.length ? data.errors.join('; ') : null;
			if (data.updatedListings > 0) {
				onupdated();
			}
		}
	}
</script>

<div class="flex items-center gap-3">
	<Button type="button" disabled={busy} onclick={update}>
		{busy ? 'Updating…' : 'Update quotes'}
	</Button>
	{#if message}
		<span class="text-sm text-gray-500">{message}</span>
	{/if}
</div>
<ErrorNote message={error} />
