import type { PageLoad } from './$types';
import { ListPortfoliosResponse, PortfolioSnapshot, type Portfolio } from '$lib/gen/mgo_pb';
import { convertError, portfolioClient } from '$lib/api/client';

export const load = (async ({ fetch }) => {
	const client = portfolioClient(fetch);

	const portfolios = (
		await client.listPortfolios({}, {}).catch<ListPortfoliosResponse>(convertError)
	).portfolios;
	const snapshots = await Promise.all(
		portfolios.map(async (p: Portfolio) => {
			if (client == undefined) {
				throw 'could not instantiate portfolio client';
			}
			return await client
				.getPortfolioSnapshot({ portfolioName: p.name })
				.catch<PortfolioSnapshot>(convertError);
		})
	);

	return { portfolios, snapshots };
}) satisfies PageLoad;
