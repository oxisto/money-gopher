<script setup lang="ts">
import { PortfolioPosition, PortfolioSnapshot } from '@/gen/mgo_pb';
import { computed, ref } from 'vue'
import { ArrowDownIcon, ArrowRightIcon, ArrowUpIcon, ChevronDownIcon } from '@heroicons/vue/20/solid'
import PortfolioPositionRow from './PortfolioPositionRow.vue';
import TableSorter from '../TableSorter.vue';

const props = defineProps({
  snapshot: { type: PortfolioSnapshot, required: true },
})

const sorters = new Map<string, (a: PortfolioPosition, b: PortfolioPosition) => number>();
sorters.set('displayName', (a: PortfolioPosition, b: PortfolioPosition) => {
  return a.security?.displayName.localeCompare(b.security?.displayName ?? "") ?? 0
})
sorters.set('amount', (a: PortfolioPosition, b: PortfolioPosition) => {
  return a.amount - b.amount
})
sorters.set('purchaseValue', (a: PortfolioPosition, b: PortfolioPosition) => {
  return a.purchaseValue - b.purchaseValue
})
sorters.set('marketValue', (a: PortfolioPosition, b: PortfolioPosition) => {
  return a.marketValue - b.marketValue
})

const sortBy = ref('displayName')
const asc = ref(true)

const positions = computed(() => {
  let positions = Object.values(props.snapshot.positions)
  return positions.sort((a: PortfolioPosition, b: PortfolioPosition) => {
    const sort = sorters.get(sortBy.value)?.call(null, a, b) ?? 0;
    return asc.value ? sort : -sort
  })
})

const perf = computed(() => {
  return (props.snapshot.totalMarketValue - props.snapshot.totalPurchaseValue) / props.snapshot.totalPurchaseValue * 100
})

function toggleSortDirection() {
  asc.value = !asc.value
}

function changeSortBy(column: string) {
  sortBy.value = column
}
</script>

<template>
  {{ asc }}
  {{ sortBy }}
  <div class="-mx-4 mt-8 sm:-mx-0">
    <table class="min-w-full divide-y divide-gray-300">
      <thead>
        <tr>
          <th scope="col" class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-0">
            <TableSorter :active="sortBy == 'displayName'" column="displayName" @change-direction="toggleSortDirection()"
              @change-sort-by="(column) => changeSortBy(column)">
              Name
            </TableSorter>
          </th>
          <th scope="col" class="hidden px-3 py-3.5 text-right text-sm font-semibold text-gray-900 md:table-cell">
            <TableSorter :active="sortBy == 'amount'" column="amount" @change-direction="toggleSortDirection()"
              @change-sort-by="(column) => changeSortBy(column)">Amount</TableSorter>
          </th>
          <th scope="col" class="hidden px-3 py-3.5 text-right text-sm font-semibold text-gray-900 lg:table-cell">
            <TableSorter :active="sortBy == 'purchaseValue'" column="purchaseValue"
              @change-direction="toggleSortDirection()" @change-sort-by="(column) => changeSortBy(column)">Purchase Value
            </TableSorter>
          </th>
          <th scope="col" class="hidden px-3 py-3.5 text-right text-sm font-semibold text-gray-900 sm:table-cell">
            <TableSorter :active="sortBy == 'marketValue'" column="marketValue" @change-direction="toggleSortDirection()"
              @change-sort-by="(column) => changeSortBy(column)">Market Value</TableSorter>
          </th>
          <th scope="col" class="px-3 py-3.5 text-right text-sm font-semibold text-gray-900">
            <TableSorter :active="sortBy == 'profit'" column="profit" @change-direction="toggleSortDirection()"
              @change-sort-by="(column) => changeSortBy(column)">Profit/Loss</TableSorter>
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
