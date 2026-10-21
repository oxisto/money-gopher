import { HoudiniClient } from '$houdini';

// Intercept HTTP-level 401 responses (emitted by the session middleware when
// there is no valid session) and redirect the browser to the login page.
// SSR is disabled so window is always available here.
const _fetch = window.fetch.bind(window);
window.fetch = async function (input, init) {
	const res = await _fetch(input, init);
	// The root layout queries `me` on every route, including /login itself
	// (which has no session yet) — without this guard that 401 redirects
	// back to /login, re-running the same query and looping forever.
	if (res.status === 401 && window.location.pathname !== '/login') {
		window.location.href = '/login';
	}
	return res;
};

// The GraphQL endpoint and other runtime options are configured in houdini.config.js.
export default new HoudiniClient();
