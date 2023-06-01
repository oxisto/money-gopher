<script setup lang="ts">
import { PortfolioServiceClientKey } from '@/symbols'
import { Disclosure, DisclosureButton, DisclosurePanel } from '@headlessui/vue'
import {
  BanknotesIcon,
  CalendarDaysIcon,
  ChartPieIcon,
  Cog6ToothIcon,
  FolderIcon,
  HomeIcon,
} from '@heroicons/vue/24/outline'
import { inject } from 'vue'

const navigation = [
  { name: 'Dashboard', routes: ['home'], href: '/', icon: HomeIcon, current: true },
  { name: 'Securities', routes: ['securities'], href: '/securities', icon: BanknotesIcon, current: false },
  {
    name: 'Portfolios',
    routes: ['portfolios', 'portfolio-detail'],
    href: '/portfolios',
    icon: FolderIcon,
    current: false,
    children: [] as { name: string, href: string, routes: string[], routeId: string }[],
  },
  { name: 'Dividends', href: '/', icon: CalendarDaysIcon, current: false },
  { name: 'Performance', href: '/', icon: ChartPieIcon, current: false },
]

// TODO: convert to store
const client = inject(PortfolioServiceClientKey)

const portfolios = await client?.listPortfolios({})
navigation[2].children = portfolios?.portfolios.map(p => {
  return {
    name: p.displayName,
    href: '/portfolios/' + p.name,
    routes: ['portfolio-detail'],
    routeId: p.name
  }
})

const teams = [
  { id: 1, name: 'Personal', href: '#', initial: 'H', current: false },
  { id: 2, name: 'Parents', href: '#', initial: 'T', current: false },
  { id: 3, name: 'Child', href: '#', initial: 'W', current: false },
]

</script>

<template>
  <div class="flex grow flex-col gap-y-5 overflow-y-auto bg-indigo-600 px-6 pb-4">
    <div class="flex h-16 shrink-0 items-center">
      <img class="h-8 w-auto" src="https://tailwindui.com/img/logos/mark.svg?color=white" alt="Your Company" />
    </div>
    <nav class="flex flex-1 flex-col">
      <ul role="list" class="flex flex-1 flex-col gap-y-7">
        <li>
          <ul role="list" class="-mx-2 space-y-1">
            <li v-for="item in navigation" :key="item.name">
              <router-link :to="item.href"
                :class="[item.routes?.includes($route.name?.toString() ?? 'home') ? 'bg-indigo-700 text-white' : 'text-indigo-200 hover:text-white hover:bg-indigo-700', 'group flex gap-x-3 rounded-md p-2 text-sm leading-6 font-semibold']">
                <component :is="item.icon"
                  :class="[item.routes?.includes($route.name?.toString() ?? 'home') ? 'text-white' : 'text-indigo-200 group-hover:text-white', 'h-6 w-6 shrink-0']"
                  aria-hidden="true" />
                {{ item.name }}
              </router-link>
              <ul class="mt-1 px-2">
                <li v-for="subItem in item.children" :key="subItem.name">
                  <!-- 44px -->
                  <router-link :to="subItem.href"
                    :class="[subItem.routes?.includes($route.name?.toString() ?? 'home') && $route.params.name == subItem.routeId ? 'bg-indigo-700 text-white' : 'text-indigo-200 hover:text-white hover:bg-indigo-700', 'block rounded-md py-2 pr-2 pl-9 text-sm leading-6 text-gray-700']">
                    {{ subItem.name }}</router-link>
                </li>
              </ul>
            </li>
          </ul>
        </li>
        <li>
          <div class="text-xs font-semibold leading-6 text-indigo-200">Your Portfolio Group</div>
          <ul role="list" class="-mx-2 mt-2 space-y-1">
            <li v-for="team in teams" :key="team.name">
              <a :href="team.href"
                :class="[team.current ? 'bg-indigo-700 text-white' : 'text-indigo-200 hover:text-white hover:bg-indigo-700', 'group flex gap-x-3 rounded-md p-2 text-sm leading-6 font-semibold']">
                <span
                  class="flex h-6 w-6 shrink-0 items-center justify-center rounded-lg border border-indigo-400 bg-indigo-500 text-[0.625rem] font-medium text-white">{{
                    team.initial }}</span>
                <span class="truncate">{{ team.name }}</span>
              </a>
            </li>
          </ul>
        </li>
        <li class="mt-auto">
          <a href="#"
            class="group -mx-2 flex gap-x-3 rounded-md p-2 text-sm font-semibold leading-6 text-indigo-200 hover:bg-indigo-700 hover:text-white">
            <Cog6ToothIcon class="h-6 w-6 shrink-0 text-indigo-200 group-hover:text-white" aria-hidden="true" />
            Settings
          </a>
        </li>
      </ul>
    </nav>
  </div>
</template>