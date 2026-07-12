<script lang="ts">
	import { graphql, TransactionType, type PortfolioDetail$result } from '$houdini';
	import type { TransactionType$options } from '$houdini';
	import { parseAmount } from '$lib/format';
	import Button from './ui/Button.svelte';
	import ErrorNote from './ui/ErrorNote.svelte';
	import SelectField from './ui/SelectField.svelte';
	import TextField from './ui/TextField.svelte';

	let {
		portfolioID,
		defaultCashAccountID = '',
		securities,
		cashAccounts,
		oncreated
	}: {
		portfolioID: string;
		defaultCashAccountID?: string;
		securities: PortfolioDetail$result['securities'];
		cashAccounts: PortfolioDetail$result['cashAccounts'];
		oncreated: () => void;
	} = $props();

	const createTransaction = graphql(`
		mutation CreateTransaction($input: CreateTransactionInput!) {
			createTransaction(input: $input) {
				id
			}
		}
	`);

	// datetime-local wants "YYYY-MM-DDTHH:MM" in local time.
	function nowLocal(): string {
		const now = new Date();
		now.setMinutes(now.getMinutes() - now.getTimezoneOffset());
		return now.toISOString().slice(0, 16);
	}

	let type = $state<TransactionType$options>('BUY');
	let time = $state(nowLocal());
	let securityID = $state('');
	let cashAccountID = $state(defaultCashAccountID);
	let units = $state('');
	let price = $state('');
	let fees = $state('');
	let taxes = $state('');
	let cashDelta = $state('');
	let currency = $state('EUR');
	let error = $state<string | null>(null);

	async function submit(event: SubmitEvent) {
		event.preventDefault();

		const result = await createTransaction.mutate({
			input: {
				type,
				time: new Date(time),
				currency: currency.toUpperCase(),
				portfolioID,
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
		if (!error) {
			units = '';
			price = '';
			fees = '';
			taxes = '';
			cashDelta = '';
			oncreated();
		}
	}
</script>

<form onsubmit={submit} class="rounded-md border border-gray-200 bg-white p-4">
	<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
		<SelectField bind:value={type} label="Type" required>
			{#each Object.values(TransactionType) as option (option)}
				<option value={option}>{option}</option>
			{/each}
		</SelectField>

		<TextField bind:value={time} label="Date" type="datetime-local" required />

		<SelectField bind:value={securityID} label="Security">
			<option value="">None</option>
			{#each securities as security (security.id)}
				<option value={security.id}>{security.displayName}</option>
			{/each}
		</SelectField>

		<SelectField bind:value={cashAccountID} label="Cash account">
			<option value="">None</option>
			{#each cashAccounts as account (account.id)}
				<option value={account.id}>{account.displayName}</option>
			{/each}
		</SelectField>

		<TextField bind:value={units} label="Units" type="number" step="any" />
		<TextField bind:value={price} label="Price per unit" placeholder="0.00" />
		<TextField bind:value={fees} label="Fees" placeholder="0.00" />
		<TextField bind:value={taxes} label="Taxes" placeholder="0.00" />
		<TextField
			bind:value={cashDelta}
			label="Cash delta (cash events only)"
			placeholder="e.g. 1000.00 or -50.00"
		/>
		<TextField bind:value={currency} label="Currency" maxlength={3} required />
	</div>

	<div class="mt-4">
		<Button type="submit">Add transaction</Button>
	</div>
	<ErrorNote message={error} />
</form>
