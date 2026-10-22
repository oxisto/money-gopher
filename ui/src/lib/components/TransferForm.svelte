<script lang="ts">
	import { graphql, mutate } from '#lib/graphql/client.js';
	import { parseAmount } from '#lib/format.js';
	import Button from './ui/Button.svelte';
	import ErrorNote from './ui/ErrorNote.svelte';
	import Segmented from './ui/Segmented.svelte';
	import SelectField from './ui/SelectField.svelte';
	import TextField from './ui/TextField.svelte';

	let {
		account,
		cashAccounts,
		oncreated
	}: {
		/** The account the form sits on: one leg is always here. */
		account: { id: string; displayName: string; currency: string };
		cashAccounts: { id: string; displayName: string; currency: string }[];
		oncreated: () => void;
	} = $props();

	const createTransfer = graphql(`
		mutation CreateTransfer($input: CreateTransferInput!) {
			createTransfer(input: $input) {
				id
			}
		}
	`);

	// Only accounts in the same currency: a transfer does not convert.
	const others = $derived(cashAccounts.filter((a) => a.id !== account.id && a.currency === account.currency));

	function nowLocal() {
		const d = new Date();
		d.setSeconds(0, 0);
		return new Date(d.getTime() - d.getTimezoneOffset() * 60000).toISOString().slice(0, 16);
	}

	let direction = $state<'out' | 'in'>('out');
	let otherID = $state('');
	let amount = $state('');
	let time = $state(nowLocal());
	let error = $state<string | null>(null);
	let saving = $state(false);

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		const minor = parseAmount(amount);
		if (!otherID || minor === null || minor <= 0) {
			error = 'Pick the other account and a positive amount.';
			return;
		}
		saving = true;
		const result = await mutate(createTransfer, {
			input: {
				time: new Date(time),
				fromCashAccountID: direction === 'out' ? account.id : otherID,
				toCashAccountID: direction === 'out' ? otherID : account.id,
				amount: minor,
				currency: account.currency
			}
		});
		saving = false;
		error = result.error;
		if (error) return;
		amount = '';
		oncreated();
	}
</script>

<form onsubmit={submit} class="flex flex-col gap-3">
	<Segmented
		label="Direction"
		bind:value={direction}
		options={[
			{ value: 'out', label: `Out of ${account.displayName}` },
			{ value: 'in', label: `Into ${account.displayName}` }
		]}
	/>
	{#if others.length === 0}
		<p class="text-sm text-muted">No other account in {account.currency} to move money to or from.</p>
	{:else}
		<SelectField bind:value={otherID} label={direction === 'out' ? 'To' : 'From'} required>
			<option value="">Choose an account…</option>
			{#each others as a (a.id)}
				<option value={a.id}>{a.displayName}</option>
			{/each}
		</SelectField>
	{/if}
	<div class="grid grid-cols-2 gap-3">
		<TextField bind:value={amount} label="Amount" placeholder="e.g. 500" required />
		<TextField bind:value={time} label="Date" type="datetime-local" required />
	</div>
	<p class="text-xs text-muted">
		Booked as two linked legs, one on each account. A loan's disbursement and repayments, a Sparrate,
		or money moved between your own accounts: moved, not spent.
	</p>
	<ErrorNote message={error} />
	<div class="flex justify-end">
		<Button type="submit" disabled={saving || others.length === 0}>{saving ? 'Booking…' : 'Transfer'}</Button>
	</div>
</form>
