import { load_HoldingDetail } from '$houdini';
import type { PageLoad } from './$types';

export const load: PageLoad = async (event) => {
	return await load_HoldingDetail({ event, variables: { id: event.params.id } });
};
