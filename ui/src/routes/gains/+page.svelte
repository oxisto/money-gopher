<script lang="ts">
	import Section from '$lib/components/ui/Section.svelte';
	import { formatDate, formatMoney } from '$lib/format';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	const RealizedGains = $derived(data.RealizedGains);

	const currentYear = new Date().getFullYear();
	const years = [currentYear, currentYear - 1, currentYear - 2, currentYear - 3, currentYear - 4];

	let selectedYear = $state(currentYear);

	function loadYear(year: number) {
		selectedYear = year;
		const from = new Date(Date.UTC(year, 0, 1));
		// Exclusive upper bound: the instant January 1 of the following year
		// begins, so December 31 of `year` is fully included.
		const to = new Date(Date.UTC(year + 1, 0, 1));
		RealizedGains.fetch({ variables: { from, to }, policy: 'NetworkOnly' });
	}

	function entryHref(entry: { id: string }) {
		return `/transactions/${entry.id}`;
	}
</script>

<div class="mb-6 flex items-center justify-between">
	<h1 class="text-2xl font-semibold">Realized Gains</h1>
	<select
		class="rounded-md border border-gray-300 px-2 py-1 text-sm"
		value={selectedYear}
		onchange={(e) => loadYear(Number(e.currentTarget.value))}
	>
		{#each years as year (year)}
			<option value={year}>{year}</option>
		{/each}
	</select>
</div>

{#if $RealizedGains.fetching && !$RealizedGains.data}
	<p class="text-sm text-gray-500">Loading…</p>
{:else if $RealizedGains.errors?.length}
	<p class="text-sm text-red-600">Failed to load: {$RealizedGains.errors[0].message}</p>
{:else if $RealizedGains.data}
	{@const gains = $RealizedGains.data.realizedGains}

	<div class="mb-8 grid grid-cols-2 gap-4 sm:grid-cols-4">
		<div class="rounded-lg border border-gray-200 p-4">
			<p class="text-xs font-medium text-gray-500 uppercase">Dividends</p>
			<p class="mt-1 text-xl font-semibold tabular-nums">{formatMoney(gains.dividends)}</p>
		</div>
		<div class="rounded-lg border border-gray-200 p-4">
			<p class="text-xs font-medium text-gray-500 uppercase">Interest</p>
			<p class="mt-1 text-xl font-semibold tabular-nums">{formatMoney(gains.interest)}</p>
		</div>
		<div class="rounded-lg border border-gray-200 p-4">
			<p class="text-xs font-medium text-gray-500 uppercase">Trading gains</p>
			<p
				class="mt-1 text-xl font-semibold tabular-nums {gains.tradingGains.amount < 0
					? 'text-red-600'
					: ''}"
			>
				{formatMoney(gains.tradingGains)}
			</p>
		</div>
		<div class="rounded-lg border border-gray-200 bg-gray-50 p-4">
			<p class="text-xs font-medium text-gray-500 uppercase">Total</p>
			<p
				class="mt-1 text-xl font-semibold tabular-nums {gains.total.amount < 0
					? 'text-red-600'
					: 'text-green-700'}"
			>
				{formatMoney(gains.total)}
			</p>
		</div>
	</div>

	<Section title="Events">
		{#if gains.entries.length === 0}
			<p class="text-sm text-gray-500">No realized gains in {selectedYear}.</p>
		{:else}
			<div class="overflow-x-auto rounded-md border border-gray-200 bg-white">
				<table class="w-full text-sm">
					<thead>
						<tr class="border-b border-gray-200 text-left text-xs text-gray-500 uppercase">
							<th class="px-4 py-2 font-medium">Date</th>
							<th class="px-4 py-2 font-medium">Type</th>
							<th class="px-4 py-2 font-medium">Security</th>
							<th class="px-4 py-2 text-right font-medium">Amount</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100">
						{#each gains.entries as entry (entry.id)}
							<tr class="hover:bg-gray-50">
								<td class="px-4 py-2 whitespace-nowrap">
									<a href={entryHref(entry)} class="text-gray-900 hover:underline"
										>{formatDate(entry.time)}</a
									>
								</td>
								<td class="px-4 py-2">
									<span class="rounded bg-gray-100 px-1.5 py-0.5 text-xs font-mono">{entry.type}</span>
								</td>
								<td class="px-4 py-2">{entry.security?.displayName ?? '—'}</td>
								<td
									class="px-4 py-2 text-right tabular-nums {entry.amount.amount < 0
										? 'text-red-600'
										: 'text-green-700'}"
								>
									{formatMoney(entry.amount)}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</Section>
{/if}
