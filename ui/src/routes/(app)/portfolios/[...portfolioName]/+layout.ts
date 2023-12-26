import type { LayoutData } from './$types';
import { error } from '@sveltejs/kit';
import { portfolioClient, convertError } from '$lib/api/client';
import { Portfolio, PortfolioSnapshot } from '$lib/gen/mgo_pb';

export const load = (async ({ fetch, params, depends }) => {
	if (params.portfolioName == undefined) {
		throw error(405, 'Required parameter missing');
	}

	const client = portfolioClient(fetch);

	const portfolio = await client
		.getPortfolio({ name: params.portfolioName })
		.catch<Portfolio>(convertError);
	const snapshot = await client
		.getPortfolioSnapshot({ portfolioName: params.portfolioName })
		.catch<PortfolioSnapshot>(convertError);

	depends(`data:portfolio-snapshot:${params.portfolioName}`);

	return { portfolio, snapshot };
}) as LayoutData;
