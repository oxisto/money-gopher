<script lang="ts">
	import { graphql, type Dashboard$result } from '$houdini';
	import { formatMoney } from '$lib/format';
	import Button from './ui/Button.svelte';
	import ErrorNote from './ui/ErrorNote.svelte';
	import GainBadge from './ui/GainBadge.svelte';
	import SelectField from './ui/SelectField.svelte';
	import TextField from './ui/TextField.svelte';

	let {
		portfolios,
		cashAccounts
	}: {
		portfolios: Dashboard$result['portfolios'];
		cashAccounts: Dashboard$result['cashAccounts'];
	} = $props();

	const createPortfolio = graphql(`
		mutation CreatePortfolio($input: CreatePortfolioInput!) {
			createPortfolio(input: $input) {
				...All_Portfolios_insert
			}
		}
	`);

	let name = $state('');
	let cashAccountID = $state('');
	let error = $state<string | null>(null);

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		if (!name.trim() || !cashAccountID) return;

		const result = await createPortfolio.mutate({
			input: { displayName: name.trim(), cashAccountID }
		});
		error = result.errors?.[0]?.message ?? null;
		if (!error) {
			name = '';
			cashAccountID = '';
		}
	}
</script>

{#if portfolios.length === 0}
	<p class="text-sm text-gray-500">No portfolios yet.</p>
{:else}
	<ul class="divide-y divide-gray-200 rounded-md border border-gray-200 bg-white">
		{#each portfolios as portfolio (portfolio.id)}
			<li>
				<a
					href="/portfolios/{portfolio.id}"
					class="flex items-center justify-between px-4 py-3 text-sm hover:bg-gray-50"
				>
					<span class="font-medium">{portfolio.displayName}</span>
					<span class="flex items-center gap-2 tabular-nums">
						{formatMoney(portfolio.snapshot.totalMarketValue)}
						<GainBadge gains={portfolio.snapshot.totalGains} />
					</span>
				</a>
			</li>
		{/each}
	</ul>
{/if}

<form onsubmit={submit} class="mt-4 grid gap-2 sm:grid-cols-3">
	<TextField bind:value={name} placeholder="Portfolio name" required />
	<SelectField bind:value={cashAccountID} required>
		<option value="">Settlement account…</option>
		{#each cashAccounts as account (account.id)}
			<option value={account.id}>{account.displayName} ({account.currency})</option>
		{/each}
	</SelectField>
	<Button type="submit">Create portfolio</Button>
</form>
<ErrorNote message={error} />
