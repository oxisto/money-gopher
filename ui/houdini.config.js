/// <reference types="houdini-svelte" />

/** @type {import('houdini').ConfigFile} */
const config = {
	// The Go backend owns the schema; Houdini generates from the same file
	// that the server verifies its resolvers against.
	schemaPath: '../api/schema.graphql',
	// Same-origin in production (the SPA is served by moneyd); the vite dev
	// server proxies /graphql to a locally running moneyd. The schema comes
	// from the file above, so never poll the endpoint for it.
	url: '/graphql',
	watchSchema: null,
	plugins: {
		'houdini-svelte': {
			client: './src/client',
			// There is no svelte.config.js (the kit config lives inline in
			// vite.config.ts), so framework detection needs a nudge.
			framework: 'kit',
			forceRunesMode: true
		}
	},
	scalars: {
		Time: {
			type: 'Date',
			unmarshal(val) {
				return new Date(val);
			},
			marshal(date) {
				return date.toISOString();
			}
		}
	}
};

export default config;
