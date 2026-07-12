<script lang="ts">
	import { graphql } from '$houdini';
	import Button from './ui/Button.svelte';
	import ErrorNote from './ui/ErrorNote.svelte';
	import SelectField from './ui/SelectField.svelte';
	import TextField from './ui/TextField.svelte';

	const createSecurity = graphql(`
		mutation CreateSecurity($input: CreateSecurityInput!) {
			createSecurity(input: $input) {
				...All_Securities_insert
			}
		}
	`);

	let name = $state('');
	let isin = $state('');
	let ticker = $state('');
	let exchange = $state('');
	let currency = $state('EUR');
	let quoteProvider = $state('');
	let error = $state<string | null>(null);

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		if (!name.trim()) return;

		const result = await createSecurity.mutate({
			input: {
				displayName: name.trim(),
				identifiers: isin.trim() ? [{ kind: 'ISIN', value: isin.trim().toUpperCase() }] : [],
				listings: ticker.trim()
					? [
							{
								ticker: ticker.trim().toUpperCase(),
								exchange: exchange.trim() || undefined,
								currency: currency.toUpperCase(),
								quoteProvider: quoteProvider || undefined
							}
						]
					: []
			}
		});
		error = result.errors?.[0]?.message ?? null;
		if (!error) {
			name = '';
			isin = '';
			ticker = '';
			exchange = '';
		}
	}
</script>

<form onsubmit={submit} class="rounded-md border border-gray-200 bg-white p-4">
	<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
		<TextField bind:value={name} label="Name" placeholder="e.g. iShares Core MSCI World" required />
		<TextField bind:value={isin} label="ISIN" placeholder="e.g. IE00B4L5Y983" maxlength={12} />
		<TextField bind:value={ticker} label="Ticker" placeholder="e.g. EUNL" />
		<TextField bind:value={exchange} label="Exchange (MIC)" placeholder="e.g. XETR" />
		<TextField bind:value={currency} label="Listing currency" maxlength={3} />
		<SelectField bind:value={quoteProvider} label="Quote provider">
			<option value="">None</option>
			<option value="yf">Yahoo Finance (by ticker)</option>
			<option value="ing">ING (by ISIN)</option>
		</SelectField>
	</div>

	<div class="mt-4">
		<Button type="submit">Create security</Button>
	</div>
	<ErrorNote message={error} />
</form>
