<script lang="ts">
	import { refreshAll } from '$app/navigation';
	import { graphql, mutate } from '#lib/graphql/client.js';
	import Button from './ui/Button.svelte';
	import ErrorNote from './ui/ErrorNote.svelte';
	import SelectField from './ui/SelectField.svelte';
	import TextField from './ui/TextField.svelte';
	import { asKind, securityKinds } from '#lib/securities.js';

	let { oncreated }: { oncreated?: () => void } = $props();

	const createSecurity = graphql(`
		mutation CreateSecurity($input: CreateSecurityInput!) {
			createSecurity(input: $input) {
				id
			}
		}
	`);

	let name = $state('');
	let kind = $state('');
	let isin = $state('');
	let ticker = $state('');
	let exchange = $state('');
	let currency = $state('');
	let quoteProvider = $state('');
	let error = $state<string | null>(null);

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		if (!name.trim()) return;
		if (ticker.trim() && currency.trim().length !== 3) {
			error = 'A listing needs the currency it is quoted in.';
			return;
		}

		const result = await mutate(createSecurity, {
			input: {
				displayName: name.trim(),
				kind: asKind(kind),
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
		error = result.error;
		if (!error) {
			name = '';
			isin = '';
			ticker = '';
			exchange = '';
			await refreshAll();
			oncreated?.();
		}
	}
</script>

<form onsubmit={submit}>
	<div class="grid gap-3 sm:grid-cols-2">
		<TextField bind:value={name} label="Name" placeholder="e.g. iShares Core MSCI World" required />
		<SelectField bind:value={kind} label="Kind">
			<option value="">Not set</option>
			{#each securityKinds as [value, label] (value)}
				<option {value}>{label}</option>
			{/each}
		</SelectField>
		<TextField bind:value={isin} label="ISIN" placeholder="e.g. IE00B4L5Y983" maxlength={12} />
		<TextField bind:value={ticker} label="Ticker" placeholder="e.g. EUNL" />
		<TextField bind:value={exchange} label="Exchange (MIC)" placeholder="e.g. XETR" />
		<TextField
			bind:value={currency}
			label="Listing currency"
			placeholder="What it is quoted in, e.g. USD"
			maxlength={3}
			required={!!ticker.trim()}
		/>
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
