import { load_Dashboard, load_NetWorthHistory } from '$houdini';
import type { PageLoad } from './$types';

export const load: PageLoad = async (event) => {
	return {
		...(await load_Dashboard({ event })),
		...(await load_NetWorthHistory({ event, variables: { period: 'YEAR_TO_DATE' } }))
	};
};
