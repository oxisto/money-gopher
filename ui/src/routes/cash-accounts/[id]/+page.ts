import { load_CashAccountDetail } from '$houdini';
import type { PageLoad } from './$types';

export const load: PageLoad = async (event) => {
	return await load_CashAccountDetail({
		event,
		variables: { id: event.params.id }
	});
};
