<script setup lang="ts">
import { ChevronDownIcon } from '@heroicons/vue/20/solid'

const props = defineProps<{ active?: boolean }>()
</script>

<template>
  <a href="#" class="group inline-flex">
    <slot></slot>
    <span
      :class="active ? 'ml-2 flex-none rounded bg-gray-100 text-gray-900 group-hover:bg-gray-200' : 'invisible ml-2 flex-none rounded text-gray-400 group-hover:visible group-focus:visible'">
      <ChevronDownIcon class="h-5 w-5" aria-hidden="true" />
    </span>
  </a>
</template>