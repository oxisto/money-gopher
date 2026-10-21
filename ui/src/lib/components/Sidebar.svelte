<script lang="ts">
	import { page } from '$app/state';
	import { cache, graphql, type Layout$result } from '$houdini';
	import { invalidateAll } from '$app/navigation';
	import { formatMoneyCompact } from '$lib/format';
	import { applyTheme, loadTheme, type ThemePreference } from '$lib/theme';

	interface Props {
		data: Layout$result | null;
		onnavigate?: () => void;
	}

	let { data, onnavigate }: Props = $props();

	const activePerson = $derived(data?.me?.activePerson ?? null);
	const persons = $derived(data?.me?.persons ?? []);
	const toReview = $derived(data?.documents?.length ?? 0);

	// Icons are 24×24 stroke paths, drawn by the template below.
	const nav = $derived([
		{ href: '/', label: 'Overview', icon: 'M3 12l9-8 9 8M5 10v10h5v-6h4v6h5V10' },
		{ href: '/holdings', label: 'Holdings', icon: 'M4 20V10M10 20V4M16 20v-7M22 20H2' },
		{ href: '/activity', label: 'Activity', icon: 'M3 6h18M3 12h18M3 18h12' },
		{
			href: '/income',
			label: 'Income',
			icon: 'M12 3a9 9 0 1 0 0 18a9 9 0 1 0 0-18M12 7v10M9.5 9.5h4a1.5 1.5 0 0 1 0 3h-3a1.5 1.5 0 0 0 0 3h4'
		},
		{ href: '/imports', label: 'Import', icon: 'M12 3v12M7 10l5 5 5-5M4 17v3h16v-3', badge: toReview }
	]);

	const footer = [
		{ href: '/securities', label: 'Securities', icon: 'M4 7h16M4 12h16M4 17h10' },
		{
			href: '/settings/people',
			label: 'People & access',
			icon: 'M9 11a4 4 0 1 0 0-8a4 4 0 0 0 0 8M3 21v-1a6 6 0 0 1 12 0v1M16 3.5a4 4 0 0 1 0 7M21 21v-1a6 6 0 0 0-3-5.2'
		}
	];

	function isActive(href: string): boolean {
		return href === '/' ? page.url.pathname === '/' : page.url.pathname.startsWith(href);
	}

	const switchPersonMutation = graphql(`
		mutation SwitchPerson($personID: ID!) {
			switchPerson(personID: $personID) {
				id
				displayName
			}
		}
	`);

	let open = $state(false);

	async function switchTo(id: string) {
		open = false;
		await switchPersonMutation.mutate({ personID: id });
		// switchPerson's response only carries the newly active Person, not
		// every other person-scoped query already cached (activePerson,
		// portfolios, cashAccounts, ...) — Houdini has no way to know those
		// need refetching just because the mutation ran, so mark the whole
		// cache stale before re-running load functions.
		cache.markStale();
		await invalidateAll();
	}

	let theme = $state<ThemePreference>(loadTheme());
	const themes: { value: ThemePreference; label: string; icon: string }[] = [
		{
			value: 'light',
			label: 'Light',
			icon: 'M12 8a4 4 0 1 0 0 8a4 4 0 0 0 0-8M12 2v2M12 20v2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M2 12h2M20 12h2M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4'
		},
		{ value: 'system', label: 'System', icon: 'M3 5h18v11H3zM8 20h8M12 16v4' },
		{ value: 'dark', label: 'Dark', icon: 'M20 14.5A8 8 0 0 1 9.5 4a8 8 0 1 0 10.5 10.5' }
	];

	function setTheme(value: ThemePreference) {
		theme = value;
		applyTheme(value);
	}
</script>

{#snippet icon(d: string, size = 16)}
	<svg
		width={size}
		height={size}
		viewBox="0 0 24 24"
		fill="none"
		stroke="currentColor"
		stroke-width="1.8"
		stroke-linecap="round"
		stroke-linejoin="round"
		aria-hidden="true"
	>
		<path {d} />
	</svg>
{/snippet}

<nav aria-label="Main" class="flex h-full flex-col gap-4 px-3 py-3.5">
	<!-- On small screens the layout's top bar already shows the logo. -->
	<a href="/" onclick={onnavigate} class="hidden items-center gap-2.5 px-1.5 font-semibold tracking-tight md:flex">
		<img src="/gopher.png" alt="" class="h-7.5 w-6.5 object-contain" />
		Money Gopher
	</a>

	{#if activePerson}
		<div class="relative">
			<button
				type="button"
				onclick={() => (open = !open)}
				aria-expanded={open}
				aria-haspopup="menu"
				disabled={persons.length < 2}
				class="flex w-full items-center gap-2.5 rounded-lg border border-line bg-surface px-2 py-1.5 text-left shadow-xs enabled:hover:bg-raised"
			>
				<span
					class="grid size-6 shrink-0 place-items-center rounded-md bg-accent text-xs font-semibold text-on-accent"
				>
					{activePerson.displayName.charAt(0)}
				</span>
				<span class="min-w-0 flex-1 truncate font-medium">{activePerson.displayName}</span>
				{#if persons.length > 1}
					<span class="text-muted">{@render icon('M8 9l4-4 4 4M8 15l4 4 4-4', 14)}</span>
				{/if}
			</button>
			{#if open}
				<div
					role="menu"
					class="absolute inset-x-0 top-full z-20 mt-1 rounded-lg border border-line bg-surface p-1 shadow-lg"
				>
					{#each persons as p (p.id)}
						<button
							type="button"
							role="menuitem"
							onclick={() => switchTo(p.id)}
							class="flex w-full items-center gap-2.5 rounded-md px-2 py-1.5 text-left hover:bg-hover
								{p.id === activePerson.id ? 'font-medium' : ''}"
						>
							<span
								class="grid size-5.5 place-items-center rounded-md border border-line bg-raised text-[11px] font-semibold"
							>
								{p.displayName.charAt(0)}
							</span>
							<span class="flex-1 truncate">{p.displayName}</span>
							{#if p.id === activePerson.id}
								<span class="text-accent">{@render icon('M5 12l5 5L20 7', 14)}</span>
							{/if}
						</button>
					{/each}
				</div>
			{/if}
		</div>
	{/if}

	<div class="flex flex-col gap-px">
		{#each nav as item (item.href)}
			<a
				href={item.href}
				onclick={onnavigate}
				aria-current={isActive(item.href) ? 'page' : undefined}
				class="flex items-center gap-2.5 rounded-md px-2 py-1.5 font-medium
					{isActive(item.href) ? 'bg-hover text-fg' : 'text-muted hover:bg-hover hover:text-fg'}"
			>
				{@render icon(item.icon)}
				<span class="flex-1">{item.label}</span>
				{#if item.badge}
					<span class="rounded-full bg-accent-soft px-1.5 text-[11px] font-semibold text-accent-fg">
						{item.badge}
					</span>
				{/if}
			</a>
		{/each}
	</div>

	{#if data && (data.portfolios.length || data.cashAccounts.length)}
		<div class="flex flex-col gap-px">
			<div class="px-2 pb-1 text-[11px] font-medium text-muted">Accounts</div>
			{#each data.portfolios as p (p.id)}
				{@const href = `/portfolios/${p.id}`}
				<a
					{href}
					onclick={onnavigate}
					aria-current={isActive(href) ? 'page' : undefined}
					class="flex items-center gap-2.5 rounded-md px-2 py-1.5 text-[13px]
						{isActive(href) ? 'bg-hover' : 'hover:bg-hover'}"
				>
					<span class="size-1.5 rounded-full bg-accent"></span>
					<span class="flex-1 truncate">{p.displayName}</span>
					<span class="num text-xs text-muted">{formatMoneyCompact(p.snapshot.totalMarketValue)}</span>
				</a>
			{/each}
			{#each data.cashAccounts as a (a.id)}
				{@const href = `/cash-accounts/${a.id}`}
				<a
					{href}
					onclick={onnavigate}
					aria-current={isActive(href) ? 'page' : undefined}
					class="flex items-center gap-2.5 rounded-md px-2 py-1.5 text-[13px]
						{isActive(href) ? 'bg-hover' : 'hover:bg-hover'}"
				>
					<span class="size-1.5 rounded-full bg-gold"></span>
					<span class="flex-1 truncate">{a.displayName}</span>
					<span class="num text-xs text-muted">{formatMoneyCompact(a.balance)}</span>
				</a>
			{/each}
		</div>
	{/if}

	<div class="mt-auto flex flex-col gap-px border-t border-line pt-2.5">
		{#each footer as item (item.href)}
			<a
				href={item.href}
				onclick={onnavigate}
				aria-current={isActive(item.href) ? 'page' : undefined}
				class="flex items-center gap-2.5 rounded-md px-2 py-1.5 text-[13px]
					{isActive(item.href) ? 'bg-hover text-fg' : 'text-muted hover:bg-hover hover:text-fg'}"
			>
				{@render icon(item.icon, 15)}
				{item.label}
			</a>
		{/each}
		<a
			href="/auth/logout"
			class="flex items-center gap-2.5 rounded-md px-2 py-1.5 text-[13px] text-muted hover:bg-hover hover:text-fg"
		>
			{@render icon('M15 4h4v16h-4M10 8l-4 4 4 4M6 12h11', 15)}
			Sign out
		</a>

		<div role="radiogroup" aria-label="Theme" class="mt-2 flex gap-0.5 rounded-lg border border-line bg-raised p-0.5">
			{#each themes as t (t.value)}
				<button
					type="button"
					role="radio"
					aria-checked={theme === t.value}
					aria-label={t.label}
					title={t.label}
					onclick={() => setTheme(t.value)}
					class="flex flex-1 items-center justify-center rounded-md py-1
						{theme === t.value ? 'bg-surface text-fg shadow-xs' : 'text-muted hover:text-fg'}"
				>
					{@render icon(t.icon, 14)}
				</button>
			{/each}
		</div>
	</div>
</nav>
