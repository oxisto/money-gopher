import { load_Layout } from '$houdini';
import type { LayoutLoad } from './$types';

// The UI is a pure SPA; rendering happens only in the browser.
export const ssr = false;
export const prerender = false;

export const load: LayoutLoad = async (event) => {
	return await load_Layout({ event });
};
