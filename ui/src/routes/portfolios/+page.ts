import type { PageLoad } from './$types';
import type { ListPortfoliosResponse, Portfolio } from '$lib/gen/mgo_pb';
import { portfolioClient } from '$lib/api/client';

export const load = (async ({ fetch }) => {
	const client = portfolioClient(fetch);

	const portfolios = ((await client.listPortfolios({}, {})) as ListPortfoliosResponse).portfolios;
	const snapshots = await Promise.all(
		portfolios.map(async (p: Portfolio) => {
			if (client == undefined) {
				throw 'could not instantiate portfolio client';
			}
			return await client.getPortfolioSnapshot({ portfolioName: p.name });
		})
	);

	return { portfolios, snapshots };
}) satisfies PageLoad;
