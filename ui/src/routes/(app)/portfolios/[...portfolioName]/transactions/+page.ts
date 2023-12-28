import { error } from '@sveltejs/kit';
import { convertError, portfolioClient } from '$lib/api/client';
import type { PageData } from './$types';
import { ListPortfolioTransactionsResponse } from '$lib/gen/mgo_pb';

export const load = (async ({ fetch, params, depends }) => {
	if (params.portfolioName == undefined) {
		throw error(405, 'Required parameter missing');
	}

	const client = portfolioClient(fetch);

	const transactions = (
		await client
			.listPortfolioTransactions({
				portfolioName: params.portfolioName
			})
			.catch<ListPortfolioTransactionsResponse>(convertError)
	).transactions;

	depends(`data:portfolio-transactions:${params.portfolioName}`);

	return { transactions };
}) as PageData;
