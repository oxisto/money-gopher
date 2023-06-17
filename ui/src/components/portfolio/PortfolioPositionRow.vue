<script setup lang="ts">
import { PortfolioPosition } from '@/gen/mgo_pb';
import { ArrowDownIcon, ArrowRightIcon, ArrowUpIcon } from '@heroicons/vue/20/solid'

const props = defineProps({
  position: { type: PortfolioPosition, required: true },
})

function shorten(text: string): string {
  let max = 30

  if (text.length > max) {
    return text.substring(0, max) + "..."
  } else {
    return text;
  }
}
</script>
<template>
  <tr v-if="position.security">
    <td class="whitespace-nowrap truncate font-medium py-2 pl-4 pr-3 text-sm sm:pl-0">
      <div class="text-gray-900">
        {{ shorten(position.security.displayName) }}
      </div>
      <div class="text-gray-400">
        {{ position.security.name }}
      </div>
    </td>
    <td class="hidden whitespace-nowrap px-3 py-2 text-right text-sm text-gray-500 md:table-cell">
      {{ Intl.NumberFormat('de', {
        maximumFractionDigits: 2
      }).format(position.amount) }}
    </td>
    <td class="hidden whitespace-nowrap px-3 py-2 text-right text-sm lg:table-cell">
      <div class="text-gray-500">
        {{ $filters.currency(position.purchasePrice, "EUR") }}
      </div>
      <div class="text-gray-400">
        {{ $filters.currency(position.purchaseValue, "EUR") }}
      </div>
    </td>
    <td class="hidden whitespace-nowrap px-3 py-2 text-right text-sm sm:table-cell">
      <div class="text-gray-500">
        {{ $filters.currency(position.marketPrice, "EUR") }}
      </div>
      <div class="text-gray-400">
        {{ $filters.currency(position.marketValue, "EUR") }}
      </div>
    </td>
    <td class="whitespace-nowrap px-3 py-2 text-right text-sm" :class="{
      'text-red-500': position.gains < 0,
      'text-gray-500': position.gains <= 0.01,
      'text-green-500': position.gains > 0.01
    }">
      <div>
        {{
          Intl.NumberFormat('de', {
            style: 'percent',
            maximumFractionDigits: 2
          }).format(position.gains)
        }}
        <component :is="position.gains < 0 ? ArrowDownIcon : position.gains < 1 ? ArrowRightIcon : ArrowUpIcon"
          class="h-4 w-4 mt-0.5 float-right" aria-hidden="true" />
      </div>
      <div class="pr-4">
        {{ $filters.currency(position.profitOrLoss, "EUR") }}
      </div>
    </td>
  </tr>
</template>
