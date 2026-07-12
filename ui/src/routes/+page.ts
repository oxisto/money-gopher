import { load_Dashboard } from '$houdini';
import type { PageLoad } from './$types';

export const load: PageLoad = async (event) => {
	return await load_Dashboard({ event });
};
