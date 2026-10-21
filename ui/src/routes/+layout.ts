import { query } from '$lib/graphql/client';
import { LayoutQuery } from '$lib/graphql/documents';
import type { LayoutLoad } from './$types';

// The UI is a pure SPA; rendering happens only in the browser.
export const ssr = false;
export const prerender = false;

export const load: LayoutLoad = async ({ fetch, url }) => {
	// /login has no session to load anything with.
	if (url.pathname === '/login') return { shell: null };
	return { shell: await query(LayoutQuery, {}, fetch) };
};
