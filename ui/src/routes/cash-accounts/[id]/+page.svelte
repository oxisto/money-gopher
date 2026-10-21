<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import ActivityFeed from '$lib/components/activity/ActivityFeed.svelte';
	import TransactionForm from '$lib/components/TransactionForm.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Dialog from '$lib/components/ui/Dialog.svelte';
	import Icon from '$lib/components/ui/Icon.svelte';
	import StatStrip from '$lib/components/ui/StatStrip.svelte';
	import { formatMoney, formatSignedMoney, toneOf } from '$lib/format';
	import type { FeedTransaction } from '$lib/transaction-types';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let addOpen = $state(false);

	function refresh() {
		invalidateAll();
	}

	const account = $derived(data.cashAccount ?? null);
	const transactions = $derived((account?.transactions ?? []) as FeedTransaction[]);

	const stats = $derived.by(() => {
		if (!account) return [];
		const year = new Date().getFullYear();
		const thisYear = transactions.filter((t) => t.time.getFullYear() === year);
		const money = (amount: number) => ({ amount, currency: account.currency });
		const sum = (f: (t: FeedTransaction) => boolean) =>
			thisYear.filter(f).reduce((s, t) => s + t.cashDelta.amount, 0);
		const inflow = sum((t) => t.cashDelta.amount > 0);
		const outflow = sum((t) => t.cashDelta.amount < 0);
		const interest = sum((t) => t.type === 'INTEREST');
		return [
			{ label: 'Balance', value: formatMoney(account.balance), sub: `${transactions.length} transactions` },
			{ label: `In ${year}`, value: formatSignedMoney(money(inflow)), tone: toneOf(inflow) },
			{ label: `Out ${year}`, value: formatSignedMoney(money(outflow)) },
			{ label: `Interest ${year}`, value: formatSignedMoney(money(interest)), tone: toneOf(interest) }
		];
	});
</script>

{#if account}
	<div class="mb-6 flex flex-wrap items-center gap-3">
		<div class="flex-1">
			<h1 class="text-lg font-semibold">{account.displayName}</h1>
			<p class="num text-[13px] text-muted">
				Cash account · {account.currency}{#if account.iban}
					· {account.iban}{/if}
			</p>
		</div>
		<Button onclick={() => (addOpen = true)}><Icon name="plus" size={15} /> Add transaction</Button>
	</div>

	<Dialog bind:open={addOpen} title="Add transaction" wide>
		<TransactionForm
			defaultCashAccountID={account.id}
			defaultType="DEPOSIT_CASH"
			securities={data.securities}
			cashAccounts={data.cashAccounts}
			oncreated={() => {
				addOpen = false;
				refresh();
			}}
		/>
	</Dialog>

	<div class="flex flex-col gap-6">
		<StatStrip {stats} />
		<ActivityFeed
			{transactions}
			securities={data.securities}
			cashAccounts={data.cashAccounts}
			showBalance
			onchanged={refresh}
		/>
	</div>
{:else}
	<p class="text-sm text-muted">Cash account not found.</p>
{/if}
