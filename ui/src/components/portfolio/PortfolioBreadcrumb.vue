<script setup lang="ts">
import { Portfolio, PortfolioSnapshot } from '@/gen/mgo_pb';
import {
  ArrowTrendingDownIcon,
  ArrowTrendingUpIcon,
  CalendarIcon,
  CheckIcon,
  ChevronDownIcon,
  ChevronRightIcon,
  CurrencyEuroIcon,
  LinkIcon,
  PencilIcon,
  ListBulletIcon,
} from '@heroicons/vue/20/solid'
import { Menu, MenuButton, MenuItem, MenuItems } from '@headlessui/vue'
import { computed } from 'vue';

const props = defineProps({
  portfolio: { type: Portfolio, required: true },
  snapshot: { type: PortfolioSnapshot, required: true },
})

const perf = computed(() => {
  return (props.snapshot.totalMarketValue - props.snapshot.totalPurchaseValue) / props.snapshot.totalPurchaseValue * 100
})

const perfAbs = computed(() => {
  return (props.snapshot.totalMarketValue - props.snapshot.totalPurchaseValue)
})
</script>

<template>
  <div class="lg:flex lg:items-center lg:justify-between">
    <div class="min-w-0 flex-1">
      <nav class="flex" aria-label="Breadcrumb">
        <ol role="list" class="flex items-center space-x-4">
          <li>
            <div class="flex">
              <router-link to="/portfolios"
                class="text-sm font-medium text-gray-500 hover:text-gray-700">Portfolios</router-link>
            </div>
          </li>
          <li>
            <div class="flex items-center">
              <ChevronRightIcon class="h-5 w-5 flex-shrink-0 text-gray-400" aria-hidden="true" />
              <a href="#" class="ml-4 text-sm font-medium text-gray-500 hover:text-gray-700">
                My Bank
              </a>
            </div>
          </li>
        </ol>
      </nav>
      <h2 class="mt-2 text-2xl font-bold leading-7 text-gray-900 sm:truncate sm:text-3xl sm:tracking-tight">
        {{ portfolio.displayName }}
      </h2>
      <div class="mt-1 flex flex-col sm:mt-0 sm:flex-row sm:flex-wrap sm:space-x-6">
        <div class="mt-2 flex items-center text-sm" :class="{ 'text-red-500': perf < 0, 'text-green-500': perf >= 0 }">
          <component :is="perf > 0 ? ArrowTrendingUpIcon : ArrowTrendingDownIcon" class="mr-1.5 h-5 w-5 flex-shrink-0"
            :class="{ 'text-red-400': perf < 0, 'text-green-400': perf >= 0 }" aria-hidden="true" />
          {{ Intl.NumberFormat('de', { maximumFractionDigits: 2 }).format(perf) }} % ({{ $filters.currency(perfAbs, "EUR")
          }})
        </div>
        <div class="mt-2 flex items-center text-sm text-gray-500">
          <ListBulletIcon class="mr-1.5 h-5 w-5 flex-shrink-0 text-gray-400" aria-hidden="true" />
          {{ Object.values(snapshot.positions).length }} Positions
        </div>
        <div class="mt-2 flex items-center text-sm text-gray-500">
          <CurrencyEuroIcon class="mr-1.5 h-5 w-5 flex-shrink-0 text-gray-400" aria-hidden="true" />
          {{ $filters.currency(snapshot.totalMarketValue, "EUR") }}
        </div>
        <div class="mt-2 flex items-center text-sm text-gray-500">
          <CalendarIcon class="mr-1.5 h-5 w-5 flex-shrink-0 text-gray-400" aria-hidden="true" />
          First transaction on {{ snapshot.firstTransactionTime }}
        </div>
      </div>
    </div>
    <div class="mt-5 flex lg:ml-4 lg:mt-0">
      <span class="hidden sm:block">
        <button type="button"
          class="inline-flex items-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50">
          <PencilIcon class="-ml-0.5 mr-1.5 h-5 w-5 text-gray-400" aria-hidden="true" />
          Edit
        </button>
      </span>

      <span class="ml-3 hidden sm:block">
        <button type="button"
          class="inline-flex items-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50">
          <LinkIcon class="-ml-0.5 mr-1.5 h-5 w-5 text-gray-400" aria-hidden="true" />
          View
        </button>
      </span>

      <span class="sm:ml-3">
        <button type="button"
          class="inline-flex items-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600">
          <CheckIcon class="-ml-0.5 mr-1.5 h-5 w-5" aria-hidden="true" />
          Publish
        </button>
      </span>

      <!-- Dropdown -->
      <Menu as="div" class="relative ml-3 sm:hidden">
        <MenuButton
          class="inline-flex items-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:ring-gray-400">
          More
          <ChevronDownIcon class="-mr-1 ml-1.5 h-5 w-5 text-gray-400" aria-hidden="true" />
        </MenuButton>

        <transition enter-active-class="transition ease-out duration-200" enter-from-class="transform opacity-0 scale-95"
          enter-to-class="transform opacity-100 scale-100" leave-active-class="transition ease-in duration-75"
          leave-from-class="transform opacity-100 scale-100" leave-to-class="transform opacity-0 scale-95">
          <MenuItems
            class="absolute right-0 z-10 -mr-1 mt-2 w-48 origin-top-right rounded-md bg-white py-1 shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none">
            <MenuItem v-slot="{ active }">
            <a href="#" :class="[active ? 'bg-gray-100' : '', 'block px-4 py-2 text-sm text-gray-700']">Edit</a>
            </MenuItem>
            <MenuItem v-slot="{ active }">
            <a href="#" :class="[active ? 'bg-gray-100' : '', 'block px-4 py-2 text-sm text-gray-700']">View</a>
            </MenuItem>
          </MenuItems>
        </transition>
      </Menu>
    </div>
  </div>
</template>