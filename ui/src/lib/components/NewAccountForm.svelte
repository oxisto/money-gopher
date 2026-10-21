<script lang="ts">
	import { graphql, mutate } from '#lib/graphql/client.js';
	import { refreshAll } from '$app/navigation';
	import Button from './ui/Button.svelte';
	import ErrorNote from './ui/ErrorNote.svelte';
	import Segmented from './ui/Segmented.svelte';
	import SelectField from './ui/SelectField.svelte';
	import TextField from './ui/TextField.svelte';

	let {
		cashAccounts,
		oncreated
	}: {
		cashAccounts: { id: string; displayName: string; currency: string }[];
		oncreated?: () => void;
	} = $props();

	// A portfolio settles against a cash account, so with none yet the only
	// sensible thing to create is a cash account.
	let kind = $state<'cash' | 'portfolio'>('cash');

	const createCashAccount = graphql(`
		mutation CreateCashAccount($input: CreateCashAccountInput!) {
			createCashAccount(input: $input) {
				id
			}
		}
	`);

	const createPortfolio = graphql(`
		mutation CreatePortfolio($input: CreatePortfolioInput!) {
			createPortfolio(input: $input) {
				id
			}
		}
	`);

	let name = $state('');
	let currency = $state('EUR');
	let iban = $state('');
	let cashAccountID = $state('');
	let error = $state<string | null>(null);

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		if (!name.trim()) return;

		const result =
			kind === 'cash'
				? await mutate(createCashAccount, {
						input: {
							displayName: name.trim(),
							currency: currency.toUpperCase(),
							iban: iban.trim() || null
						}
					})
				: await mutate(createPortfolio, {
						input: { displayName: name.trim(), cashAccountID }
					});
		error = result.error;
		if (error) return;

		name = '';
		iban = '';
		cashAccountID = '';
		// Reloads the page and the sidebar's account list alike.
		await refreshAll();
		oncreated?.();
	}
</script>

<form onsubmit={submit} class="flex flex-col gap-3">
	<Segmented
		label="Account type"
		bind:value={kind}
		options={[
			{ value: 'cash', label: 'Cash account' },
			{ value: 'portfolio', label: 'Portfolio' }
		]}
	/>

	{#if kind === 'cash'}
		<p class="text-xs text-muted">A bank account holding money: Giro, Tagesgeld, a broker's settlement account.</p>
		<div class="grid grid-cols-[1fr_5rem] gap-2">
			<TextField label="Name" bind:value={name} placeholder="ING Tagesgeld" required />
			<TextField label="Currency" bind:value={currency} maxlength={3} required />
		</div>
		<TextField label="IBAN (optional)" bind:value={iban} placeholder="Matches imported statements to this account" />
	{:else if cashAccounts.length === 0}
		<p class="text-sm text-muted">A portfolio settles its trades against a cash account. Create one of those first.</p>
	{:else}
		<p class="text-xs text-muted">A securities account (Depot). Trades and dividends settle against a cash account.</p>
		<TextField label="Name" bind:value={name} placeholder="ING Depot" required />
		<SelectField label="Settles against" bind:value={cashAccountID} required>
			<option value="">Choose a cash account…</option>
			{#each cashAccounts as account (account.id)}
				<option value={account.id}>{account.displayName} ({account.currency})</option>
			{/each}
		</SelectField>
	{/if}

	<ErrorNote message={error} />
	<div class="flex justify-end">
		<Button type="submit" disabled={kind === 'portfolio' && cashAccounts.length === 0}>
			Create {kind === 'cash' ? 'cash account' : 'portfolio'}
		</Button>
	</div>
</form>
