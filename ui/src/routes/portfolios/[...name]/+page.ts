import type { PageLoad } from './$types';
import { error } from '@sveltejs/kit';
import type { Portfolio } from '$lib/gen/mgo_pb';
import { portfolioClient } from '$lib/api/client';

export const load = (async ({ fetch, params }) => {
	if (params.name == undefined) {
		throw error(405, 'Required parameter missing');
	}

	const client = portfolioClient(fetch);
	console.log(params.name);

	const portfolio = client.getPortfolio({ name: params.name });
	const snapshot = client.getPortfolioSnapshot({ portfolioName: params.name });

	return { portfolio, snapshot };
}) as PageLoad;
