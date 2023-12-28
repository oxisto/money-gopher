import 'inter-ui/inter-variable.css';
import '../app.css';
import { securitiesClient } from '$lib/api/client';
import type { LayoutLoad } from './$types';

export const ssr = false;

export const load = (async ({ fetch }) => {
	const client = securitiesClient(fetch);

	const securities = (await client.listSecurities({})).securities;

	return { securities };
}) satisfies LayoutLoad;
