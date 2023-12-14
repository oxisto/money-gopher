<script lang="ts">
	import Date from '$lib/components/Date.svelte';
	import Performance from '$lib/components/Performance.svelte';
	import type { Portfolio, PortfolioSnapshot } from '$lib/gen/mgo_pb';
	import { currency } from '$lib/intl';
	import {
		Calendar,
		Check,
		ChevronDown,
		ChevronRight,
		CurrencyEuro,
		Link,
		ListBullet,
		Pencil
	} from '@steeze-ui/heroicons';
	import { Icon } from '@steeze-ui/svelte-icon';
	import { createMenu } from 'svelte-headlessui';
	import { Transition } from 'svelte-transition';

	const menu = createMenu({ label: 'More' });

	export let portfolio: Portfolio;
	export let snapshot: PortfolioSnapshot;
</script>

<div class="lg:flex lg:items-center lg:justify-between">
	<div class="min-w-0 flex-1">
		<nav class="flex" aria-label="Breadcrumb">
			<ol role="list" class="flex items-center space-x-4">
				<li>
					<div class="flex">
						<a href="/portfolios" class="text-sm font-medium text-gray-500 hover:text-gray-700"
							>Portfolios
						</a>
					</div>
				</li>
				<li>
					<div class="flex items-center">
						<Icon
							src={ChevronRight}
							class="h-5 w-5 flex-shrink-0 text-gray-400"
							aria-hidden="true"
						/>
						<a href="." class="ml-4 text-sm font-medium text-gray-500 hover:text-gray-700">
							My Bank
						</a>
					</div>
				</li>
			</ol>
		</nav>
		<h2
			class="mt-2 text-2xl font-bold leading-7 text-gray-900 sm:truncate sm:text-3xl sm:tracking-tight"
		>
			{portfolio.displayName}
		</h2>
		<div class="mt-1 flex flex-col sm:mt-0 sm:flex-row sm:flex-wrap sm:space-x-6">
			<div class="mt-2 flex items-center text-sm">
				<Performance {snapshot} />
			</div>
			<div class="mt-2 flex items-center text-sm text-gray-500">
				<Icon
					src={ListBullet}
					class="mr-1.5 h-5 w-5 flex-shrink-0 text-gray-400"
					aria-hidden="true"
				/>
				{Object.values(snapshot.positions).length} Positions
			</div>
			<div class="mt-2 flex items-center text-sm text-gray-500">
				<Icon
					src={CurrencyEuro}
					class="mr-1.5 h-5 w-5 flex-shrink-0 text-gray-400"
					aria-hidden="true"
				/>
				{currency(snapshot.totalMarketValue, 'EUR')}
			</div>
			<div class="mt-2 flex items-center text-sm text-gray-500">
				<Icon
					src={Calendar}
					class="mr-1.5 h-5 w-5 flex-shrink-0 text-gray-400"
					aria-hidden="true"
				/>
				First transaction on&nbsp;<Date date={snapshot.firstTransactionTime?.toDate()} />
			</div>
		</div>
	</div>
	<div class="mt-5 flex lg:ml-4 lg:mt-0">
		<span class="hidden sm:block">
			<button
				type="button"
				class="inline-flex items-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
			>
				<Icon src={Pencil} class="-ml-0.5 mr-1.5 h-5 w-5 text-gray-400" aria-hidden="true" />
				Edit
			</button>
		</span>

		<span class="ml-3 hidden sm:block">
			<button
				type="button"
				class="inline-flex items-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
			>
				<Icon src={Link} class="-ml-0.5 mr-1.5 h-5 w-5 text-gray-400" aria-hidden="true" />
				View
			</button>
		</span>

		<span class="sm:ml-3">
			<button
				type="button"
				class="inline-flex items-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
			>
				<Icon src={Check} class="-ml-0.5 mr-1.5 h-5 w-5" aria-hidden="true" />
				Publish
			</button>
		</span>

		<!-- Dropdown -->
		<div class="relative ml-3 sm:hidden">
			<button
				on:click={menu.open}
				class="inline-flex items-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:ring-gray-400"
			>
				More
				<Icon src={ChevronDown} class="-mr-1 ml-1.5 h-5 w-5 text-gray-400" aria-hidden="true" />
			</button>

			<Transition
				show={$menu.expanded}
				enter="transition ease-out duration-200"
				enterFrom="transform opacity-0 scale-95"
				enterTo="transform opacity-100 scale-100"
				leave="transition ease-in duration-75"
				leaveFrom="transform opacity-100 scale-100"
				leaveTo="transform opacity-0 scale-95"
			>
				<div
					use:menu.items
					class="absolute right-0 z-10 -mr-1 mt-2 w-48 origin-top-right rounded-md bg-white py-1 shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none"
				>
					<div use:menu.item>
						<a
							href="."
							class="{$menu.active === 'Edit'
								? 'bg-gray-100'
								: ''} block px-4 py-2 text-sm text-gray-700">Edit</a
						>
					</div>
					<div use:menu.item>
						<a
							href="."
							class="{$menu.active === 'View'
								? 'bg-gray-100'
								: ''} block px-4 py-2 text-sm text-gray-700">View</a
						>
					</div>
				</div>
			</Transition>
		</div>
	</div>
</div>
