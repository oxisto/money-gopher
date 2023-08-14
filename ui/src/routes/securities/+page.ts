import type { PageLoad } from './$types';
import { securitiesClient } from '$lib/api/client';

export const load = (async ({ fetch }) => {
	const client = securitiesClient(fetch);

	const securities = (await client.listSecurities({})).securities;

	return { securities };
}) satisfies PageLoad;
