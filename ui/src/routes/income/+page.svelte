<script lang="ts">
	import Icon from '$lib/components/ui/Icon.svelte';
	import Panel from '$lib/components/ui/Panel.svelte';
	import Segmented from '$lib/components/ui/Segmented.svelte';
	import StatStrip from '$lib/components/ui/StatStrip.svelte';
	import { describeTransaction } from '$lib/describe';
	import { formatDate, formatMoney, formatSignedMoney, toneOf } from '$lib/format';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	const RealizedGains = $derived(data.RealizedGains);

	const currentYear = new Date().getFullYear();
	const years = Array.from({ length: 5 }, (_, i) => String(currentYear - i));
	let year = $state(String(currentYear));

	function loadYear(value: string) {
		const y = Number(value);
		// Exclusive upper bound: the instant January 1 of the following year
		// begins, so December 31 is fully included.
		RealizedGains.fetch({
			variables: { from: new Date(Date.UTC(y, 0, 1)), to: new Date(Date.UTC(y + 1, 0, 1)) },
			policy: 'NetworkOnly'
		});
	}

	const g = $derived($RealizedGains.data?.realizedGains ?? null);
	const entries = $derived(g?.entries ?? []);
	const money = (amount: number) => ({ amount, currency: g?.currency ?? 'EUR' });

	const stats = $derived(
		g
			? [
					{ label: 'Dividends', value: formatSignedMoney(g.dividends), tone: toneOf(g.dividends.amount) },
					{ label: 'Interest', value: formatSignedMoney(g.interest), tone: toneOf(g.interest.amount) },
					{
						label: 'Sales',
						value: formatSignedMoney(g.tradingGains),
						sub: 'profit/loss of completed sales',
						tone: toneOf(g.tradingGains.amount)
					},
					{ label: 'Total', value: formatSignedMoney(g.total), sub: `realized in ${year}`, tone: toneOf(g.total.amount) }
				]
			: []
	);

	// Dividends and interest per month, stacked; sales stay out of the bars
	// because a loss would have to point down and muddle the picture.
	const months = $derived.by(() => {
		const rows = Array.from({ length: 12 }, () => ({ dividends: 0, interest: 0 }));
		for (const e of entries) {
			const m = (e.time as Date).getMonth();
			if (e.type === 'DIVIDEND') rows[m].dividends += e.amount.amount;
			if (e.type === 'INTEREST') rows[m].interest += e.amount.amount;
		}
		const max = Math.max(...rows.map((r) => r.dividends + r.interest), 1);
		const label = new Intl.DateTimeFormat(undefined, { month: 'short' });
		const long = new Intl.DateTimeFormat(undefined, { month: 'long' });
		return rows.map((r, i) => ({
			label: label.format(new Date(2000, i, 1)),
			title: `${long.format(new Date(2000, i, 1))}: ${formatMoney(money(r.dividends))} dividends, ${formatMoney(money(r.interest))} interest`,
			dividends: (r.dividends / max) * 100,
			interest: (r.interest / max) * 100
		}));
	});

	const payers = $derived.by(() => {
		const bySecurity = new Map<string, { id: string; name: string; amount: number; count: number }>();
		for (const e of entries) {
			if (e.type !== 'DIVIDEND' || !e.security) continue;
			const p = bySecurity.get(e.security.id) ?? { id: e.security.id, name: e.security.displayName, amount: 0, count: 0 };
			p.amount += e.amount.amount;
			p.count += 1;
			bySecurity.set(e.security.id, p);
		}
		const sorted = [...bySecurity.values()].sort((a, b) => b.amount - a.amount).slice(0, 8);
		const max = sorted[0]?.amount || 1;
		return sorted.map((p) => ({ ...p, width: (p.amount / max) * 100 }));
	});

	const groups = $derived.by(() => {
		const fmt = new Intl.DateTimeFormat(undefined, { month: 'long' });
		const out: { label: string; total: number; rows: typeof entries }[] = [];
		for (const e of entries) {
			const label = fmt.format(e.time as Date);
			let m = out[out.length - 1];
			if (!m || m.label !== label) out.push((m = { label, total: 0, rows: [] }));
			m.rows.push(e);
			m.total += e.amount.amount;
		}
		return out;
	});

	function title(e: (typeof entries)[number]) {
		if (e.type === 'SELL') return `Sold ${e.security?.displayName ?? ''}`;
		return describeTransaction({ type: e.type, units: 0, security: e.security }).title;
	}
</script>

<div class="mb-6 flex flex-wrap items-center gap-3">
	<div class="flex-1">
		<h1 class="text-lg font-semibold">Income</h1>
		<p class="text-[13px] text-muted">What your money paid out: dividends, interest and the result of completed sales.</p>
	</div>
	<Segmented label="Year" options={years.map((y) => ({ value: y, label: y }))} bind:value={year} onchange={loadYear} />
</div>

{#if $RealizedGains.fetching && !$RealizedGains.data}
	<p class="text-sm text-muted">Loading…</p>
{:else if $RealizedGains.errors?.length}
	<p class="text-sm text-neg">Failed to load: {$RealizedGains.errors[0].message}</p>
{:else if g}
	<div class="flex flex-col gap-6 {$RealizedGains.fetching ? 'opacity-60' : ''}">
		<StatStrip {stats} />

		<div class="grid gap-5 lg:grid-cols-[3fr_2fr]">
			<Panel title="Dividends & interest by month">
				<div class="flex min-h-44 flex-1 items-end gap-1.5 pt-2" role="img" aria-label="Dividends and interest per month">
					{#each months as m (m.label)}
						<div class="flex h-full flex-1 flex-col items-center justify-end gap-1" title={m.title}>
							<div class="flex w-full max-w-8 flex-col justify-end" style="height: {m.dividends + m.interest}%">
								<span class="rounded-t-[3px] bg-accent/45" style="flex: {m.interest} 1 0"></span>
								<span class="bg-accent {m.interest ? '' : 'rounded-t-[3px]'}" style="flex: {m.dividends} 1 0"></span>
							</div>
							<span class="text-[10px] text-muted">{m.label}</span>
						</div>
					{/each}
				</div>
				<div class="flex gap-4 text-xs text-muted">
					<span class="flex items-center gap-1.5"><span class="size-2 rounded-[2px] bg-accent"></span>Dividends</span>
					<span class="flex items-center gap-1.5"><span class="size-2 rounded-[2px] bg-accent/45"></span>Interest</span>
				</div>
			</Panel>

			<Panel title="Top dividend payers">
				{#if payers.length}
					<ul class="num flex flex-col gap-2 text-[13px]">
						{#each payers as p (p.id)}
							<li>
								<a href="/holdings/{p.id}" class="flex justify-between gap-3 hover:underline">
									<span class="truncate">{p.name}</span>
									<span class="shrink-0">{formatMoney(money(p.amount))}</span>
								</a>
								<div class="mt-1 flex items-center gap-2">
									<span class="h-1.5 flex-1 overflow-hidden rounded-full bg-line">
										<span class="block h-full rounded-full bg-accent" style="width: {p.width}%"></span>
									</span>
									<span class="w-16 text-right text-[11px] text-muted">{p.count} {p.count === 1 ? 'payment' : 'payments'}</span>
								</div>
							</li>
						{/each}
					</ul>
				{:else}
					<p class="text-sm text-muted">No dividends in {year}.</p>
				{/if}
			</Panel>
		</div>

		<section class="flex flex-col gap-4">
			<h2 class="text-sm font-semibold">Events</h2>
			{#if groups.length === 0}
				<p class="rounded-xl border border-dashed border-line px-4 py-8 text-center text-sm text-muted">
					Nothing realized in {year}.
				</p>
			{/if}
			{#each groups as m (m.label)}
				<section aria-label={m.label}>
					<div class="num flex items-baseline justify-between border-b border-line px-1 pb-1.5">
						<h3 class="text-[13px] font-semibold">{m.label}</h3>
						<span class="text-xs font-semibold {toneOf(m.total)}">{formatSignedMoney(money(m.total))}</span>
					</div>
					<ul class="mt-1 flex flex-col gap-px">
						{#each m.rows as e (e.id)}
							<li>
								<a
									href="/transactions/{e.id}"
									class="num grid grid-cols-[2.75rem_1.75rem_1fr_auto] items-center gap-3 rounded-lg px-2.5 py-2 hover:bg-raised"
								>
									<span class="text-xs text-muted">
										{new Intl.DateTimeFormat(undefined, { day: 'numeric', month: 'short' }).format(e.time as Date)}
									</span>
									<span class="grid size-7 place-items-center rounded-md border border-line bg-raised text-muted">
										<Icon name={e.type === 'SELL' ? 'minus' : describeTransaction({ type: e.type, units: 0, security: null }).icon} size={14} />
									</span>
									<span class="min-w-0">
										<span class="block truncate">{title(e)}</span>
										<span class="block truncate text-xs text-muted">
											{e.portfolio?.displayName ?? e.cashAccount?.displayName ?? ''}
										</span>
									</span>
									<span class="font-medium {toneOf(e.amount.amount)}">{formatSignedMoney(e.amount)}</span>
								</a>
							</li>
						{/each}
					</ul>
				</section>
			{/each}
		</section>
	</div>
{/if}
