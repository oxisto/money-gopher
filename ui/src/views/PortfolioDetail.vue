<script setup lang="ts">
import PortfolioBreadcrumb from '@/components/PortfolioBreadcrumb.vue';
import { Portfolio, PortfolioSnapshot } from '@/gen/mgo_pb';
import { PortfolioServiceClientKey } from '@/symbols';
import { watch, inject, ref } from 'vue';

const props = defineProps<{ name: string }>();
const client = inject(PortfolioServiceClientKey)
var portfolio = ref<Portfolio | undefined>()
var snapshot = ref<PortfolioSnapshot | undefined>()

watch(props, async (props) => {
  portfolio.value = await client?.getPortfolio({ name: props.name })
  snapshot.value = await client?.getPortfolioSnapshot({ portfolioName: props.name })
}, { immediate: true });
</script>

<template>
  <PortfolioBreadcrumb :portfolio="portfolio" :snapshot="snapshot" v-if="portfolio && snapshot" />
</template>