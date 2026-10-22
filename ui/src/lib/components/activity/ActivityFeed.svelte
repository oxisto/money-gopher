<script lang="ts">
	import { onMount } from 'svelte';
	import { graphql, mutate } from '#lib/graphql/client.js';
	import TransactionEditForm from '#lib/components/TransactionEditForm.svelte';
	import Button from '#lib/components/ui/Button.svelte';
	import Dialog from '#lib/components/ui/Dialog.svelte';
	import ErrorNote from '#lib/components/ui/ErrorNote.svelte';
	import Icon from '#lib/components/ui/Icon.svelte';
	import { describeTransaction, formatUnits } from '#lib/describe.js';
	import { formatDate, formatMoney, formatPrice, formatSignedMoney, toneOf } from '#lib/format.js';
	import type { CashAccountOption, FeedTransaction, AssetOption } from '#lib/transaction-types.js';

	let {
		transactions,
		securities,
		cashAccounts,
		showBalance = false,
		showAccountFilter = false,
		onchanged
	}: {
		/** Newest first. */
		transactions: FeedTransaction[];
		securities: AssetOption[];
		cashAccounts: CashAccountOption[];
		// Only meaningful when transactions is one cash account's full
		// history: the running sum then ends exactly at its balance. A
		// portfolio's list may miss movements on its settlement account, so
		// callers opt in rather than this component guessing.
		showBalance?: boolean;
		showAccountFilter?: boolean;
		onchanged: () => void;
	} = $props();

	const groups = {
		all: { label: 'All', types: null },
		trades: { label: 'Trades', types: ['BUY', 'SELL', 'DELIVERY_INBOUND', 'DELIVERY_OUTBOUND'] },
		income: { label: 'Income', types: ['DIVIDEND', 'INTEREST'] },
		cash: { label: 'Cash & transfers', types: ['DEPOSIT_CASH', 'WITHDRAW_CASH'] },
		fees: { label: 'Fees & taxes', types: ['ACCOUNT_FEES', 'TAX_REFUND'] }
	} as const;
	type Group = keyof typeof groups;

	let group = $state<Group>('all');
	let query = $state('');
	let account = $state('all');
	let limit = $state(60);

	// Newest first, so the balance after the newest row is the sum of all
	// deltas; walking down subtracts each row's own delta. Computed over the
	// full list so filtering never shifts a row's balance.
	const balances = $derived.by(() => {
		const m = new Map<string, number>();
		let running = transactions.reduce((s, t) => s + t.cashDelta.amount, 0);
		for (const t of transactions) {
			m.set(t.id, running);
			running -= t.cashDelta.amount;
		}
		return m;
	});

	const accounts = $derived.by(() => {
		const m = new Map<string, string>();
		for (const t of transactions) {
			if (t.portfolio) m.set(t.portfolio.id, t.portfolio.displayName);
			if (t.cashAccount) m.set(t.cashAccount.id, t.cashAccount.displayName);
		}
		return [...m].map(([id, name]) => ({ id, name })).sort((a, b) => a.name.localeCompare(b.name));
	});

	const filtered = $derived.by(() => {
		const types = groups[group].types as readonly string[] | null;
		const q = query.trim().toLowerCase();
		return transactions.filter(
			(t) =>
				(!types || types.includes(t.type)) &&
				(account === 'all' || t.portfolio?.id === account || t.cashAccount?.id === account) &&
				(!q ||
					describeTransaction(t).title.toLowerCase().includes(q) ||
					(t.note ?? '').toLowerCase().includes(q) ||
					t.document?.filename.toLowerCase().includes(q))
		);
	});

	const months = $derived.by(() => {
		const fmt = new Intl.DateTimeFormat(undefined, { month: 'long', year: 'numeric' });
		const out: { label: string; net: number; rows: FeedTransaction[] }[] = [];
		for (const t of filtered.slice(0, limit)) {
			const label = fmt.format(t.time);
			let m = out[out.length - 1];
			if (!m || m.label !== label) {
				m = { label, net: 0, rows: [] };
				out.push(m);
			}
			m.rows.push(t);
			m.net += t.cashDelta.amount;
		}
		return out;
	});

	const currency = $derived(transactions[0]?.cashDelta.currency ?? 'EUR');

	let selectedID = $state<string | null>(null);
	const selected = $derived(
		transactions.find((t) => t.id === selectedID) ?? (filtered.length ? filtered[0] : null)
	);
	// Below lg there is no room for the side panel; details open in a dialog.
	let detailOpen = $state(false);
	let narrow = $state(false);

	function pick(id: string) {
		selectedID = id;
		confirmingDelete = false;
		deleteError = null;
		if (narrow) detailOpen = true;
	}

	onMount(() => {
		const media = window.matchMedia('(max-width: 1023px)');
		narrow = media.matches;
		media.addEventListener('change', (e) => (narrow = e.matches));

		// Links from elsewhere can point at one transaction via #tx-<id>.
		const id = window.location.hash.replace(/^#tx-/, '');
		if (id && transactions.some((t) => t.id === id)) {
			selectedID = id;
			const index = filtered.findIndex((t) => t.id === id);
			if (index >= limit) limit = index + 20;
			setTimeout(() => document.getElementById(`tx-${id}`)?.scrollIntoView({ block: 'center' }), 0);
		}
	});

	let editOpen = $state(false);
	let confirmingDelete = $state(false);
	let deleteError = $state<string | null>(null);

	const deleteTransaction = graphql(`
		mutation DeleteTransaction($id: ID!) {
			deleteTransaction(id: $id)
		}
	`);

	async function del(id: string) {
		const result = await mutate(deleteTransaction, { id });
		if (result.error) {
			deleteError = result.error;
			return;
		}
		confirmingDelete = false;
		detailOpen = false;
		selectedID = null;
		onchanged();
	}
</script>

{#snippet detail(t: FeedTransaction)}
	{@const d = describeTransaction(t)}
	{@const fields = [
		['Date', formatDate(t.time)],
		['Portfolio', t.portfolio?.displayName],
		['Cash account', t.cashAccount?.displayName],
		[t.cashDelta.amount < 0 ? 'Transferred to' : 'Transferred from', t.counterpart?.cashAccount?.displayName],
		['Units', t.units ? formatUnits(t.units) : null],
		['Price', t.price ? formatPrice(t.price) : null],
		['FX rate', t.fxRate ? `1 ${t.cashDelta.currency} = ${t.fxRate} ${t.price?.currency ?? ''}` : null],
		['Fees', t.fees.amount ? formatMoney(t.fees) : null],
		['Taxes', t.taxes.amount ? formatMoney(t.taxes) : null],
		['Realized gain', t.gain ? formatSignedMoney(t.gain) : null],
		[
			'Expected',
			t.schedule
				? `${t.schedule.displayName} · ${t.document ? 'confirmed by the statement' : 'booked, not confirmed yet'}`
				: null
		]
	].filter(([, v]) => v)}
	<div class="num flex flex-col gap-3">
		<div class="flex items-center gap-2.5">
			<span class="grid size-8 place-items-center rounded-lg border border-line bg-raised text-muted">
				<Icon name={d.icon} size={15} />
			</span>
			<div class="min-w-0 flex-1 font-semibold">{d.title}</div>
		</div>
		<div class="text-2xl font-semibold tracking-tight {t.cashDelta.amount > 0 ? 'text-pos' : ''}">
			{t.cashDelta.amount ? formatSignedMoney(t.cashDelta) : 'No cash movement'}
		</div>
		<dl class="text-[13px]">
			{#each fields as [k, v] (k)}
				<div class="flex justify-between gap-3 border-t border-line py-1.5">
					<dt class="text-muted">{k}</dt>
					<dd class="text-right">{v}</dd>
				</div>
			{/each}
		</dl>
		<a
			href="/transactions/{t.id}"
			class="flex items-center gap-2 rounded-lg border border-line bg-raised px-3 py-2.5 text-[13px] hover:bg-hover"
		>
			<Icon name="document" size={15} />
			<span class="min-w-0 flex-1 truncate">
				{t.document
					? t.document.filename
					: t.scheduleID
						? 'Booked by a schedule · not yet confirmed by a statement'
						: 'Entered manually'}
			</span>
			<span class="shrink-0 text-muted">Open →</span>
		</a>
		<div class="flex gap-2">
			<Button variant="secondary" onclick={() => (editOpen = true)}>Edit</Button>
			{#if confirmingDelete}
				<Button variant="danger" onclick={() => del(t.id)}>{t.transferID ? 'Really delete both legs' : 'Really delete'}</Button>
				<Button variant="secondary" onclick={() => (confirmingDelete = false)}>Keep</Button>
			{:else}
				<Button variant="danger" onclick={() => (confirmingDelete = true)}>Delete</Button>
			{/if}
		</div>
		<ErrorNote message={deleteError} />
	</div>
{/snippet}

<div class="flex flex-col gap-4">
	<div class="flex flex-wrap items-center gap-2.5">
		<label class="relative block w-full max-w-64">
			<span class="sr-only">Search activity</span>
			<span class="pointer-events-none absolute inset-y-0 left-2.5 grid place-items-center text-muted">
				<Icon name="search" size={15} />
			</span>
			<input type="search" bind:value={query} placeholder="Search security or document" class="h-9 w-full pl-8" />
		</label>
		<div role="radiogroup" aria-label="Type" class="flex flex-wrap gap-1.5">
			{#each Object.entries(groups) as [key, g] (key)}
				<button
					type="button"
					role="radio"
					aria-checked={group === key}
					onclick={() => (group = key as Group)}
					class="h-7.5 rounded-full border px-3 text-xs font-medium
						{group === key ? 'border-accent bg-accent-soft text-accent-fg' : 'border-line text-muted hover:text-fg'}"
				>
					{g.label}
				</button>
			{/each}
		</div>
		{#if showAccountFilter && accounts.length > 1}
			<label class="ml-auto flex items-center gap-2 text-[13px] text-muted">
				Account
				<select bind:value={account} class="h-8 py-0 pr-8 pl-2.5 text-[13px]">
					<option value="all">All accounts</option>
					{#each accounts as a (a.id)}
						<option value={a.id}>{a.name}</option>
					{/each}
				</select>
			</label>
		{/if}
	</div>

	{#if transactions.length === 0}
		<p class="rounded-xl border border-dashed border-line px-4 py-8 text-center text-sm text-muted">
			No transactions yet. Import a bank statement or add one by hand.
		</p>
	{:else if filtered.length === 0}
		<p class="rounded-xl border border-dashed border-line px-4 py-8 text-center text-sm text-muted">
			Nothing matches these filters.
		</p>
	{:else}
		<div class="flex items-start gap-5">
			<div class="flex min-w-0 flex-1 flex-col gap-5">
				{#each months as m (m.label)}
					<section aria-label={m.label}>
						<div class="num flex items-baseline justify-between border-b border-line px-1 pb-1.5">
							<h2 class="text-[13px] font-semibold">{m.label}</h2>
							<span class="text-xs text-muted">
								{m.rows.length}
								{m.rows.length === 1 ? 'entry' : 'entries'} · net
								<span class="font-semibold {m.net > 0 ? 'text-pos' : 'text-fg'}">
									{formatSignedMoney({ amount: m.net, currency })}
								</span>
							</span>
						</div>
						<ul class="mt-1 flex flex-col gap-px">
							{#each m.rows as t (t.id)}
								{@const d = describeTransaction(t)}
								{@const isSelected = !narrow && selected?.id === t.id}
								<li id="tx-{t.id}">
									<button
										type="button"
										onclick={() => pick(t.id)}
										aria-current={isSelected ? 'true' : undefined}
										class="num grid w-full grid-cols-[2.75rem_1.75rem_1fr_auto] items-center gap-3 rounded-lg px-2.5 py-2 text-left
											{isSelected ? 'bg-hover shadow-[inset_2px_0_0_var(--mg-accent)]' : 'hover:bg-raised'}"
									>
										<span class="text-xs text-muted">
											{new Intl.DateTimeFormat(undefined, { day: 'numeric', month: 'short' }).format(t.time)}
										</span>
										<span class="grid size-7 place-items-center rounded-md border border-line bg-raised text-muted">
											<Icon name={d.icon} size={14} />
										</span>
										<span class="min-w-0">
											<span class="block truncate">
												{d.title}
												{#if t.schedule}
													<span
														class="ml-1 rounded-full px-1.5 text-[11px] font-semibold {t.document ? 'bg-pos-soft text-pos' : 'bg-accent-soft text-accent-fg'}"
														title={t.document ? `Expected (${t.schedule.displayName}), confirmed by the statement` : `Expected (${t.schedule.displayName}), not confirmed by a statement yet`}
													>{t.document ? '✓ expected' : '≈ expected'}</span>
												{/if}
											</span>
											<span class="block truncate text-xs text-muted">
												{[d.subtitle, t.portfolio?.displayName ?? t.cashAccount?.displayName, t.price ? `${formatPrice(t.price, 4)} each` : null]
													.filter(Boolean)
													.join(' · ')}
											</span>
										</span>
										<span class="text-right">
											<span class="block font-medium {t.cashDelta.amount > 0 ? 'text-pos' : ''}">
												{t.cashDelta.amount ? formatSignedMoney(t.cashDelta) : '—'}
											</span>
											{#if showBalance}
												<span class="block text-[11px] text-muted">
													{formatMoney({ amount: balances.get(t.id) ?? 0, currency })}
												</span>
											{:else if t.gain}
												<span class="block text-[11px] {toneOf(t.gain.amount)}">
													realized {formatSignedMoney(t.gain)}
												</span>
											{/if}
										</span>
									</button>
								</li>
							{/each}
						</ul>
					</section>
				{/each}
				{#if filtered.length > limit}
					<button
						type="button"
						onclick={() => (limit += 60)}
						class="h-8 self-center rounded-lg border border-line px-3.5 text-[13px] text-muted hover:text-fg"
					>
						Show more · {filtered.length - limit} older
					</button>
				{/if}
			</div>

			{#if selected && !narrow}
				<aside
					aria-label="Transaction detail"
					class="sticky top-4 hidden w-80 shrink-0 rounded-xl border border-line p-4 lg:block"
				>
					{@render detail(selected)}
				</aside>
			{/if}
		</div>
	{/if}
</div>

{#if selected}
	<Dialog bind:open={detailOpen} title="Transaction">
		{@render detail(selected)}
	</Dialog>
	<Dialog bind:open={editOpen} title="Edit transaction" wide>
		{#if editOpen}
			<TransactionEditForm
				transaction={{ ...selected, cashAccount: selected.cashAccount ? { id: selected.cashAccount.id } : null }}
				{securities}
				{cashAccounts}
				onsaved={() => {
					editOpen = false;
					onchanged();
				}}
				oncancel={() => (editOpen = false)}
			/>
		{/if}
	</Dialog>
{/if}
