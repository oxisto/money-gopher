import { createRouter, createWebHistory } from 'vue-router'
import Home from '@/views/Home.vue'
import Portfolios from '@/views/Portfolios.vue'
import Securities from '@/views/Securities.vue'

const routes = [
  { path: '/', component: Home },
  { path: '/portfolios', component: Portfolios },
  { path: '/securities', component: Securities },
]
const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes, // short for `routes: routes`
})

export default router