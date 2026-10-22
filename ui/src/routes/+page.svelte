<script lang="ts">
	import NetWorthChart from '#lib/components/overview/NetWorthChart.svelte';
	import NewAccountForm from '#lib/components/NewAccountForm.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import Dialog from '#lib/components/ui/Dialog.svelte';
	import Icon from '#lib/components/ui/Icon.svelte';
	import Panel from '#lib/components/ui/Panel.svelte';
	import Segmented from '#lib/components/ui/Segmented.svelte';
	import StatStrip from '#lib/components/ui/StatStrip.svelte';
	import { describeTransaction } from '#lib/describe.js';
	import { formatDate, formatMoney, formatPercent, formatSignedMoney, toneOf, type Money } from '#lib/format.js';
	import { goto } from '$app/navigation';
	import { navigating } from '$app/state';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	// The range lives in the URL (?range=1Y), so changing it reloads the page
	// data through +page.ts and the view can be bookmarked.
	const periods = [
		{ value: '1M', label: '1M', since: 'past month' },
		{ value: '6M', label: '6M', since: 'past 6 months' },
		{ value: 'YTD', label: 'YTD', since: 'since Jan 1' },
		{ value: '1Y', label: '1Y', since: 'past year' },
		{ value: 'Max', label: 'Max', since: 'since the first transaction' }
	];
	let period = $derived(data.range);

	let newAccountOpen = $state(false);

	// Everything below assumes one currency across a person's accounts, like
	// the backend's totals do.
	const sum = (values: Money[]): Money => ({
		amount: values.reduce((s, m) => s + m.amount, 0),
		currency: values[0]?.currency ?? 'EUR'
	});
	const signed = formatSignedMoney;
	const tone = toneOf;

	const d = $derived(data.dashboard);
	const portfolios = $derived(d.portfolios);
	const cashAccounts = $derived(d.cashAccounts);
	const empty = $derived(portfolios.length === 0 && cashAccounts.length === 0);

	const invested = $derived(sum(portfolios.map((p) => p.snapshot.totalMarketValue)));
	const cash = $derived(sum(cashAccounts.map((a) => a.balance)));
	const netWorth = $derived(sum([invested, cash]));

	const history = $derived(data.history);
	const change = $derived.by(() => {
		if (history.length < 2) return null;
		const first = history[0];
		const last = history[history.length - 1];
		const total = last.netWorth.amount - first.netWorth.amount;
		const paidIn = last.invested.amount - first.invested.amount;
		return {
			total: { amount: total, currency: last.netWorth.currency },
			paidIn: { amount: paidIn, currency: last.netWorth.currency },
			earned: { amount: total - paidIn, currency: last.netWorth.currency }
		};
	});

	function pickPeriod(value: string) {
		goto(`?range=${value}`, { reset: false, replace: true });
	}

	const tiles = $derived.by(() => {
		const pol = sum(portfolios.map((p) => p.snapshot.totalProfitOrLoss));
		const cost = sum(portfolios.map((p) => p.snapshot.totalPurchaseValue));
		// With several portfolios, weight each one's return by its value — an
		// approximation, but a stable one, until the backend computes it.
		const twr = invested.amount
			? portfolios.reduce(
					(s, p) => s + p.snapshot.performance.timeWeightedReturn * p.snapshot.totalMarketValue.amount,
					0
				) / invested.amount
			: 0;
		const g = d.realizedGains;
		return [
			{
				label: 'Unrealized gain',
				value: signed(pol),
				tone: tone(pol.amount),
				sub: cost.amount ? `${formatPercent(pol.amount / cost.amount)} on ${formatMoney(cost)} cost` : 'no positions'
			},
			{
				label: `Realized ${new Date().getFullYear()}`,
				value: signed(g.total),
				tone: tone(g.total.amount),
				sub: 'dividends, interest and sales'
			},
			{
				label: 'Return YTD',
				value: formatPercent(twr),
				tone: tone(twr),
				sub: 'time-weighted, ignores deposits'
			},
			{
				label: 'Cash',
				value: formatMoney(cash),
				tone: '',
				sub: netWorth.amount ? `${formatPercent(cash.amount / netWorth.amount).replace('+', '')} of net worth` : ''
			}
		];
	});

	// The biggest holdings across all portfolios, then everything else, then
	// cash. Shades of the accent keep it calm; cash gets its own color.
	const SHADES = ['bg-accent', 'bg-accent/80', 'bg-accent/65', 'bg-accent/50', 'bg-accent/38', 'bg-accent/28'];
	const allocation = $derived.by(() => {
		const byAsset = new Map<string, { id: string; name: string; value: number }>();
		for (const p of portfolios) {
			for (const pos of p.snapshot.positions) {
				const e = byAsset.get(pos.asset.id) ?? {
					id: pos.asset.id,
					name: pos.asset.displayName,
					value: 0
				};
				e.value += pos.marketValue.amount;
				byAsset.set(pos.asset.id, e);
			}
		}
		const sorted = [...byAsset.values()].sort((a, b) => b.value - a.value);
		const top = sorted.slice(0, SHADES.length);
		const rest = sorted.slice(SHADES.length);
		// Keyed by ID: different securities can share a display name.
		const rows = top.map((e, i) => ({ key: e.id, label: e.name, value: e.value, color: SHADES[i] }));
		if (rest.length) {
			rows.push({
				key: 'other',
				label: `${rest.length} other holdings`,
				value: rest.reduce((s, e) => s + e.value, 0),
				color: 'bg-line-strong'
			});
		}
		if (cash.amount > 0) {
			rows.push({ key: 'cash', label: 'Cash', value: cash.amount, color: 'bg-gold' });
		}
		const total = rows.reduce((s, r) => s + r.value, 0) || 1;
		return rows.map((r) => ({ ...r, share: r.value / total }));
	});

	const income = $derived.by(() => {
		const months = Array.from({ length: 12 }, () => 0);
		for (const e of d.realizedGains.entries) {
			if (e.type === 'DIVIDEND' || e.type === 'INTEREST') {
				months[e.time.getMonth()] += e.amount.amount;
			}
		}
		const max = Math.max(...months, 1);
		const now = new Date().getMonth();
		const fmt = new Intl.DateTimeFormat(undefined, { month: 'narrow' });
		return {
			total: sum([d.realizedGains.dividends, d.realizedGains.interest]),
			bars: months.map((v, i) => ({
				label: fmt.format(new Date(2000, i, 1)),
				title: `${new Intl.DateTimeFormat(undefined, { month: 'long' }).format(new Date(2000, i, 1))}: ${formatMoney({ amount: v, currency: cash.currency })}`,
				height: v > 0 ? Math.max(4, (v / max) * 100) : 4,
				future: i > now
			}))
		};
	});

	const recent = $derived.by(() => {
		const seen = new Map<string, (typeof portfolios)[number]['transactions'][number]>();
		for (const p of portfolios) for (const t of p.transactions) seen.set(t.id, t);
		for (const a of cashAccounts) for (const t of a.transactions) seen.set(t.id, t);
		return [...seen.values()]
			.sort((a, b) => b.time.getTime() - a.time.getTime())
			.slice(0, 6)
			.map((t) => ({
				...t,
				...describeTransaction(t),
				account: t.portfolio?.displayName ?? t.cashAccount?.displayName ?? ''
			}));
	});

	const inbox = $derived(
		d.documents.slice(0, 3).map((doc) => {
			const staged = doc.stagedTransactions;
			const first = staged[0];
			const matches = staged.every(
				(s) => !s.calculatedCashDelta || s.calculatedCashDelta.amount === s.cashDelta.amount
			);
			return {
				id: doc.id,
				title: first
					? staged.length === 1
						? describeTransaction({ ...first, asset: first.asset ?? (first.assetHint ? { displayName: first.assetHint } : null) }).title
						: `${staged.length} transactions`
					: doc.filename,
				meta: [doc.transactionDate ? formatDate(doc.transactionDate) : null, doc.suggestedCashAccount?.displayName]
					.filter(Boolean)
					.join(' · '),
				amount: sum(staged.map((s) => s.cashDelta)),
				matches
			};
		})
	);
</script>

{#snippet header()}
	<div class="mb-6 flex flex-wrap items-center gap-3">
		<h1 class="flex-1 text-lg font-semibold">Overview</h1>
		<Button variant="secondary" onclick={() => (newAccountOpen = true)}>New account</Button>
		<a
			href="/imports"
			class="inline-flex h-8.5 items-center gap-1.5 rounded-lg bg-accent px-3 text-sm font-medium text-on-accent hover:bg-accent-hover"
		>
			<Icon name="upload" size={15} /> Import
		</a>
	</div>
{/snippet}

<Dialog bind:open={newAccountOpen} title="New account">
	<NewAccountForm {cashAccounts} oncreated={() => (newAccountOpen = false)} />
</Dialog>

{#if empty}
	{@render header()}
	<div class="mx-auto mt-10 max-w-md rounded-xl border border-line p-8 text-center">
		<img src="/gopher.png" alt="" class="mx-auto mb-4 h-16 object-contain" />
		<h2 class="text-base font-semibold">Let's get your money in here</h2>
		<p class="mt-1 text-sm text-muted">
			Create the cash account your bank statements belong to, then a portfolio for your securities — or just drop
			your bank's PDFs on the import page.
		</p>
		<div class="mt-5 flex justify-center gap-2">
			<Button onclick={() => (newAccountOpen = true)}>Create an account</Button>
		</div>
	</div>
{:else}
	{@render header()}

	<div class="flex flex-col gap-7">
		<section class="flex flex-col gap-4">
			<div class="flex flex-wrap items-end gap-4">
				<div class="min-w-64 flex-1">
					<div class="text-[13px] text-muted">Net worth</div>
					<div class="num mt-1 text-4xl font-semibold tracking-tight">{formatMoney(netWorth)}</div>
					{#if change}
						{@const since = periods.find((p) => p.value === period)?.since}
						<div class="num mt-1.5 flex flex-wrap items-center gap-x-2.5 gap-y-1 text-sm">
							<!-- No percentage here: with money paid in during the period it
							     would mostly measure deposits, not performance. -->
							<span class="font-semibold {tone(change.total.amount)}">{signed(change.total)}</span>
							<span class="text-muted">{since}</span>
							<span class="text-muted" aria-hidden="true">·</span>
							<span class="text-muted">
								<span class="font-medium {tone(change.earned.amount)}">{signed(change.earned)}</span> earned
							</span>
							<span class="text-muted" aria-hidden="true">·</span>
							<span class="text-muted">{signed(change.paidIn)} paid in</span>
						</div>
					{/if}
				</div>
				<Segmented label="Time range" options={periods} bind:value={period} onchange={pickPeriod} />
			</div>

			{#if history.length > 1}
				<div class="transition-opacity {navigating.to?.route.id === '/' ? 'opacity-60' : ''}">
					<NetWorthChart points={history} />
				</div>
			{:else}
				<div class="grid h-[260px] place-items-center rounded-xl border border-dashed border-line text-sm text-muted">
					Not enough history yet.
				</div>
			{/if}
		</section>

		<StatStrip stats={tiles} />

		<div class="grid gap-5 lg:grid-cols-2">
			<Panel title="Allocation">
				{#if allocation.length}
					<div class="flex h-2.5 gap-0.5 overflow-hidden rounded-full" aria-hidden="true">
						{#each allocation as row (row.key)}
							<span class={row.color} style="flex: {row.share} 1 0"></span>
						{/each}
					</div>
					<ul class="num text-[13px]">
						{#each allocation as row (row.key)}
							<li class="grid grid-cols-[0.75rem_1fr_auto_3.5rem] items-center gap-2.5 border-t border-line py-1.5 first:border-t-0">
								<span class="size-2.5 rounded-[3px] {row.color}"></span>
								<span class="truncate">{row.label}</span>
								<span>{formatMoney({ amount: row.value, currency: cash.currency })}</span>
								<span class="text-right text-muted">{(row.share * 100).toFixed(1)}%</span>
							</li>
						{/each}
					</ul>
				{:else}
					<p class="text-sm text-muted">Nothing held yet.</p>
				{/if}
			</Panel>

			<Panel title="Income">
				{#snippet aside()}
					<a href="/income" class="num text-xs text-muted hover:text-fg">
						{new Date().getFullYear()} · <strong class="font-semibold text-fg">{formatMoney(income.total)}</strong> dividends
						&amp; interest →
					</a>
				{/snippet}
				<div class="flex min-h-36 flex-1 items-end gap-1.5 pt-2" role="img" aria-label="Dividends and interest per month this year">
					{#each income.bars as bar, i (i)}
						<div class="flex h-full flex-1 flex-col items-center justify-end gap-1" title={bar.title}>
							<span
								class="w-full max-w-7 rounded-t-[3px] {bar.future ? 'bg-line' : 'bg-accent'}"
								style="height: {bar.height}%"
							></span>
							<span class="text-[10px] text-muted">{bar.label}</span>
						</div>
					{/each}
				</div>
			</Panel>

			<Panel title="Recent activity">
				{#if recent.length}
					<ul class="flex flex-col">
						{#each recent as tx (tx.id)}
							<li>
								<a href="/transactions/{tx.id}" class="num -mx-2 grid grid-cols-[1.75rem_1fr_auto] items-center gap-2.5 rounded-lg px-2 py-1.5 hover:bg-raised">
									<span class="grid size-7 place-items-center rounded-md border border-line bg-raised text-muted">
										<Icon name={tx.icon} size={14} />
									</span>
									<span class="min-w-0">
										<span class="block truncate">{tx.title}</span>
										<span class="block text-xs text-muted">{formatDate(tx.time)} · {tx.account}</span>
									</span>
									<span class="font-medium {tx.cashDelta.amount > 0 ? 'text-pos' : ''}">
										{tx.cashDelta.amount === 0 ? '—' : signed(tx.cashDelta)}
									</span>
								</a>
							</li>
						{/each}
					</ul>
				{:else}
					<p class="text-sm text-muted">No transactions yet.</p>
				{/if}
			</Panel>

			<Panel title="Import inbox">
				{#snippet aside()}
					{#if d.documents.length}
						<span class="rounded-full bg-accent-soft px-2 text-xs font-semibold text-accent-fg">
							{d.documents.length} to review
						</span>
					{/if}
				{/snippet}
				{#if inbox.length}
					<ul class="flex flex-col gap-2">
						{#each inbox as doc (doc.id)}
							<li class="num grid grid-cols-[1fr_auto] gap-x-3 rounded-lg border border-line bg-raised px-3 py-2.5">
								<span class="truncate font-medium">{doc.title}</span>
								<span class="text-right font-medium">{signed(doc.amount)}</span>
								<span class="truncate text-xs text-muted">{doc.meta}</span>
								<span class="text-right text-xs {doc.matches ? 'text-pos' : 'text-neg'}">
									{doc.matches ? '✓ amounts match' : 'amounts differ'}
								</span>
							</li>
						{/each}
					</ul>
					<a href="/imports" class="inline-flex h-8 items-center self-start rounded-lg bg-accent px-3 text-sm font-medium text-on-accent hover:bg-accent-hover">
						Review
					</a>
				{:else}
					<div class="flex items-center gap-3 rounded-lg border border-dashed border-line px-4 py-5 text-sm text-muted">
						<Icon name="check" />
						<span class="flex-1">All caught up. New bank PDFs show up here for review.</span>
						<a href="/imports" class="text-fg underline">Upload</a>
					</div>
				{/if}
			</Panel>
		</div>
	</div>
{/if}
