<script setup lang="ts">
import SecuritiesTable from '../components/SecuritiesTable.vue';
import Button from '@/components/Button.vue';
</script>

<template>
  <div class="px-4 sm:px-6 lg:px-8">
    <div class="sm:flex sm:items-center">
      <div class="sm:flex-auto">
        <h1 class="text-base font-semibold leading-6 text-gray-900">Securities</h1>
        <p class="mt-2 text-sm text-gray-700">A list of all securities currently configured.</p>
      </div>
      <div class="mt-4 sm:ml-16 sm:mt-0 sm:flex-none">
        <Button>Add user</Button>
      </div>
    </div>
    <SecuritiesTable></SecuritiesTable>
  </div>
</template>