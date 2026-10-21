<script lang="ts">
	import { graphql, mutate } from '#lib/graphql/client.js';
	import { TRANSACTION_TYPES, type TransactionType } from '#lib/transaction-types.js';
	import { labelOf } from '#lib/describe.js';
	import { formatMoney, parseAmount, parsePrice } from '#lib/format.js';
	import type { SecurityOption, CashAccountOption } from '#lib/transaction-types.js';
	import Button from './ui/Button.svelte';
	import ErrorNote from './ui/ErrorNote.svelte';
	import SelectField from './ui/SelectField.svelte';
	import TextField from './ui/TextField.svelte';

	let {
		portfolioID = '',
		defaultCashAccountID = '',
		defaultType = 'BUY',
		securities,
		cashAccounts,
		oncreated
	}: {
		portfolioID?: string;
		defaultCashAccountID?: string;
		defaultType?: TransactionType;
		securities: SecurityOption[];
		cashAccounts: CashAccountOption[];
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

	let type = $state<TransactionType>(defaultType);
	let time = $state(nowLocal());
	let securityID = $state('');
	let cashAccountID = $state(defaultCashAccountID);
	let units = $state('');
	let price = $state('');
	let fees = $state('');
	let taxes = $state('');
	let cashDelta = $state('');
	let gross = $state('');
	let currency = $state('EUR');
	let error = $state<string | null>(null);

	// DEPOSIT_CASH/INTEREST credit the account net of any withheld tax;
	// WITHDRAW_CASH/ACCOUNT_FEES debit it the same way — matching the
	// gross/taxes/net convention the Kontoauszug importer already applies
	// (see internal/importer/ing_konto.go), so a manual entry ends up
	// looking the same as an imported one instead of needing the net
	// amount worked out by hand first.
	const creditTypes = new Set(['DEPOSIT_CASH', 'INTEREST']);
	const debitTypes = new Set(['WITHDRAW_CASH', 'ACCOUNT_FEES']);
	const isGrossDerived = $derived(creditTypes.has(type) || debitTypes.has(type));

	const derivedCashDelta = $derived.by(() => {
		const net = (parseAmount(gross) ?? 0) - (parseAmount(taxes) ?? 0);
		return debitTypes.has(type) ? -net : net;
	});

	async function submit(event: SubmitEvent) {
		event.preventDefault();

		const result = await mutate(createTransaction, {
			input: {
				type,
				time: new Date(time),
				currency: currency.toUpperCase(),
				portfolioID: portfolioID || undefined,
				securityID: securityID || undefined,
				cashAccountID: cashAccountID || undefined,
				units: units ? parseFloat(units) : undefined,
				price: parsePrice(price) ?? undefined,
				fees: parseAmount(fees) ?? undefined,
				taxes: parseAmount(taxes) ?? undefined,
				cashDelta: isGrossDerived ? derivedCashDelta : (parseAmount(cashDelta) ?? undefined)
			}
		});
		error = result.error;
		if (!error) {
			units = '';
			price = '';
			fees = '';
			taxes = '';
			cashDelta = '';
			gross = '';
			oncreated();
		}
	}
</script>

<form onsubmit={submit}>
	<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
		<SelectField bind:value={type} label="Type" required>
			{#each TRANSACTION_TYPES as option (option)}
				<option value={option}>{labelOf(option)}</option>
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

		{#if isGrossDerived}
			<TextField bind:value={gross} label="Gross amount" placeholder="0.00" />
			<TextField bind:value={taxes} label="Taxes withheld" placeholder="0.00" />
			<div class="flex flex-col justify-end pb-1.5">
				<span class="mb-1 block text-sm font-medium text-fg">Cash delta</span>
				<span
					class="tabular-nums {derivedCashDelta < 0 ? 'text-neg' : 'text-pos'}"
				>
					{formatMoney({ amount: derivedCashDelta, currency: currency.toUpperCase() || 'EUR' })}
				</span>
			</div>
		{:else}
			<TextField bind:value={units} label="Units" type="number" step="any" />
			<TextField bind:value={price} label="Price per unit" placeholder="0.00" />
			<TextField bind:value={fees} label="Fees" placeholder="0.00" />
			<TextField bind:value={taxes} label="Taxes" placeholder="0.00" />
			<TextField
				bind:value={cashDelta}
				label="Cash delta (cash events only)"
				placeholder="e.g. 1000.00 or -50.00"
			/>
		{/if}
		<TextField bind:value={currency} label="Currency" maxlength={3} required />
	</div>

	<div class="mt-4">
		<Button type="submit">Add transaction</Button>
	</div>
	<ErrorNote message={error} />
</form>
