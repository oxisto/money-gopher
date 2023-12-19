<script lang="ts">
	import type { Currency } from '$lib/gen/mgo_pb';

	export let value: Currency | undefined;
	export let name: string;

	let internal: number;

	function output(x: number | undefined) {
		if (value !== undefined && x !== undefined) {
			value.value = x * 100;
		}
	}

	function input(x: Currency | undefined) {
		internal = (x?.value ?? 0) / 100;
	}

	$: input(value);
	$: output(internal);
</script>

{#if value}
	<div class="sm:grid sm:grid-cols-3 sm:items-start sm:gap-4 sm:py-6">
		<label for={name} class="block text-sm font-medium leading-6 text-gray-900 sm:pt-1.5">
			<slot />
		</label>
		<div class="relative mt-2 rounded-md shadow-sm">
			<div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
				<span class="text-gray-500 sm:text-sm">€</span>
			</div>
			<input
				type="number"
				{name}
				id={name}
				bind:value={internal}
				class="block w-full rounded-md border-0 py-1.5 pl-7 pr-12 text-gray-900 ring-1 ring-inset ring-gray-300 placeholder:text-gray-400
      focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
				placeholder="0.00"
				aria-describedby="price-currency"
			/>
			<div class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-3">
				<span class="text-gray-500 sm:text-sm" id="price-currency">{value.symbol}</span>
			</div>
		</div>
	</div>
{/if}
