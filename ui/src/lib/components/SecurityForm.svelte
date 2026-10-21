<script lang="ts">
	import { graphql } from '$houdini';
	import Button from './ui/Button.svelte';
	import ErrorNote from './ui/ErrorNote.svelte';
	import SelectField from './ui/SelectField.svelte';
	import TextField from './ui/TextField.svelte';

	let { oncreated }: { oncreated?: () => void } = $props();

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
			oncreated?.();
		}
	}
</script>

<form onsubmit={submit}>
	<div class="grid gap-3 sm:grid-cols-2">
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

	<ErrorNote message={error} />
	<div class="mt-4 flex justify-end">
		<Button type="submit">Create security</Button>
	</div>
</form>
