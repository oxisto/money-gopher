<script setup lang="ts">
import Button from '@/components/Button.vue';
import PortfolioCard from '@/components/PortfolioCard.vue';
import { PortfolioServiceClientKey } from '@/symbols'
import { inject } from 'vue';

let client = inject(PortfolioServiceClientKey)
let portfolios = ((await client?.listPortfolios({}, {}))?.portfolios) ?? [];
</script>

<template>
  <div class="border-b border-gray-200 bg-white px-4 py-5 sm:px-6 my-4">
    <div class="-ml-4 -mt-4 flex flex-wrap items-center justify-between sm:flex-nowrap">
      <div class="ml-4 mt-4">
        <h3 class="text-base font-semibold leading-6 text-gray-900">Portfolios</h3>
        <p class="mt-1 text-sm text-gray-500">Lorem ipsum dolor sit amet consectetur adipisicing elit quam corrupti
          consectetur.</p>
      </div>
      <div class="ml-4 mt-4 flex-shrink-0">
        <button type="button"
          class="relative inline-flex items-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600">Create
          new job</button>
      </div>
    </div>
  </div>
  <ul role="list" class="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
    <li v-for="portfolio in portfolios" :key="portfolio.name"
      class="col-span-1 divide-y divide-gray-200 rounded-lg bg-white shadow">
      <PortfolioCard :portfolio="portfolio"></PortfolioCard>
    </li>
  </ul>
</template>