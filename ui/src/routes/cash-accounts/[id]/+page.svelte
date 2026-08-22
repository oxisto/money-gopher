<script lang="ts">
	import { formatMoney } from '$lib/format';
	import TransactionForm from '$lib/components/TransactionForm.svelte';
	import TransactionTable from '$lib/components/TransactionTable.svelte';
	import Section from '$lib/components/ui/Section.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	const CashAccountDetail = $derived(data.CashAccountDetail);

	function refresh() {
		CashAccountDetail.fetch({ policy: 'NetworkOnly' });
	}
</script>

<!-- Only the initial load blanks the page; refetches keep the old data
     (and the components' state) on screen. -->
{#if $CashAccountDetail.fetching && !$CashAccountDetail.data}
	<p class="text-sm text-gray-500">Loading…</p>
{:else if $CashAccountDetail.errors?.length}
	<p class="text-sm text-red-600">Failed to load: {$CashAccountDetail.errors[0].message}</p>
{:else if $CashAccountDetail.data}
	{@const account = $CashAccountDetail.data.cashAccount}
	{#if account}
		<div class="mb-6 flex items-baseline justify-between">
			<div>
				<h1 class="text-2xl font-semibold">{account.displayName}</h1>
				{#if account.iban}
					<p class="mt-1 font-mono text-sm text-gray-500">{account.iban}</p>
				{/if}
			</div>
			<span class="text-xl font-semibold tabular-nums">{formatMoney(account.balance)}</span>
		</div>

		<div class="flex flex-col gap-8">
			<Section title="Transactions">
				<TransactionTable
					transactions={account.transactions}
					securities={$CashAccountDetail.data.securities}
					cashAccounts={$CashAccountDetail.data.cashAccounts}
					showBalance
					onchanged={refresh}
				/>
			</Section>

			<Section title="Add transaction">
				<TransactionForm
					defaultCashAccountID={account.id}
					defaultType="DEPOSIT_CASH"
					securities={$CashAccountDetail.data.securities}
					cashAccounts={$CashAccountDetail.data.cashAccounts}
					oncreated={refresh}
				/>
			</Section>
		</div>
	{:else}
		<p class="text-sm text-gray-500">Cash account not found.</p>
	{/if}
{/if}
