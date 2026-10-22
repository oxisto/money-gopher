<script lang="ts">
	import { graphql, mutate } from '#lib/graphql/client.js';
	import { parseAmount } from '#lib/format.js';
	import type { ScheduleFields } from '#lib/graphql/documents.js';
	import type { ResultOf } from 'gql.tada';
	import Button from './ui/Button.svelte';
	import Checkbox from './ui/Checkbox.svelte';
	import ErrorNote from './ui/ErrorNote.svelte';
	import Segmented from './ui/Segmented.svelte';
	import SelectField from './ui/SelectField.svelte';
	import TextField from './ui/TextField.svelte';

	type Schedule = ResultOf<typeof ScheduleFields>;

	let {
		account,
		cashAccounts,
		assets,
		schedule = null,
		onsaved
	}: {
		/** The account whose page the form sits on. */
		account: { id: string; displayName: string; currency: string; kind?: string | null };
		cashAccounts: { id: string; displayName: string; currency: string }[];
		assets: { id: string; displayName: string }[];
		/** Editing an existing schedule; null creates one. */
		schedule?: Schedule | null;
		onsaved: () => void;
	} = $props();

	const createSchedule = graphql(`
		mutation CreateSchedule($input: CreateScheduleInput!) {
			createSchedule(input: $input) {
				id
			}
		}
	`);
	const updateSchedule = graphql(`
		mutation UpdateSchedule($id: ID!, $input: UpdateScheduleInput!) {
			updateSchedule(id: $id, input: $input) {
				id
			}
		}
	`);

	const types: { value: string; label: string }[] = [
		{ value: 'TRANSFER', label: 'Transfer between my accounts' },
		{ value: 'DEPOSIT_CASH', label: 'Deposit (salary, rent received)' },
		{ value: 'WITHDRAW_CASH', label: 'Withdrawal (card bill, rent paid)' },
		{ value: 'INTEREST', label: account.kind === 'LOAN' ? 'Interest charged' : 'Interest' },
		{ value: 'ACCOUNT_FEES', label: 'Fees' },
		{ value: 'DIVIDEND', label: 'Dividend or rent from an asset' }
	];

	// An existing transfer schedule lands on its cashAccount; seen from the
	// other account it is money going out.
	const editingOut = $derived(schedule?.type === 'TRANSFER' && schedule.cashAccount.id !== account.id);

	function dateOf(t: Date | string | null | undefined): string {
		if (!t) return '';
		const d = new Date(t);
		return d.toISOString().slice(0, 10);
	}
	function dateInput(s: string): Date {
		return new Date(`${s}T00:00:00Z`);
	}

	let name = $state(schedule?.displayName ?? '');
	let type = $state(schedule?.type ?? 'TRANSFER');
	let direction = $state<'in' | 'out'>(editingOut ? 'out' : 'in');
	let otherID = $state(
		editingOut ? (schedule?.cashAccount.id ?? '') : (schedule?.counterpartAccount?.id ?? '')
	);
	let assetID = $state(schedule?.asset?.id ?? '');
	let interval = $state(schedule?.interval ?? 'MONTHLY');
	let day = $state(String(schedule?.dayOfMonth ?? 1));
	let start = $state(dateOf(schedule?.start) || new Date().toISOString().slice(0, 10));
	let end = $state(dateOf(schedule?.end));
	let books = $state(schedule?.books ?? true);
	let steps = $state<{ from: string; amount: string }[]>(
		schedule?.steps.map((s) => ({ from: dateOf(s.from), amount: String(Math.abs(s.amount.amount) / 100) })) ?? [
			{ from: new Date().toISOString().slice(0, 10), amount: '' }
		]
	);
	let error = $state<string | null>(null);
	let saving = $state(false);

	const others = $derived(cashAccounts.filter((a) => a.id !== account.id && a.currency === account.currency));

	function addStep() {
		steps = [...steps, { from: '', amount: '' }];
	}
	function removeStep(i: number) {
		steps = steps.filter((_, j) => j !== i);
	}

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		error = null;
		const parsedSteps: { from: Date; amount: number }[] = [];
		for (const s of steps) {
			const minor = parseAmount(s.amount);
			if (!s.from || minor === null || minor <= 0) {
				error = 'Every amount needs a date and a positive value.';
				return;
			}
			parsedSteps.push({ from: dateInput(s.from), amount: minor });
		}
		if (type === 'TRANSFER' && !otherID) {
			error = 'Pick the other account.';
			return;
		}
		if (type === 'DIVIDEND' && !assetID) {
			error = 'Pick the asset that pays.';
			return;
		}
		const dayOfMonth = parseInt(day, 10);
		if (!(dayOfMonth >= 1 && dayOfMonth <= 31)) {
			error = 'The day must be between 1 and 31.';
			return;
		}
		saving = true;
		const common = {
			displayName: name.trim(),
			interval: interval as 'MONTHLY' | 'QUARTERLY' | 'YEARLY',
			dayOfMonth,
			start: dateInput(start),
			end: end ? dateInput(end) : null,
			steps: parsedSteps,
			books: type === 'DIVIDEND' ? false : books
		};
		const result = schedule
			? await mutate(updateSchedule, {
					id: schedule.id,
					input: {
						...common,
						clearEnd: !end,
						counterpartAccountID:
							type === 'TRANSFER' ? (direction === 'in' ? otherID : account.id) : undefined
					}
				})
			: await mutate(createSchedule, {
					input: {
						...common,
						type: type as 'TRANSFER' | 'DEPOSIT_CASH' | 'WITHDRAW_CASH' | 'INTEREST' | 'ACCOUNT_FEES' | 'DIVIDEND',
						// A transfer lands on cashAccountID and is paid from the counterpart.
						cashAccountID: type === 'TRANSFER' && direction === 'out' ? otherID : account.id,
						counterpartAccountID:
							type === 'TRANSFER' ? (direction === 'in' ? otherID : account.id) : undefined,
						assetID: type === 'DIVIDEND' ? assetID : undefined
					}
				});
		saving = false;
		error = result.error;
		if (!error) onsaved();
	}
</script>

<form onsubmit={submit} class="flex flex-col gap-3">
	<TextField bind:value={name} label="Name" placeholder="e.g. Sparrate, Gehalt, Zinsen" required />
	{#if schedule}
		<p class="text-xs text-muted">{types.find((t) => t.value === type)?.label}. The kind and the account stay; make a new schedule for a different one.</p>
	{:else}
		<SelectField bind:value={type} label="Kind" required>
			{#each types as t (t.value)}
				<option value={t.value}>{t.label}</option>
			{/each}
		</SelectField>
	{/if}
	{#if type === 'TRANSFER'}
		{#if !schedule}
			<Segmented
				label="Direction"
				bind:value={direction}
				options={[
					{ value: 'in', label: `Into ${account.displayName}` },
					{ value: 'out', label: `Out of ${account.displayName}` }
				]}
			/>
		{/if}
		<SelectField bind:value={otherID} label={direction === 'in' ? 'Paid from' : 'Paid into'} required>
			<option value="">Choose an account…</option>
			{#each others as a (a.id)}
				<option value={a.id}>{a.displayName}</option>
			{/each}
		</SelectField>
	{:else if type === 'DIVIDEND' && !schedule}
		<SelectField bind:value={assetID} label="Paid by" required>
			<option value="">Choose an asset…</option>
			{#each assets as a (a.id)}
				<option value={a.id}>{a.displayName}</option>
			{/each}
		</SelectField>
	{/if}
	{#if type !== 'DIVIDEND'}
		<Checkbox bind:checked={books} label="Book automatically" />
		<p class="text-xs text-muted">
			Each occurrence is written as a transaction once its day has passed, past ones right away. A
			statement that lists the same movement later confirms it instead of adding a second one.
		</p>
	{:else}
		<p class="text-xs text-muted">
			An expectation only: the dividend or rent confirmations book the real income.
		</p>
	{/if}
	<div class="grid grid-cols-2 gap-3">
		<SelectField bind:value={interval} label="Every">
			<option value="MONTHLY">Month</option>
			<option value="QUARTERLY">Quarter</option>
			<option value="YEARLY">Year</option>
		</SelectField>
		<TextField bind:value={day} label="On day" type="number" required />
	</div>
	<div class="grid grid-cols-2 gap-3">
		<TextField bind:value={start} label="From" type="date" required />
		<TextField bind:value={end} label="Until (optional)" type="date" />
	</div>

	<div class="flex flex-col gap-2">
		<span class="text-xs font-medium text-muted">Amount, and from when</span>
		{#each steps as step, i (i)}
			<div class="flex items-end gap-2">
				<TextField bind:value={step.from} label={i === 0 ? 'From' : undefined} type="date" required />
				<TextField bind:value={step.amount} label={i === 0 ? `Amount (${account.currency})` : undefined} placeholder="e.g. 535.82" required />
				{#if steps.length > 1}
					<Button variant="secondary" type="button" onclick={() => removeStep(i)}>Remove</Button>
				{/if}
			</div>
		{/each}
		<button type="button" class="self-start text-xs text-accent-fg hover:underline" onclick={addStep}>
			+ Add a change (a raise, a new Sparrate)
		</button>
	</div>

	<ErrorNote message={error} />
	<div class="flex justify-end">
		<Button type="submit" disabled={saving || !name.trim()}>{saving ? 'Saving…' : schedule ? 'Save' : 'Create schedule'}</Button>
	</div>
</form>
