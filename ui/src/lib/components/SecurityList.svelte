<script lang="ts">
	import type { Securities$result } from '$houdini';
	import { formatDate, formatMoney } from '$lib/format';

	let { securities }: { securities: Securities$result['securities'] } = $props();
</script>

{#if securities.length === 0}
	<p class="text-sm text-gray-500">No securities yet.</p>
{:else}
	<ul class="divide-y divide-gray-200 rounded-md border border-gray-200 bg-white">
		{#each securities as security (security.id)}
			<li class="px-4 py-3">
				<div class="flex items-center justify-between">
					<span class="text-sm font-medium">{security.displayName}</span>
					<span class="flex gap-1">
						{#each security.identifiers as identifier (identifier.kind)}
							<span class="rounded bg-gray-100 px-1.5 py-0.5 text-xs text-gray-600">
								{identifier.kind}: {identifier.value}
							</span>
						{/each}
					</span>
				</div>
				{#if security.listings.length > 0}
					<div class="mt-1 flex flex-col gap-0.5 text-xs text-gray-500">
						{#each security.listings as listing (listing.id)}
							<div class="flex items-center justify-between">
								<span>
									{listing.ticker}{listing.exchange ? ` @ ${listing.exchange}` : ''}
									({listing.currency}){listing.quoteProvider
										? ` · quotes: ${listing.quoteProvider}`
										: ''}
								</span>
								{#if listing.latestQuote}
									<span class="tabular-nums">
										{formatMoney(listing.latestQuote.price)}
										<span class="text-gray-400">({formatDate(listing.latestQuote.time)})</span>
									</span>
								{:else}
									<span class="text-gray-400">no quote yet</span>
								{/if}
							</div>
						{/each}
					</div>
				{/if}
			</li>
		{/each}
	</ul>
{/if}
