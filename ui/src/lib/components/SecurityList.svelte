<script lang="ts">
	import { graphql } from '$houdini';
	import type { Securities$result } from '$houdini';
	import { formatDate, formatPrice } from '$lib/format';
	import Button from './ui/Button.svelte';
	import Icon from './ui/Icon.svelte';
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

	let query = $state('');
	const shown = $derived.by(() => {
		const q = query.trim().toLowerCase();
		if (!q) return securities;
		return securities.filter(
			(s) =>
				s.displayName.toLowerCase().includes(q) ||
				s.identifiers.some((i) => i.value.toLowerCase().includes(q)) ||
				s.listings.some((l) => l.ticker.toLowerCase().includes(q))
		);
	});

	// ING and Yahoo quote on trading days; anything older than a long
	// weekend means the provider stopped answering for this listing.
	const STALE_AFTER = 3 * 864e5;
	const isStale = (time: Date) => Date.now() - time.getTime() > STALE_AFTER;
</script>

{#snippet editor(security: (typeof securities)[number])}
	{@const editingListingID = editingKey?.startsWith('new:') ? null : editingKey}
	<div class="grid gap-2 sm:grid-cols-4">
		<TextField bind:value={ticker} label="Ticker" placeholder="e.g. EUNL" />
		<TextField bind:value={exchange} label="Exchange (MIC)" placeholder="e.g. XETR" />
		<TextField bind:value={currency} label="Currency" maxlength={3} />
		<SelectField bind:value={quoteProvider} label="Quote provider">
			<option value="">None</option>
			<option value="yf">Yahoo Finance (by ticker)</option>
			<option value="ing">ING (by ISIN)</option>
		</SelectField>
	</div>
	<div class="mt-3 flex items-center gap-2">
		<Button onclick={() => save(security.id, security.listings, editingListingID)}>Save listing</Button>
		<Button variant="secondary" onclick={cancel}>Cancel</Button>
	</div>
	<ErrorNote message={error} />
{/snippet}

<div class="flex flex-col gap-3">
	<label class="relative block max-w-xs">
		<span class="sr-only">Filter securities</span>
		<span class="pointer-events-none absolute inset-y-0 left-2.5 grid place-items-center text-muted">
			<Icon name="search" size={15} />
		</span>
		<input type="search" bind:value={query} placeholder="Filter by name, ISIN or ticker" class="h-9 w-full pl-8" />
	</label>

	{#if securities.length === 0}
		<p class="rounded-xl border border-dashed border-line px-4 py-8 text-center text-sm text-muted">
			No securities yet. Importing a statement creates them, or add one by hand.
		</p>
	{:else}
		<div class="overflow-x-auto rounded-xl border border-line">
			<table class="num w-full min-w-[46rem] text-[13px]">
				<thead class="bg-raised text-xs text-muted">
					<tr>
						<th scope="col" class="py-2.5 pr-3 pl-4 text-left font-medium">Security</th>
						<th scope="col" class="px-3 py-2.5 text-left font-medium">Listing</th>
						<th scope="col" class="px-3 py-2.5 text-right font-medium">Latest quote</th>
						<th scope="col" class="py-2.5 pr-4 pl-3"><span class="sr-only">Actions</span></th>
					</tr>
				</thead>
				<tbody>
					{#each shown as security (security.id)}
						{@const isin = security.identifiers.find((i) => i.kind === 'ISIN')?.value}
						{@const listing = security.listings[0]}
						<tr class="border-t border-line align-top hover:bg-raised">
							<td class="py-2.5 pr-3 pl-4">
								<a href="/holdings/{security.id}" class="font-medium hover:underline">{security.displayName}</a>
								<div class="text-[11px] text-muted">
									{security.identifiers.map((i) => (i.kind === 'ISIN' ? i.value : `${i.kind} ${i.value}`)).join(' · ')}
									{#if security.underlyingSecurity}
										· derives from {security.underlyingSecurity.displayName}
									{/if}
								</div>
							</td>
							<td class="px-3 py-2.5">
								{#each security.listings as l (l.id)}
									<div>
										{l.ticker !== isin ? l.ticker : ''}{l.exchange ? ` ${l.exchange}` : ''}
										<span class="text-muted">{l.currency} · {l.quoteProvider ? `quotes: ${l.quoteProvider}` : 'no quotes'}</span>
									</div>
								{:else}
									<span class="text-muted">No listing</span>
								{/each}
							</td>
							<td class="px-3 py-2.5 text-right">
								{#if listing?.latestQuote}
									{@const stale = isStale(listing.latestQuote.time as Date)}
									<div>{formatPrice(listing.latestQuote.price, 4)}</div>
									<div class="text-[11px] {stale ? 'text-gold' : 'text-muted'}" title={stale ? 'No new quote for a while' : ''}>
										{formatDate(listing.latestQuote.time as Date)}{stale ? ' · stale' : ''}
									</div>
								{:else if listing?.quoteProvider}
									<span class="text-gold">no quote yet</span>
								{:else}
									<span class="text-muted">—</span>
								{/if}
							</td>
							<td class="py-2.5 pr-4 pl-3 text-right whitespace-nowrap">
								{#if listing}
									<button
										type="button"
										onclick={() => (editingKey === listing.id ? cancel() : startEdit(listing))}
										class="rounded-md px-2 py-1 text-xs text-muted hover:bg-hover hover:text-fg"
									>
										{editingKey === listing.id ? 'Close' : 'Edit listing'}
									</button>
								{:else}
									<button
										type="button"
										onclick={() => startAdd(security.id)}
										class="rounded-md px-2 py-1 text-xs text-muted hover:bg-hover hover:text-fg"
									>
										Add listing
									</button>
								{/if}
							</td>
						</tr>
						{#if security.listings.some((l) => l.id === editingKey) || editingKey === `new:${security.id}`}
							<tr class="bg-raised">
								<td colspan="4" class="px-4 py-3">{@render editor(security)}</td>
							</tr>
						{/if}
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>
