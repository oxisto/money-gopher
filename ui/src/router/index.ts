import { createRouter, createWebHistory } from 'vue-router'
import Home from '@/views/Home.vue'
import Portfolios from '@/views/Portfolios.vue'
import Securities from '@/views/Securities.vue'
import PortfolioDetail from '@/views/PortfolioDetail.vue'

const routes = [
  { path: '/', name: 'home', component: Home },
  { path: '/portfolios', name: 'portfolios', component: Portfolios },
  { path: '/portfolios/:name(.*)', name: 'portfolio-detail', component: PortfolioDetail, props: true },
  { path: '/securities', name: 'securities', component: Securities },
]
const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes, // short for `routes: routes`
})

export default router