<script setup lang="ts">
import { PortfolioSnapshot } from '@/gen/mgo_pb';
import { computed } from 'vue'
import { ArrowDownIcon, ArrowRightIcon, ArrowUpIcon, ChevronDownIcon } from '@heroicons/vue/20/solid'
import PortfolioPositionRow from './PortfolioPositionRow.vue';
import TableSorter from '../TableSorter.vue';

const props = defineProps({
  snapshot: { type: PortfolioSnapshot, required: true },
})

const positions = computed(() => {
  let positions = Object.values(props.snapshot.positions)
  return positions.sort((a, b) => {
    return a.security?.displayName.localeCompare(b.security?.displayName ?? "") ?? 0
  })
})

const perf = computed(() => {
  return (props.snapshot.totalMarketValue - props.snapshot.totalPurchaseValue) / props.snapshot.totalPurchaseValue * 100
})
</script>

<template>
  <div class="-mx-4 mt-8 sm:-mx-0">
    <table class="min-w-full divide-y divide-gray-300">
      <thead>
        <tr>
          <th scope="col" class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-0">
            <TableSorter active>Name</TableSorter>
          </th>
          <th scope="col" class="hidden px-3 py-3.5 text-right text-sm font-semibold text-gray-900 md:table-cell">
            <TableSorter>Amount</TableSorter>
          </th>
          <th scope="col" class="hidden px-3 py-3.5 text-right text-sm font-semibold text-gray-900 lg:table-cell">
            <TableSorter>Purchase Value</TableSorter>
          </th>
          <th scope="col" class="hidden px-3 py-3.5 text-right text-sm font-semibold text-gray-900 sm:table-cell">
            <TableSorter>Market Value</TableSorter>
          </th>
          <th scope="col" class="px-3 py-3.5 text-right text-sm font-semibold text-gray-900">
            <TableSorter>Profit/Loss</TableSorter>
          </th>
        </tr>
      </thead>
      <tbody class="divide-y divide-gray-200">
        <PortfolioPositionRow v-for="position in positions" :key="position.security?.name" :position="position" />
      </tbody>
      <tfoot>
        <th scope="col" class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-0">Total</th>
        <th scope="col" class="hidden px-3 py-3.5 text-right text-sm font-semibold text-gray-900 md:table-cell"></th>
        <th scope="col" class="hidden px-3 py-3.5 text-right text-sm font-semibold text-gray-900 lg:table-cell">
          {{ $filters.currency(snapshot.totalPurchaseValue, "EUR") }}
        </th>
        <th scope="col" class="hidden px-3 py-3.5 text-right text-sm font-semibold text-gray-900 sm:table-cell">
          {{ $filters.currency(snapshot.totalMarketValue, "EUR") }}
        </th>
        <th scope="col" class="px-3 py-3.5 text-right text-sm font-semibold" :class="{
          'text-red-500': perf < 0,
          'text-gray-500': perf <= 1,
          'text-green-500': perf > 1
        }">
          <div>
            {{
              Intl.NumberFormat('de', {
                maximumFractionDigits: 2
              }).format(perf)
            }} %
            <component :is="perf < 0 ? ArrowDownIcon : perf < 1 ? ArrowRightIcon : ArrowUpIcon"
              class="h-4 w-4 mt-0.5 float-right" aria-hidden="true" />
          </div>
          <div class="pr-4">
            {{ $filters.currency(snapshot.totalMarketValue - snapshot.totalPurchaseValue, "EUR") }}
          </div>
        </th>
      </tfoot>
    </table>
  </div>
</template>
