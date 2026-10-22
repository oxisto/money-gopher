<script lang="ts">
	import Panel from '#lib/components/ui/Panel.svelte';
	import StatStrip from '#lib/components/ui/StatStrip.svelte';
	import { intervalLabel, scheduleTypeLabel } from '#lib/describe.js';
	import { formatDate, formatMoney, formatSignedMoney, toneOf } from '#lib/format.js';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	const cf = $derived(data.cashFlow);

	const monthLabel = (t: Date | string) =>
		new Intl.DateTimeFormat(undefined, { month: 'short', year: '2-digit', timeZone: 'UTC' }).format(new Date(t));
	const thisMonth = $derived.by(() => {
		const now = new Date();
		const key = `${now.getUTCFullYear()}-${now.getUTCMonth()}`;
		return cf.months.find((m) => {
			const d = new Date(m.month);
			return `${d.getUTCFullYear()}-${d.getUTCMonth()}` === key;
		}) ?? cf.months[0];
	});
	const scale = $derived(
		Math.max(
			1,
			...cf.months.map((m) => Math.max(m.expectedIn.amount, m.expectedOut.amount + m.expectedSaved.amount, m.actualIn?.amount ?? 0, (m.actualOut?.amount ?? 0) + (m.actualSaved?.amount ?? 0)))
		)
	);
	let open = $state<string | null>(null);

	/** How the schedule shows on the cash accounts: in, out, saved, or not at all. */
	function bucket(s: (typeof cf.schedules)[number]): string {
		const target = s.cashAccount.kind ?? 'CASH';
		if (s.type === 'TRANSFER') {
			if ((s.counterpartAccount?.kind ?? 'CASH') !== 'CASH') return '—';
			return target === 'SAVINGS' ? 'saved' : target === 'LOAN' ? 'out' : '—';
		}
		if (target !== 'CASH') return '—';
		return ['WITHDRAW_CASH', 'ACCOUNT_FEES'].includes(s.type) ? 'out' : 'in';
	}
</script>

<div class="mb-6 flex flex-wrap items-center gap-3">
	<div class="flex-1">
		<h1 class="text-lg font-semibold">
			Cash flow
			<span class="ml-1 rounded-full bg-accent-soft px-1.5 align-middle text-[11px] font-semibold text-accent-fg" title="From your schedules, not from tracking every expense">≈ estimated</span>
		</h1>
		<p class="text-[13px] text-muted">
			The household's month as your schedules expect it, seen from the cash accounts: in, out, saved, and what is left.
		</p>
	</div>
</div>

<div class="flex flex-col gap-6">
	<StatStrip
		stats={[
			{ label: `In, ${monthLabel(thisMonth.month)}`, value: formatSignedMoney(thisMonth.expectedIn), sub: 'salary, rent, interest', tone: toneOf(thisMonth.expectedIn.amount) },
			{ label: 'Out', value: formatSignedMoney({ amount: -thisMonth.expectedOut.amount, currency: cf.currency }), sub: 'bills, fees, loan payments' },
			{ label: 'Saved', value: formatSignedMoney({ amount: -thisMonth.expectedSaved.amount, currency: cf.currency }), sub: 'into savings, not spent', tone: 'muted' },
			{ label: 'Free', value: formatSignedMoney(thisMonth.expectedFree), sub: 'in − out − saved', tone: toneOf(thisMonth.expectedFree.amount) }
		]}
	/>

	<Panel title="Month by month">
		{#snippet aside()}
			<span class="flex items-center gap-3 text-xs text-muted">
				<span><span class="mr-1.5 inline-block size-2 rounded-sm bg-pos"></span>in</span>
				<span><span class="mr-1.5 inline-block size-2 rounded-sm bg-neg"></span>out</span>
				<span><span class="mr-1.5 inline-block size-2 rounded-sm bg-accent"></span>saved</span>
			</span>
		{/snippet}
		<table class="num w-full text-[13px]">
			<tbody class="divide-y divide-line">
				{#each cf.months as m (m.month)}
					{@const key = String(m.month)}
					<tr class="cursor-pointer hover:bg-raised" onclick={() => (open = open === key ? null : key)}>
						<td class="w-16 py-2 pr-3 text-muted">{monthLabel(m.month)}</td>
						<td class="py-2 pr-3">
							<div class="flex flex-col gap-1">
								<div class="flex h-2.5 w-full gap-px overflow-hidden rounded-sm bg-raised" title={`expected: in ${formatMoney(m.expectedIn)} · out ${formatMoney(m.expectedOut)} · saved ${formatMoney(m.expectedSaved)}`}>
									<div class="bg-pos" style="width: {(m.expectedIn.amount / scale) * 50}%"></div>
									<div class="bg-neg" style="width: {(m.expectedOut.amount / scale) * 50}%"></div>
									<div class="bg-accent" style="width: {(m.expectedSaved.amount / scale) * 50}%"></div>
								</div>
								{#if m.actualIn}
									<div class="flex h-1.5 w-full gap-px overflow-hidden rounded-sm bg-raised opacity-70" title={`actual: in ${formatMoney(m.actualIn)} · out ${formatMoney(m.actualOut ?? m.actualIn)} · saved ${formatMoney(m.actualSaved ?? m.actualIn)}`}>
										<div class="bg-pos" style="width: {(m.actualIn.amount / scale) * 50}%"></div>
										<div class="bg-neg" style="width: {((m.actualOut?.amount ?? 0) / scale) * 50}%"></div>
										<div class="bg-accent" style="width: {((m.actualSaved?.amount ?? 0) / scale) * 50}%"></div>
									</div>
								{/if}
							</div>
						</td>
						<td class="w-28 py-2 text-right tabular-nums font-medium {m.expectedFree.amount > 0 ? 'text-pos' : m.expectedFree.amount < 0 ? 'text-neg' : ''}">
							{formatSignedMoney(m.expectedFree)}
							{#if m.actualFree}
								<div class="text-[11px] font-normal text-muted" title="What the tracked cash accounts actually show">actual {formatSignedMoney(m.actualFree)}</div>
							{/if}
						</td>
					</tr>
					{#if open === key}
						<tr>
							<td colspan="3" class="bg-raised/50 px-3 py-2">
								{#if m.flows.length === 0}
									<p class="text-xs text-muted">Nothing expected this month.</p>
								{:else}
									<ul class="grid gap-x-6 gap-y-1 text-xs sm:grid-cols-2 lg:grid-cols-3">
										{#each m.flows as f (f.schedule.id + f.date)}
											<li class="flex justify-between gap-3">
												<span class="truncate">
													<span class="text-muted">{formatDate(f.date)}</span>
													{f.schedule.displayName}
													{#if f.booked}<span class="text-pos" title="A transaction exists for this occurrence">✓</span>{/if}
												</span>
												<span class="tabular-nums {f.amount.amount > 0 ? 'text-pos' : ''}">{formatSignedMoney(f.amount)}</span>
											</li>
										{/each}
									</ul>
								{/if}
							</td>
						</tr>
					{/if}
				{/each}
			</tbody>
		</table>
		<p class="text-xs text-muted">
			Expected from your schedules; the thin bar is what the tracked cash accounts actually show, for months that are over. Click a month for its occurrences.
		</p>
	</Panel>

	<Panel title="Schedules" href="/cash-accounts" linkLabel="Accounts">
		{#if cf.schedules.length === 0}
			<p class="text-sm text-muted">No schedules yet. Add them on the account pages: a salary, the card bill, a Sparrate.</p>
		{:else}
			<table class="num w-full text-[13px]">
				<thead class="text-xs text-muted">
					<tr>
						<th class="py-1 pr-3 text-left font-medium">Schedule</th>
						<th class="py-1 pr-3 text-left font-medium">On</th>
						<th class="py-1 pr-3 text-left font-medium">Counts as</th>
						<th class="py-1 pr-3 text-right font-medium">Amount</th>
						<th class="py-1 text-right font-medium">Next</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-line">
					{#each cf.schedules as s (s.id)}
						<tr>
							<td class="py-1.5 pr-3">
								<div class="font-medium">{s.displayName}</div>
								<div class="text-xs text-muted">{scheduleTypeLabel(s.type, s.cashAccount.kind)} · {intervalLabel(s.interval)} on the {s.dayOfMonth}.{#if s.books} · books{/if}</div>
							</td>
							<td class="py-1.5 pr-3">
								<a href="/cash-accounts/{s.cashAccount.id}" class="hover:underline">{s.cashAccount.displayName}</a>
								{#if s.counterpartAccount}<span class="text-xs text-muted"> from {s.counterpartAccount.displayName}</span>{/if}
							</td>
							<td class="py-1.5 pr-3 text-xs text-muted">{bucket(s)}</td>
							<td class="py-1.5 pr-3 text-right tabular-nums {s.currentAmount.amount > 0 ? 'text-pos' : ''}">{formatSignedMoney(s.currentAmount)}</td>
							<td class="py-1.5 text-right text-xs text-muted">{s.nextDate ? formatDate(s.nextDate) : 'ended'}</td>
						</tr>
					{/each}
				</tbody>
			</table>
			<p class="text-xs text-muted">"—" means the schedule does not touch a cash account: a loan's interest charge, a fee inside a savings account.</p>
		{/if}
	</Panel>
</div>
