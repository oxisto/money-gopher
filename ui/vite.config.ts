import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [
		tailwindcss(),
		sveltekit({
			compilerOptions: {
				// Force runes mode for the project, except for libraries. Can be removed in svelte 6.
				runes: ({ filename }) =>
					filename.split(/[/\\]/).includes('node_modules') ? undefined : true
			},

			// The UI is a pure SPA: no SSR, everything falls back to index.html.
			// Output goes under build/site, which this wipes on every build;
			// build/.gitkeep next to it keeps go:embed in ui.go valid without.
			adapter: adapter({ pages: 'build/site', assets: 'build/site', fallback: 'index.html' })
		})
	],
	server: {
		// The dev server proxies API routes to a locally running moneyd.
		proxy: {
			'/graphql': 'http://localhost:8080',
			'/upload': 'http://localhost:8080',
			'/documents': 'http://localhost:8080',
			'/auth': 'http://localhost:8080'
		}
	}
});
