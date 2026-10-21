<script lang="ts">
	import Section from '$lib/components/ui/Section.svelte';
	import { formatDate, formatMoney, formatPrice } from '$lib/format';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	const TransactionDetail = $derived(data.TransactionDetail);
</script>

{#if $TransactionDetail.fetching && !$TransactionDetail.data}
	<p class="text-sm text-muted">Loading…</p>
{:else if $TransactionDetail.errors?.length}
	<p class="text-sm text-neg">Failed to load: {$TransactionDetail.errors[0].message}</p>
{:else if $TransactionDetail.data}
	{@const tx = $TransactionDetail.data.transaction}
	{#if !tx}
		<p class="text-sm text-muted">Transaction not found.</p>
	{:else}
		<div class="mb-6 flex items-start justify-between">
			<div>
				<div class="flex items-center gap-2">
					<span class="rounded bg-hover px-1.5 py-0.5 text-xs font-mono">{tx.type}</span>
					<h1 class="text-2xl font-semibold">
						{tx.security?.displayName ?? tx.cashAccount?.displayName ?? tx.type}
					</h1>
				</div>
				<p class="mt-1 text-sm text-muted">{formatDate(tx.time)}</p>
			</div>
			<div class="text-right">
				<p
					class="text-2xl font-semibold tabular-nums {tx.cashDelta.amount < 0
						? 'text-neg'
						: 'text-pos'}"
				>
					{formatMoney(tx.cashDelta)}
				</p>
				{#if tx.gain}
					<p
						class="mt-0.5 text-sm tabular-nums {tx.gain.amount < 0 ? 'text-neg' : 'text-pos'}"
					>
						Gain: {formatMoney(tx.gain)}
					</p>
				{/if}
			</div>
		</div>

		<div class="mb-4 flex gap-4 text-sm">
			{#if tx.portfolio}
				<a href={`/portfolios/${tx.portfolio.id}#tx-${tx.id}`} class="text-accent hover:underline">
					← {tx.portfolio.displayName}
				</a>
			{/if}
			{#if tx.cashAccount}
				<a
					href={`/cash-accounts/${tx.cashAccount.id}#tx-${tx.id}`}
					class="text-accent hover:underline"
				>
					← {tx.cashAccount.displayName}
				</a>
			{/if}
		</div>

		<div class="grid grid-cols-1 gap-8 {tx.document ? 'lg:grid-cols-2' : ''}">
			<div class="flex flex-col gap-8">
				<Section title="Details">
					<div class="grid grid-cols-2 gap-4 sm:grid-cols-4">
						<div class="rounded-md border border-line bg-surface px-4 py-3">
							<div class="text-xs text-muted uppercase">Units</div>
							<div class="mt-1 font-semibold tabular-nums">{tx.units || '—'}</div>
						</div>
						<div class="rounded-md border border-line bg-surface px-4 py-3">
							<div class="text-xs text-muted uppercase">Price</div>
							<div class="mt-1 font-semibold tabular-nums">{tx.price ? formatPrice(tx.price) : '—'}</div>
							{#if tx.fxRate}
								<div class="mt-0.5 text-xs text-muted">
									FX 1 {tx.cashDelta.currency} = {tx.fxRate} {tx.price?.currency}
								</div>
							{/if}
						</div>
						<div class="rounded-md border border-line bg-surface px-4 py-3">
							<div class="text-xs text-muted uppercase">Fees</div>
							<div class="mt-1 font-semibold tabular-nums">{formatMoney(tx.fees)}</div>
						</div>
						<div class="rounded-md border border-line bg-surface px-4 py-3">
							<div class="text-xs text-muted uppercase">Taxes</div>
							<div class="mt-1 font-semibold tabular-nums">{formatMoney(tx.taxes)}</div>
						</div>
					</div>
					<p class="mt-3 text-xs text-muted">Source: {tx.source}</p>
				</Section>

				{#if tx.security}
					<Section title="Security">
						<div class="rounded-md border border-line bg-surface px-4 py-3">
							<p class="font-medium">{tx.security.displayName}</p>
							{#if tx.security.identifiers.length > 0}
								<dl class="mt-2 flex flex-wrap gap-x-6 gap-y-1 text-sm text-muted">
									{#each tx.security.identifiers as ident (ident.kind + ident.value)}
										<div class="flex gap-1">
											<dt class="text-muted">{ident.kind}</dt>
											<dd class="font-mono">{ident.value}</dd>
										</div>
									{/each}
								</dl>
							{/if}
						</div>
					</Section>
				{/if}
			</div>

			{#if tx.document}
				<div class="lg:sticky lg:top-4 lg:self-start">
					<Section title="Source document">
						<div class="rounded-md border border-line bg-surface p-3">
							<div class="mb-2 flex items-center justify-between text-sm">
								<div>
									<span class="font-medium">{tx.document.filename}</span>
									{#if tx.document.detectedBank}
										<span class="ml-2 text-muted">({tx.document.detectedBank})</span>
									{/if}
								</div>
								<a
									href={`/documents/${tx.document.id}`}
									target="_blank"
									rel="noopener"
									class="text-accent hover:underline"
								>
									Open in new tab ↗
								</a>
							</div>
							{#if tx.document.contentType === 'application/pdf'}
								<iframe
									src={`/documents/${tx.document.id}`}
									title={tx.document.filename}
									class="h-[calc(100vh-10rem)] w-full rounded border border-line"
								></iframe>
							{:else}
								<p class="text-sm text-muted">
									No inline preview for {tx.document.contentType || 'this file type'} — use "Open in
									new tab" above.
								</p>
							{/if}
						</div>
					</Section>
				</div>
			{/if}
		</div>
	{/if}
{/if}
