import { securitiesClient, convertError } from '$lib/api/client';
import { ListSecuritiesResponse } from '$lib/gen/mgo_pb';
import type { LayoutLoad } from './$types';

export const load = (async ({ fetch }) => {
	const client = securitiesClient(fetch);

	try {
		const securities = (await client.listSecurities({}).catch<ListSecuritiesResponse>(convertError))
			.securities;
		return { securities };
	} catch (err) {
		convertError(err);
	}
}) satisfies LayoutLoad;
