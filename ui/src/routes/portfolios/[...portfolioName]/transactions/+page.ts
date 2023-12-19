import { error } from '@sveltejs/kit';
import { portfolioClient } from '$lib/api/client';
import type { PageData } from './$types';

export const load = (async ({ fetch, params, depends }) => {
	if (params.portfolioName == undefined) {
		throw error(405, 'Required parameter missing');
	}

	const client = portfolioClient(fetch);

	const transactions = (
		await client.listPortfolioTransactions({
			portfolioName: params.portfolioName
		})
	).transactions;

	depends(`data:portfolio-transactions:${params.portfolioName}`)

	return { transactions };
}) as PageData;