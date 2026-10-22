<script lang="ts">
	import { refreshAll } from '$app/navigation';
	import ActivityFeed from '#lib/components/activity/ActivityFeed.svelte';
	import TransactionForm from '#lib/components/TransactionForm.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import Dialog from '#lib/components/ui/Dialog.svelte';
	import ErrorNote from '#lib/components/ui/ErrorNote.svelte';
	import Icon from '#lib/components/ui/Icon.svelte';
	import StatStrip from '#lib/components/ui/StatStrip.svelte';
	import TextField from '#lib/components/ui/TextField.svelte';
	import { graphql, mutate } from '#lib/graphql/client.js';
	import { formatMoney, formatSignedMoney, toneOf } from '#lib/format.js';
	import type { FeedTransaction } from '#lib/transaction-types.js';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let addOpen = $state(false);

	const updateCashAccount = graphql(`
		mutation EditCashAccount($id: ID!, $input: UpdateCashAccountInput!) {
			updateCashAccount(id: $id, input: $input) {
				id
				displayName
				iban
			}
		}
	`);
	let editOpen = $state(false);
	let editName = $state('');
	let editIBAN = $state('');
	let editError = $state<string | null>(null);

	function openEdit() {
		editName = account?.displayName ?? '';
		editIBAN = account?.iban ?? '';
		editError = null;
		editOpen = true;
	}

	async function saveEdit(event: SubmitEvent) {
		event.preventDefault();
		if (!account || !editName.trim()) return;
		const result = await mutate(updateCashAccount, {
			id: account.id,
			input: { displayName: editName.trim(), iban: editIBAN.trim() }
		});
		editError = result.error;
		if (!editError) {
			editOpen = false;
			await refreshAll();
		}
	}

	function refresh() {
		refreshAll();
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
		<Button variant="secondary" onclick={openEdit}>Edit</Button>
		<Button onclick={() => (addOpen = true)}><Icon name="plus" size={15} /> Add transaction</Button>
	</div>

	<Dialog bind:open={editOpen} title="Edit account">
		<form onsubmit={saveEdit} class="flex flex-col gap-3">
			<TextField bind:value={editName} label="Name" required />
			<TextField bind:value={editIBAN} label="IBAN" placeholder="DE00 0000 0000 0000 0000 00" />
			<p class="text-xs text-muted">
				Statements and confirmations find this account by its IBAN; spaces don't matter.
			</p>
			<ErrorNote message={editError} />
			<div class="mt-2 flex justify-end gap-2">
				<Button variant="secondary" onclick={() => (editOpen = false)}>Cancel</Button>
				<Button type="submit" disabled={!editName.trim()}>Save</Button>
			</div>
		</form>
	</Dialog>

	<Dialog bind:open={addOpen} title="Add transaction" wide>
		<TransactionForm
			defaultCashAccountID={account.id}
			defaultType="DEPOSIT_CASH"
			securities={[...data.securities, ...data.assets]}
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
			securities={[...data.securities, ...data.assets]}
			cashAccounts={data.cashAccounts}
			showBalance
			onchanged={refresh}
		/>
	</div>
{:else}
	<p class="text-sm text-muted">Cash account not found.</p>
{/if}
