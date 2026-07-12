import { load_PortfolioDetail } from '$houdini';
import type { PageLoad } from './$types';

export const load: PageLoad = async (event) => {
	return await load_PortfolioDetail({
		event,
		variables: { id: event.params.id }
	});
};
