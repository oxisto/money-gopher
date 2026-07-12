<script lang="ts">
	import { graphql, type Dashboard$result } from '$houdini';
	import { formatMoney } from '$lib/format';
	import Button from './ui/Button.svelte';
	import ErrorNote from './ui/ErrorNote.svelte';
	import TextField from './ui/TextField.svelte';

	let { accounts }: { accounts: Dashboard$result['cashAccounts'] } = $props();

	const createCashAccount = graphql(`
		mutation CreateCashAccount($input: CreateCashAccountInput!) {
			createCashAccount(input: $input) {
				...All_CashAccounts_insert
			}
		}
	`);

	let name = $state('');
	let currency = $state('EUR');
	let iban = $state('');
	let error = $state<string | null>(null);

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		if (!name.trim()) return;

		const result = await createCashAccount.mutate({
			input: {
				displayName: name.trim(),
				currency: currency.toUpperCase(),
				iban: iban.trim() || null
			}
		});
		error = result.errors?.[0]?.message ?? null;
		if (!error) {
			name = '';
			iban = '';
		}
	}
</script>

{#if accounts.length === 0}
	<p class="text-sm text-gray-500">No cash accounts yet.</p>
{:else}
	<ul class="divide-y divide-gray-200 rounded-md border border-gray-200 bg-white">
		{#each accounts as account (account.id)}
			<li class="flex items-center justify-between px-4 py-3 text-sm">
				<span class="font-medium">{account.displayName}</span>
				<span class="tabular-nums">{formatMoney(account.balance)}</span>
			</li>
		{/each}
	</ul>
{/if}

<form onsubmit={submit} class="mt-4 flex flex-col gap-2">
	<div class="flex items-start gap-2">
		<div class="grow">
			<TextField bind:value={name} placeholder="Account name" required />
		</div>
		<div class="w-20">
			<TextField bind:value={currency} placeholder="EUR" maxlength={3} required />
		</div>
		<Button type="submit">Create</Button>
	</div>
	<TextField bind:value={iban} placeholder="IBAN (optional — for auto-matching imports)" />
</form>
<ErrorNote message={error} />
