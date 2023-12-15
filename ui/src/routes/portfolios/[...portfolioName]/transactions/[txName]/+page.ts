import { portfolioClient } from '$lib/api/client';
import { PortfolioEvent, PortfolioEventType } from '$lib/gen/mgo_pb';
import type { PageLoad } from './$types';
import {Timestamp } from "@bufbuild/protobuf";

export const load = (async ({ params, parent }) => {
	const data = await parent();
	const txName = params.txName;
	const add = txName == 'add';

	let transaction: PortfolioEvent;
	if (add) {
		// Construct a new time, based on "now" but reset the minutes to 0
		var time = new Date();
		time.setMinutes(0);

		// Create a new default import template
		transaction = new PortfolioEvent({
			amount: 1,
			type: PortfolioEventType.BUY,
			portfolioName: data.portfolio.name,
			time: Timestamp.fromDate(time),
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
