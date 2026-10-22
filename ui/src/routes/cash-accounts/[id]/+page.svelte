<script lang="ts">
	import { refreshAll } from '$app/navigation';
	import ActivityFeed from '#lib/components/activity/ActivityFeed.svelte';
	import TransactionForm from '#lib/components/TransactionForm.svelte';
	import TransferForm from '#lib/components/TransferForm.svelte';
	import ScheduleForm from '#lib/components/ScheduleForm.svelte';
	type ScheduleFormPreset = { name: string; type: string; dayOfMonth: number; start: string; steps: { from: string; amount: string }[]; books: boolean };
	import Panel from '#lib/components/ui/Panel.svelte';
	import Checkbox from '#lib/components/ui/Checkbox.svelte';
	import { intervalLabel, scheduleTypeLabel } from '#lib/describe.js';
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

	// Schedules: the panel lists what lands here or is paid from here.
	type Schedule = NonNullable<typeof account>['schedules'][number];
	let scheduleOpen = $state(false);
	let editingSchedule = $state<Schedule | null>(null);
	type Suggestion = NonNullable<typeof account>['suggestedSchedules'][number];
	let preset = $state<ScheduleFormPreset | null>(null);
	const suggestions = $derived((data.cashAccount?.suggestedSchedules ?? []).filter((s) => !s.covered));
	function dateOnly(t: Date | string): string {
		return new Date(t).toISOString().slice(0, 10);
	}
	/** Opens the form for a new schedule shaped like the suggestion. An estimated account gets expectations only. */
	function adopt(s: Suggestion) {
		preset = {
			name: s.name,
			type: s.incoming ? 'DEPOSIT_CASH' : 'WITHDRAW_CASH',
			dayOfMonth: s.dayOfMonth,
			start: dateOnly(s.start),
			steps: s.steps.map((st) => ({ from: dateOnly(st.from), amount: String(st.amount.amount / 100) })),
			books: !data.cashAccount?.estimated
		};
		editingSchedule = null;
		scheduleError = null;
		scheduleOpen = true;
	}
	let deletingSchedule = $state<string | null>(null);
	let scheduleError = $state<string | null>(null);
	const deleteSchedule = graphql(`
		mutation DeleteSchedule($id: ID!) {
			deleteSchedule(id: $id)
		}
	`);
	const bookSchedules = graphql(`
		mutation BookSchedules {
			bookSchedules
		}
	`);
	function openSchedule(s: Schedule | null) {
		editingSchedule = s;
		preset = null;
		scheduleError = null;
		scheduleOpen = true;
	}
	async function removeSchedule(id: string) {
		const result = await mutate(deleteSchedule, { id });
		scheduleError = result.error;
		deletingSchedule = null;
		if (!scheduleError) refresh();
	}
	async function bookNow() {
		const result = await mutate(bookSchedules, {});
		scheduleError = result.error;
		if (!scheduleError) refresh();
	}
	/** The schedule's amount as it hits this account: a transfer paid from here leaves. */
	function amountHere(s: Schedule) {
		const out = s.type === 'TRANSFER' && s.cashAccount.id !== account?.id;
		return { amount: out ? -s.currentAmount.amount : s.currentAmount.amount, currency: s.currentAmount.currency };
	}
	function describeSchedule(s: Schedule): string {
		if (s.type === 'TRANSFER') {
			return s.cashAccount.id === account?.id
				? `from ${s.counterpartAccount?.displayName ?? '…'}`
				: `to ${s.cashAccount.displayName}`;
		}
		if (s.type === 'DIVIDEND') return `from ${s.asset?.displayName ?? '…'}`;
		return scheduleTypeLabel(s.type, account?.kind).toLowerCase();
	}

	const updateCashAccount = graphql(`
		mutation EditCashAccount($id: ID!, $input: UpdateCashAccountInput!) {
			updateCashAccount(id: $id, input: $input) {
				id
				displayName
				iban
				reference
			}
		}
	`);
	let editOpen = $state(false);
	let editName = $state('');
	let editIBAN = $state('');
	let editReference = $state('');
	let editKind = $state('CASH');
	let editEstimated = $state(false);
	let editDeclared = $state('');
	let editError = $state<string | null>(null);

	function openEdit() {
		editName = account?.displayName ?? '';
		editIBAN = account?.iban ?? '';
		editReference = account?.reference ?? '';
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
				reference: editReference.trim(),
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
					· {account.iban}{/if}{#if account.reference}
					· No. {account.reference}{/if}
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

	<Dialog bind:open={scheduleOpen} title={editingSchedule ? 'Edit schedule' : 'New schedule'}>
		{#key editingSchedule?.id ?? preset?.name ?? 'new'}
			<ScheduleForm
				{account}
				cashAccounts={data.cashAccounts}
				assets={[...data.securities, ...data.assets]}
				schedule={editingSchedule}
				{preset}
				onsaved={() => {
					scheduleOpen = false;
					refresh();
				}}
			/>
		{/key}
	</Dialog>

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
			<TextField bind:value={editReference} label="Account or contract number" placeholder="e.g. a Bausparvertrag's Vertragsnummer" />
			<p class="text-xs text-muted">
				For statements that print a contract or account number instead of your IBAN. Two
				accounts may share one, like a Bausparvertrag and the loan against it.
			</p>
			<SelectField bind:value={editKind} label="Kind">
				<option value="CASH">Cash account</option>
				<option value="SAVINGS">Savings account</option>
				<option value="LOAN">Loan</option>
			</SelectField>
			<Checkbox bind:checked={editEstimated} label="The balance is an estimate" />
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
		<Panel title="Expected cash flows">
			{#snippet aside()}
				<div class="flex items-center gap-3 text-xs">
					{#if account.schedules.some((s) => s.books)}
						<button class="text-muted hover:text-fg" onclick={bookNow} title="Write every due occurrence now; moneyd does this by itself a few times a day">Book due now</button>
					{/if}
					<button class="text-accent-fg hover:underline" onclick={() => openSchedule(null)}>+ Add</button>
				</div>
			{/snippet}
			{#if account.schedules.length === 0}
				<p class="text-sm text-muted">
					Nothing recurring yet. A salary, a Sparrate, a loan's interest: ≈ estimates for the
					cash-flow view, and booked for you when you switch that on.
				</p>
			{:else}
				<table class="num w-full text-[13px]">
					<tbody class="divide-y divide-line">
						{#each account.schedules as s (s.id)}
							{@const amount = amountHere(s)}
							<tr>
								<td class="py-2 pr-3">
									<div class="font-medium">{s.displayName}</div>
									<div class="text-xs text-muted">
										{describeSchedule(s)} · {intervalLabel(s.interval)} on the {s.dayOfMonth}.{#if s.end}
											· until {formatDate(s.end)}{/if}
									</div>
								</td>
								<td class="py-2 pr-3 text-right tabular-nums {amount.amount > 0 ? 'text-pos' : ''}">
									{formatSignedMoney(amount)}
								</td>
								<td class="py-2 pr-3 text-right text-xs text-muted whitespace-nowrap">
									{#if s.nextDate}next {formatDate(s.nextDate)}{:else}ended{/if}
								</td>
								<td class="py-2 pr-3 text-right text-xs whitespace-nowrap">
									{#if s.books}
										<span class="text-muted">{s.booked} booked</span>
										{#if s.unconfirmed.length > 0}
											<span class="ml-1 rounded-full bg-accent-soft px-1.5 text-[11px] font-semibold text-accent-fg" title="Booked by the schedule; no statement has confirmed these yet">{s.unconfirmed.length} unconfirmed</span>
										{/if}
									{:else}
										<span class="rounded-full bg-raised px-1.5 text-[11px] font-semibold text-muted">≈ estimate</span>
									{/if}
								</td>
								<td class="py-2 text-right text-xs whitespace-nowrap">
									<button class="text-muted hover:text-fg" onclick={() => openSchedule(s)}>Edit</button>
									{#if deletingSchedule === s.id}
										<button class="ml-2 text-neg hover:underline" onclick={() => removeSchedule(s.id)}>Really delete</button>
										<button class="ml-2 text-muted hover:text-fg" onclick={() => (deletingSchedule = null)}>Keep</button>
									{:else}
										<button class="ml-2 text-muted hover:text-neg" onclick={() => (deletingSchedule = s.id)}>Delete</button>
									{/if}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
				<p class="text-xs text-muted">Booked transactions stay when a schedule is deleted.</p>
			{/if}
			{#if suggestions.length > 0}
				<div class="mt-2 border-t border-line pt-3">
					<p class="mb-2 text-xs font-medium text-muted" title="Read off the statements: a counterparty that recurs on a rhythm, with its amounts split where they changed">
						Seen recurring on this account
					</p>
					<table class="num w-full text-[13px]">
						<tbody class="divide-y divide-line">
							{#each suggestions as s (s.name + s.incoming)}
								{@const latest = s.steps[s.steps.length - 1]}
								<tr>
									<td class="py-1.5 pr-3">
										<div class="truncate font-medium" title={s.name}>{s.name}</div>
										<div class="text-xs text-muted">
											{s.incoming ? 'in' : 'out'} · {intervalLabel(s.interval)} on the {s.dayOfMonth}. · {s.count}× since {formatDate(s.start)}{#if s.steps.length > 1}
												· {s.steps.length} amounts{/if}{#if s.variable}
												· varies{/if}
										</div>
									</td>
									<td class="py-1.5 pr-3 text-right tabular-nums {s.incoming ? 'text-pos' : ''}">
										{s.variable ? '≈ ' : ''}{formatSignedMoney({ amount: (s.incoming ? 1 : -1) * latest.amount.amount, currency: latest.amount.currency })}
									</td>
									<td class="py-1.5 text-right text-xs whitespace-nowrap">
										<button class="text-accent-fg hover:underline" onclick={() => adopt(s)}>Create schedule</button>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}
			<ErrorNote message={scheduleError} />
		</Panel>
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
