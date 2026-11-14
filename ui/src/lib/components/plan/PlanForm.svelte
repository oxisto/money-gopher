<script lang="ts">
	import { graphql, mutate } from '#lib/graphql/client.js';
	import { formatMoney, parseAmount, type Money } from '#lib/format.js';
	import Button from '../ui/Button.svelte';
	import Checkbox from '../ui/Checkbox.svelte';
	import ErrorNote from '../ui/ErrorNote.svelte';
	import TextField from '../ui/TextField.svelte';

	let {
		plan = null,
		netWorth,
		surplus,
		onsaved
	}: {
		/** Editing an existing plan; null creates one. */
		plan?: {
			id: string;
			displayName: string;
			start: Date | string;
			horizonMonths: number;
			annualCost: number;
			startingAmount: Money;
			startingAmountDeclared: boolean;
			contribution: Money;
			contributionDeclared: boolean;
		} | null;
		/** The net worth now, what an undeclared starting amount follows. */
		netWorth: Money;
		/** The schedules' monthly in − out, what an undeclared contribution follows. */
		surplus: Money;
		onsaved: () => void;
	} = $props();

	const createPlan = graphql(`
		mutation CreatePlan($input: CreatePlanInput!) {
			createPlan(input: $input) {
				id
			}
		}
	`);
	const updatePlan = graphql(`
		mutation UpdatePlan($id: ID!, $input: UpdatePlanInput!) {
			updatePlan(id: $id, input: $input) {
				id
			}
		}
	`);

	const monthOf = (t: Date | string) => new Date(t).toISOString().slice(0, 7);
	const major = (m: Money) => String(m.amount / 100);

	let name = $state(plan?.displayName ?? 'My plan');
	let start = $state(plan ? monthOf(plan.start) : monthOf(new Date()));
	let years = $state(String(Math.round((plan?.horizonMonths ?? 240) / 12)));
	let followNetWorth = $state(!(plan?.startingAmountDeclared ?? false));
	let startingAmount = $state(plan?.startingAmountDeclared ? major(plan.startingAmount) : major(netWorth));
	let followSchedules = $state(!(plan?.contributionDeclared ?? false));
	let contribution = $state(plan?.contributionDeclared ? major(plan.contribution) : major(surplus));
	let cost = $state(String((plan?.annualCost ?? 0) * 100));
	let error = $state<string | null>(null);
	let saving = $state(false);

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		error = null;
		const horizonMonths = Math.round(parseFloat(years) * 12);
		if (!(horizonMonths >= 1 && horizonMonths <= 600)) {
			error = 'The horizon must be between one month and fifty years.';
			return;
		}
		const amount = followNetWorth ? null : parseAmount(startingAmount);
		if (!followNetWorth && amount === null) {
			error = 'Enter the starting amount, or let it follow your net worth.';
			return;
		}
		const monthly = followSchedules ? null : parseAmount(contribution.replace(/^-/, ''));
		if (!followSchedules && monthly === null) {
			error = 'Enter the monthly contribution, or let it follow your schedules.';
			return;
		}
		const annualCost = parseFloat(cost || '0') / 100;
		if (!(annualCost >= 0 && annualCost <= 0.1)) {
			error = 'The annual cost must be between 0 and 10 %.';
			return;
		}
		if (!/^\d{4}-\d{2}$/.test(start)) {
			error = 'Pick the month the plan starts in.';
			return;
		}
		saving = true;
		const common = {
			displayName: name.trim(),
			start: new Date(`${start}-01T00:00:00Z`),
			horizonMonths,
			annualCost,
			startingAmount: amount,
			contribution: monthly === null ? null : contribution.startsWith('-') ? -monthly : monthly
		};
		const result = plan
			? await mutate(updatePlan, {
					id: plan.id,
					input: { ...common, clearStartingAmount: followNetWorth, clearContribution: followSchedules }
				})
			: await mutate(createPlan, { input: common });
		saving = false;
		error = result.error;
		if (!error) onsaved();
	}
</script>

<form onsubmit={submit} class="flex flex-col gap-3">
	<TextField bind:value={name} label="Name" required />
	<div class="grid grid-cols-2 gap-3">
		<label class="block text-sm">
			<span class="mb-1 block text-xs font-medium text-muted">Starts in</span>
			<input type="month" bind:value={start} required class="h-9 w-full px-3 py-1.5" />
		</label>
		<TextField bind:value={years} label="Horizon (years)" type="number" step="1" required />
	</div>

	<div class="flex flex-col gap-2 rounded-lg border border-line p-3">
		<Checkbox bind:checked={followNetWorth} label="Start from my net worth" />
		{#if followNetWorth}
			<p class="text-xs text-muted">
				Your net worth as the books have it at the start month; ≈ {formatMoney(netWorth)} today.
			</p>
		{:else}
			<TextField bind:value={startingAmount} label="Starting amount ({netWorth.currency})" placeholder="e.g. 25000" />
		{/if}
	</div>

	<div class="flex flex-col gap-2 rounded-lg border border-line p-3">
		<Checkbox bind:checked={followSchedules} label="Contribute what my schedules leave" />
		{#if followSchedules}
			<p class="text-xs text-muted">
				The next twelve months' in − out from your schedules, averaged: {formatMoney(surplus)} a month right now. It follows the schedules as they change.
			</p>
		{:else}
			<TextField bind:value={contribution} label="Monthly contribution ({surplus.currency})" placeholder="e.g. 1500" />
		{/if}
	</div>

	<TextField bind:value={cost} label="Annual cost (%)" type="number" step="0.01" placeholder="e.g. 0.2" />
	<p class="text-xs text-muted">Fees and taxes taken off every growing bucket's return each year; 0.2 % for a cheap ETF portfolio.</p>

	<ErrorNote message={error} />
	<div class="flex justify-end">
		<Button type="submit" disabled={saving || !name.trim()}>{saving ? 'Saving…' : plan ? 'Save' : 'Create plan'}</Button>
	</div>
</form>
