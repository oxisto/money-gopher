<script lang="ts">
	import { goto, invalidate } from '$app/navigation';
	import { portfolioClient } from '$lib/api/client';
	import Button from '$lib/components/Button.svelte';
	import CurrencyInput from '$lib/components/CurrencyInput.svelte';
	import DateInput from '$lib/components/DateTimeInput.svelte';
	import SecurityListBox from '$lib/components/SecurityListBox.svelte';
	import TransactionTypeListBox from '$lib/components/TransactionTypeListBox.svelte';
	import { Currency, PortfolioEventType } from '$lib/gen/mgo_pb';
	import { currency } from '$lib/intl';
	import { FieldMask } from '@bufbuild/protobuf';
	import { ConnectError } from '@connectrpc/connect';
	import type { PageData } from './$types';

	export let data: PageData;

	$: total = new Currency({
		symbol: 'EUR',
		value:
			data.transaction.amount *
			((data.transaction.price?.value ?? 0) +
				(data.transaction.fees?.value ?? 0) +
				(data.transaction.taxes?.value ?? 0))
	});

	$: isSecurityTransaction =
		data.transaction.type == PortfolioEventType.BUY ||
		data.transaction.type == PortfolioEventType.SELL;

	async function save() {
		try {
			let client = portfolioClient();
			error = undefined;

			if (data.add) {
				await client.createPortfolioTransaction({ transaction: data.transaction });

				// Invalidate the portfolio snapshot
				await invalidate(`data:portfolio-snapshot:${data.transaction.portfolioName}`);
				goto('/portfolios/' + data.transaction.portfolioName);
			} else {
				await client.updatePortfolioTransaction({
					transaction: data.transaction,
					updateMask: new FieldMask({
						paths: ['amount', 'price', 'fees', 'taxes', 'security_name', 'time']
					})
				});

				// Invalidate the portfolio transaction list
				await invalidate(`data:portfolio-transactions:${data.transaction.portfolioName}`);
				goto('/portfolios/' + data.transaction.portfolioName + '/transactions');
			}
		} catch (err) {
			error = ConnectError.from(err);
		}
	}

	let error: undefined | ConnectError = undefined;
</script>

<form>
	<div class="space-y-12 sm:space-y-16">
		<div>
			<h2 class="text-base font-semibold leading-7 text-gray-900">
				{#if data.add}
					Add Transaction
				{:else}
					Edit Transaction
				{/if}
			</h2>
			<p class="mt-1 max-w-2xl text-sm leading-6 text-gray-600">
				{#if data.add}
					This allows you to add a transaction to portfolio <b>{data.transaction.portfolioName}</b>.
				{:else}
					This allows you to edit the existing transaction <b>{data.transaction.name}</b> in
					portfolio <b>{data.transaction.portfolioName}</b>.
				{/if}
			</p>

			<div
				class="mt-10 space-y-8 border-b border-gray-900/10 pb-12 sm:space-y-0 sm:divide-y sm:divide-gray-900/10 sm:border-t sm:pb-0"
			>
				<div class="sm:grid sm:grid-cols-3 sm:items-start sm:gap-4 sm:py-6">
					<label for="username" class="block text-sm font-medium leading-6 text-gray-900 sm:pt-1.5">
						Transaction Type
					</label>
					<div class="mt-2 sm:col-span-2 sm:mt-0">
						<TransactionTypeListBox bind:type={data.transaction.type} />
					</div>
				</div>

				<div class="sm:grid sm:grid-cols-3 sm:items-start sm:gap-4 sm:py-6">
					<label for="username" class="block text-sm font-medium leading-6 text-gray-900 sm:pt-1.5">
						Date
					</label>
					<div class="mt-2 sm:col-span-2 sm:mt-0">
						<DateInput bind:date={data.transaction.time} />
					</div>
				</div>

				{#if isSecurityTransaction}
					<div class="sm:grid sm:grid-cols-3 sm:items-start sm:gap-4 sm:py-6">
						<label
							for="username"
							class="block text-sm font-medium leading-6 text-gray-900 sm:pt-1.5"
						>
							Security
						</label>
						<div class="mt-2 sm:col-span-2 sm:mt-0">
							<SecurityListBox
								securities={data.securities}
								bind:securityName={data.transaction.securityName}
							/>
						</div>
					</div>

					<div class="sm:grid sm:grid-cols-3 sm:items-start sm:gap-4 sm:py-6">
						<label for="amount" class="block text-sm font-medium leading-6 text-gray-900 sm:pt-1.5">
							Amount
						</label>
						<div class="mt-2 sm:col-span-2 sm:mt-0">
							<input
								type="number"
								name="amount"
								id="amount"
								min="1"
								bind:value={data.transaction.amount}
								class="
              block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300
              placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:max-w-xs sm:text-sm sm:leading-6"
							/>
						</div>
					</div>
				{/if}

				<CurrencyInput name="price" bind:value={data.transaction.price}>Price</CurrencyInput>
				{#if isSecurityTransaction}
					<CurrencyInput name="fees" bind:value={data.transaction.fees}>Fees</CurrencyInput>
					<CurrencyInput name="taxes" bind:value={data.transaction.taxes}>Taxes</CurrencyInput>
				{/if}

				<div class="sm:grid sm:grid-cols-3 sm:items-start sm:gap-4 sm:py-6">
					<div class="block text-sm font-medium leading-6 text-gray-900 sm:pt-1.5">Total</div>
					<div class="mt-2 sm:col-span-2 sm:mt-0">
						<div class="block w-full text-gray-900 sm:max-w-xs sm:text-sm sm:leading-6">
							{currency(total)}
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>

	<Button on:click={save}>
		{#if data.add}
			Add
		{:else}
			Save
		{/if}
	</Button>
</form>

{#if error}
	{error}
{/if}
