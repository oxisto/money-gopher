<script lang="ts">
	import { graphql } from '$houdini';
	import type { Securities$result } from '$houdini';
	import { formatDate, formatPrice } from '$lib/format';
	import Button from './ui/Button.svelte';
	import ErrorNote from './ui/ErrorNote.svelte';
	import SelectField from './ui/SelectField.svelte';
	import TextField from './ui/TextField.svelte';

	let { securities }: { securities: Securities$result['securities'] } = $props();

	type Listing = Securities$result['securities'][number]['listings'][number];

	const updateSecurity = graphql(`
		mutation UpdateSecurityListing($id: ID!, $listings: [ListingInput!]!) {
			updateSecurity(id: $id, input: { listings: $listings }) {
				id
				listings {
					id
					ticker
					exchange
					currency
					quoteProvider
					latestQuote {
						time
						price {
							amount
							currency
						}
					}
				}
			}
		}
	`);

	// Keyed by listing ID for an existing listing, or `new:<securityID>` for
	// the not-yet-created one — at most one editor open at a time.
	let editingKey = $state<string | null>(null);
	let ticker = $state('');
	let exchange = $state('');
	let currency = $state('EUR');
	let quoteProvider = $state('');
	let error = $state<string | null>(null);

	function startEdit(listing: Listing) {
		editingKey = listing.id;
		ticker = listing.ticker;
		exchange = listing.exchange ?? '';
		currency = listing.currency;
		quoteProvider = listing.quoteProvider ?? '';
		error = null;
	}

	function startAdd(securityID: string) {
		editingKey = `new:${securityID}`;
		ticker = '';
		exchange = '';
		currency = 'EUR';
		quoteProvider = '';
		error = null;
	}

	function cancel() {
		editingKey = null;
		error = null;
	}

	interface ListingInputData {
		id?: string;
		ticker: string;
		exchange?: string;
		currency: string;
		quoteProvider?: string;
	}

	// updateSecurity's listings input is the *complete* desired set — send
	// every other existing listing back unchanged (by ID, so its quote
	// history is preserved too) alongside the one actually being edited or
	// added.
	async function save(securityID: string, listings: readonly Listing[], editingListingID: string | null) {
		if (!ticker.trim()) {
			error = 'Ticker is required.';
			return;
		}

		const input: ListingInputData[] = listings.map((l) =>
			l.id === editingListingID
				? {
						id: l.id,
						ticker: ticker.trim().toUpperCase(),
						exchange: exchange.trim() || undefined,
						currency: currency.trim().toUpperCase(),
						quoteProvider: quoteProvider || undefined
					}
				: {
						id: l.id,
						ticker: l.ticker,
						exchange: l.exchange ?? undefined,
						currency: l.currency,
						quoteProvider: l.quoteProvider ?? undefined
					}
		);
		if (editingListingID === null) {
			input.push({
				ticker: ticker.trim().toUpperCase(),
				exchange: exchange.trim() || undefined,
				currency: currency.trim().toUpperCase(),
				quoteProvider: quoteProvider || undefined
			});
		}

		const result = await updateSecurity.mutate({ id: securityID, listings: input });
		if (result.errors?.length) {
			error = result.errors[0].message;
			return;
		}
		editingKey = null;
	}
</script>

{#if securities.length === 0}
	<p class="text-sm text-muted">No securities yet.</p>
{:else}
	<ul class="divide-y divide-line rounded-md border border-line bg-surface">
		{#each securities as security (security.id)}
			<li class="px-4 py-3">
				<div class="flex items-center justify-between">
					<span class="text-sm font-medium">
						{security.displayName}
						{#if security.underlyingSecurity}
							<span class="ml-1 text-xs font-normal text-muted">
								→ {security.underlyingSecurity.displayName}
							</span>
						{/if}
					</span>
					<span class="flex gap-1">
						{#each security.identifiers as identifier (identifier.kind)}
							<span class="rounded bg-hover px-1.5 py-0.5 text-xs text-muted">
								{identifier.kind}: {identifier.value}
							</span>
						{/each}
					</span>
				</div>

				{#if security.listings.length > 0}
					<div class="mt-1 flex flex-col gap-0.5 text-xs text-muted">
						{#each security.listings as listing (listing.id)}
							<div class="flex items-center justify-between">
								<span>
									{listing.ticker}{listing.exchange ? ` @ ${listing.exchange}` : ''}
									({listing.currency}){listing.quoteProvider
										? ` · quotes: ${listing.quoteProvider}`
										: ''}
								</span>
								<span class="flex items-center gap-2">
									{#if listing.latestQuote}
										<span class="tabular-nums">
											{formatPrice(listing.latestQuote.price)}
											<span class="text-muted">({formatDate(listing.latestQuote.time)})</span>
										</span>
									{:else}
										<span class="text-muted">no quote yet</span>
									{/if}
									<button
										onclick={() => (editingKey === listing.id ? cancel() : startEdit(listing))}
										class="rounded px-1.5 py-0.5 text-muted hover:bg-hover hover:text-fg"
									>
										{editingKey === listing.id ? 'Cancel' : 'Edit'}
									</button>
								</span>
							</div>
						{/each}
					</div>
				{:else if editingKey !== `new:${security.id}`}
					<button
						onclick={() => startAdd(security.id)}
						class="mt-1 text-xs text-muted hover:text-fg"
					>
						+ Add listing
					</button>
				{/if}

				{#if security.listings.some((l) => l.id === editingKey) || editingKey === `new:${security.id}`}
					{@const editingListingID = editingKey?.startsWith('new:') ? null : editingKey}
					<div class="mt-2 grid gap-2 rounded-md bg-raised p-3 sm:grid-cols-4">
						<TextField bind:value={ticker} label="Ticker" placeholder="e.g. EUNL" />
						<TextField bind:value={exchange} label="Exchange (MIC)" placeholder="e.g. XETR" />
						<TextField bind:value={currency} label="Currency" maxlength={3} />
						<SelectField bind:value={quoteProvider} label="Quote provider">
							<option value="">None</option>
							<option value="yf">Yahoo Finance (by ticker)</option>
							<option value="ing">ING (by ISIN)</option>
						</SelectField>
					</div>
					<div class="mt-2 flex items-center gap-2">
						<Button
							type="button"
							onclick={() => save(security.id, security.listings, editingListingID)}
						>
							Save
						</Button>
						<button onclick={cancel} class="text-xs text-muted hover:text-fg">Cancel</button>
					</div>
					<ErrorNote message={error} />
				{/if}
			</li>
		{/each}
	</ul>
{/if}
