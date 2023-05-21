<script setup lang="ts">
import type { Portfolio } from '@/gen/mgo_pb';
import { PortfolioServiceClientKey } from '@/symbols';
import { watch, inject } from 'vue';

const props = defineProps<{ name: string }>();
const client = inject(PortfolioServiceClientKey)
var portfolio: Portfolio | undefined

watch(props, async (props) => {
  portfolio = await client?.getPortfolio({ name: props.name })
}, { immediate: true });
</script>

<template>
  hallo
  {{ portfolio }}
</template>