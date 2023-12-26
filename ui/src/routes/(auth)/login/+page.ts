import { redirectLogin } from '$lib/oauth';
import type { PageLoad } from './$types';

export const load = (async () => {
	return redirectLogin();
}) satisfies PageLoad;
