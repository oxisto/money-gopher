import { PortfolioEvent, PortfolioEventType } from '$lib/gen/mgo_pb';
import type { PageLoad } from './$types';

export const load = (async ({ params, parent }) => {
	const data = await parent();
	const transactionName = params.txName;
	const add = transactionName == 'add';

	let transaction: PortfolioEvent;
	if (add) {
		// Create a new default import template
		transaction = new PortfolioEvent({
			amount: 1,
			type: PortfolioEventType.BUY,
			portfolioName: data.portfolio.name
		});
	} else {
		//transaction = await data.client.getImportTemplate({ id: templateId });
		transactionName;
		transaction = new PortfolioEvent();
	}

	return {
		transaction,
		add
	};
}) satisfies PageLoad;
