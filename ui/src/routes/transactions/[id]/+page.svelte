<script lang="ts">
	import { refreshAll } from '$app/navigation';
	import TransactionEditForm from '#lib/components/TransactionEditForm.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import Dialog from '#lib/components/ui/Dialog.svelte';
	import Icon from '#lib/components/ui/Icon.svelte';
	import { describeTransaction, formatUnits } from '#lib/describe.js';
	import { formatDate, formatMoney, formatPrice, formatSignedMoney, toneOf } from '#lib/format.js';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	const tx = $derived(data.transaction);

	let editOpen = $state(false);
</script>

{#if !tx}
	<p class="text-sm text-muted">Transaction not found.</p>
{:else}
	{@const d = describeTransaction(tx)}
	{@const fields = [
		['Date', formatDate(tx.time)],
		['Units', tx.units ? formatUnits(tx.units) : null],
		['Price', tx.price ? formatPrice(tx.price) : null],
		['FX rate', tx.fxRate ? `1 ${tx.cashDelta.currency} = ${tx.fxRate} ${tx.price?.currency ?? ''}` : null],
		['Fees', formatMoney(tx.fees)],
		['Taxes', formatMoney(tx.taxes)],
		['Source', tx.source]
	].filter(([, v]) => v)}
	<nav aria-label="Breadcrumb" class="mb-4 flex gap-1.5 text-[13px] text-muted">
		<a href="/activity#tx-{tx.id}" class="hover:text-fg">Activity</a>
		<span aria-hidden="true">/</span>
		<span class="truncate text-fg">{d.title}</span>
	</nav>

	<div class="mb-6 flex flex-wrap items-start gap-4">
		<span class="grid size-10 place-items-center rounded-xl border border-line bg-raised text-muted">
			<Icon name={d.icon} size={18} />
		</span>
		<div class="min-w-48 flex-1">
			<h1 class="text-lg font-semibold">{d.title}</h1>
			<p class="text-[13px] text-muted">{formatDate(tx.time)}</p>
		</div>
		<Button variant="secondary" onclick={() => (editOpen = true)}>Edit</Button>
		<div class="num text-right">
			<div class="text-2xl font-semibold tracking-tight {tx.cashDelta.amount > 0 ? 'text-pos' : ''}">
				{tx.cashDelta.amount ? formatSignedMoney(tx.cashDelta) : 'No cash movement'}
			</div>
			{#if tx.gain}
				<div class="text-[13px] {toneOf(tx.gain.amount)}">realized {formatSignedMoney(tx.gain)}</div>
			{/if}
		</div>
	</div>

	<Dialog bind:open={editOpen} title="Edit transaction" wide>
		{#if editOpen}
			<TransactionEditForm
				transaction={{ ...tx, cashAccount: tx.cashAccount ? { id: tx.cashAccount.id } : null }}
				securities={data.securities}
				cashAccounts={data.cashAccounts}
				onsaved={() => {
					editOpen = false;
					refreshAll();
				}}
				oncancel={() => (editOpen = false)}
			/>
		{/if}
	</Dialog>

	<div class="flex flex-wrap items-start gap-5">
		<div class="flex min-w-0 flex-[1_1_20rem] flex-col gap-5">
			<section class="num rounded-xl border border-line p-4">
				<h2 class="mb-1 text-sm font-semibold">Details</h2>
				<dl class="text-[13px]">
					{#each fields as [k, v] (k)}
						<div class="flex justify-between gap-3 border-t border-line py-2 first:border-t-0">
							<dt class="text-muted">{k}</dt>
							<dd class="text-right">{v}</dd>
						</div>
					{/each}
				</dl>
			</section>

			<section class="rounded-xl border border-line p-4">
				<h2 class="mb-2 text-sm font-semibold">Related</h2>
				<ul class="flex flex-col gap-1 text-[13px]">
					{#if tx.security}
						<li>
							<a href="/holdings/{tx.security.id}" class="flex items-center justify-between gap-3 rounded-lg px-2 py-1.5 hover:bg-raised">
								<span>
									{tx.security.displayName}
									<span class="block text-[11px] text-muted">
										{tx.security.identifiers.map((i) => `${i.kind} ${i.value}`).join(' · ')}
									</span>
								</span>
								<span class="text-muted">→</span>
							</a>
						</li>
					{/if}
					{#if tx.portfolio}
						<li>
							<a href="/portfolios/{tx.portfolio.id}" class="flex items-center justify-between gap-3 rounded-lg px-2 py-1.5 hover:bg-raised">
								<span>{tx.portfolio.displayName} <span class="text-muted">· portfolio</span></span>
								<span class="text-muted">→</span>
							</a>
						</li>
					{/if}
					{#if tx.cashAccount}
						<li>
							<a href="/cash-accounts/{tx.cashAccount.id}#tx-{tx.id}" class="flex items-center justify-between gap-3 rounded-lg px-2 py-1.5 hover:bg-raised">
								<span>{tx.cashAccount.displayName} <span class="text-muted">· cash account</span></span>
								<span class="text-muted">→</span>
							</a>
						</li>
					{/if}
				</ul>
			</section>
		</div>

		<section class="min-w-0 flex-[2_1_28rem] rounded-xl border border-line p-4 lg:sticky lg:top-4">
			<div class="mb-3 flex items-center gap-3">
				<h2 class="text-sm font-semibold">Source document</h2>
				{#if tx.document}
					<a
						href="/documents/{tx.document.id}"
						target="_blank"
						rel="noopener"
						class="ml-auto text-[13px] text-muted hover:text-fg"
					>
						Open in new tab ↗
					</a>
				{/if}
			</div>
			{#if tx.document}
				<p class="mb-3 truncate text-xs text-muted" title={tx.document.filename}>
					{tx.document.filename}{tx.document.detectedBank ? ` · ${tx.document.detectedBank}` : ''}
				</p>
				{#if tx.document.contentType === 'application/pdf'}
					<iframe
						src="/documents/{tx.document.id}"
						title={tx.document.filename}
						class="h-[calc(100vh-14rem)] min-h-96 w-full rounded-lg border border-line"
					></iframe>
				{:else}
					<p class="text-sm text-muted">No preview for {tx.document.contentType || 'this file type'}.</p>
				{/if}
			{:else}
				<p class="rounded-lg border border-dashed border-line px-4 py-6 text-center text-sm text-muted">
					Entered by hand, so there's no document.
				</p>
			{/if}
		</section>
	</div>
{/if}
