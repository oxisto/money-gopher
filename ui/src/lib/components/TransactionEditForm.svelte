<script lang="ts">
	import { graphql, TransactionType, type PortfolioDetail$result } from '$houdini';
	import type { TransactionType$options } from '$houdini';
	import { parseAmount } from '$lib/format';
	import Button from './ui/Button.svelte';
	import ErrorNote from './ui/ErrorNote.svelte';
	import SelectField from './ui/SelectField.svelte';
	import TextField from './ui/TextField.svelte';

	type Transaction = NonNullable<PortfolioDetail$result['portfolio']>['transactions'][number];

	let {
		transaction: tx,
		securities,
		cashAccounts,
		onsaved,
		oncancel
	}: {
		transaction: Transaction;
		securities: PortfolioDetail$result['securities'];
		cashAccounts: PortfolioDetail$result['cashAccounts'];
		onsaved: () => void;
		oncancel: () => void;
	} = $props();

	const updateTransaction = graphql(`
		mutation UpdateTransaction($id: ID!, $input: UpdateTransactionInput!) {
			updateTransaction(id: $id, input: $input) {
				id
			}
		}
	`);

	// Convert a Date to the datetime-local input format (local time).
	function toDatetimeLocal(d: Date): string {
		const copy = new Date(d);
		copy.setMinutes(copy.getMinutes() - copy.getTimezoneOffset());
		return copy.toISOString().slice(0, 16);
	}

	// Minor units → decimal string for display.
	function minorToStr(amount: number | null | undefined): string {
		if (amount == null) return '';
		return (amount / 100).toFixed(2);
	}

	let type = $state<TransactionType$options>(tx.type as TransactionType$options);
	let time = $state(toDatetimeLocal(tx.time));
	let securityID = $state(tx.security?.id ?? '');
	let cashAccountID = $state(tx.cashAccount?.id ?? '');
	let units = $state(tx.units ? String(tx.units) : '');
	let price = $state(minorToStr(tx.price?.amount));
	let fees = $state(minorToStr(tx.fees?.amount));
	let taxes = $state(minorToStr(tx.taxes?.amount));
	let cashDelta = $state(minorToStr(tx.cashDelta?.amount));
	let currency = $state(tx.cashDelta?.currency ?? 'EUR');
	let error = $state<string | null>(null);

	async function submit(event: SubmitEvent) {
		event.preventDefault();

		const result = await updateTransaction.mutate({
			id: tx.id,
			input: {
				type,
				time: new Date(time),
				currency: currency.toUpperCase(),
				securityID: securityID || undefined,
				cashAccountID: cashAccountID || undefined,
				units: units ? parseFloat(units) : undefined,
				price: parseAmount(price) ?? undefined,
				fees: parseAmount(fees) ?? undefined,
				taxes: parseAmount(taxes) ?? undefined,
				cashDelta: parseAmount(cashDelta) ?? undefined
			}
		});

		error = result.errors?.[0]?.message ?? null;
		if (!error) onsaved();
	}
</script>

<form onsubmit={submit} class="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
	<SelectField bind:value={type} label="Type" required>
		{#each Object.values(TransactionType) as option (option)}
			<option value={option}>{option}</option>
		{/each}
	</SelectField>

	<TextField bind:value={time} label="Date" type="datetime-local" required />

	<SelectField bind:value={securityID} label="Security">
		<option value="">None</option>
		{#each securities as s (s.id)}
			<option value={s.id}>{s.displayName}</option>
		{/each}
	</SelectField>

	<SelectField bind:value={cashAccountID} label="Cash account">
		<option value="">None</option>
		{#each cashAccounts as a (a.id)}
			<option value={a.id}>{a.displayName}</option>
		{/each}
	</SelectField>

	<TextField bind:value={units} label="Units" type="number" step="any" />
	<TextField bind:value={price} label="Price per unit" placeholder="0.00" />
	<TextField bind:value={fees} label="Fees" placeholder="0.00" />
	<TextField bind:value={taxes} label="Taxes" placeholder="0.00" />
	<TextField bind:value={cashDelta} label="Cash delta" placeholder="e.g. 1000.00 or -50.00" />
	<TextField bind:value={currency} label="Currency" maxlength={3} required />

	<div class="flex items-end gap-2 sm:col-span-2 lg:col-span-4">
		<Button type="submit">Save changes</Button>
		<button
			type="button"
			onclick={oncancel}
			class="rounded-md px-3 py-1.5 text-sm text-gray-500 hover:text-gray-800"
		>
			Cancel
		</button>
		<ErrorNote message={error} />
	</div>
</form>
