<script lang="ts">
	import { currency } from '$lib/intl';
	import type { PageData } from './$types';

	export let data: PageData;
</script>

<div class="px-4 sm:px-6 lg:px-8">
	<div class="sm:flex sm:items-center">
		<div class="sm:flex-auto">
			<h1 class="text-base font-semibold leading-6 text-gray-900">Securities</h1>
			<p class="mt-2 text-sm text-gray-700">A list of all securities currently configured.</p>
		</div>
		<div class="mt-4 sm:ml-16 sm:mt-0 sm:flex-none"></div>
	</div>
	<div class="mt-8 flow-root">
		<div class="-mx-4 -my-2 overflow-x-auto sm:-mx-6 lg:-mx-8">
			<div class="inline-block min-w-full py-2 align-middle sm:px-6 lg:px-8">
				<table class="min-w-full divide-y divide-gray-300">
					<thead>
						<tr>
							<th
								scope="col"
								class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-0"
								>ISIN</th
							>
							<th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900"
								>Display Name</th
							>
							<th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900"
								>Listed On</th
							>
							<th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900"
								>Role</th
							>
							<th scope="col" class="relative py-3.5 pl-3 pr-4 sm:pr-0">
								<span class="sr-only">Edit</span>
							</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-200">
						{#each data.securities as security (security.name)}
							<tr>
								<td
									class="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-0"
								>
									{security.name}</td
								>
								<td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500"
									>{security.displayName}</td
								>
								<td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
									{#each security.listedOn as ls (ls.ticker)}
										<div>
											{ls.ticker} | 
											{#if ls.latestQuote}
												{currency(ls.latestQuote)}
											{/if}
										</div>
									{/each}
								</td>
								<td
									class="relative whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-0"
								>
									<a href="#" class="text-indigo-600 hover:text-indigo-900"
										>Edit<span class="sr-only">, {security.name}</span></a
									>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	</div>
</div>
