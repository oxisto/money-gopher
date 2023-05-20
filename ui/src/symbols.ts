import type { InjectionKey } from 'vue';
import type { PromiseClient } from '@bufbuild/connect'
import { PortfolioService } from './gen/mgo_connect'

export const PortfolioServiceClientKey: InjectionKey<PromiseClient<typeof PortfolioService>> = Symbol('PortfolioServiceClient');
