import { PortfolioService, SecuritiesService } from '$lib/gen/mgo_connect';
import { createPromiseClient } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import type { PromiseClient } from '@connectrpc/connect';

export function portfolioClient(fetch = window.fetch): PromiseClient<typeof PortfolioService> {
	return createPromiseClient(
		PortfolioService,
		createConnectTransport({
			baseUrl: '/',
			useHttpGet: true,
			fetch: fetch
		})
	);
}

export function securitiesClient(fetch = window.fetch): PromiseClient<typeof SecuritiesService> {
	return createPromiseClient(
		SecuritiesService,
		createConnectTransport({
			baseUrl: '/',
			useHttpGet: true,
			fetch: fetch
		})
	);
}
