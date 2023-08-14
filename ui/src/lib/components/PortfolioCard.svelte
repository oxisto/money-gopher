<script lang="ts">
	import type { Portfolio, PortfolioSnapshot } from '$lib/gen/mgo_pb';
	import { currency } from '$lib/intl';
	import { CalendarDays, CreditCard, UserCircle } from '@steeze-ui/heroicons';
	import { Icon } from '@steeze-ui/svelte-icon';

	export let portfolio: Portfolio;
	export let snapshot: PortfolioSnapshot;
</script>

<div class="lg:col-start-3 lg:row-end-1">
	<h2 class="sr-only">Summary</h2>
	<div class="rounded-lg bg-gray-50 shadow-sm ring-1 ring-gray-900/5">
		<dl class="flex flex-wrap">
			<div class="flex-auto pl-6 pt-6">
				<dt class="text-sm font-semibold leading-6 text-gray-900">{portfolio.displayName}</dt>
				<dd class="mt-1 text-base font-semibold leading-6 text-gray-900">
					{currency(snapshot?.totalMarketValue, 'EUR')}
				</dd>
			</div>
			<div class="flex-none self-end px-6 pt-4">
				<dd
					class="mx-1 inline-flex items-center rounded-md bg-green-50 px-2 py-1 text-xs font-medium text-green-700 ring-1 ring-inset ring-green-600/20"
				>
					Stocks
				</dd>
				<dd
					class="mx-1 inline-flex items-center rounded-md bg-green-50 px-2 py-1 text-xs font-medium text-green-700 ring-1 ring-inset ring-green-600/20"
				>
					ETFs
				</dd>
			</div>
			<div class="mt-6 flex w-full flex-none gap-x-4 border-t border-gray-900/5 px-6 pt-6">
				<dt class="flex-none">
					<span class="sr-only">Client</span>
					<Icon src={UserCircle} class="h-6 w-5 text-gray-400" aria-hidden="true" />
				</dt>
				<dd class="text-sm font-medium leading-6 text-gray-900">Current User</dd>
			</div>
			<div class="mt-4 flex w-full flex-none gap-x-4 px-6">
				<dt class="flex-none">
					<span class="sr-only">Due date</span>
					<Icon src={CalendarDays} class="h-6 w-5 text-gray-400" aria-hidden="true" />
				</dt>
				<dd class="text-sm leading-6 text-gray-500">
					<time datetime="2023-01-31">{snapshot?.firstTransactionTime}</time>
				</dd>
			</div>
			<div class="mt-4 flex w-full flex-none gap-x-4 px-6">
				<dt class="flex-none">
					<span class="sr-only">Status</span>
					<Icon src={CreditCard} class="h-6 w-5 text-gray-400" aria-hidden="true" />
				</dt>
				<dd class="text-sm leading-6 text-gray-500">Associated with <i>My Bank Account</i></dd>
			</div>
		</dl>
		<div class="mt-6 border-t border-gray-900/5 px-6 py-6">
			<a
				href={'/portfolios/' + portfolio.name}
				class="text-sm font-semibold leading-6 text-gray-900"
			>
				Show positions
				<span aria-hidden="true">&rarr;</span>
			</a>
		</div>
	</div>
</div>
