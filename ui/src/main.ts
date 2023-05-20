import { createApp } from 'vue'

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
    baseUrl: 'http://localhost:8080',
  })
)

createApp(App)
  .use(router)
  .provide(PortfolioServiceClientKey, client)
  .mount('#app')
