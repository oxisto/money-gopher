import type { LayoutData } from './$types';
import { error } from '@sveltejs/kit';
import { portfolioClient } from '$lib/api/client';

export const load = (async ({ fetch, params }) => {
	if (params.portfolioName == undefined) {
		throw error(405, 'Required parameter missing');
	}

	const client = portfolioClient(fetch);
	console.log(params.portfolioName);

	const portfolio = client.getPortfolio({ name: params.portfolioName });
	const snapshot = client.getPortfolioSnapshot({ portfolioName: params.portfolioName });

	return { portfolio, snapshot };
}) as LayoutData;
