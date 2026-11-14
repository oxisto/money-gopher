<script lang="ts">
	import PlanChart from '#lib/components/plan/PlanChart.svelte';
	import PlanForm from '#lib/components/plan/PlanForm.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import Dialog from '#lib/components/ui/Dialog.svelte';
	import ErrorNote from '#lib/components/ui/ErrorNote.svelte';
	import Panel from '#lib/components/ui/Panel.svelte';
	import Segmented from '#lib/components/ui/Segmented.svelte';
	import StatStrip from '#lib/components/ui/StatStrip.svelte';
	import { goto, refreshAll } from '$app/navigation';
	import { planBucketLabel } from '#lib/describe.js';
	import { formatMoney, formatMoneyCompact, formatPercent, formatSignedMoney, toneOf } from '#lib/format.js';
	import { graphql, mutate } from '#lib/graphql/client.js';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	const plan = $derived(data.plan);
	type Plan = NonNullable<typeof data.plan>;

	let formOpen = $state(false);
	let editing = $state<Plan | null>(null);
	let deleteOpen = $state(false);
	let error = $state<string | null>(null);

	const deletePlan = graphql(`
		mutation DeletePlan($id: ID!) {
			deletePlan(id: $id)
		}
	`);
	const updateAssumptions = graphql(`
		mutation UpdatePlanAssumptions($id: ID!, $assumptions: [PlanAssumptionInput!]!) {
			updatePlan(id: $id, input: { assumptions: $assumptions }) {
				id
			}
		}
	`);

	function openForm(p: Plan | null) {
		editing = p;
		formOpen = true;
	}
	async function remove() {
		if (!plan) return;
		const result = await mutate(deletePlan, { id: plan.id });
		error = result.error;
		deleteOpen = false;
		if (!error) await goto('/plan', { invalidateAll: true });
	}

	// Assumptions are edited in place: every shown bucket's return,
	// volatility and share, saved in one go.
	let editingAssumptions = $state(false);
	let showEmpty = $state(false);
	let rows = $state<{ bucket: string; weight: string; annualReturn: string; volatility: string }[]>([]);
	let savingAssumptions = $state(false);
	const pct = (f: number) => String(Math.round(f * 10000) / 100);
	function startEditing() {
		if (!plan) return;
		rows = plan.assumptions.map((a) => ({
			bucket: a.bucket,
			weight: pct(a.weight),
			annualReturn: pct(a.annualReturn),
			volatility: pct(a.volatility)
		}));
		showEmpty = true;
		editingAssumptions = true;
	}
	async function saveAssumptions() {
		if (!plan) return;
		error = null;
		const assumptions = rows.map((r) => ({
			bucket: r.bucket as Plan['assumptions'][number]['bucket'],
			weight: parseFloat(r.weight || '0') / 100,
			annualReturn: parseFloat(r.annualReturn || '0') / 100,
			volatility: parseFloat(r.volatility || '0') / 100
		}));
		if (assumptions.some((a) => [a.weight, a.annualReturn, a.volatility].some((v) => Number.isNaN(v)))) {
			error = 'Every field needs a number.';
			return;
		}
		savingAssumptions = true;
		const result = await mutate(updateAssumptions, { id: plan.id, assumptions });
		savingAssumptions = false;
		error = result.error;
		if (!error) {
			editingAssumptions = false;
			await refreshAll();
		}
	}
	async function resetAssumptions() {
		if (!plan) return;
		const result = await mutate(updateAssumptions, { id: plan.id, assumptions: [] });
		error = result.error;
		editingAssumptions = false;
		if (!error) await refreshAll();
	}

	const visibleAssumptions = $derived(
		plan ? plan.assumptions.filter((a) => showEmpty || a.value.amount !== 0 || a.weight > 0) : []
	);
	const hiddenCount = $derived(plan ? plan.assumptions.length - visibleAssumptions.length : 0);
	const last = $derived(plan ? plan.projection[plan.projection.length - 1] : null);
	const years = $derived(plan ? Math.round(plan.horizonMonths / 12) : 0);
	const monthFmt = new Intl.DateTimeFormat(undefined, { month: 'long', year: 'numeric', timeZone: 'UTC' });
	const horizonLabel = $derived(
		plan ? monthFmt.format(new Date(new Date(plan.start).getTime() + 0).setUTCMonth(new Date(plan.start).getUTCMonth() + plan.horizonMonths)) : ''
	);
	const stats = $derived.by(() => {
		if (!plan || !last) return [];
		const cur = plan.startingAmount.currency;
		const out = [];
		if (plan.status) {
			out.push({
				label: 'Today',
				value: formatMoney(plan.status.actual),
				sub: `plan ${formatMoney(plan.status.expected)} · ${plan.status.gap.amount ? formatSignedMoney(plan.status.gap) : 'on plan'}`,
				tone: toneOf(plan.status.gap.amount)
			});
		} else {
			out.push({ label: 'Starting amount', value: formatMoney(plan.startingAmount), sub: plan.startingAmountDeclared ? 'declared' : `your net worth in ${monthFmt.format(new Date(plan.start))}` });
		}
		out.push({
			label: `In ${horizonLabel}`,
			value: formatMoney(last.p50),
			sub: `10–90 %: ${formatMoneyCompact(last.p10)} – ${formatMoneyCompact(last.p90)}`
		});
		out.push({
			label: 'Put in by then',
			value: formatMoney({ amount: plan.startingAmount.amount + last.contributed.amount, currency: cur }),
			sub: `${formatMoney(plan.startingAmount)} ${plan.startingAmountDeclared ? 'declared' : 'at the start'} + ${formatMoney(last.contributed)} contributed`
		});
		out.push({
			label: 'Contribution',
			value: `${formatMoney(plan.contribution)} / mo`,
			sub: plan.contributionDeclared ? 'declared' : 'what your schedules leave each month'
		});
		return out;
	});
</script>

<div class="mb-6 flex flex-wrap items-center gap-3">
	<div class="flex-1">
		<h1 class="text-lg font-semibold">
			Plan
			<span class="ml-1 rounded-full bg-accent-soft px-1.5 align-middle text-[11px] font-semibold text-accent-fg" title="A projection from your assumptions, not a promise">≈ projection</span>
		</h1>
		<p class="text-[13px] text-muted">
			Where your net worth is heading if the assumptions hold: securities, cash, savings, property and debt, each at its own expected return.
		</p>
	</div>
	{#if data.plans.length > 1 && plan}
		<Segmented
			label="Plan"
			options={data.plans.map((p) => ({ value: p.id, label: p.displayName }))}
			value={plan.id}
			onchange={(id) => goto(`?plan=${id}`, { replace: true, reset: false })}
		/>
	{/if}
	{#if plan}
		<Button variant="secondary" onclick={() => openForm(plan)}>Edit</Button>
		<Button variant="danger" onclick={() => (deleteOpen = true)}>Delete</Button>
	{/if}
	<Button onclick={() => openForm(null)}>New plan</Button>
</div>

<ErrorNote message={error} />

{#if !plan || !last}
	<Panel title="No plan yet">
		<p class="text-sm text-muted">
			A plan starts from your net worth and adds what your schedules leave each month, then grows each kind of wealth at a rate you choose: shares and ETFs at a long-run market return, savings at their interest, the house slowly, debt not at all. A thousand simulated paths show how wide the outcome can get.
		</p>
		<div><Button onclick={() => openForm(null)}>Create plan</Button></div>
	</Panel>
{:else}
	<div class="flex flex-col gap-6">
		<StatStrip {stats} />

		<Panel title="{plan.displayName} · {monthFmt.format(new Date(plan.start))} to {horizonLabel}">
			{#snippet aside()}
				<span class="text-xs text-muted">{years} years{plan.annualCost ? ` · ${formatPercent(plan.annualCost).replace('+', '')} cost a year` : ''}</span>
			{/snippet}
			<PlanChart projection={plan.projection} actual={plan.actual} startingAmount={plan.startingAmount} />
		</Panel>

		<Panel title="Assumptions">
			{#snippet aside()}
				<span class="flex items-center gap-3 text-xs">
					{#if editingAssumptions}
						<button type="button" class="text-muted hover:text-fg" onclick={() => (editingAssumptions = false)}>Cancel</button>
						<button type="button" class="text-muted hover:text-fg" onclick={resetAssumptions}>Reset to defaults</button>
						<Button onclick={saveAssumptions} disabled={savingAssumptions}>{savingAssumptions ? 'Saving…' : 'Save'}</Button>
					{:else}
						{#if hiddenCount > 0}
							<button type="button" class="text-muted hover:text-fg" onclick={() => (showEmpty = !showEmpty)}>{showEmpty ? 'Hide empty kinds' : `Show ${hiddenCount} empty kinds`}</button>
						{/if}
						<Button variant="secondary" onclick={startEditing}>Edit</Button>
					{/if}
				</span>
			{/snippet}
			<p class="text-xs text-muted">
				Each kind of wealth grows at its compound annual return, net of cost; the volatility widens the fan. The share says where each month's contribution goes; left alone it follows what you hold in the liquid kinds. Debt stays flat: its interest is already an expense in your cash flow, its repayment a transfer.
			</p>
			<table class="num w-full text-[13px]">
				<thead class="text-left text-xs text-muted">
					<tr>
						<th class="py-1.5 pr-3 font-medium">Kind</th>
						<th class="py-1.5 pr-3 text-right font-medium">Held today</th>
						<th class="py-1.5 pr-3 text-right font-medium">Share of contribution</th>
						<th class="py-1.5 pr-3 text-right font-medium">Return / year</th>
						<th class="py-1.5 text-right font-medium">Volatility</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-line">
					{#if editingAssumptions}
						{#each rows as row (row.bucket)}
							{@const a = plan.assumptions.find((x) => x.bucket === row.bucket)}
							<tr>
								<td class="py-1.5 pr-3">{planBucketLabel(row.bucket)}</td>
								<td class="py-1.5 pr-3 text-right {a ? toneOf(a.value.amount) : ''}">{a ? formatMoney(a.value) : ''}</td>
								<td class="py-1.5 pr-3 text-right"><input type="number" step="1" min="0" max="100" bind:value={row.weight} class="h-8 w-20 px-2 text-right" aria-label="Share of contribution in percent" /> %</td>
								<td class="py-1.5 pr-3 text-right"><input type="number" step="0.1" bind:value={row.annualReturn} class="h-8 w-20 px-2 text-right" aria-label="Return per year in percent" /> %</td>
								<td class="py-1.5 text-right"><input type="number" step="0.1" min="0" bind:value={row.volatility} class="h-8 w-20 px-2 text-right" aria-label="Volatility in percent" /> %</td>
							</tr>
						{/each}
					{:else}
						{#each visibleAssumptions as a (a.bucket)}
							<tr>
								<td class="py-1.5 pr-3">{planBucketLabel(a.bucket)}</td>
								<td class="py-1.5 pr-3 text-right {toneOf(a.value.amount)}">{formatMoney(a.value)}</td>
								<td class="py-1.5 pr-3 text-right text-muted">{a.weight ? `${Math.round(a.weight * 100)} %${a.weightDeclared ? '' : ' · follows holdings'}` : '—'}</td>
								<td class="py-1.5 pr-3 text-right">{formatPercent(a.annualReturn)}</td>
								<td class="py-1.5 text-right text-muted">{a.volatility ? formatPercent(a.volatility).replace('+', '') : '—'}</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</Panel>
	</div>
{/if}

<Dialog bind:open={formOpen} title={editing ? 'Edit plan' : 'New plan'}>
	{#key editing?.id ?? 'new'}
		<PlanForm
			plan={editing}
			netWorth={data.netWorth}
			surplus={data.surplus}
			onsaved={async () => {
				formOpen = false;
				await refreshAll();
			}}
		/>
	{/key}
</Dialog>

<Dialog bind:open={deleteOpen} title="Delete plan">
	<p class="text-sm">Delete “{plan?.displayName}”? Its assumptions go with it; nothing in the books changes.</p>
	<div class="mt-4 flex justify-end gap-2">
		<Button variant="secondary" onclick={() => (deleteOpen = false)}>Keep</Button>
		<Button variant="danger" onclick={remove}>Delete</Button>
	</div>
</Dialog>
