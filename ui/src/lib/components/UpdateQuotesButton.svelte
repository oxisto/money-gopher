<script lang="ts">
	import { graphql, mutate } from '#lib/graphql/client.js';
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

		const result = await mutate(triggerQuoteUpdate, {});
		busy = false;

		if (result.error) {
			error = result.error;
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
	{#if message}
		<span class="text-[14px] text-muted">{message}</span>
	{/if}
	<Button variant="secondary" disabled={busy} onclick={update}>
		{busy ? 'Updating…' : 'Update quotes'}
	</Button>
</div>
<ErrorNote message={error} />
