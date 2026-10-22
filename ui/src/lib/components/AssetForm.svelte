<script lang="ts">
	import { refreshAll } from '$app/navigation';
	import { graphql, mutate } from '#lib/graphql/client.js';
	import { asKind, handValuedKinds } from '#lib/securities.js';
	import Button from './ui/Button.svelte';
	import ErrorNote from './ui/ErrorNote.svelte';
	import SelectField from './ui/SelectField.svelte';
	import TextField from './ui/TextField.svelte';

	let {
		cashAccounts,
		oncreated
	}: {
		cashAccounts: { id: string; displayName: string; currency: string }[];
		oncreated?: () => void;
	} = $props();

	const bookAsset = graphql(`
		mutation BookAsset($input: BookAssetInput!) {
			bookAsset(input: $input) {
				id
			}
		}
	`);

	const today = new Date().toISOString().slice(0, 10);

	let name = $state('');
	let kind = $state('PROPERTY');
	let currency = $state('EUR');
	let acquiredOn = $state(today);
	let cost = $state('');
	let valueToday = $state('');
	let cashAccountID = $state(cashAccounts[0]?.id ?? '');
	let error = $state<string | null>(null);
	let saving = $state(false);

	const decimal = (s: string) => s.replace(',', '.').trim();

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		const assetKind = asKind(kind);
		if (!name.trim() || !assetKind) return;
		if (currency.trim().length !== 3) {
			error = 'The currency is the three-letter code the cost and valuations are in.';
			return;
		}

		saving = true;
		const result = await mutate(bookAsset, {
			input: {
				displayName: name.trim(),
				kind: assetKind,
				currency: currency.trim().toUpperCase(),
				acquiredAt: new Date(acquiredOn + 'T12:00:00Z'),
				cost: decimal(cost),
				valueToday: valueToday.trim() ? decimal(valueToday) : undefined,
				cashAccountID: cashAccountID || undefined
			}
		});
		saving = false;
		error = result.error;
		if (error) return;

		name = '';
		cost = '';
		valueToday = '';
		// Reloads the page and the sidebar's account list, which gains the
		// "Other assets" portfolio the first time.
		await refreshAll();
		oncreated?.();
	}
</script>

<form onsubmit={submit} class="flex flex-col gap-3">
	<div class="grid gap-3 sm:grid-cols-2">
		<TextField bind:value={name} label="Name" placeholder="e.g. Office room" required />
		<SelectField bind:value={kind} label="Kind" required>
			{#each handValuedKinds as [value, label] (value)}
				<option {value}>{label}</option>
			{/each}
		</SelectField>
		<TextField bind:value={acquiredOn} label="Acquired on" type="date" required />
		<TextField bind:value={currency} label="Currency" maxlength={3} required />
		<TextField bind:value={cost} label="What it cost" placeholder="e.g. 42000" required />
		<TextField bind:value={valueToday} label="Worth today (optional)" placeholder="leave empty to use the cost" />
		<SelectField bind:value={cashAccountID} label="Settles against">
			<option value="">Choose…</option>
			{#each cashAccounts as account (account.id)}
				<option value={account.id}>{account.displayName} ({account.currency})</option>
			{/each}
		</SelectField>
	</div>
	<p class="text-xs text-muted">
		The asset is yours alone. It is booked at cost into your "Other assets" portfolio, which is
		created the first time and settles rent or a sale against the account chosen here. Later
		valuations go on the asset's own page.
	</p>

	<ErrorNote message={error} />
	<div class="mt-1 flex justify-end">
		<Button type="submit" disabled={saving || !name.trim() || !cost.trim()}>
			{saving ? 'Booking…' : 'Book asset'}
		</Button>
	</div>
</form>
