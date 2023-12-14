import { portfolioClient } from '$lib/api/client';
import { PortfolioEvent, PortfolioEventType } from '$lib/gen/mgo_pb';
import type { PageLoad } from './$types';

export const load = (async ({ params, parent }) => {
	const data = await parent();
	const txName = params.txName;
	const add = txName == 'add';

	let transaction: PortfolioEvent;
	if (add) {
		// Create a new default import template
		transaction = new PortfolioEvent({
			amount: 1,
			type: PortfolioEventType.BUY,
			portfolioName: data.portfolio.name
		});
	} else {
		const client = portfolioClient(fetch);
		transaction = await client.getPortfolioTransaction({ name: txName });
	}

	return {
		transaction,
		add
	};
}) satisfies PageLoad;
