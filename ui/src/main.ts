import { createApp } from 'vue'

import "inter-ui/inter.css";
import './style.css'
import App from './App.vue'
import router from './router'
import { PortfolioServiceClientKey } from '@/symbols'
import { createPromiseClient } from '@bufbuild/connect'
import { createConnectTransport } from '@bufbuild/connect-web'
import type { PromiseClient } from '@bufbuild/connect'
import { PortfolioService } from './gen/mgo_connect'

let client: PromiseClient<typeof PortfolioService> = createPromiseClient(
  PortfolioService,
  createConnectTransport({
    baseUrl: '/',
    useHttpGet: true
  })
)

let app = createApp(App)
  .use(router)
  .provide(PortfolioServiceClientKey, client)

app.config.globalProperties.$filters = {
  currency(value: number, currency: string): string {
    var formatter = Intl.NumberFormat('de', {
      style: "currency",
      currency: currency
    })

    return formatter.format(value)
  }
}

app.mount('#app')

declare module 'vue' {
  interface ComponentCustomProperties {
    $filters: { currency(value: number, currency: string): string }
  }
}