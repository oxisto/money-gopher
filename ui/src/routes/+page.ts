import { redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load = (async () => {
	throw redirect(301, '/dashboard');
}) satisfies PageLoad;
