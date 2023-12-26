import { PortfolioService, SecuritiesService } from '$lib/gen/mgo_connect';
import { Code, ConnectError, createPromiseClient } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import type { Interceptor, PromiseClient } from '@connectrpc/connect';
import { error } from '@sveltejs/kit';

/**
 * This function converts an {@link ConnectError} into a error suitable for
 * catching in SvelteKit (using {@link error}).
 *
 * @param err an error
 */
export function convertError<T>(err: unknown): Promise<T> {
	if (err instanceof ConnectError) {
		// convert it into a svelekit error and adjust the code for some errors
		if (err.code == Code.Unauthenticated) {
			throw error(401, err.rawMessage);
		} else {
			throw error(500, err.rawMessage);
		}
	} else {
		// otherwise, just rethrow it
		throw err;
	}
}

const authorizer: Interceptor = (next) => async (req) => {
	req.header.set('Authorization', `Bearer ${localStorage.token}`);
	return await next(req);
};

export function portfolioClient(fetch = window.fetch): PromiseClient<typeof PortfolioService> {
	return createPromiseClient(
		PortfolioService,
		createConnectTransport({
			baseUrl: '/',
			useHttpGet: true,
			fetch: fetch,
			interceptors: [authorizer]
		})
	);
}

export function securitiesClient(fetch = window.fetch): PromiseClient<typeof SecuritiesService> {
	return createPromiseClient(
		SecuritiesService,
		createConnectTransport({
			baseUrl: '/',
			useHttpGet: true,
			fetch: fetch,
			interceptors: [authorizer]
		})
	);
}
