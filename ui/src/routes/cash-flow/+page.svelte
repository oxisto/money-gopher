<script lang="ts">
	import Panel from '#lib/components/ui/Panel.svelte';
	import Segmented from '#lib/components/ui/Segmented.svelte';
	import StatStrip from '#lib/components/ui/StatStrip.svelte';
	import { goto } from '$app/navigation';
	import { navigating } from '$app/state';
	import { describeTransaction, intervalLabel, scheduleTypeLabel } from '#lib/describe.js';
	import { formatDate, formatMoney, formatSignedMoney, toneOf } from '#lib/format.js';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	const cf = $derived(data.cashFlow);

	// The year lives in the URL (?year=2027); next year is all expectation.
	const currentYear = new Date().getFullYear();
	const years = Array.from({ length: 5 }, (_, i) => String(currentYear + 1 - i));
	let year = $derived(String(data.year));
	function loadYear(value: string) {
		open = null;
		goto(`?year=${value}`, { reset: false, replace: true });
	}

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
	// One month at a time: the strip and the opened row show the selected
	// one — the current month on entry, else the last with a statement.
	const lastActual = $derived([...cf.months].reverse().find((m) => m.actualFree));
	const isCurrent = $derived(String(data.year) === String(currentYear));
	let open = $state<string | null>(null);
	const selected = $derived(
		cf.months.find((m) => String(m.month) === open) ?? (isCurrent ? thisMonth : (lastActual ?? cf.months[0]))
	);
	const isOpen = (m: (typeof cf.months)[number]) => (open === null ? selected === m : open === String(m.month));
	const monthName = (t: Date | string) =>
		new Intl.DateTimeFormat(undefined, { month: 'long', year: 'numeric', timeZone: 'UTC' }).format(new Date(t));
	type M = (typeof cf.months)[number];
	/** A tile: the actual when the month has a statement, the expectation beneath; else the expectation. */
	function tile(m: M, label: string, expected: number, actual: number | null | undefined, sub: string, sign = 1) {
		const cur = cf.currency;
		if (actual === null || actual === undefined) {
			return { label, value: formatSignedMoney({ amount: sign * expected, currency: cur }), sub: `expected · ${sub}`, tone: toneOf(sign * expected) };
		}
		const diff = actual - expected;
		return {
			label,
			value: formatSignedMoney({ amount: sign * actual, currency: cur }),
			sub: `expected ${formatSignedMoney({ amount: sign * expected, currency: cur })}${diff ? ` · ${diff > 0 ? '+' : '−'}${formatMoney({ amount: Math.abs(diff), currency: cur })}` : ''}`,
			tone: toneOf(sign * actual)
		};
	}

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
			Cash flow {year}
			<span class="ml-1 rounded-full bg-accent-soft px-1.5 align-middle text-[11px] font-semibold text-accent-fg" title="From your schedules, not from tracking every expense">≈ estimated</span>
		</h1>
		<p class="text-[13px] text-muted">
			The household's month as your schedules expect it, seen from the cash accounts: in, out, saved, and what is left.
		</p>
	</div>
	<Segmented label="Year" options={years.map((y) => ({ value: y, label: y }))} bind:value={year} onchange={loadYear} />
</div>

<div class="flex flex-col gap-6 transition-opacity {navigating.to?.route.id === '/cash-flow' ? 'opacity-60' : ''}">
	<div class="flex flex-col gap-2">
		<h2 class="text-sm font-semibold">
			{monthName(selected.month)}
			<span class="ml-1 font-normal text-muted">
				{selected.actualFree ? '· what happened, against what the schedules expected' : '· what the schedules expect'}
			</span>
		</h2>
		<StatStrip
			stats={[
				tile(selected, 'In', selected.expectedIn.amount, selected.actualIn?.amount, 'salary, rent, interest'),
				tile(selected, 'Out', selected.expectedOut.amount, selected.actualOut?.amount, 'bills, fees, loan payments', -1),
				tile(selected, 'Saved', selected.expectedSaved.amount, selected.actualSaved?.amount, 'into savings, not spent', -1),
				tile(selected, 'Free', selected.expectedFree.amount, selected.actualFree?.amount, 'in − out − saved'),
				...(selected.unscheduledOut
					? [
							{
								label: 'Outside any schedule',
								value: formatSignedMoney({ amount: -selected.unscheduledOut.amount, currency: cf.currency }),
								sub: `${selected.unscheduled.length} movements the schedules do not know`,
								tone: ''
							}
						]
					: [])
			]}
		/>
	</div>

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
					<tr class="cursor-pointer hover:bg-raised {selected === m ? 'bg-raised' : ''}" onclick={() => (open = isOpen(m) ? '' : key)}>
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
					{#if isOpen(m)}
						<tr>
							<td colspan="3" class="bg-raised/50 px-3 py-2">
								<div class="grid gap-4 text-xs {m.actualIn ? 'lg:grid-cols-2' : ''}">
									<div>
										<p class="mb-1 font-medium text-muted">Expected</p>
										{#if m.flows.length === 0}
											<p class="text-muted">Nothing expected this month.</p>
										{:else}
											<ul class="flex flex-col gap-1">
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
									</div>
									{#if m.actualIn && m.actualOut && m.actualSaved && m.unscheduledIn && m.unscheduledOut && m.unscheduledSaved}
										<div>
											<p class="mb-1 font-medium text-muted">
												Happened · in {formatMoney(m.actualIn)} · out {formatMoney(m.actualOut)} · saved {formatMoney(m.actualSaved)}
											</p>
											{#if m.unscheduled.length === 0}
												<p class="text-muted">Every movement belonged to a schedule.</p>
											{:else}
												<p class="mb-1 text-muted">
													Outside any schedule: in {formatMoney(m.unscheduledIn)} · out {formatMoney(m.unscheduledOut)}{#if m.unscheduledSaved.amount}
														· saved {formatMoney(m.unscheduledSaved)}{/if}
												</p>
												<ul class="flex flex-col gap-1">
													{#each m.unscheduled as t (t.id)}
														{@const d = describeTransaction(t)}
														<li class="flex justify-between gap-3">
															<a href="/transactions/{t.id}" class="truncate hover:underline" title={d.subtitle}>
																<span class="text-muted">{formatDate(t.time)}</span>
																{d.title}
															</a>
															<span class="tabular-nums {t.cashDelta.amount > 0 ? 'text-pos' : ''}">{formatSignedMoney(t.cashDelta)}</span>
														</li>
													{/each}
												</ul>
											{/if}
										</div>
									{/if}
								</div>
							</td>
						</tr>
					{/if}
				{/each}
			</tbody>
		</table>
		<p class="text-xs text-muted">
			Thick bar: what the schedules expect. Thin bar: what the cash accounts actually show, for the months a statement covers. Click a month to put it in the strip above and see its details.
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
