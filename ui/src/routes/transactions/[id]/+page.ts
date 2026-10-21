import { load_TransactionDetail } from '$houdini';
import type { PageLoad } from './$types';

export const load: PageLoad = async (event) => {
	return await load_TransactionDetail({
		event,
		variables: { id: event.params.id }
	});
};
