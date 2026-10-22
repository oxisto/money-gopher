<script lang="ts">
	import { refreshAll } from '$app/navigation';
	import ActivityFeed from '#lib/components/activity/ActivityFeed.svelte';
	import TransactionForm from '#lib/components/TransactionForm.svelte';
	import TransferForm from '#lib/components/TransferForm.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import Dialog from '#lib/components/ui/Dialog.svelte';
	import ErrorNote from '#lib/components/ui/ErrorNote.svelte';
	import Icon from '#lib/components/ui/Icon.svelte';
	import StatStrip from '#lib/components/ui/StatStrip.svelte';
	import SelectField from '#lib/components/ui/SelectField.svelte';
	import TextField from '#lib/components/ui/TextField.svelte';
	import { graphql, mutate } from '#lib/graphql/client.js';
	import { accountKindLabel } from '#lib/describe.js';
	import { formatDate, formatMoney, formatSignedMoney, parseAmount, toneOf } from '#lib/format.js';
	import type { FeedTransaction } from '#lib/transaction-types.js';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let addOpen = $state(false);
	let transferOpen = $state(false);

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
	let editKind = $state('CASH');
	let editEstimated = $state(false);
	let editDeclared = $state('');
	let editError = $state<string | null>(null);

	function openEdit() {
		editName = account?.displayName ?? '';
		editIBAN = account?.iban ?? '';
		editKind = account?.kind ?? 'CASH';
		editEstimated = account?.estimated ?? false;
		editDeclared = account?.declaredBalance ? String(account.declaredBalance.amount / 100) : '';
		editError = null;
		editOpen = true;
	}

	async function saveEdit(event: SubmitEvent) {
		event.preventDefault();
		if (!account || !editName.trim()) return;
		// A declared balance is sent only when it changed, so an unchanged
		// estimate keeps its declaration date.
		const declared = editEstimated ? parseAmount(editDeclared) : null;
		if (editEstimated && declared === null) {
			editError = 'An estimated account needs the balance you believe it has.';
			return;
		}
		const unchanged = declared !== null && declared === (account.declaredBalance?.amount ?? null);
		const result = await mutate(updateCashAccount, {
			id: account.id,
			input: {
				displayName: editName.trim(),
				iban: editIBAN.trim(),
				kind: editKind as 'CASH' | 'SAVINGS' | 'LOAN',
				estimated: editEstimated,
				declaredBalance: editEstimated && !unchanged ? declared : undefined
			}
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
		const balanceTile = account.estimated
			? {
					label: 'Balance',
					value: `≈ ${formatMoney(account.balance)}`,
					sub: `declared ${account.declaredAt ? formatDate(account.declaredAt) : ''} · ledger ${formatMoney(account.ledgerBalance)}`
				}
			: {
					label: account.kind === 'LOAN' ? 'Outstanding' : 'Balance',
					value: formatMoney(account.balance),
					sub: `${transactions.length} transactions`
				};
		return [
			balanceTile,
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
				{accountKindLabel(account.kind)} · {account.currency}{#if account.iban}
					· {account.iban}{/if}
				{#if account.estimated}
					<span class="ml-1 rounded-full bg-accent-soft px-1.5 text-[11px] font-semibold text-accent-fg" title="The balance is what you declared, not a sum of transactions">
						≈ estimated
					</span>
				{/if}
			</p>
		</div>
		<Button variant="secondary" onclick={openEdit}>Edit</Button>
		<Button variant="secondary" onclick={() => (transferOpen = true)}><Icon name="transferOut" size={15} /> Transfer</Button>
		<Button onclick={() => (addOpen = true)}><Icon name="plus" size={15} /> Add transaction</Button>
	</div>

	<Dialog bind:open={transferOpen} title="Transfer between accounts">
		<TransferForm
			{account}
			cashAccounts={data.cashAccounts}
			oncreated={() => {
				transferOpen = false;
				refresh();
			}}
		/>
	</Dialog>

	<Dialog bind:open={editOpen} title="Edit account">
		<form onsubmit={saveEdit} class="flex flex-col gap-3">
			<TextField bind:value={editName} label="Name" required />
			<TextField bind:value={editIBAN} label="IBAN" placeholder="DE00 0000 0000 0000 0000 00" />
			<p class="text-xs text-muted">
				Statements and confirmations find this account by its IBAN; spaces don't matter.
			</p>
			<SelectField bind:value={editKind} label="Kind">
				<option value="CASH">Cash account</option>
				<option value="SAVINGS">Savings account</option>
				<option value="LOAN">Loan</option>
			</SelectField>
			<label class="flex items-center gap-2 text-sm">
				<input type="checkbox" bind:checked={editEstimated} />
				The balance is an estimate
			</label>
			{#if editEstimated}
				<TextField bind:value={editDeclared} label="Approximate balance" placeholder="e.g. 1500" required />
				<p class="text-xs text-muted">
					For an account you don't track day to day: the balance is what you declare here, and
					transactions on it no longer move it. Net worth shows it marked as an estimate.
				</p>
			{/if}
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
