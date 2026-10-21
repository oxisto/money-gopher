import { redirect } from '@sveltejs/kit';

// Realized gains moved to /income; keep old links working.
export const load = () => {
	redirect(301, '/income');
};
