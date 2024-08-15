<script lang="ts">
	import { PortfolioEventType, Security } from '$lib/gen/mgo_pb';
	import { Check, ChevronUpDown } from '@steeze-ui/heroicons';
	import { Icon } from '@steeze-ui/svelte-icon';
	import { createListbox } from 'svelte-headlessui';
	import { Transition } from 'svelte-transition';

	let { type = $bindable() }: {type: PortfolioEventType} = $props();

	type PortfolioEventTypeStrings = keyof typeof PortfolioEventType;

	let types = [
		PortfolioEventType[PortfolioEventType.BUY],
		PortfolioEventType[PortfolioEventType.SELL],
		PortfolioEventType[PortfolioEventType.DEPOSIT_CASH],
		PortfolioEventType[PortfolioEventType.WITHDRAW_CASH]
	] as PortfolioEventTypeStrings[];

	const listbox = createListbox({
		label: 'Transaction Type',
		selected: types.find((t) => (type !== undefined ? t == PortfolioEventType[type] : undefined))
	});

	$effect(() => {
		$inspect($listbox.selected);
		let newType = PortfolioEventType[$listbox.selected as PortfolioEventTypeStrings]
		if(type != newType) {
			type = newType
		}
	})
</script>

<div class="relative mt-2">
	<button
		use:listbox.button
		class="
      relative w-full cursor-default rounded-md bg-white py-1.5 pl-3 pr-10 text-left text-gray-900 shadow-sm ring-1
      ring-inset ring-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-500 sm:text-sm sm:leading-6"
	>
		<span class="inline-flex w-full truncate">
			{#if $listbox.selected}
				<span class="truncate">{$listbox.selected}</span>
			{:else}
				<span class="truncate">Please select a security</span>
			{/if}
		</span>
		<span class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2">
			<Icon src={ChevronUpDown} class="h-5 w-5 text-gray-400" aria-hidden="true" />
		</span>
	</button>

	<Transition
		show={$listbox.expanded}
		leave="transition ease-in duration-100"
		leaveFrom="opacity-100"
		leaveTo="opacity-0"
	>
		<ul
			use:listbox.items
			class="
        absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1
        ring-black ring-opacity-5 focus:outline-none sm:text-sm"
		>
			{#each types as value (value)}
				{@const active = $listbox.active === value}
				{@const selected = $listbox.selected === value}
				<li
					use:listbox.item={{ value }}
					class="{active
						? 'bg-indigo-600 text-white'
						: 'text-gray-900'} relative cursor-default select-none py-2 pl-3 pr-9"
				>
					<div class="flex">
						<span class="{selected ? 'font-semibold' : 'font-normal'} truncate">
							{value}
						</span>
					</div>

					{#if selected}
						<span
							class="{active
								? 'text-white'
								: 'text-indigo-600'} absolute inset-y-0 right-0 flex items-center pr-4"
						>
							<Icon src={Check} class="h-5 w-5" aria-hidden="true" />
						</span>
					{/if}
				</li>
			{/each}
		</ul>
	</Transition>
</div>
