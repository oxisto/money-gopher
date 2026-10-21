<script lang="ts">
	import { refreshAll } from '$app/navigation';
	import ActivityFeed from '#lib/components/activity/ActivityFeed.svelte';
	import TransactionForm from '#lib/components/TransactionForm.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import Dialog from '#lib/components/ui/Dialog.svelte';
	import Icon from '#lib/components/ui/Icon.svelte';
	import type { FeedTransaction } from '#lib/transaction-types.js';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	// A trade belongs to its portfolio and settles on a cash account, so it
	// comes back from both; keep one of each.
	const transactions = $derived.by(() => {
		const byID = new Map<string, FeedTransaction>();
		for (const p of data.portfolios ?? []) for (const t of p.transactions) byID.set(t.id, t as FeedTransaction);
		for (const a of data.cashAccounts ?? []) for (const t of a.transactions) byID.set(t.id, t as FeedTransaction);
		return [...byID.values()].sort((a, b) => b.time.getTime() - a.time.getTime());
	});

	let addOpen = $state(false);

	function refresh() {
		refreshAll();
	}
</script>

<div class="mb-6 flex flex-wrap items-center gap-3">
	<h1 class="flex-1 text-lg font-semibold">Activity</h1>
	<Button onclick={() => (addOpen = true)}><Icon name="plus" size={15} /> Add transaction</Button>
</div>

<Dialog bind:open={addOpen} title="Add transaction" wide>
	<TransactionForm
		defaultType="DEPOSIT_CASH"
		securities={data.securities}
		cashAccounts={data.cashAccounts}
		oncreated={() => {
			addOpen = false;
			refresh();
		}}
	/>
</Dialog>
<ActivityFeed
	{transactions}
	securities={data.securities}
	cashAccounts={data.cashAccounts}
	showAccountFilter
	onchanged={refresh}
/>
