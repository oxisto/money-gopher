import { PortfolioService, SecuritiesService } from '$lib/gen/mgo_connect';
import { createPromiseClient } from '@bufbuild/connect';
import { createConnectTransport } from '@bufbuild/connect-web';
import type { PromiseClient } from '@bufbuild/connect';

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
