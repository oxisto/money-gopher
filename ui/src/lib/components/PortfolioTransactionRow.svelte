<script lang="ts">
	import { PortfolioEventType, type PortfolioEvent, Security } from '$lib/gen/mgo_pb';
	import { currency as formatCurrency } from '$lib/intl';
	import { PencilSquare } from '@steeze-ui/heroicons';
	import { Icon } from '@steeze-ui/svelte-icon';
	import Date from './Date.svelte';

	export let tx: PortfolioEvent;
	export let security: Security | undefined;
	export let currency = 'EUR';

	$: total = tx.amount * (tx.price + tx.fees + tx.taxes);
</script>

<tr>
	<td class="truncate whitespace-nowrap py-2 pl-4 pr-3 text-sm font-medium sm:pl-0">
		<Date date={tx.time?.toDate()} />
	</td>
	<td class="truncate whitespace-nowrap py-2 pl-4 pr-3 text-sm font-medium sm:pl-0">
		{PortfolioEventType[tx.type]}
	</td>
	<td class="truncate whitespace-nowrap py-2 pl-4 pr-3 text-sm font-medium sm:pl-0">
		{#if security}
			<div class="text-gray-900">{security.displayName}</div>
		{/if}
		<div class="text-gray-400">
			{tx.securityName}
		</div>
	</td>
	<td class="hidden whitespace-nowrap px-3 py-2 text-right text-sm text-gray-500 md:table-cell">
		{Intl.NumberFormat(navigator.language, {
			maximumFractionDigits: 2
		}).format(tx.amount)}
	</td>
	<td class="hidden whitespace-nowrap px-3 py-2 text-right text-sm lg:table-cell">
		<div class="text-gray-500">
			{formatCurrency(tx.price, currency)}
		</div>
	</td>
	<td class="hidden whitespace-nowrap px-3 py-2 text-right text-sm sm:table-cell">
		<div class="text-gray-500">
			{formatCurrency(tx.fees, currency)}
		</div>
	</td>
	<td class="hidden whitespace-nowrap px-3 py-2 text-right text-sm sm:table-cell">
		<div class="text-gray-500">
			{formatCurrency(tx.taxes, currency)}
		</div>
	</td>
	<td class="hidden whitespace-nowrap px-3 py-2 text-right text-sm sm:table-cell">
		<div class="text-gray-500">
			{formatCurrency(total, currency)}
		</div>
	</td>
	<td>
		<a href="/portfolios/{tx.portfolioName}/transactions/{tx.name}">
			<Icon src={PencilSquare} class="h-5 w-5 text-gray-400" theme="mini" />
		</a>
	</td>
</tr>
