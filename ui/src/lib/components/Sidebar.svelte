<script setup lang="ts">
	import { page } from '$app/stores';
	import { Banknotes, CalendarDays, ChartPie, Cog6Tooth, Folder, Home } from '@steeze-ui/heroicons';
	import { Icon, type IconSource } from '@steeze-ui/svelte-icon';

	interface NavigationItemData {
		name: string;
		href: string;
		icon?: IconSource;
		isSub?: boolean;
		isNew?: boolean;
		children?: NavigationItemData[];
		disabled?: boolean;
	}

	const navigation: NavigationItemData[] = [
		{ name: 'Dashboard', href: '/dashboard', icon: Home },
		{ name: 'Securities', href: '/securities', icon: Banknotes },
		{
			name: 'Portfolios',
			href: '/portfolios',
			icon: Folder,
			children: [] as { name: string; href: string; routes: string[]; routeId: string }[]
		},
		{ name: 'Dividends', href: '/dividends', icon: CalendarDays },
		{ name: 'Performance', href: '/performance', icon: ChartPie }
	];

	// TODO: convert to store
	/*const client = inject(PortfolioServiceClientKey)
  
  const portfolios = await client?.listPortfolios({})
  navigation[2].children = portfolios?.portfolios.map(p => {
    return {
      name: p.displayName,
      href: '/portfolios/' + p.name,
      routes: ['portfolio-detail'],
      routeId: p.name
    }
  })*/

	const teams = [
		{ id: 1, name: 'Personal', href: '#', initial: 'H', current: false },
		{ id: 2, name: 'Parents', href: '#', initial: 'T', current: false },
		{ id: 3, name: 'Child', href: '#', initial: 'W', current: false }
	];
</script>

<div class="flex grow flex-col gap-y-5 overflow-y-auto bg-indigo-600 px-6 pb-4">
	<div class="flex h-16 shrink-0 items-center">
		<img
			class="h-8 w-auto"
			src="https://tailwindui.com/img/logos/mark.svg?color=white"
			alt="Your Company"
		/>
	</div>
	<nav class="flex flex-1 flex-col">
		<ul role="list" class="flex flex-1 flex-col gap-y-7">
			<li>
				<ul role="list" class="-mx-2 space-y-1">
					{#each navigation as item (item.name)}
						<li>
							<a
								href={item.href}
								class="{$page.url.pathname.startsWith(item.href)
									? 'bg-indigo-700 text-white'
									: 'text-indigo-200 hover:bg-indigo-700 hover:text-white'} group flex gap-x-3 rounded-md p-2 text-sm font-semibold leading-6"
							>
								{#if item.icon}
									<Icon
										src={item.icon}
										class="{$page.url.pathname.startsWith(item.href)
											? 'text-white'
											: 'text-indigo-200 group-hover:text-white'} h-6 w-6 shrink-0"
										aria-hidden="true"
									/>
								{/if}
								{item.name}
							</a>
							<ul class="mt-1 px-2">
								{#each item.children ?? [] as subItem (subItem.name)}
									<li>
										<!-- 44px -->
										<a
											href={subItem.href}
											class="{$page.url.pathname.startsWith(subItem.href)
												? 'bg-indigo-700 text-white'
												: 'text-indigo-200 hover:bg-indigo-700 hover:text-white'} block rounded-md py-2 pl-9 pr-2 text-sm leading-6 text-gray-700"
										>
											{subItem.name}</a
										>
									</li>
								{/each}
							</ul>
						</li>
					{/each}
				</ul>
			</li>
			<li>
				<div class="text-xs font-semibold leading-6 text-indigo-200">Your Portfolio Group</div>
				<ul role="list" class="-mx-2 mt-2 space-y-1">
					{#each teams as team (team.name)}
						<li>
							<a
								href={team.href}
								class="{team.current
									? 'bg-indigo-700 text-white'
									: 'text-indigo-200 hover:bg-indigo-700 hover:text-white'} 
                   group flex gap-x-3 rounded-md p-2 text-sm font-semibold leading-6"
							>
								<span
									class="flex h-6 w-6 shrink-0 items-center justify-center rounded-lg border border-indigo-400 bg-indigo-500 text-[0.625rem] font-medium text-white"
									>{team.initial}</span
								>
								<span class="truncate">{team.name}</span>
							</a>
						</li>
					{/each}
				</ul>
			</li>
			<li class="mt-auto">
				<a
					href="#"
					class="group -mx-2 flex gap-x-3 rounded-md p-2 text-sm font-semibold leading-6 text-indigo-200 hover:bg-indigo-700 hover:text-white"
				>
					<Icon
						src={Cog6Tooth}
						class="h-6 w-6 shrink-0 text-indigo-200 group-hover:text-white"
						aria-hidden="true"
					/>
					Settings
				</a>
			</li>
		</ul>
	</nav>
</div>
